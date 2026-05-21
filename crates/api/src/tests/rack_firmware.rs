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

use common::api_fixtures::create_test_env;
use rpc::forge::{
    RackFirmwareCreateRequest, RackFirmwareDeleteRequest, RackFirmwareGetRequest,
    RackFirmwareSearchFilter, RackFirmwareSetDefaultRequest,
};
use rpc::protos::forge::forge_server::Forge;

use crate::tests::common;

fn firmware_json(id: &str) -> String {
    serde_json::json!({
        "Id": id,
        "Name": "Test Firmware Object",
        "BoardSKUs": []
    })
    .to_string()
}

fn create_request(id: &str, hardware_type: &str) -> RackFirmwareCreateRequest {
    RackFirmwareCreateRequest {
        rack_hardware_type: Some(rpc::common::RackHardwareType {
            value: hardware_type.to_string(),
        }),
        config_json: firmware_json(id),
        artifactory_token: "test-token".to_string(),
        set_default: false,
    }
}

#[crate::sqlx_test()]
async fn test_create_rack_firmware_proxies_to_rms(
    pool: sqlx::PgPool,
) -> Result<(), Box<dyn std::error::Error>> {
    let env = create_test_env(pool).await;

    let response = env
        .api
        .create_rack_firmware(tonic::Request::new(create_request("fw-001", "any")))
        .await?
        .into_inner();

    assert_eq!(response.id, "fw-001");
    assert_eq!(response.config_json, firmware_json("fw-001"));
    assert_eq!(response.rack_hardware_type.unwrap().value, "any");
    assert!(response.is_default);

    Ok(())
}

#[crate::sqlx_test()]
async fn test_get_and_list_rack_firmware_proxy_to_rms(
    pool: sqlx::PgPool,
) -> Result<(), Box<dyn std::error::Error>> {
    let env = create_test_env(pool).await;

    env.api
        .create_rack_firmware(tonic::Request::new(create_request("fw-001", "any")))
        .await?;
    env.api
        .create_rack_firmware(tonic::Request::new(create_request("fw-002", "gb200")))
        .await?;

    let got = env
        .api
        .get_rack_firmware(tonic::Request::new(RackFirmwareGetRequest {
            id: "fw-001".to_string(),
        }))
        .await?
        .into_inner();
    assert_eq!(got.id, "fw-001");

    let listed = env
        .api
        .list_rack_firmware(tonic::Request::new(RackFirmwareSearchFilter {
            only_available: false,
            rack_hardware_type: Some(rpc::common::RackHardwareType {
                value: "gb200".to_string(),
            }),
        }))
        .await?
        .into_inner();
    assert_eq!(listed.configs.len(), 1);
    assert_eq!(listed.configs[0].id, "fw-002");

    Ok(())
}

#[crate::sqlx_test()]
async fn test_delete_rack_firmware_proxies_to_rms(
    pool: sqlx::PgPool,
) -> Result<(), Box<dyn std::error::Error>> {
    let env = create_test_env(pool).await;

    env.api
        .create_rack_firmware(tonic::Request::new(create_request("fw-delete", "any")))
        .await?;
    env.api
        .delete_rack_firmware(tonic::Request::new(RackFirmwareDeleteRequest {
            id: "fw-delete".to_string(),
        }))
        .await?;

    let err = env
        .api
        .get_rack_firmware(tonic::Request::new(RackFirmwareGetRequest {
            id: "fw-delete".to_string(),
        }))
        .await
        .expect_err("deleted firmware object should not be returned");
    assert_eq!(err.code(), tonic::Code::NotFound);
    assert!(err.message().contains("not found"));

    Ok(())
}

#[crate::sqlx_test()]
async fn test_set_default_rack_firmware_proxies_to_rms(
    pool: sqlx::PgPool,
) -> Result<(), Box<dyn std::error::Error>> {
    let env = create_test_env(pool).await;

    env.api
        .create_rack_firmware(tonic::Request::new(create_request("fw-a", "any")))
        .await?;
    env.api
        .create_rack_firmware(tonic::Request::new(create_request("fw-b", "any")))
        .await?;
    env.api
        .rack_firmware_set_default(tonic::Request::new(RackFirmwareSetDefaultRequest {
            firmware_id: "fw-b".to_string(),
        }))
        .await?;

    let got = env
        .api
        .get_rack_firmware(tonic::Request::new(RackFirmwareGetRequest {
            id: "fw-b".to_string(),
        }))
        .await?
        .into_inner();
    assert!(got.is_default);

    Ok(())
}
