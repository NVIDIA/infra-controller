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

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"

	cClient "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"

	swe "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/error"
	"github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/util"
)

// ManageNetworkSecurityGroup is an activity wrapper for NetworkSecurityGroup management tasks that allows injecting DB access
type ManageNetworkSecurityGroup struct {
	NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
}

// Function to Create NICo NetworkSecurityGroup with the Site Controller
func (mm *ManageNetworkSecurityGroup) CreateNetworkSecurityGroupOnSite(ctx context.Context, request *cwssaws.CreateNetworkSecurityGroupRequest) error {
	logger := log.With().Str("Activity", "CreateNetworkSecurityGroupOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	switch {
	case request == nil:
		err = errors.New("received empty create NetworkSecurityGroup request")
	case request.Id == nil || *request.Id == "":
		err = errors.New("received create NetworkSecurityGroup request without ID")
	case request.TenantOrganizationId == "":
		err = errors.New("received create NetworkSecurityGroup request with empty Tenant ID")

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

	_, err = rpcClient.CreateNetworkSecurityGroup(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create NetworkSecurityGroup using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

// Function Update NICo NetworkSecurityGroup with the Site Controller
func (mm *ManageNetworkSecurityGroup) UpdateNetworkSecurityGroupOnSite(ctx context.Context, request *cwssaws.UpdateNetworkSecurityGroupRequest) error {
	logger := log.With().Str("Activity", "UpdateNetworkSecurityGroupOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	switch {
	case request == nil:
		err = errors.New("received empty NetworkSecurityGroup config update request")
	case request.Id == "":
		err = errors.New("received NetworkSecurityGroup config update request without NetworkSecurityGroup ID")
	case request.TenantOrganizationId == "":
		err = errors.New("received NetworkSecurityGroup config update request without Tenant ID")
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

	_, err = rpcClient.UpdateNetworkSecurityGroup(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to update config for NetworkSecurityGroup using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

// Function to Delete NICo NetworkSecurityGroup with the Site Controller
func (mm *ManageNetworkSecurityGroup) DeleteNetworkSecurityGroupOnSite(ctx context.Context, request *cwssaws.DeleteNetworkSecurityGroupRequest) error {
	logger := log.With().Str("Activity", "DeleteNetworkSecurityGroupOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	switch {
	case request == nil:
		err = errors.New("received empty delete NetworkSecurityGroup request")
	case request.Id == "":
		err = errors.New("received delete NetworkSecurityGroup request without NetworkSecurityGroup ID")
	case request.TenantOrganizationId == "":
		err = errors.New("received delete NetworkSecurityGroup request without empty Tenant ID")
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

	_, err = rpcClient.DeleteNetworkSecurityGroup(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to delete NetworkSecurityGroup using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

// NewManageNetworkSecurityGroup returns a new ManageNetworkSecurityGroup activity
func NewManageNetworkSecurityGroup(nicoClient *cClient.NICoCoreAtomicClient) ManageNetworkSecurityGroup {
	return ManageNetworkSecurityGroup{
		NICoCoreAtomicClient: nicoClient,
	}
}

// ManageNetworkSecurityGroupInventory is an activity wrapper for NetworkSecurityGroup inventory collection and publishing
type ManageNetworkSecurityGroupInventory struct {
	config ManageInventoryConfig
}

// DiscoverNetworkSecurityGroupInventory is an activity to collect NetworkSecurityGroup inventory and publish to Temporal queue
func (mmi *ManageNetworkSecurityGroupInventory) DiscoverNetworkSecurityGroupInventory(ctx context.Context) error {
	logger := log.With().Str("Activity", "DiscoverNetworkSecurityGroupInventory").Logger()
	logger.Info().Msg("Starting activity")
	inventoryImpl := manageInventoryImpl[*cwssaws.UUID, *cwssaws.NetworkSecurityGroup, *cwssaws.NetworkSecurityGroupInventory]{
		itemType:               "NetworkSecurityGroup",
		config:                 mmi.config,
		internalFindIDs:        networkSecurityGroupFindIDs,
		internalFindByIDs:      networkSecurityGroupFindByIDs,
		internalPagedInventory: networkSecurityGroupPagedInventory,
	}
	return inventoryImpl.CollectAndPublishInventory(ctx, &logger)
}

// NewManageNetworkSecurityGroupInventory returns a ManageInventory implementation for NetworkSecurityGroup activity
func NewManageNetworkSecurityGroupInventory(config ManageInventoryConfig) ManageNetworkSecurityGroupInventory {
	return ManageNetworkSecurityGroupInventory{
		config: config,
	}
}

func networkSecurityGroupFindIDs(ctx context.Context, nicoClient *cClient.NICoCoreClient) ([]*cwssaws.UUID, error) {
	// Call Site Controller gRPC endpoint
	rpcClient := nicoClient.NICo()
	networkSecurityGroupIdList, err := rpcClient.FindNetworkSecurityGroupIds(ctx, &cwssaws.FindNetworkSecurityGroupIdsRequest{})
	if err != nil {
		return nil, err
	}
	return util.StringsToProtobufUUIDList(networkSecurityGroupIdList.GetNetworkSecurityGroupIds()), nil
}

func networkSecurityGroupFindByIDs(ctx context.Context, nicoClient *cClient.NICoCoreClient, ids []*cwssaws.UUID) ([]*cwssaws.NetworkSecurityGroup, error) {
	nsgIDs := make([]string, len(ids))

	for i, id := range ids {
		nsgIDs[i] = id.GetValue()
	}

	// Call Site Controller gRPC endpoint
	rpcClient := nicoClient.NICo()
	networkSecurityGroupList, err := rpcClient.FindNetworkSecurityGroupsByIds(ctx, &cwssaws.FindNetworkSecurityGroupsByIdsRequest{
		NetworkSecurityGroupIds: nsgIDs,
	})
	if err != nil {
		return nil, err
	}
	return networkSecurityGroupList.GetNetworkSecurityGroups(), nil
}

func networkSecurityGroupPagedInventory(allItemIDs []*cwssaws.UUID, pagedItems []*cwssaws.NetworkSecurityGroup, input *pagedInventoryInput) *cwssaws.NetworkSecurityGroupInventory {
	itemIDs := []string{}
	for _, id := range allItemIDs {
		itemIDs = append(itemIDs, id.GetValue())
	}

	// Create an inventory page with the subset of Machines
	networkSecurityGroupInventory := &cwssaws.NetworkSecurityGroupInventory{
		NetworkSecurityGroups: pagedItems,
		Timestamp: &timestamppb.Timestamp{
			Seconds: time.Now().Unix(),
		},
		InventoryStatus: input.status,
		StatusMsg:       input.statusMessage,
		InventoryPage:   input.buildPage(),
	}
	if networkSecurityGroupInventory.InventoryPage != nil {
		networkSecurityGroupInventory.InventoryPage.ItemIds = itemIDs
	}
	return networkSecurityGroupInventory
}
