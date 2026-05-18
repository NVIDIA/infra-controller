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
	"errors"
	"time"

	swe "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/error"
	cClient "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"
	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ManageNVLinkLogicalPartitionInventory is an activity wrapper for NVLinkLogical Partition inventory collection and publishing
type ManageNVLinkLogicalPartitionInventory struct {
	config ManageInventoryConfig
}

// DiscoverNVLinkLogicalPartitionInventory is an activity to collect NVLinkLogical Partition inventory and publish to Temporal queue
func (mmi *ManageNVLinkLogicalPartitionInventory) DiscoverNVLinkLogicalPartitionInventory(ctx context.Context) error {
	logger := log.With().Str("Activity", "DiscoverNVLinkLogicalPartitionInventory").Logger()
	logger.Info().Msg("Starting activity")
	inventoryImpl := manageInventoryImpl[*cwssaws.NVLinkLogicalPartitionId, *cwssaws.NVLinkLogicalPartition, *cwssaws.NVLinkLogicalPartitionInventory]{
		itemType:               "NVLinkLogicalPartition",
		config:                 mmi.config,
		internalFindIDs:        nvllpFindIDs,
		internalFindByIDs:      nvllpFindByIDs,
		internalPagedInventory: nvllpPagedInventory,
	}
	return inventoryImpl.CollectAndPublishInventory(ctx, &logger)
}

// NewManageNVLinkLogicalPartitionInventory returns a ManageInventory implementation for NVLinkLogical Partition activity
func NewManageNVLinkLogicalPartitionInventory(config ManageInventoryConfig) ManageNVLinkLogicalPartitionInventory {
	return ManageNVLinkLogicalPartitionInventory{
		config: config,
	}
}

func nvllpFindIDs(ctx context.Context, nicoClient *cClient.NICoCoreClient) ([]*cwssaws.NVLinkLogicalPartitionId, error) {
	resp, err := nicoClient.NICo().FindNVLinkLogicalPartitionIds(ctx, &cwssaws.NVLinkLogicalPartitionSearchFilter{})
	if err != nil {
		return nil, err
	}
	ids := make([]*cwssaws.NVLinkLogicalPartitionId, len(resp.GetPartitionIds()))
	for i, id := range resp.GetPartitionIds() {
		ids[i] = id
	}
	return ids, nil
}

func nvllpFindByIDs(ctx context.Context, nicoClient *cClient.NICoCoreClient, ids []*cwssaws.NVLinkLogicalPartitionId) ([]*cwssaws.NVLinkLogicalPartition, error) {
	req := &cwssaws.NVLinkLogicalPartitionsByIdsRequest{
		PartitionIds: ids,
	}
	resp, err := nicoClient.NICo().FindNVLinkLogicalPartitionsByIds(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetPartitions(), nil
}

func nvllpPagedInventory(allItemIDs []*cwssaws.NVLinkLogicalPartitionId, pagedItems []*cwssaws.NVLinkLogicalPartition, input *pagedInventoryInput) *cwssaws.NVLinkLogicalPartitionInventory {
	itemIDs := []string{}
	for _, id := range allItemIDs {
		itemIDs = append(itemIDs, id.GetValue())
	}

	// Create an inventory page with the subset of NVLinkLogicalPartitions
	inventory := &cwssaws.NVLinkLogicalPartitionInventory{
		Partitions: pagedItems,
		Timestamp: &timestamppb.Timestamp{
			Seconds: time.Now().Unix(),
		},
		InventoryStatus: input.status,
		StatusMsg:       input.statusMessage,
		InventoryPage:   input.buildPage(),
	}

	if inventory.InventoryPage != nil {
		inventory.InventoryPage.ItemIds = itemIDs
	}

	return inventory
}

// ManageNVLinkLogicalPartition is an activity wrapper for NVLinkLogical Partition management
type ManageNVLinkLogicalPartition struct {
	NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
}

// NewManageNVLinkLogicalPartition returns a new ManageNVLinkLogicalPartition client
func NewManageNVLinkLogicalPartition(nicoClient *cClient.NICoCoreAtomicClient) ManageNVLinkLogicalPartition {
	return ManageNVLinkLogicalPartition{
		NICoCoreAtomicClient: nicoClient,
	}
}

// Function to create NVLinkLogical Partition with NICo
func (mnvllp *ManageNVLinkLogicalPartition) CreateNVLinkLogicalPartitionOnSite(ctx context.Context, request *cwssaws.NVLinkLogicalPartitionCreationRequest) (*cwssaws.NVLinkLogicalPartition, error) {
	logger := log.With().Str("Activity", "CreateNVLinkLogicalPartitionOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty create NVLink Logical Partition request")
	} else if request.Id == nil || request.GetId().GetValue() == "" {
		err = errors.New("received create NVLink Logical Partition request missing ID")
	} else if request.Config == nil {
		err = errors.New("received create NVLink Logical Partition request missing Config")
	} else if request.Config.Metadata == nil {
		err = errors.New("received create NVLink Logical Partition request missing Metadata")
	} else if request.Config.Metadata.Name == "" {
		err = errors.New("received create NVLink Logical Partition request missing Name")
	} else if request.Config.TenantOrganizationId == "" {
		err = errors.New("received create NVLink Logical Partition request missing TenantOrganizationId")
	}

	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mnvllp.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return nil, cClient.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	// Call NICo gRPC endpoint
	nvLinkLogicalPartition, err := rpcClient.CreateNVLinkLogicalPartition(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create NVLink Logical Partition using Site Controller API")
		return nil, swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")
	return nvLinkLogicalPartition, nil
}

// Function to update NVLinkLogical Partition with NICo
func (mnvllp *ManageNVLinkLogicalPartition) UpdateNVLinkLogicalPartitionOnSite(ctx context.Context, request *cwssaws.NVLinkLogicalPartitionUpdateRequest) error {
	logger := log.With().Str("Activity", "UpdateNVLinkLogicalPartitionOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty update NVLink Logical Partition request")
	} else if request.Id == nil || request.GetId().GetValue() == "" {
		err = errors.New("received update NVLink Logical Partition request missing ID")
	} else if request.Config == nil {
		err = errors.New("received update NVLink Logical Partition request missing Config")
	} else if request.Config.Metadata == nil {
		err = errors.New("received update NVLink Logical Partition request missing Metadata")
	} else if request.Config.Metadata.Name == "" {
		err = errors.New("received update NVLink Logical Partition request missing Name")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mnvllp.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return cClient.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	// Call NICo gRPC endpoint
	_, err = rpcClient.UpdateNVLinkLogicalPartition(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to update NVLink Logical Partition using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

// Function to delete NVLinkLogical Partition on NICo
func (mnvllp *ManageNVLinkLogicalPartition) DeleteNVLinkLogicalPartitionOnSite(ctx context.Context, request *cwssaws.NVLinkLogicalPartitionDeletionRequest) error {
	logger := log.With().Str("Activity", "DeleteNVLinkLogicalPartitionOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty delete NVLink Logical Partition request")
	} else if request.Id == nil || request.Id.GetValue() == "" {
		err = errors.New("received delete NVLink Logical Partition request without ID")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mnvllp.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return cClient.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	_, err = rpcClient.DeleteNVLinkLogicalPartition(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to delete NVLink Logical Partition using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}
