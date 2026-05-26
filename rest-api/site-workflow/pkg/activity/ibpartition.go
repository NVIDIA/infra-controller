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
	"github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"
	cClient "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"
	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ManageInfiniBandPartitionInventory is an activity wrapper for InfiniBand Partition inventory collection and publishing
type ManageInfiniBandPartitionInventory struct {
	config ManageInventoryConfig
}

// DiscoverInfiniBandPartitionInventory is an activity to collect InfiniBand Partition inventory and publish to Temporal queue
func (mmi *ManageInfiniBandPartitionInventory) DiscoverInfiniBandPartitionInventory(ctx context.Context) error {
	logger := log.With().Str("Activity", "DiscoverIBPartitionInventory").Logger()
	logger.Info().Msg("Starting activity")
	inventoryImpl := manageInventoryImpl[*cwssaws.IBPartitionId, *cwssaws.IBPartition, *cwssaws.InfiniBandPartitionInventory]{
		itemType:               "InfiniBandPartition",
		config:                 mmi.config,
		internalFindIDs:        ibpFindIDs,
		internalFindByIDs:      ibpFindByIDs,
		internalPagedInventory: ibpPagedInventory,
	}
	return inventoryImpl.CollectAndPublishInventory(ctx, &logger)
}

// NewManageInfiniBandPartitionInventory returns a ManageInventory implementation for InfiniBand Partition activity
func NewManageInfiniBandPartitionInventory(config ManageInventoryConfig) ManageInfiniBandPartitionInventory {
	return ManageInfiniBandPartitionInventory{
		config: config,
	}
}

func ibpFindIDs(ctx context.Context, nicoClient *cClient.NICoCoreClient) ([]*cwssaws.IBPartitionId, error) {
	idList, err := nicoClient.NICo().FindIBPartitionIds(ctx, &cwssaws.IBPartitionSearchFilter{})
	if err != nil {
		return nil, err
	}
	return idList.GetIbPartitionIds(), nil
}

func ibpFindByIDs(ctx context.Context, nicoClient *cClient.NICoCoreClient, ids []*cwssaws.IBPartitionId) ([]*cwssaws.IBPartition, error) {
	list, err := nicoClient.NICo().FindIBPartitionsByIds(ctx, &cwssaws.IBPartitionsByIdsRequest{
		IbPartitionIds: ids,
	})
	if err != nil {
		return nil, err
	}
	return list.GetIbPartitions(), nil
}

func ibpPagedInventory(allItemIDs []*cwssaws.IBPartitionId, pagedItems []*cwssaws.IBPartition, input *pagedInventoryInput) *cwssaws.InfiniBandPartitionInventory {
	itemIDs := []string{}
	for _, id := range allItemIDs {
		itemIDs = append(itemIDs, id.GetValue())
	}

	// Create an inventory page with the subset of VPCs
	inventory := &cwssaws.InfiniBandPartitionInventory{
		IbPartitions: pagedItems,
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

// ManageInfiniBandPartition is an activity wrapper for InfiniBand Partition management
type ManageInfiniBandPartition struct {
	NICoCoreAtomicClient *client.NICoCoreAtomicClient
}

// NewManageInfiniBandPartition returns a new ManageInfiniBandPartition client
func NewManageInfiniBandPartition(nicoClient *client.NICoCoreAtomicClient) ManageInfiniBandPartition {
	return ManageInfiniBandPartition{
		NICoCoreAtomicClient: nicoClient,
	}
}

// Function to create InfiniBand Partition with NICo
func (mibp *ManageInfiniBandPartition) CreateInfiniBandPartitionOnSite(ctx context.Context, request *cwssaws.IBPartitionCreationRequest) error {
	logger := log.With().Str("Activity", "CreateInfiniBandPartitionOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty create InfiniBand Partition request")
	} else if request.Id == nil || request.GetId().GetValue() == "" {
		err = errors.New("received create InfiniBand Partition request without ID")
	} else if request.GetConfig() == nil {
		err = errors.New("received create InfiniBand Partition request without Config")
	} else if request.GetMetadata().GetName() == "" && request.GetConfig().GetName() == "" {
		// Backward compatibility: both Metadata.Name and Config.Name are accepted
		err = errors.New("received create InfiniBand Partition request without Name")
	} else if request.GetConfig().GetTenantOrganizationId() == "" {
		err = errors.New("received create InfiniBand Partition request without TenantOrganizationId")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mibp.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return client.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	// Call NICo gRPC endpoint
	_, err = rpcClient.CreateIBPartition(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create InfiniBand Partition using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

// UpdateInfiniBandPartitionOnSite applies an IB partition update on the site NICo controller
func (mibp *ManageInfiniBandPartition) UpdateInfiniBandPartitionOnSite(ctx context.Context, request *cwssaws.IBPartitionUpdateRequest) error {
	logger := log.With().Str("Activity", "UpdateInfiniBandPartitionOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	if request == nil {
		err = errors.New("received empty update InfiniBand Partition request")
	} else if request.Id == nil || request.GetId().GetValue() == "" {
		err = errors.New("received update InfiniBand Partition request without ID")
	} else if request.GetConfig() == nil && request.GetMetadata() == nil {
		err = errors.New("received update InfiniBand Partition request without config or metadata")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	nicoClient := mibp.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return client.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	_, err = rpcClient.UpdateIBPartition(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to update InfiniBand Partition using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

// Function to delete InfiniBand Partition on NICo
func (mipb *ManageInfiniBandPartition) DeleteInfiniBandPartitionOnSite(ctx context.Context, request *cwssaws.IBPartitionDeletionRequest) error {
	logger := log.With().Str("Activity", "DeleteInfiniBandPartitionOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty delete InfiniBand Partition request")
	} else if request.Id == nil || request.Id.GetValue() == "" {
		err = errors.New("received delete InfiniBand Partition request without ID")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mipb.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return client.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	_, err = rpcClient.DeleteIBPartition(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to delete InfiniBand Partition using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}
