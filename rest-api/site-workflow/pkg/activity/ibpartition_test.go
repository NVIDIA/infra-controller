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
	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	tmocks "go.temporal.io/sdk/mocks"
)

func TestManageIBPartitionInventory_DiscoverInfiniBandPartitionInventory(t *testing.T) {
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
			name: "test collecting and publishing ib partition inventory, empty inventory",
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
			name: "test collecting and publishing ib partition inventory, normal inventory",
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

			manageInstance := NewManageInfiniBandPartitionInventory(ManageInventoryConfig{
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

			err := manageInstance.DiscoverInfiniBandPartitionInventory(ctx)
			assert.NoError(t, err)

			if tt.args.wantTotalItems == 0 {
				tc.AssertNumberOfCalls(t, "ExecuteWorkflow", 1)
			} else {
				tc.AssertNumberOfCalls(t, "ExecuteWorkflow", totalPages)
			}

			inventory, ok := tc.Calls[0].Arguments[4].(*cwssaws.InfiniBandPartitionInventory)
			assert.True(t, ok)

			if tt.args.wantTotalItems == 0 {
				assert.Equal(t, 0, len(inventory.IbPartitions))
			} else {
				assert.Equal(t, tt.fields.cloudPageSize, len(inventory.IbPartitions))
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

func TestManageIBPartition_CreateIBPartitionOnSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	name := "best_ibpartition_ever"
	org := "best_org_ever"

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.IBPartitionCreationRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test create ibpartition success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.IBPartitionCreationRequest{
					Id: &cwssaws.IBPartitionId{Value: "b410867c-655a-11ef-bc4a-0393098e5d09"},
					Config: &cwssaws.IBPartitionConfig{
						Name:                 name,
						TenantOrganizationId: org,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test create ibpartition fail on missing name",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.IBPartitionCreationRequest{
					Id: &cwssaws.IBPartitionId{Value: "b410867c-655a-11ef-bc4a-0393098e5d09"},
					Config: &cwssaws.IBPartitionConfig{
						Name:                 "",
						TenantOrganizationId: org,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test create ibpartition fail on missing org",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.IBPartitionCreationRequest{
					Id: &cwssaws.IBPartitionId{Value: "b410867c-655a-11ef-bc4a-0393098e5d09"},
					Config: &cwssaws.IBPartitionConfig{
						Name:                 name,
						TenantOrganizationId: "",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test create ibpartition fail on missing id",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.IBPartitionCreationRequest{
					Id: nil,
					Config: &cwssaws.IBPartitionConfig{
						Name:                 name,
						TenantOrganizationId: org,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test create ibpartition fail on missing request",
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
			name: "test create ibpartition success with Metadata only (no deprecated Config)",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.IBPartitionCreationRequest{
					Id: &cwssaws.IBPartitionId{Value: "b410867c-655a-11ef-bc4a-0393098e5d09"},
					Config: &cwssaws.IBPartitionConfig{
						TenantOrganizationId: org,
					},
					Metadata: &cwssaws.Metadata{
						Name: name,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test create ibpartition success with deprecated Config only (backward compatibility)",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.IBPartitionCreationRequest{
					Id: &cwssaws.IBPartitionId{Value: "b410867c-655a-11ef-bc4a-0393098e5d09"},
					Config: &cwssaws.IBPartitionConfig{
						Name:                 name,
						TenantOrganizationId: org,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := NewManageInfiniBandPartition(tt.fields.NICoCoreAtomicClient)
			err := mm.CreateInfiniBandPartitionOnSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManageIBPartition_DeleteIBPartitionOnSite(t *testing.T) {
	mockNICo := cClient.NewMockNICoClient()

	nicoCoreAtomicClient := cClient.NewNICoCoreAtomicClient(&cClient.NICoCoreClientConfig{})
	nicoCoreAtomicClient.SwapClient(mockNICo)

	type fields struct {
		NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
	}
	type args struct {
		ctx     context.Context
		request *cwssaws.IBPartitionDeletionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test delete ibpartition success",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.IBPartitionDeletionRequest{
					Id: &cwssaws.IBPartitionId{Value: "b410867c-655a-11ef-bc4a-0393098e5d09"},
				},
			},
			wantErr: false,
		},
		{
			name: "test delete ibpartition fail on blank id",
			fields: fields{
				NICoCoreAtomicClient: nicoCoreAtomicClient,
			},
			args: args{
				ctx: context.Background(),
				request: &cwssaws.IBPartitionDeletionRequest{
					Id: &cwssaws.IBPartitionId{Value: ""},
				},
			},
			wantErr: true,
		},
		{
			name: "test delete ibpartition fail on missing request",
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
			mm := NewManageInfiniBandPartition(tt.fields.NICoCoreAtomicClient)
			err := mm.DeleteInfiniBandPartitionOnSite(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
