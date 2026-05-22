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

use ::rpc::forge as rpc;

use super::args::MaintenanceOptions;
use crate::errors::{CarbideCliError, CarbideCliResult};
use crate::rpc::ApiClient;

fn resolve_firmware_upgrade_source(
    args: &MaintenanceOptions,
) -> CarbideCliResult<(String, Option<String>)> {
    if args.firmware_version.is_some() && args.sot_json_file.is_some() {
        return Err(CarbideCliError::ChooseOneError(
            "--firmware-version",
            "--sot-json-file",
        ));
    }

    let firmware_version = if let Some(path) = args.sot_json_file.as_ref() {
        let config_json = std::fs::read_to_string(path)?;
        serde_json::from_str::<serde_json::Value>(&config_json)?;
        config_json
    } else {
        args.firmware_version.clone().unwrap_or_default()
    };

    let access_token = args.access_token.as_ref().and_then(|token| {
        if token.trim().is_empty() {
            None
        } else {
            Some(token.clone())
        }
    });

    if args.sot_json_file.is_some() && access_token.is_none() {
        return Err(CarbideCliError::GenericError(
            "--access-token is required with --sot-json-file".to_string(),
        ));
    }
    if args.access_token.is_some() && firmware_version.trim().is_empty() {
        return Err(CarbideCliError::GenericError(
            "--access-token requires SOT JSON from --sot-json-file or --firmware-version"
                .to_string(),
        ));
    }
    if access_token.is_some() {
        serde_json::from_str::<serde_json::Value>(&firmware_version)?;
    }

    Ok((firmware_version, access_token))
}

pub async fn on_demand_rack_maintenance(
    api_client: &ApiClient,
    args: MaintenanceOptions,
) -> CarbideCliResult<()> {
    use rpc::maintenance_activity_config::Activity as ProtoActivity;

    let (firmware_version, access_token) = resolve_firmware_upgrade_source(&args)?;
    let components = args.components.unwrap_or_default();
    let firmware_object_id = args.firmware_object_id.unwrap_or_default();
    let force_update = args.force_update;

    let activities: Vec<rpc::MaintenanceActivityConfig> = args
         .activities
         .unwrap_or_default()
         .iter()
         .map(|s| {
             let activity = match s.as_str() {
                 "firmware-upgrade" => Ok(ProtoActivity::FirmwareUpgrade(
                     rpc::FirmwareUpgradeActivity {
                         firmware_version: firmware_version.clone(),
                         components: components.clone(),
                         access_token: access_token.clone(),
                         force_update,
                     },
                 )),
                 "nvos-update" => Ok(ProtoActivity::NvosUpdate(
                     rpc::NvosUpdateActivity {
                         firmware_object_id: firmware_object_id.clone(),
                     },
                 )),
                 "configure-nmx-cluster" => Ok(ProtoActivity::ConfigureNmxCluster(
                     rpc::ConfigureNmxClusterActivity {},
                 )),
                 "power-sequence" => Ok(ProtoActivity::PowerSequence(
                     rpc::PowerSequenceActivity {},
                 )),
                 other => Err(eyre::eyre!(
                     "Unknown activity '{}'. Valid values: firmware-upgrade, nvos-update, configure-nmx-cluster, power-sequence",
                     other
                 )),
             }?;
            Ok::<_, eyre::Report>(rpc::MaintenanceActivityConfig {
                activity: Some(activity),
            })
         })
         .collect::<Result<Vec<_>, _>>()?;

    api_client
        .on_demand_rack_maintenance(
            args.rack,
            args.machine_ids.unwrap_or_default(),
            args.switch_ids.unwrap_or_default(),
            args.power_shelf_ids.unwrap_or_default(),
            activities,
        )
        .await?;
    println!("On-demand rack maintenance scheduled successfully.");
    Ok(())
}
