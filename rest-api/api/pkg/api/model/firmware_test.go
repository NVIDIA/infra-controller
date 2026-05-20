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
	"testing"

	flowv1 "github.com/NVIDIA/infra-controller-rest/workflow-schema/flow/protobuf/v1"
	"github.com/stretchr/testify/assert"
)

func TestAPIUpdateFirmwareRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request APIUpdateFirmwareRequest
		wantErr bool
	}{
		{
			name:    "valid - with siteId and version",
			request: APIUpdateFirmwareRequest{SiteID: "site-1", Version: strPtr("24.11.0")},
			wantErr: false,
		},
		{
			name:    "valid - with siteId only (no version)",
			request: APIUpdateFirmwareRequest{SiteID: "site-1"},
			wantErr: false,
		},
		{
			name:    "invalid - missing siteId",
			request: APIUpdateFirmwareRequest{Version: strPtr("24.11.0")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewAPIUpdateFirmwareResponse(t *testing.T) {
	tests := []struct {
		name     string
		resp     *flowv1.SubmitTaskResponse
		expected *APIUpdateFirmwareResponse
	}{
		{
			name:     "nil response returns empty task IDs",
			resp:     nil,
			expected: &APIUpdateFirmwareResponse{TaskIDs: []string{}},
		},
		{
			name: "single task ID",
			resp: &flowv1.SubmitTaskResponse{
				TaskIds: []*flowv1.UUID{{Id: "task-1"}},
			},
			expected: &APIUpdateFirmwareResponse{TaskIDs: []string{"task-1"}},
		},
		{
			name: "multiple task IDs",
			resp: &flowv1.SubmitTaskResponse{
				TaskIds: []*flowv1.UUID{{Id: "task-1"}, {Id: "task-2"}},
			},
			expected: &APIUpdateFirmwareResponse{TaskIDs: []string{"task-1", "task-2"}},
		},
		{
			name:     "empty task IDs",
			resp:     &flowv1.SubmitTaskResponse{},
			expected: &APIUpdateFirmwareResponse{TaskIDs: []string{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewAPIUpdateFirmwareResponse(tt.resp)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAPIBatchRackFirmwareUpdateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request APIBatchRackFirmwareUpdateRequest
		wantErr bool
	}{
		{
			name:    "valid - with siteId only",
			request: APIBatchRackFirmwareUpdateRequest{SiteID: "site-1"},
			wantErr: false,
		},
		{
			name: "valid - with filter and version",
			request: APIBatchRackFirmwareUpdateRequest{
				SiteID:  "site-1",
				Filter:  &RackFilter{Names: []string{"rack-1"}},
				Version: strPtr("1.0"),
			},
			wantErr: false,
		},
		{
			name:    "invalid - missing siteId",
			request: APIBatchRackFirmwareUpdateRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
