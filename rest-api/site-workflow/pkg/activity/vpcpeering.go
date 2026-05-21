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

// ManageVpcPeering is an activity wrapper for VpcPeering management
type ManageVpcPeering struct {
	NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
}

// NewManageVpcPeering returns a new ManageVpcPeering client
func NewManageVpcPeering(nicoClient *cClient.NICoCoreAtomicClient) ManageVpcPeering {
	return ManageVpcPeering{
		NICoCoreAtomicClient: nicoClient,
	}
}

// Function to create VpcPeering with NICo
func (mvp *ManageVpcPeering) CreateVpcPeeringOnSite(ctx context.Context, request *cwssaws.VpcPeeringCreationRequest) error {
	logger := log.With().Str("Activity", "CreateVpcPeeringOnSite").Logger()

	logger.Info().Msg("Starting activity'")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty create VpcPeering request")
	} else if request.VpcId == nil || request.VpcId.Value == "" {
		err = errors.New("received create VpcPeering request missing VpcId")
	} else if request.PeerVpcId == nil || request.PeerVpcId.Value == "" {
		err = errors.New("received create VpcPeering request missing PeerVpcId")
	} else if request.Id == nil || request.Id.Value == "" {
		// Don't let a request come in without a cloud-provided ID
		// or carbide will generate one and cloud won't know the relationship.
		err = errors.New("received create VpcPeering request missing VPC peering ID")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller API
	nicoClient := mvp.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return cClient.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	_, err = rpcClient.CreateVpcPeering(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create VpcPeering using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

// Function to delete VpcPeering on NICo
func (mvp *ManageVpcPeering) DeleteVpcPeeringOnSite(ctx context.Context, request *cwssaws.VpcPeeringDeletionRequest) error {
	logger := log.With().Str("Activity", "DeleteVpcPeeringOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty delete VpcPeering request")
	} else if request.Id == nil || request.Id.Value == "" {
		err = errors.New("received delete VpcPeering request missing VPC peering ID")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller API
	nicoClient := mvp.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return cClient.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	_, err = rpcClient.DeleteVpcPeering(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to delete VpcPeering using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

// ManageVpcPeeringInventory is an activity wrapper for VpcPeering inventory collection and publishing
type ManageVpcPeeringInventory struct {
	config ManageInventoryConfig
}

func NewManageVpcPeeringInventory(config ManageInventoryConfig) ManageVpcPeeringInventory {
	return ManageVpcPeeringInventory{
		config: config,
	}
}

// DiscoverVpcPeeringInventory is an activity to collect VpcPeering inventory and publish to Temporal queue
func (mvi *ManageVpcPeeringInventory) DiscoverVpcPeeringInventory(ctx context.Context) error {
	logger := log.With().Str("Activity", "DiscoverVpcPeeringInventory").Logger()
	logger.Info().Msg("Starting activity")

	inventoryImpl := manageInventoryImpl[*cwssaws.VpcPeeringId, *cwssaws.VpcPeering, *cwssaws.VPCPeeringInventory]{
		itemType:               "VpcPeering",
		config:                 mvi.config,
		internalFindIDs:        VpcPeeringFindIDs,
		internalFindByIDs:      VpcPeeringFindByIDs,
		internalPagedInventory: VpcPeeringPagedInventory,
	}
	return inventoryImpl.CollectAndPublishInventory(ctx, &logger)
}

func VpcPeeringFindIDs(ctx context.Context, nicoClient *cClient.NICoCoreClient) ([]*cwssaws.VpcPeeringId, error) {
	resp, err := nicoClient.NICo().FindVpcPeeringIds(ctx, &cwssaws.VpcPeeringSearchFilter{})
	if err != nil {
		return nil, err
	}
	return resp.VpcPeeringIds, nil
}

func VpcPeeringFindByIDs(ctx context.Context, nicoClient *cClient.NICoCoreClient, ids []*cwssaws.VpcPeeringId) ([]*cwssaws.VpcPeering, error) {
	req := &cwssaws.VpcPeeringsByIdsRequest{
		VpcPeeringIds: ids,
	}
	resp, err := nicoClient.NICo().FindVpcPeeringsByIds(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetVpcPeerings(), nil
}

func VpcPeeringPagedInventory(allItemIDs []*cwssaws.VpcPeeringId, pagedItems []*cwssaws.VpcPeering, input *pagedInventoryInput) *cwssaws.VPCPeeringInventory {
	itemIDs := []string{}
	for _, id := range allItemIDs {
		itemIDs = append(itemIDs, id.GetValue())
	}

	// Create an inventory page with the subset of VpcPeerings
	inventory := &cwssaws.VPCPeeringInventory{
		VpcPeerings: pagedItems,
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
