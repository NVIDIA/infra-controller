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

use eyre::ContextCompat;

use crate::grpcurl::grpcurl;

/// Simulate a DHCP discovery for `mac_address` arriving on `relay_address`, and return
/// the IP address NICo assigned to it. This is how a BMC that isn't driven by
/// machine-a-tron (e.g. a power shelf or switch) gets an underlay address that the
/// site explorer can then probe -- the same way a real DHCP relay announces a freshly
/// cabled BMC.
pub async fn discover(
    addrs: &[SocketAddr],
    mac_address: &str,
    relay_address: &str,
) -> eyre::Result<String> {
    let data = serde_json::json!({
        "mac_address": mac_address,
        "relay_address": relay_address,
    });
    let response = grpcurl(addrs, "DiscoverDhcp", Some(&data)).await?;
    let record: serde_json::Value = serde_json::from_str(&response)?;
    let address = record["address"]
        .as_str()
        .with_context(|| format!("DiscoverDhcp returned no address: {response}"))?;
    Ok(address.to_string())
}
