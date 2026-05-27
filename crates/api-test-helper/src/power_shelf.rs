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

use std::net::SocketAddr;
use std::time::{Duration, Instant};

use eyre::ContextCompat;
use tokio::time::sleep;

use crate::grpcurl::grpcurl;

/// Upper bound on waiting for the site explorer to discover a BMC and link its
/// expected entity to a managed one. The happy path is a few seconds (the explorer
/// runs on a one-second loop), but giving it some extra time, since we give 90
/// seconds for equivalent machine runs.
const LINK_TIMEOUT: Duration = Duration::from_secs(60);

/// Register an expected power shelf. The site explorer matches a discovered BMC to
/// this entry by `bmc_mac_address`, then creates and links a managed PowerShelf.
pub async fn add_expected(
    addrs: &[SocketAddr],
    bmc_mac_address: &str,
    bmc_username: &str,
    bmc_password: &str,
    serial_number: &str,
) -> eyre::Result<()> {
    let data = serde_json::json!({
        "bmc_mac_address": bmc_mac_address,
        "bmc_username": bmc_username,
        "bmc_password": bmc_password,
        "shelf_serial_number": serial_number,
        "metadata": { "name": serial_number },
    });
    grpcurl(addrs, "AddExpectedPowerShelf", Some(&data)).await?;
    Ok(())
}

/// Poll until the expected power shelf identified by `bmc_mac_address` is linked to a
/// managed PowerShelf, returning that PowerShelf's id. Fails if it is not linked
/// within [`LINK_TIMEOUT`].
pub async fn wait_for_linked(addrs: &[SocketAddr], bmc_mac_address: &str) -> eyre::Result<String> {
    let start = Instant::now();
    loop {
        let response = grpcurl(addrs, "GetAllExpectedPowerShelvesLinked", Some("{}")).await?;
        let linked: serde_json::Value = serde_json::from_str(&response)?;
        if let Some(entries) = linked["expectedPowerShelves"].as_array() {
            for entry in entries {
                let matches_mac = entry["bmcMacAddress"]
                    .as_str()
                    .is_some_and(|m| m.eq_ignore_ascii_case(bmc_mac_address));
                if matches_mac && let Some(id) = entry["powerShelfId"]["id"].as_str() {
                    return Ok(id.to_string());
                }
            }
        }
        if start.elapsed() > LINK_TIMEOUT {
            eyre::bail!(
                "expected power shelf {bmc_mac_address} was not linked to a managed \
                 power shelf within {LINK_TIMEOUT:?}"
            );
        }
        sleep(Duration::from_secs(2)).await;
    }
}

/// Fetch a managed power shelf by id via FindPowerShelvesByIds, returning the id the
/// API echoes back. Used to confirm a linked PowerShelf is actually retrievable.
pub async fn find_by_id(addrs: &[SocketAddr], power_shelf_id: &str) -> eyre::Result<String> {
    let data = serde_json::json!({
        "power_shelf_ids": [{ "id": power_shelf_id }],
    });
    let response = grpcurl(addrs, "FindPowerShelvesByIds", Some(&data)).await?;
    let list: serde_json::Value = serde_json::from_str(&response)?;
    let id = list["powerShelves"][0]["id"]["id"]
        .as_str()
        .with_context(|| format!("FindPowerShelvesByIds returned no power shelf: {response}"))?;
    Ok(id.to_string())
}
