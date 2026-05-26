/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//! BIOS configuration: machine_setup, Dell job wait/recovery, and PollingBiosSetup escalation.

use carbide_uuid::machine::MachineId;
use chrono::Utc;
use eyre::eyre;
use libredfish::{Redfish, SystemPowerControl};
use model::machine::{
    BiosConfigInfo, BiosConfigState, ManagedHostState, ManagedHostStateSnapshot, PowerState,
};

use super::{
    ReachabilityParams, RebootStatus, call_machine_setup_and_handle_no_dpu_error,
    handler_host_power_control, trigger_reboot_if_needed,
};
use crate::state_controller::external_service_error::redfish_error;
use crate::state_controller::machine::context::MachineStateHandlerContextObjects;
use crate::state_controller::state_handler::{
    StateHandlerContext, StateHandlerError, StateHandlerOutcome,
};

/// Outcome of configure_host_bios function.
pub(super) enum BiosConfigOutcome {
    Done,
    WaitingForReboot(String),
    /// Dell BIOS PATCH returned a job ID; wait for it to complete before boot order.
    WaitingForBiosJob(BiosConfigInfo),
}

/// Outcome of advancing the BIOS config job state machine (Dell: wait for BIOS PATCH job before boot order).
pub(super) enum BiosConfigJobAdvanceOutcome {
    Continue(BiosConfigInfo),
    /// Dell BIOS job completed; proceed to verify settings via PollingBiosSetup.
    Done,
    Failed {
        failure: String,
    },
    /// Same state, but wait (e.g. waiting for power down or BMC to come back).
    Wait(String),
    /// After successful power/BMC recovery from a failed BIOS job: re-run machine_setup (not PollingBiosSetup).
    RetryPlatformConfiguration {
        retry_count: u32,
    },
}

#[derive(Debug)]
pub(super) enum PollingBiosSetupOutcome {
    Verified,
    Wait(String),
    EnterRecovery(BiosConfigInfo),
    Failed { failure: String },
}

/// Max configure_host_bios retry cycles through HandleBiosJobFailure recovery (matches boot-order retry budget).
const MAX_BIOS_CONFIG_RETRIES: u32 = 3;

/// How long PollingBiosSetup may sit on Ok(false) before escalating into HandleBiosJobFailure recovery.
///
/// From `machine_state_history` (4 sites, ~4500 samples): HostInit/PollingBiosSetup usually
/// finishes within ~11 min p95; wedged hosts sit 90+ min. 15 min keeps the first recovery attempt
/// inside the 30-min HOST_INIT SLA.
const POLLING_BIOS_SETUP_STUCK_THRESHOLD: chrono::Duration = chrono::Duration::minutes(15);

pub(super) async fn configure_host_bios(
    ctx: &mut StateHandlerContext<'_, MachineStateHandlerContextObjects>,
    reachability_params: &ReachabilityParams,
    redfish_client: &dyn Redfish,
    mh_snapshot: &ManagedHostStateSnapshot,
    retry_count: u32,
) -> Result<BiosConfigOutcome, StateHandlerError> {
    let boot_interface_mac = mh_snapshot.boot_interface_mac().map(|m| m.to_string());

    let bios_job_id = match call_machine_setup_and_handle_no_dpu_error(
        redfish_client,
        boot_interface_mac.as_deref(),
        mh_snapshot.host_snapshot.associated_dpu_machine_ids().len(),
        &ctx.services.site_config,
    )
    .await
    {
        Err(e) => {
            tracing::warn!(
                "redfish machine_setup failed for {}, potentially due to known race condition between UEFI POST and BMC. triggering force-restart if needed. err: {}",
                mh_snapshot.host_snapshot.id,
                e
            );

            // if machine_setup failed, reboot to potentially work around
            // a known race between the DPU UEFI and the BMC, where if
            // the BMC is not up when DPU UEFI runs, then Attributes might
            // not come through. The fix is to force-restart the DPU to
            // re-POST.
            //
            // As of July 2024, Josh Price said there's an NBU FR to fix
            // this, but it wasn't target to a release yet.
            let reboot_status = if mh_snapshot.host_snapshot.last_reboot_requested.is_none() {
                handler_host_power_control(mh_snapshot, ctx, SystemPowerControl::ForceRestart)
                    .await?;

                RebootStatus {
                    increase_retry_count: true,
                    status: "Restarted host".to_string(),
                }
            } else {
                trigger_reboot_if_needed(
                    &mh_snapshot.host_snapshot,
                    mh_snapshot,
                    None,
                    reachability_params,
                    ctx,
                )
                .await?
            };
            return Ok(BiosConfigOutcome::WaitingForReboot(format!(
                "redfish machine_setup failed: {e}; triggered host reboot: {reboot_status:#?}"
            )));
        }
        Ok(jid) => jid,
    };

    if let Some(job_id) = &bios_job_id {
        return Ok(BiosConfigOutcome::WaitingForBiosJob(BiosConfigInfo {
            bios_job_id: Some(job_id.clone()),
            bios_config_state: BiosConfigState::WaitForBiosJobScheduled,
            retry_count,
        }));
    }

    // No job to wait for (non-Dell or vendor that doesn't return job); reboot to apply and continue.
    handler_host_power_control(mh_snapshot, ctx, SystemPowerControl::ForceRestart).await?;
    Ok(BiosConfigOutcome::Done)
}

/// Advance one step of the BIOS config job wait state machine.
pub(super) async fn advance_bios_config_job(
    ctx: &mut StateHandlerContext<'_, MachineStateHandlerContextObjects>,
    redfish_client: &dyn Redfish,
    mh_snapshot: &ManagedHostStateSnapshot,
    info: BiosConfigInfo,
) -> Result<BiosConfigJobAdvanceOutcome, StateHandlerError> {
    match info.bios_config_state {
        BiosConfigState::WaitForBiosJobScheduled => {
            let job_id = info.bios_job_id.as_ref().ok_or_else(|| {
                StateHandlerError::GenericError(eyre!(
                    "WaitForBiosJobScheduled requires bios_job_id for host {}",
                    mh_snapshot.host_snapshot.id
                ))
            })?;
            let job_state = redfish_client
                .get_job_state(job_id)
                .await
                .map_err(|e| redfish_error("get_job_state", e))?;
            if matches!(
                job_state,
                libredfish::JobState::ScheduledWithErrors
                    | libredfish::JobState::CompletedWithErrors
            ) {
                let failure = format!("BIOS job {} failed with state {job_state:#?}", job_id);
                tracing::warn!(
                    "{} for {}, transitioning to HandleBiosJobFailure (power cycle + BMC reset)",
                    failure,
                    mh_snapshot.host_snapshot.id
                );
                return try_bios_recovery_attempt(
                    info.retry_count,
                    info.bios_job_id.clone(),
                    failure,
                    &mh_snapshot.host_snapshot.id,
                );
            }
            if !matches!(job_state, libredfish::JobState::Scheduled) {
                return Err(StateHandlerError::GenericError(eyre!(
                    "waiting for BIOS job {:#?} to be scheduled; current state: {job_state:#?}",
                    job_id
                )));
            }
            Ok(BiosConfigJobAdvanceOutcome::Continue(BiosConfigInfo {
                bios_job_id: info.bios_job_id.clone(),
                bios_config_state: BiosConfigState::RebootHost,
                retry_count: info.retry_count,
            }))
        }
        BiosConfigState::RebootHost => {
            handler_host_power_control(mh_snapshot, ctx, SystemPowerControl::ForceRestart).await?;
            Ok(BiosConfigJobAdvanceOutcome::Continue(BiosConfigInfo {
                bios_job_id: info.bios_job_id.clone(),
                bios_config_state: BiosConfigState::WaitForBiosJobCompletion,
                retry_count: info.retry_count,
            }))
        }
        BiosConfigState::WaitForBiosJobCompletion => {
            const JOB_QUERY_WAIT_MINUTES: i64 = 5;
            let job_id = info.bios_job_id.as_ref().ok_or_else(|| {
                StateHandlerError::GenericError(eyre!(
                    "WaitForBiosJobCompletion requires bios_job_id for host {}",
                    mh_snapshot.host_snapshot.id
                ))
            })?;
            let job_state = match redfish_client.get_job_state(job_id).await {
                Ok(s) => s,
                Err(e) => {
                    let minutes_since_state_change = mh_snapshot
                        .host_snapshot
                        .state
                        .version
                        .since_state_change()
                        .num_minutes();
                    if minutes_since_state_change < JOB_QUERY_WAIT_MINUTES {
                        return Err(redfish_error("get_job_state", e));
                    }
                    let failure = format!(
                        "BIOS config job {} lookup failed after {} min: {}",
                        job_id, minutes_since_state_change, e
                    );
                    tracing::warn!(
                        "{} for {}, transitioning to HandleBiosJobFailure (power cycle + BMC reset)",
                        failure,
                        mh_snapshot.host_snapshot.id
                    );
                    return try_bios_recovery_attempt(
                        info.retry_count,
                        info.bios_job_id.clone(),
                        failure,
                        &mh_snapshot.host_snapshot.id,
                    );
                }
            };
            match job_state {
                libredfish::JobState::Completed => Ok(BiosConfigJobAdvanceOutcome::Done),
                libredfish::JobState::ScheduledWithErrors
                | libredfish::JobState::CompletedWithErrors => {
                    let failure = format!(
                        "BIOS config job {} failed with state {job_state:#?}",
                        job_id
                    );
                    tracing::warn!(
                        "{} for {}, transitioning to HandleBiosJobFailure (power cycle + BMC reset)",
                        failure,
                        mh_snapshot.host_snapshot.id,
                    );
                    try_bios_recovery_attempt(
                        info.retry_count,
                        info.bios_job_id.clone(),
                        failure,
                        &mh_snapshot.host_snapshot.id,
                    )
                }
                _ => Err(StateHandlerError::GenericError(eyre!(
                    "waiting for BIOS job {:#?} to complete; current state: {job_state:#?}",
                    job_id
                ))),
            }
        }
        BiosConfigState::HandleBiosJobFailure {
            failure,
            power_state,
        } => {
            let current_power_state = redfish_client
                .get_power_state()
                .await
                .map_err(|e| redfish_error("get_power_state", e))?;

            match power_state {
                PowerState::Off => {
                    if current_power_state != libredfish::PowerState::Off {
                        handler_host_power_control(mh_snapshot, ctx, SystemPowerControl::ForceOff)
                            .await?;
                        return Ok(BiosConfigJobAdvanceOutcome::Wait(format!(
                            "HandleBiosJobFailure: waiting for {} to power down; current power state: {current_power_state}; failure: {}",
                            mh_snapshot.host_snapshot.id, failure
                        )));
                    }
                    tracing::info!(
                        "HandleBiosJobFailure: Resetting BMC for {} after BIOS job failure: {}",
                        mh_snapshot.host_snapshot.id,
                        failure
                    );
                    redfish_client
                        .bmc_reset()
                        .await
                        .map_err(|e| redfish_error("bmc_reset", e))?;
                    Ok(BiosConfigJobAdvanceOutcome::Continue(BiosConfigInfo {
                        bios_job_id: info.bios_job_id.clone(),
                        bios_config_state: BiosConfigState::HandleBiosJobFailure {
                            failure: failure.clone(),
                            power_state: PowerState::On,
                        },
                        retry_count: info.retry_count,
                    }))
                }
                PowerState::On => {
                    if current_power_state != libredfish::PowerState::On {
                        let basetime = mh_snapshot
                            .host_snapshot
                            .last_reboot_requested
                            .as_ref()
                            .map(|x| x.time)
                            .unwrap_or(mh_snapshot.host_snapshot.state.version.timestamp());
                        let power_down_wait = ctx
                            .services
                            .site_config
                            .machine_state_controller
                            .power_down_wait;
                        if Utc::now().signed_duration_since(basetime) < power_down_wait {
                            return Ok(BiosConfigJobAdvanceOutcome::Wait(format!(
                                "HandleBiosJobFailure: waiting for BMC to come back online for {}; failure: {}",
                                mh_snapshot.host_snapshot.id, failure
                            )));
                        }
                        handler_host_power_control(mh_snapshot, ctx, SystemPowerControl::On)
                            .await?;
                        return Ok(BiosConfigJobAdvanceOutcome::Wait(format!(
                            "HandleBiosJobFailure: powering on {} after BMC reset; failure: {}",
                            mh_snapshot.host_snapshot.id, failure
                        )));
                    }
                    tracing::info!(
                        machine_id = %mh_snapshot.host_snapshot.id,
                        retry_count = info.retry_count,
                        "HandleBiosJobFailure: BMC reset complete; re-running platform configuration (machine_setup) — power cycle does not apply BIOS attributes",
                    );
                    Ok(BiosConfigJobAdvanceOutcome::RetryPlatformConfiguration {
                        retry_count: info.retry_count,
                    })
                }
                _ => Err(StateHandlerError::GenericError(eyre!(
                    "HandleBiosJobFailure: unexpected power state {power_state:#?} for {}",
                    mh_snapshot.host_snapshot.id
                ))),
            }
        }
    }
}

/// Enter HandleBiosJobFailure recovery, or move to Failed when budget is exhausted.
fn try_bios_recovery_attempt(
    retry_count: u32,
    bios_job_id: Option<String>,
    failure: String,
    host_id: &MachineId,
) -> Result<BiosConfigJobAdvanceOutcome, StateHandlerError> {
    if retry_count >= MAX_BIOS_CONFIG_RETRIES {
        tracing::warn!(
            machine_id = %host_id,
            retry_count,
            max_retries = MAX_BIOS_CONFIG_RETRIES,
            %failure,
            "BIOS recovery budget exhausted; moving host to Failed for manual remediation"
        );
        return Ok(BiosConfigJobAdvanceOutcome::Failed {
            failure: format!(
                "{failure} (automated BIOS recovery exhausted after {MAX_BIOS_CONFIG_RETRIES} attempts)"
            ),
        });
    }
    Ok(BiosConfigJobAdvanceOutcome::Continue(BiosConfigInfo {
        bios_job_id,
        bios_config_state: BiosConfigState::HandleBiosJobFailure {
            failure,
            power_state: PowerState::Off,
        },
        retry_count: retry_count + 1,
    }))
}

pub(super) async fn advance_polling_bios_setup(
    redfish_client: &dyn Redfish,
    mh_snapshot: &ManagedHostStateSnapshot,
    retry_count: u32,
) -> Result<PollingBiosSetupOutcome, StateHandlerError> {
    let boot_interface_mac = mh_snapshot.boot_interface_mac().map(|m| m.to_string());
    let stuck_for = mh_snapshot.host_snapshot.state.version.since_state_change();
    let host_id = &mh_snapshot.host_snapshot.id;

    match redfish_client
        .is_bios_setup(boot_interface_mac.as_deref())
        .await
    {
        Ok(true) => {
            tracing::info!(
                machine_id = %host_id,
                "BIOS setup verified successfully"
            );
            Ok(PollingBiosSetupOutcome::Verified)
        }
        Ok(false) => {
            if let Some(outcome) =
                escalate_stuck_polling_bios_setup(retry_count, stuck_for, host_id)?
            {
                return Ok(outcome);
            }
            Ok(PollingBiosSetupOutcome::Wait(format!(
                "Polling BIOS setup status, waiting for settings to be applied (retry_count={retry_count})"
            )))
        }
        Err(e) => {
            tracing::warn!(
                machine_id = %host_id,
                error = %e,
                "Failed to check BIOS setup status, will retry"
            );
            Ok(PollingBiosSetupOutcome::Wait(format!(
                "Failed to check BIOS setup status: {e}. Will retry."
            )))
        }
    }
}

fn escalate_stuck_polling_bios_setup(
    retry_count: u32,
    stuck_for: chrono::Duration,
    host_id: &MachineId,
) -> Result<Option<PollingBiosSetupOutcome>, StateHandlerError> {
    if stuck_for <= POLLING_BIOS_SETUP_STUCK_THRESHOLD {
        return Ok(None);
    }

    tracing::warn!(
        machine_id = %host_id,
        ?stuck_for,
        retry_count,
        "PollingBiosSetup stuck; attempting HandleBiosJobFailure recovery (power-off + BMC reset + power-on + re-run machine_setup)"
    );

    let failure = format!(
        "PollingBiosSetup stuck for {} minutes (is_bios_setup returned false)",
        stuck_for.num_minutes()
    );

    Ok(Some(
        match try_bios_recovery_attempt(retry_count, None, failure, host_id)? {
            BiosConfigJobAdvanceOutcome::Continue(info) => {
                PollingBiosSetupOutcome::EnterRecovery(info)
            }
            BiosConfigJobAdvanceOutcome::Failed { failure } => {
                PollingBiosSetupOutcome::Failed { failure }
            }
            _ => unreachable!("recovery attempt only returns Continue or Failed"),
        },
    ))
}

pub(super) async fn handle_bios_setup_failed_recovery(
    ctx: &mut StateHandlerContext<'_, MachineStateHandlerContextObjects>,
    mh_snapshot: &ManagedHostStateSnapshot,
    machine_id: &MachineId,
    recovered_state: ManagedHostState,
) -> Result<StateHandlerOutcome<ManagedHostState>, StateHandlerError> {
    let boot_interface_mac = mh_snapshot.boot_interface_mac().map(|m| m.to_string());
    let redfish_client = ctx
        .services
        .create_redfish_client_from_machine(&mh_snapshot.host_snapshot)
        .await?;
    match redfish_client
        .is_bios_setup(boot_interface_mac.as_deref())
        .await
    {
        Ok(true) => {
            tracing::info!(
                machine_id = %machine_id,
                "BIOS setup verified after manual remediation; resuming state machine"
            );
            Ok(StateHandlerOutcome::transition(recovered_state))
        }
        Ok(false) => Ok(StateHandlerOutcome::do_nothing()),
        Err(e) => {
            tracing::warn!(
                machine_id = %machine_id,
                error = %e,
                "Failed to check BIOS setup status, will retry"
            );
            Ok(StateHandlerOutcome::do_nothing())
        }
    }
}

#[cfg(test)]
mod tests {
    use std::str::FromStr;

    use super::*;

    #[test]
    fn escalate_stuck_polling_bios_setup_not_triggered_before_threshold() {
        let host_id =
            MachineId::from_str("fm100ht7blqjsadm2uuh3qqbf1h7k8pmf47um6v9uckrg7l03po8mhqgvng")
                .unwrap();

        let result =
            escalate_stuck_polling_bios_setup(0, chrono::Duration::minutes(10), &host_id).unwrap();

        assert!(result.is_none());
    }

    #[test]
    fn escalate_stuck_polling_bios_setup_enters_handle_bios_job_failure_when_stuck() {
        let host_id =
            MachineId::from_str("fm100ht7blqjsadm2uuh3qqbf1h7k8pmf47um6v9uckrg7l03po8mhqgvng")
                .unwrap();

        let info = escalate_stuck_polling_bios_setup(0, chrono::Duration::minutes(16), &host_id)
            .unwrap()
            .expect("recovery should be triggered");
        let PollingBiosSetupOutcome::EnterRecovery(info) = info else {
            panic!("expected EnterRecovery");
        };
        assert_eq!(info.bios_job_id, None);
        assert_eq!(info.retry_count, 1);
        assert!(matches!(
            info.bios_config_state,
            BiosConfigState::HandleBiosJobFailure {
                power_state: PowerState::Off,
                ..
            }
        ));
    }

    #[test]
    fn escalate_stuck_polling_bios_setup_respects_shared_retry_budget() {
        let host_id =
            MachineId::from_str("fm100ht7blqjsadm2uuh3qqbf1h7k8pmf47um6v9uckrg7l03po8mhqgvng")
                .unwrap();

        let result = escalate_stuck_polling_bios_setup(
            MAX_BIOS_CONFIG_RETRIES,
            chrono::Duration::minutes(20),
            &host_id,
        )
        .unwrap()
        .expect("expected Failed outcome");

        assert!(matches!(result, PollingBiosSetupOutcome::Failed { .. }));
    }

    #[test]
    fn try_bios_recovery_attempt_fails_when_budget_exhausted() {
        let host_id =
            MachineId::from_str("fm100ht7blqjsadm2uuh3qqbf1h7k8pmf47um6v9uckrg7l03po8mhqgvng")
                .unwrap();

        let result = try_bios_recovery_attempt(
            MAX_BIOS_CONFIG_RETRIES,
            Some("job-1".to_string()),
            "job failed".to_string(),
            &host_id,
        )
        .unwrap();

        assert!(matches!(result, BiosConfigJobAdvanceOutcome::Failed { .. }));
    }

    #[test]
    fn escalate_stuck_polling_bios_setup_allows_last_budgeted_attempt() {
        let host_id =
            MachineId::from_str("fm100ht7blqjsadm2uuh3qqbf1h7k8pmf47um6v9uckrg7l03po8mhqgvng")
                .unwrap();

        let outcome = escalate_stuck_polling_bios_setup(
            MAX_BIOS_CONFIG_RETRIES - 1,
            chrono::Duration::minutes(20),
            &host_id,
        )
        .unwrap()
        .expect("last budgeted recovery should be allowed");

        let PollingBiosSetupOutcome::EnterRecovery(info) = outcome else {
            panic!("expected EnterRecovery");
        };
        assert_eq!(info.retry_count, MAX_BIOS_CONFIG_RETRIES);
    }
}
