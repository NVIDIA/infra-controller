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

use std::collections::HashMap;

use db::ObjectColumnFilter;
use librms::protos::rack_manager as rms;
use rpc::forge::{
    ApplyFirmwareObjectRequest, ApplyFirmwareObjectResponse, CreateFirmwareObjectRequest,
    DeleteFirmwareObjectRequest, DeviceUpdateResult, FirmwareObject, FirmwareObjectHistoryRecord,
    FirmwareObjectHistoryRecords, FirmwareObjectHistoryRequest, FirmwareObjectHistoryResponse,
    FirmwareObjectJobStatusRequest, FirmwareObjectJobStatusResponse, FirmwareObjectList,
    FirmwareObjectSearchFilter, GetFirmwareObjectRequest, NodeJobInfo,
    SetDefaultFirmwareObjectRequest,
};
use tonic::{Request, Response, Status};

use crate::api::Api;
use crate::errors::CarbideError;
use crate::rack::firmware_object::profile_hardware_type_wire_value;
use crate::rack::firmware_update::{
    build_new_node_info, firmware_type_for_profile, load_rack_firmware_inventory,
};

fn rms_client(api: &Api) -> Result<&dyn librms::RmsApi, CarbideError> {
    api.rms_client
        .as_deref()
        .ok_or_else(|| CarbideError::FailedPrecondition("RMS client not configured".to_string()))
}

fn rms_error(operation: &str, error: librms::RackManagerError) -> Status {
    match error {
        librms::RackManagerError::ApiInvocationError(status) => status,
        error => CarbideError::Internal {
            message: format!("RMS {operation} failed: {error}"),
        }
        .into(),
    }
}

fn operation_failed(operation: &str, message: &str) -> CarbideError {
    CarbideError::FailedPrecondition(if message.is_empty() {
        format!("RMS {operation} failed")
    } else {
        format!("RMS {operation} failed: {message}")
    })
}

fn hardware_type_value(value: Option<rpc::common::RackHardwareType>) -> String {
    value.map(|value| value.value).unwrap_or_default()
}

fn rack_hardware_type(value: impl Into<String>) -> Option<rpc::common::RackHardwareType> {
    Some(rpc::common::RackHardwareType {
        value: value.into(),
    })
}

fn timestamp_to_string(timestamp: Option<&prost_types::Timestamp>) -> String {
    timestamp
        .and_then(|timestamp| {
            chrono::DateTime::<chrono::Utc>::from_timestamp(
                timestamp.seconds,
                timestamp.nanos as u32,
            )
        })
        .map(|timestamp| timestamp.format("%Y-%m-%d %H:%M:%S").to_string())
        .unwrap_or_default()
}

fn firmware_object_to_forge(object: rms::FirmwareObject) -> FirmwareObject {
    let parsed_components = object
        .metadata
        .as_ref()
        .and_then(|metadata| serde_json::to_string(metadata).ok())
        .unwrap_or_else(|| "{}".to_string());

    FirmwareObject {
        id: object.id,
        config_json: object.config_json,
        available: object.available,
        created: timestamp_to_string(object.created.as_ref()),
        updated: timestamp_to_string(object.updated.as_ref()),
        parsed_components,
        rack_hardware_type: rack_hardware_type(object.hardware_type),
        is_default: object.is_default,
    }
}

fn node_jobs_from_firmware(jobs: &[rms::NodeFirmwareJobInfo]) -> Vec<NodeJobInfo> {
    jobs.iter()
        .map(|job| NodeJobInfo {
            node_id: job.node_id.clone(),
            job_id: job.job_id.clone(),
        })
        .collect()
}

fn apply_response_to_forge(
    rack_id: &str,
    object_id: &str,
    response: rms::ApplyFirmwareObjectResponse,
) -> ApplyFirmwareObjectResponse {
    let batch = response.response.as_ref();
    let success = batch
        .map(|batch| batch.status == rms::ReturnCode::Success as i32)
        .unwrap_or(false);
    let message = batch
        .map(|batch| batch.message.clone())
        .unwrap_or_else(|| "RMS did not return a batch response".to_string());
    let total_nodes = batch
        .map(|batch| batch.total_nodes)
        .unwrap_or(response.node_jobs.len() as i32);
    let successful_nodes = batch
        .map(|batch| batch.successful_nodes)
        .unwrap_or(if success { total_nodes } else { 0 });
    let failed_nodes =
        batch
            .map(|batch| batch.failed_nodes)
            .unwrap_or(if success { 0 } else { total_nodes });
    let job_id = batch.map(|batch| batch.job_id.clone()).unwrap_or_default();

    let device_results = vec![DeviceUpdateResult {
        device_id: rack_id.to_string(),
        device_type: "rack".to_string(),
        success,
        message: format!("Firmware object {object_id} apply: {message}"),
        job_id,
        node_jobs: node_jobs_from_firmware(&response.node_jobs),
    }];

    ApplyFirmwareObjectResponse {
        total_updates: total_nodes,
        successful_updates: successful_nodes,
        failed_updates: failed_nodes,
        device_results,
    }
}

/// Create a new firmware object by proxying the SOT JSON to RMS.
pub async fn create(
    api: &Api,
    request: Request<CreateFirmwareObjectRequest>,
) -> Result<Response<FirmwareObject>, Status> {
    let req = request.into_inner();
    if req.artifactory_token.is_empty() {
        return Err(CarbideError::InvalidArgument("access token is required".to_string()).into());
    }

    let object = rms_client(api)?
        .add_firmware_object(rms::AddFirmwareObjectRequest {
            metadata: None,
            config_json: req.config_json,
            access_token: req.artifactory_token,
            hardware_type: hardware_type_value(req.rack_hardware_type),
            set_default: req.set_default,
        })
        .await
        .map_err(|error| rms_error("AddFirmwareObject", error))?;

    Ok(Response::new(firmware_object_to_forge(object)))
}

/// Get a firmware object by ID.
pub async fn get(
    api: &Api,
    request: Request<GetFirmwareObjectRequest>,
) -> Result<Response<FirmwareObject>, Status> {
    let req = request.into_inner();
    let object = rms_client(api)?
        .get_firmware_object(rms::GetFirmwareObjectRequest {
            metadata: None,
            id: req.id,
        })
        .await
        .map_err(|error| rms_error("GetFirmwareObject", error))?;

    Ok(Response::new(firmware_object_to_forge(object)))
}

/// List firmware objects.
pub async fn list(
    api: &Api,
    request: Request<FirmwareObjectSearchFilter>,
) -> Result<Response<FirmwareObjectList>, Status> {
    let req = request.into_inner();
    let response = rms_client(api)?
        .list_firmware_objects(rms::ListFirmwareObjectsRequest {
            metadata: None,
            only_available: req.only_available,
            hardware_type: hardware_type_value(req.rack_hardware_type),
        })
        .await
        .map_err(|error| rms_error("ListFirmwareObjects", error))?;

    Ok(Response::new(FirmwareObjectList {
        objects: response
            .objects
            .into_iter()
            .map(firmware_object_to_forge)
            .collect(),
    }))
}

/// Delete a firmware object.
pub async fn delete(
    api: &Api,
    request: Request<DeleteFirmwareObjectRequest>,
) -> Result<Response<()>, Status> {
    let req = request.into_inner();
    let response = rms_client(api)?
        .delete_firmware_object(rms::DeleteFirmwareObjectRequest {
            metadata: None,
            id: req.id,
        })
        .await
        .map_err(|error| rms_error("DeleteFirmwareObject", error))?;

    if response.status != rms::ReturnCode::Success as i32 {
        return Err(operation_failed("DeleteFirmwareObject", &response.message).into());
    }

    Ok(Response::new(()))
}

/// Apply a firmware object to all supported rack devices. BMM supplies inventory
/// and credentials; RMS resolves targets from the stored firmware object.
pub async fn apply(
    api: &Api,
    request: Request<ApplyFirmwareObjectRequest>,
) -> Result<Response<ApplyFirmwareObjectResponse>, Status> {
    let req = request.into_inner();
    let rack_id = req
        .rack_id
        .ok_or_else(|| CarbideError::InvalidArgument("rack_id is required".to_string()))?;

    let rack = db::rack::find_by(
        api.db_reader().as_mut(),
        ObjectColumnFilter::One(db::rack::IdColumn, &rack_id),
    )
    .await
    .map_err(CarbideError::from)?
    .pop()
    .ok_or_else(|| CarbideError::NotFoundError {
        kind: "rack",
        id: rack_id.to_string(),
    })?;

    let profile = rack
        .rack_profile_id
        .as_ref()
        .and_then(|profile_id| api.runtime_config.rack_profiles.get(profile_id.as_str()))
        .ok_or_else(|| {
            CarbideError::FailedPrecondition(format!(
                "rack '{rack_id}' has no known rack profile for firmware object apply"
            ))
        })?;
    let rack_hardware_type = profile_hardware_type_wire_value(profile);
    let expected_firmware_type = firmware_type_for_profile(profile);
    let firmware_type = if req.firmware_type.is_empty() {
        expected_firmware_type.to_string()
    } else {
        req.firmware_type
    };
    if firmware_type != expected_firmware_type {
        return Err(CarbideError::FailedPrecondition(format!(
            "Firmware type '{}' does not match rack '{}' hardware class expected '{}'",
            firmware_type, rack_id, expected_firmware_type
        ))
        .into());
    }

    let inventory = load_rack_firmware_inventory(
        &api.database_connection,
        api.credential_manager.as_ref(),
        &rack_id,
    )
    .await
    .map_err(|error| CarbideError::Internal {
        message: format!("failed to load rack firmware inventory: {error}"),
    })?;

    if inventory.machines.is_empty() && inventory.switches.is_empty() {
        return Err(CarbideError::FailedPrecondition(format!(
            "rack '{rack_id}' contains no compute or switch devices"
        ))
        .into());
    }

    let mut devices = Vec::with_capacity(inventory.machines.len() + inventory.switches.len());
    devices.extend(
        inventory
            .machines
            .iter()
            .map(|device| build_new_node_info(&rack_id, device, rms::NodeType::Compute)),
    );
    devices.extend(
        inventory
            .switches
            .iter()
            .map(|device| build_new_node_info(&rack_id, device, rms::NodeType::Switch)),
    );

    let object_id = req.object_id;
    tracing::info!(
        rack_id = %rack_id,
        object_id = %object_id,
        firmware_type = %firmware_type,
        hardware_type = %rack_hardware_type,
        node_count = devices.len(),
        "Firmware object apply starting"
    );
    let response = rms_client(api)?
        .apply_firmware_object(rms::ApplyFirmwareObjectRequest {
            metadata: None,
            rack_id: rack_id.to_string(),
            object_id: object_id.clone(),
            firmware_type,
            hardware_type: rack_hardware_type,
            components: Vec::new(),
            nodes: Some(rms::NodeSet { devices }),
            force_update: false,
            component_filters: HashMap::new(),
        })
        .await
        .map_err(|error| rms_error("ApplyFirmwareObject", error))?;

    let object_id = response.object_id.clone();
    Ok(Response::new(apply_response_to_forge(
        rack_id.as_ref(),
        &object_id,
        response,
    )))
}

/// Get the status of an async firmware update job by proxying to RMS.
pub async fn get_job_status(
    api: &Api,
    request: Request<FirmwareObjectJobStatusRequest>,
) -> Result<Response<FirmwareObjectJobStatusResponse>, Status> {
    let req = request.into_inner();

    if req.job_id.is_empty() {
        return Err(CarbideError::InvalidArgument("job_id is required".to_string()).into());
    }

    let rms_response = rms_client(api)?
        .get_firmware_job_status(rms::GetFirmwareJobStatusRequest {
            metadata: None,
            job_id: req.job_id,
        })
        .await
        .map_err(|error| rms_error("GetFirmwareJobStatus", error))?;

    let state = match rms_response.job_state {
        0 => "QUEUED",
        1 => "RUNNING",
        2 => "COMPLETED",
        3 => "FAILED",
        _ => "UNKNOWN",
    };

    Ok(Response::new(FirmwareObjectJobStatusResponse {
        job_id: rms_response.job_id,
        state: state.to_string(),
        state_description: rms_response.state_description,
        rack_id: rms_response.rack_id,
        node_id: rms_response.node_id,
        error_message: rms_response.error_message,
        result_json: rms_response.result_json,
    }))
}

/// Get firmware object apply history from RMS.
pub async fn get_history(
    api: &Api,
    request: Request<FirmwareObjectHistoryRequest>,
) -> Result<Response<FirmwareObjectHistoryResponse>, Status> {
    let req = request.into_inner();
    let response = rms_client(api)?
        .get_firmware_object_history(rms::GetFirmwareObjectHistoryRequest {
            metadata: None,
            object_id: req.object_id,
            rack_ids: req.rack_ids,
        })
        .await
        .map_err(|error| rms_error("GetFirmwareObjectHistory", error))?;

    let mut histories: HashMap<String, FirmwareObjectHistoryRecords> = HashMap::new();
    for record in response.records {
        let rack_id = record.rack_id.clone();
        histories
            .entry(rack_id)
            .or_default()
            .records
            .push(FirmwareObjectHistoryRecord {
                object_id: record.object_id,
                rack_id: record.rack_id,
                firmware_type: record.firmware_type,
                applied_at: timestamp_to_string(record.applied_at.as_ref()),
                firmware_available: record.firmware_available,
                rack_hardware_type: rack_hardware_type(record.hardware_type),
            });
    }

    Ok(Response::new(FirmwareObjectHistoryResponse { histories }))
}

/// Set a firmware object as default for its hardware type.
pub async fn set_default(
    api: &Api,
    request: Request<SetDefaultFirmwareObjectRequest>,
) -> Result<Response<()>, Status> {
    let req = request.into_inner();

    if req.object_id.is_empty() {
        return Err(CarbideError::InvalidArgument("object_id is required".to_string()).into());
    }

    rms_client(api)?
        .set_default_firmware_object(rms::SetDefaultFirmwareObjectRequest {
            metadata: None,
            object_id: req.object_id,
        })
        .await
        .map_err(|error| rms_error("SetDefaultFirmwareObject", error))?;

    Ok(Response::new(()))
}
