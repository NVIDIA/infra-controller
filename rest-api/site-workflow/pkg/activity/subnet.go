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
	"net"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"

	swe "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/error"
	cClient "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"
	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

// ManageSubnet is an activity wrapper for Subnet management tasks that allows injecting DB access
type ManageSubnet struct {
	NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
}

// Function to Create Subnets with the Site Controller
func (mm *ManageSubnet) CreateSubnetOnSite(ctx context.Context, request *cwssaws.NetworkSegmentCreationRequest) error {
	logger := log.With().Str("Activity", "CreateSubnetOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	switch {
	case request == nil:
		err = errors.New("received empty create Subnet request")
	case request.Name == "":
		err = errors.New("received create Subnet request without name")
	case request.VpcId == nil:
		err = errors.New("received create Subnet request without VPC ID")
	case len(request.Prefixes) == 0:
		err = errors.New("received create Subnet request with empty prefix list")
	case len(request.Prefixes) > 0:
		for _, prefix := range request.Prefixes {
			if prefix == nil {
				err = errors.New("received create Subnet request with a nil prefix in the prefix list")
				break
			}
			if _, _, err = net.ParseCIDR(prefix.Prefix); err != nil {
				err = errors.New("received create Subnet request with an invalid prefix in the prefix list")
				break
			}
		}
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mm.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return cClient.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	_, err = rpcClient.CreateNetworkSegment(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create Subnet using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

// Function to Delete Subnets with the Site Controller
func (mm *ManageSubnet) DeleteSubnetOnSite(ctx context.Context, request *cwssaws.NetworkSegmentDeletionRequest) error {
	logger := log.With().Str("Activity", "DeleteSubnetOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	switch {
	case request == nil:
		err = errors.New("received empty delete Subnet request")
	case request.Id == nil:
		err = errors.New("received delete Subnet request without subnet ID")
	case request.Id.Value == "":
		err = errors.New("received delete Subnet request with empty subnet ID")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mm.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return cClient.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	_, err = rpcClient.DeleteNetworkSegment(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to delete Subnet using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

// NewManageSubnet returns a new ManageSubnet client
func NewManageSubnet(nicoClient *cClient.NICoCoreAtomicClient) ManageSubnet {
	return ManageSubnet{
		NICoCoreAtomicClient: nicoClient,
	}
}

// ManageSubnetInventory is an activity wrapper for Subnet inventory collection and publishing
type ManageSubnetInventory struct {
	config ManageInventoryConfig
}

// DiscoverSubnetInventory is an activity to collect Subnet inventory and publish to Temporal queue
func (mmi *ManageSubnetInventory) DiscoverSubnetInventory(ctx context.Context) error {
	logger := log.With().Str("Activity", "DiscoverSubnetInventory").Logger()
	logger.Info().Msg("Starting activity")
	inventoryImpl := manageInventoryImpl[*cwssaws.NetworkSegmentId, *cwssaws.NetworkSegment, *cwssaws.SubnetInventory]{
		itemType:               "Subnet",
		config:                 mmi.config,
		internalFindIDs:        subnetFindIDs,
		internalFindByIDs:      subnetFindByIDs,
		internalPagedInventory: subnetPagedInventory,
	}
	return inventoryImpl.CollectAndPublishInventory(ctx, &logger)
}

// NewManageSubnetInventory returns a ManageInventory implementation for Subnet activity
func NewManageSubnetInventory(config ManageInventoryConfig) ManageSubnetInventory {
	return ManageSubnetInventory{
		config: config,
	}
}

func subnetFindIDs(ctx context.Context, nicoClient *cClient.NICoCoreClient) ([]*cwssaws.NetworkSegmentId, error) {
	idList, err := nicoClient.NICo().FindNetworkSegmentIds(ctx, &cwssaws.NetworkSegmentSearchFilter{})
	if err != nil {
		return nil, err
	}
	return idList.GetNetworkSegmentsIds(), nil
}

func subnetFindByIDs(ctx context.Context, nicoClient *cClient.NICoCoreClient, ids []*cwssaws.NetworkSegmentId) ([]*cwssaws.NetworkSegment, error) {
	list, err := nicoClient.NICo().FindNetworkSegmentsByIds(ctx, &cwssaws.NetworkSegmentsByIdsRequest{
		NetworkSegmentsIds: ids,
	})
	if err != nil {
		return nil, err
	}
	return list.GetNetworkSegments(), nil
}

func subnetPagedInventory(allItemIDs []*cwssaws.NetworkSegmentId, pagedItems []*cwssaws.NetworkSegment, input *pagedInventoryInput) *cwssaws.SubnetInventory {
	itemIDs := []string{}
	for _, id := range allItemIDs {
		itemIDs = append(itemIDs, id.GetValue())
	}

	// Create an inventory page
	inventory := &cwssaws.SubnetInventory{
		Segments: pagedItems,
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
