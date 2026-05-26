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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tmocks "go.temporal.io/sdk/mocks"
)

func TestManageDpuExtensionService_CreateDpuExtensionServiceOnSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.CreateDpuExtensionServiceRequest
	}

	serviceID := "test-service-id"
	serviceName := "test-service-name"
	tenantOrgID := "test-tenant-org-id"

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test create DpuExtensionService success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.CreateDpuExtensionServiceRequest{
					ServiceId:            &serviceID,
					ServiceName:          serviceName,
					TenantOrganizationId: tenantOrgID,
				},
			},
			wantErr: false,
		},
		{
			name: "test create DpuExtensionService fail on missing ServiceId",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.CreateDpuExtensionServiceRequest{
					ServiceId:            nil,
					ServiceName:          serviceName,
					TenantOrganizationId: tenantOrgID,
				},
			},
			wantErr: true,
		},
		{
			name: "test create DpuExtensionService fail on empty ServiceId",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.CreateDpuExtensionServiceRequest{
					ServiceId:            util.GetStrPtr(""),
					ServiceName:          serviceName,
					TenantOrganizationId: tenantOrgID,
				},
			},
			wantErr: true,
		},
		{
			name: "test create DpuExtensionService fail on missing ServiceName",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.CreateDpuExtensionServiceRequest{
					ServiceId:            &serviceID,
					ServiceName:          "",
					TenantOrganizationId: tenantOrgID,
				},
			},
			wantErr: true,
		},
		{
			name: "test create DpuExtensionService fail on missing TenantOrganizationId",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.CreateDpuExtensionServiceRequest{
					ServiceId:            &serviceID,
					ServiceName:          serviceName,
					TenantOrganizationId: "",
				},
			},
			wantErr: true,
		},
		{
			name: "test create DpuExtensionService fail on missing request",
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
			mm := NewManageDpuExtensionService(tt.fields.NICoCoreAtomicClient)
			result, err := mm.CreateDpuExtensionServiceOnSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.args.request.ServiceId != nil {
				assert.Equal(t, *tt.args.request.ServiceId, result.ServiceId)
			}
			assert.Equal(t, tt.args.request.ServiceName, result.ServiceName)
			assert.Equal(t, tt.args.request.TenantOrganizationId, result.TenantOrganizationId)
		})
	}
}

func TestManageDpuExtensionService_UpdateDpuExtensionServiceOnSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.UpdateDpuExtensionServiceRequest
	}

	serviceID := "test-service-id"

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test update DpuExtensionService success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.UpdateDpuExtensionServiceRequest{
					ServiceId:   serviceID,
					ServiceName: util.GetStrPtr("test-service-name"),
				},
			},
			wantErr: false,
		},
		{
			name: "test update DpuExtensionService fail on missing ServiceId",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.UpdateDpuExtensionServiceRequest{
					ServiceId: "",
				},
			},
			wantErr: true,
		},
		{
			name: "test update DpuExtensionService fail on missing request",
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
			mm := NewManageDpuExtensionService(tt.fields.NICoCoreAtomicClient)
			result, err := mm.UpdateDpuExtensionServiceOnSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.args.request.ServiceId, result.ServiceId)

			if tt.args.request.ServiceName != nil {
				assert.Equal(t, *tt.args.request.ServiceName, result.ServiceName)
			}
		})
	}
}

func TestManageDpuExtensionService_DeleteDpuExtensionServiceOnSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.DeleteDpuExtensionServiceRequest
	}

	serviceID := "test-service-id"

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test delete DpuExtensionService success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.DeleteDpuExtensionServiceRequest{
					ServiceId: serviceID,
				},
			},
			wantErr: false,
		},
		{
			name: "test delete DpuExtensionService fail on missing ServiceId",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.DeleteDpuExtensionServiceRequest{
					ServiceId: "",
				},
			},
			wantErr: true,
		},
		{
			name: "test delete DpuExtensionService fail on missing request",
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
			mm := NewManageDpuExtensionService(tt.fields.NICoCoreAtomicClient)
			err := mm.DeleteDpuExtensionServiceOnSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManageDpuExtensionService_GetDpuExtensionServiceVersionsInfoOnSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.GetDpuExtensionServiceVersionsInfoRequest
	}

	serviceID := "test-service-id"

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantCount int
		wantErr   bool
	}{
		{
			name: "test get DpuExtensionService versions info success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.WithValue(context.Background(), "wantCount", 20),
				request: &cwssaws.GetDpuExtensionServiceVersionsInfoRequest{
					ServiceId: serviceID,
				},
			},
			wantCount: 20,
			wantErr:   false,
		},
		{
			name: "test get DpuExtensionService versions info fail on missing ServiceId",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.GetDpuExtensionServiceVersionsInfoRequest{
					ServiceId: "",
				},
			},
			wantErr: true,
		},
		{
			name: "test get DpuExtensionService versions info fail on missing request",
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
			mm := NewManageDpuExtensionService(tt.fields.NICoCoreAtomicClient)
			versionInfoList, err := mm.GetDpuExtensionServiceVersionsInfoOnSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(versionInfoList.VersionInfos))
			}
		})
	}
}

func TestManageDpuExtensionServiceInventory_DiscoverDpuExtensionServiceInventory(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	wid := "test-workflow-id"
	wrun := &tmocks.WorkflowRun{}
	wrun.On("GetID").Return(wid)

	type fields struct {
		siteID               uuid.UUID
		nicoCoreAtomicClient *cClient.NICoCoreAtomicClient
		temporalPublishQueue string
		sitePageSize         int
		cloudPageSize        int
	}
	type args struct {
		wantTotalItems int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test collecting and publishing dpu extension service inventory, empty inventory",
			fields: fields{
				siteID:               uuid.New(),
				nicoCoreAtomicClient: nicoCoreAtomicClient,
				temporalPublishQueue: "test-queue",
				sitePageSize:         100,
				cloudPageSize:        25,
			},
			args: args{
				wantTotalItems: 0,
			},
		},
		{
			name: "test collecting and publishing dpu extension service inventory, normal inventory",
			fields: fields{
				siteID:               uuid.New(),
				nicoCoreAtomicClient: nicoCoreAtomicClient,
				temporalPublishQueue: "test-queue",
				sitePageSize:         100,
				cloudPageSize:        25,
			},
			args: args{
				wantTotalItems: 195,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &tmocks.Client{}
			tc.Mock.On("ExecuteWorkflow", mock.Anything, mock.AnythingOfType("internal.StartWorkflowOptions"),
				mock.AnythingOfType("string"), mock.AnythingOfType("uuid.UUID"), mock.Anything).Return(wrun, nil)
			tc.AssertNumberOfCalls(t, "ExecuteWorkflow", 0)

			manageDpuExtensionService := NewManageDpuExtensionServiceInventory(ManageInventoryConfig{
				SiteID:                tt.fields.siteID,
				NICoCoreAtomicClient:  tt.fields.nicoCoreAtomicClient,
				TemporalPublishClient: tc,
				TemporalPublishQueue:  tt.fields.temporalPublishQueue,
				SitePageSize:          tt.fields.sitePageSize,
				CloudPageSize:         tt.fields.cloudPageSize,
			})

			ctx := context.Background()
			ctx = context.WithValue(ctx, "wantCount", tt.args.wantTotalItems)

			totalPages := tt.args.wantTotalItems / tt.fields.cloudPageSize
			if tt.args.wantTotalItems%tt.fields.cloudPageSize > 0 {
				totalPages++
			}

			err := manageDpuExtensionService.DiscoverDpuExtensionServiceInventory(ctx)
			assert.NoError(t, err)

			if tt.args.wantTotalItems == 0 {
				tc.AssertNumberOfCalls(t, "ExecuteWorkflow", 1)
			} else {
				tc.AssertNumberOfCalls(t, "ExecuteWorkflow", totalPages)
			}

			inventory, ok := tc.Calls[0].Arguments[4].(*cwssaws.DpuExtensionServiceInventory)
			assert.True(t, ok)

			if tt.args.wantTotalItems == 0 {
				assert.Equal(t, 0, len(inventory.DpuExtensionServices))
			} else {
				assert.Equal(t, tt.fields.cloudPageSize, len(inventory.DpuExtensionServices))
			}

			assert.Equal(t, cwssaws.InventoryStatus_INVENTORY_STATUS_SUCCESS, inventory.InventoryStatus)
			assert.Equal(t, totalPages, int(inventory.InventoryPage.TotalPages))
			assert.Equal(t, 1, int(inventory.InventoryPage.CurrentPage))
			assert.Equal(t, tt.fields.cloudPageSize, int(inventory.InventoryPage.PageSize))
			assert.Equal(t, tt.args.wantTotalItems, int(inventory.InventoryPage.TotalItems))
			assert.Equal(t, tt.args.wantTotalItems, len(inventory.InventoryPage.ItemIds))
		})
	}
}
