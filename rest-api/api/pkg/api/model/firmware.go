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

package model

import (
	"fmt"

	flowv1 "github.com/NVIDIA/infra-controller-rest/workflow-schema/flow/protobuf/v1"
)

// ========== Firmware Update Request ==========

// APIUpdateFirmwareRequest is the request body for firmware update operations
type APIUpdateFirmwareRequest struct {
	SiteID  string  `json:"siteId"`
	Version *string `json:"version,omitempty"`
}

// Validate validates the firmware update request
func (r *APIUpdateFirmwareRequest) Validate() error {
	if r.SiteID == "" {
		return fmt.Errorf("siteId is required")
	}
	return nil
}

// ========== Firmware Update Response ==========

// APIUpdateFirmwareResponse is the API response for firmware update operations
type APIUpdateFirmwareResponse struct {
	TaskIDs []string `json:"taskIds"`
}

// FromProto converts an Flow SubmitTaskResponse to an APIUpdateFirmwareResponse
func (r *APIUpdateFirmwareResponse) FromProto(resp *flowv1.SubmitTaskResponse) {
	if resp == nil {
		r.TaskIDs = []string{}
		return
	}
	r.TaskIDs = make([]string, 0, len(resp.GetTaskIds()))
	for _, id := range resp.GetTaskIds() {
		r.TaskIDs = append(r.TaskIDs, id.GetId())
	}
}

// NewAPIUpdateFirmwareResponse creates an APIUpdateFirmwareResponse from an Flow SubmitTaskResponse
func NewAPIUpdateFirmwareResponse(resp *flowv1.SubmitTaskResponse) *APIUpdateFirmwareResponse {
	r := &APIUpdateFirmwareResponse{}
	r.FromProto(resp)
	return r
}

// ========== Batch Rack Firmware Update Request ==========

// APIBatchRackFirmwareUpdateRequest is the JSON body for batch rack firmware update.
type APIBatchRackFirmwareUpdateRequest struct {
	SiteID  string      `json:"siteId"`
	Filter  *RackFilter `json:"filter,omitempty"`
	Version *string     `json:"version,omitempty"`
}

// Validate checks required fields.
func (r *APIBatchRackFirmwareUpdateRequest) Validate() error {
	if r.SiteID == "" {
		return fmt.Errorf("siteId is required")
	}
	return nil
}

// ========== Batch Tray Firmware Update Request ==========

// APIBatchTrayFirmwareUpdateRequest is the JSON body for batch tray firmware update.
type APIBatchTrayFirmwareUpdateRequest struct {
	SiteID  string      `json:"siteId"`
	Filter  *TrayFilter `json:"filter,omitempty"`
	Version *string     `json:"version,omitempty"`
}

// Validate checks required fields and filter constraints.
func (r *APIBatchTrayFirmwareUpdateRequest) Validate() error {
	if r.SiteID == "" {
		return fmt.Errorf("siteId is required")
	}
	return r.Filter.Validate()
}
