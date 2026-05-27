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

package activity

import (
	"context"
	"testing"

	cClient "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"
	"github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/util"
	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
	"github.com/stretchr/testify/assert"
)

func TestManageMachineValidation_EnableDisableMachineValidationTestOnSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	testID := "test-id-1"
	testVersion := "test-version-1"

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.MachineValidationTestEnableDisableTestRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test enable validation test success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.MachineValidationTestEnableDisableTestRequest{
					TestId:    testID,
					Version:   testVersion,
					IsEnabled: true,
				},
			},
			wantErr: false,
		},
		{
			name: "test enable validation test fails on missing ID",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.MachineValidationTestEnableDisableTestRequest{
					Version:   testVersion,
					IsEnabled: true,
				},
			},
			wantErr: true,
		},
		{
			name: "test enable validation test fails on missing Version",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.MachineValidationTestEnableDisableTestRequest{
					TestId:    testID,
					IsEnabled: true,
				},
			},
			wantErr: true,
		},
		{
			name: "test enable validation test fails on missing request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := NewManageMachineValidation(tt.fields.NICoCoreAtomicClient)
			err := mt.EnableDisableMachineValidationTestOnSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManageMachineValidation_PersistValidationResultOnSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.MachineValidationResultPostRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test persist validation result success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.MachineValidationResultPostRequest{
					Result: &cwssaws.MachineValidationResult{
						Name: "test-1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test persist validation result fails on missing Result",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: &cwssaws.MachineValidationResultPostRequest{},
			},
			wantErr: true,
		},
		{
			name: "test persist validation result fails on missing request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := NewManageMachineValidation(tt.fields.NICoCoreAtomicClient)
			err := mt.PersistValidationResultOnSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManageMachineValidation_GetMachineValidationResultsFromSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.MachineValidationGetRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test get machine validation results success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.MachineValidationGetRequest{
					MachineId: &cwssaws.MachineId{
						Id: "machine-id-1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test get machine validation results fails on missing request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := NewManageMachineValidation(tt.fields.NICoCoreAtomicClient)
			_, err := mt.GetMachineValidationResultsFromSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManageMachineValidation_GetMachineValidationRunsFromSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.MachineValidationRunListGetRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test get machine validation runs success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.MachineValidationRunListGetRequest{
					MachineId: &cwssaws.MachineId{
						Id: "machine-id-1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test get machine validation runs fails on missing request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: nil,
			},
			wantErr: true,
		},
		{
			name: "test get machine validation runs fails on missing MachineId in request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: &cwssaws.MachineValidationRunListGetRequest{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := NewManageMachineValidation(tt.fields.NICoCoreAtomicClient)
			_, err := mt.GetMachineValidationRunsFromSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManageMachineValidation_GetMachineValidationTestsFromSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.MachineValidationTestsGetRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test get machine validation tests success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: &cwssaws.MachineValidationTestsGetRequest{},
			},
			wantErr: false,
		},
		{
			name: "test get machine validation tests fails on missing request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := NewManageMachineValidation(tt.fields.NICoCoreAtomicClient)
			_, err := mt.GetMachineValidationTestsFromSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManageMachineValidation_AddMachineValidationTestOnSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		request *cwssaws.MachineValidationTestAddRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		wantID  string
	}{
		{
			name: "test add machine validation test success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				request: &cwssaws.MachineValidationTestAddRequest{
					Name:    "test-1",
					Command: "test-command",
					Args:    "test-args",
				},
			},
			wantID:  "test-id-1",
			wantErr: false,
		},
		{
			name: "test add machine validation test fails on missing request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				request: nil,
			},
			wantErr: true,
		},
		{
			name: "test add machine validation test fails on empty request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				request: &cwssaws.MachineValidationTestAddRequest{},
			},
			wantErr: true,
		},
		{
			name: "test add machine validation test fails on missing Name in request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				request: &cwssaws.MachineValidationTestAddRequest{
					Command: "test-command",
					Args:    "test-args",
				},
			},
			wantErr: true,
		},
		{
			name: "test add machine validation test fails on missing Command in request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				request: &cwssaws.MachineValidationTestAddRequest{
					Name: "test-1",
					Args: "test-args",
				},
			},
			wantErr: true,
		},
		{
			name: "test add machine validation test fails on missing Args in request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				request: &cwssaws.MachineValidationTestAddRequest{
					Name:    "test-1",
					Command: "test-command",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.wantID != "" {
				ctx = context.WithValue(ctx, "wantID", tt.wantID)
			}
			mt := NewManageMachineValidation(tt.fields.NICoCoreAtomicClient)
			response, err := mt.AddMachineValidationTestOnSite(ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.wantID, response.TestId)
			}
		})
	}
}

func TestManageMachineValidation_UpdateMachineValidationTestOnSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.MachineValidationTestUpdateRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test update machine validation test success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.MachineValidationTestUpdateRequest{
					TestId:  "test-1",
					Version: "version-1",
					Payload: &cwssaws.MachineValidationTestUpdateRequest_Payload{
						Name: util.GetStrPtr("name-2"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test update machine validation test fails on missing request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: nil,
			},
			wantErr: true,
		},
		{
			name: "test update machine validation test fails on empty request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: &cwssaws.MachineValidationTestUpdateRequest{},
			},
			wantErr: true,
		},
		{
			name: "test update machine validation test fails on missing TestId in request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.MachineValidationTestUpdateRequest{
					Version: "version-1",
					Payload: &cwssaws.MachineValidationTestUpdateRequest_Payload{
						Name: util.GetStrPtr("name-2"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test update machine validation test fails on missing Version in request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.MachineValidationTestUpdateRequest{
					TestId: "test-1",
					Payload: &cwssaws.MachineValidationTestUpdateRequest_Payload{
						Name: util.GetStrPtr("name-2"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test update machine validation test fails on missing Payload in request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.MachineValidationTestUpdateRequest{
					TestId:  "test-1",
					Version: "version-1",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := NewManageMachineValidation(tt.fields.NICoCoreAtomicClient)
			err := mt.UpdateMachineValidationTestOnSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManageMachineValidation_GetMachineValidationExternalConfigsFromSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.GetMachineValidationExternalConfigsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test get machine validation external configs success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: &cwssaws.GetMachineValidationExternalConfigsRequest{},
			},
			wantErr: false,
		},
		{
			name: "test get machine validation external configs fails on missing request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := NewManageMachineValidation(tt.fields.NICoCoreAtomicClient)
			_, err := mt.GetMachineValidationExternalConfigsFromSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManageMachineValidation_AddUpdateMachineValidationExternalConfigOnSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.AddUpdateMachineValidationExternalConfigRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test add/update machine validation external config success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.AddUpdateMachineValidationExternalConfigRequest{
					Name: "test-1",
				},
			},
			wantErr: false,
		},
		{
			name: "test add/update machine validation external config fails on missing request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: nil,
			},
			wantErr: true,
		},
		{
			name: "test add/update machine validation external config fails on empty request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: &cwssaws.AddUpdateMachineValidationExternalConfigRequest{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := NewManageMachineValidation(tt.fields.NICoCoreAtomicClient)
			err := mt.AddUpdateMachineValidationExternalConfigOnSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManageMachineValidation_RemoveMachineValidationExternalConfigOnSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.RemoveMachineValidationExternalConfigRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test remove machine validation external config success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.RemoveMachineValidationExternalConfigRequest{
					Name: "test-1",
				},
			},
			wantErr: false,
		},
		{
			name: "test remove machine validation external config fails on missing request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: nil,
			},
			wantErr: true,
		},
		{
			name: "test remove machine validation external config fails on empty request",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx:     context.Background(),
				request: &cwssaws.RemoveMachineValidationExternalConfigRequest{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := NewManageMachineValidation(tt.fields.NICoCoreAtomicClient)
			err := mt.RemoveMachineValidationExternalConfigOnSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
