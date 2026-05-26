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

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.temporal.io/sdk/temporal"

	cClient "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"

	swe "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/error"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

// ManageTenant is activity to manage a Tenant on Site
type ManageTenant struct {
	NICoCoreAtomicClient *cClient.NICoCoreAtomicClient
}

// CreateTenantOnSite creates a Tenant by calling Site Controller gRPC API
func (mt *ManageTenant) CreateTenantOnSite(ctx context.Context, request *cwssaws.CreateTenantRequest) error {
	logger := log.With().Str("Activity", "CreateTenantOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty Tenant request")
	} else if request.OrganizationId == "" {
		err = errors.New("received Tenant creation request without Organization ID")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mt.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return cClient.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	_, err = rpcClient.CreateTenant(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create Tenant using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return err
}

// UpdateTenantOnSite creates a Tenant by calling Site Controller gRPC API
func (mt *ManageTenant) UpdateTenantOnSite(ctx context.Context, request *cwssaws.UpdateTenantRequest) error {
	logger := log.With().Str("Activity", "UpdateTenantOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty Tenant request")
	} else if request.OrganizationId == "" {
		err = errors.New("received Tenant update request without Organization ID")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mt.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return cClient.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	_, err = rpcClient.UpdateTenant(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to update Tenant using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return err
}

// NewManageTenant returns a new ManageTenant activity
func NewManageTenant(nicoClient *cClient.NICoCoreAtomicClient) ManageTenant {
	return ManageTenant{
		NICoCoreAtomicClient: nicoClient,
	}
}

// ManageTenantInventory is an activity wrapper for VPC inventory collection and publishing
type ManageTenantInventory struct {
	config ManageInventoryConfig
}

// ManageTenantInventory is an activity to collect Tenant inventory and publish to Cloud
func (mti *ManageTenantInventory) DiscoverTenantInventory(ctx context.Context) error {
	logger := log.With().Str("Activity", "DiscoverTenantInventory").Logger()
	logger.Info().Msg("Starting activity")
	inventoryImpl := manageInventoryImpl[string, *cwssaws.Tenant, *cwssaws.TenantInventory]{
		itemType:               "Tenant",
		config:                 mti.config,
		internalFindIDs:        tenantFindIDs,
		internalFindByIDs:      tenantFindByIDs,
		internalPagedInventory: tenantPagedInventory,
	}
	return inventoryImpl.CollectAndPublishInventory(ctx, &logger)
}

// NewManageTenantInventory returns a ManageInventory implementation for VPC activity
func NewManageTenantInventory(config ManageInventoryConfig) ManageTenantInventory {
	return ManageTenantInventory{
		config: config,
	}
}

func tenantFindIDs(ctx context.Context, nicoClient *cClient.NICoCoreClient) ([]string, error) {
	idList, err := nicoClient.NICo().FindTenantOrganizationIds(ctx, &cwssaws.TenantSearchFilter{})
	if err != nil {
		return nil, err
	}
	return idList.GetTenantOrganizationIds(), nil
}

func tenantFindByIDs(ctx context.Context, nicoClient *cClient.NICoCoreClient, ids []string) ([]*cwssaws.Tenant, error) {
	list, err := nicoClient.NICo().FindTenantsByOrganizationIds(ctx, &cwssaws.TenantByOrganizationIdsRequest{
		OrganizationIds: ids,
	})
	if err != nil {
		return nil, err
	}
	return list.GetTenants(), nil
}

func tenantPagedInventory(allItemIDs []string, pagedItems []*cwssaws.Tenant, input *pagedInventoryInput) *cwssaws.TenantInventory {
	// Create an inventory page with the subset of Tenants
	inventory := &cwssaws.TenantInventory{
		Tenants: pagedItems,
		Timestamp: &timestamppb.Timestamp{
			Seconds: time.Now().Unix(),
		},
		InventoryStatus: input.status,
		StatusMsg:       input.statusMessage,
		InventoryPage:   input.buildPage(),
	}
	if inventory.InventoryPage != nil {
		inventory.InventoryPage.ItemIds = allItemIDs
	}
	return inventory
}
