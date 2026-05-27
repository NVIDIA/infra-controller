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
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/client"
	tClient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"google.golang.org/protobuf/types/known/timestamppb"

	cClient "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"

	swe "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/error"
)

// ManageMachine is an activity wrapper for Machine management tasks that allows injecting DB access
type ManageMachine struct {
	nicoCoreAtomicClient *cClient.NICoCoreAtomicClient
}

// SetMachineMaintenanceOnSite is an activity to set Machine maintenance mode using Site Controller API
func (mm *ManageMachine) SetMachineMaintenanceOnSite(ctx context.Context, request *cwssaws.MaintenanceRequest) error {
	logger := log.With().Str("Activity", "SetMachineMaintenanceActivity").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty Machine maintenance request")
	} else if request.HostId == nil || request.HostId.Id == "" {
		err = errors.New("received Machine maintenance request without Machine ID")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint to set SetMaintenance request
	nicoClient := mm.nicoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return cClient.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	_, err = rpcClient.SetMaintenance(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to set Maintenance mode for Machine using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return err
}

// UpdateMachineMetadataOnSite is an activity to update Machine metadata using Site Controller API
func (mm *ManageMachine) UpdateMachineMetadataOnSite(ctx context.Context, request *cwssaws.MachineMetadataUpdateRequest) error {
	logger := log.With().Str("Activity", "UpdateMachineMetadataOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty Machine metadata update request")
	} else if request.MachineId == nil || request.MachineId.Id == "" {
		err = errors.New("received Machine metadata update request without Machine ID")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint to update Machine metadata
	nicoClient := mm.nicoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return cClient.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	_, err = rpcClient.UpdateMachineMetadata(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to update Machine metadata using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return err
}

// GetDpuMachinesByIDs is an activity to retrieve DPU Machines by IDs with network configuration
func (mm *ManageMachine) GetDpuMachinesByIDs(ctx context.Context, dpuMachineIDs []string) ([]*cwssaws.DpuMachine, error) {
	logger := log.With().Str("Activity", "GetDpuMachinesByIDs").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if len(dpuMachineIDs) == 0 {
		err = errors.New("received GetDpuMachinesByIDs request without DPU Machine IDs")
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint to get DPU Machines by IDs
	nicoClient := mm.nicoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return nil, cClient.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	// Convert string IDs to MachineId objects
	machineIDs := make([]*cwssaws.MachineId, 0, len(dpuMachineIDs))
	for _, id := range dpuMachineIDs {
		machineIDs = append(machineIDs, &cwssaws.MachineId{Id: id})
	}

	request := &cwssaws.MachinesByIdsRequest{
		MachineIds: machineIDs,
	}

	machineList, err := rpcClient.FindMachinesByIds(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to retrieve DPU Machines by IDs using Site Controller API")
		return nil, swe.WrapErr(err)
	}

	// For each DPU machine, fetch the network configuration
	dpuMachines := make([]*cwssaws.DpuMachine, 0, len(machineList.Machines))
	for _, machine := range machineList.Machines {
		if machine.MachineType == cwssaws.MachineType_DPU {
			networkConfigReq := &cwssaws.ManagedHostNetworkConfigRequest{
				DpuMachineId: machine.Id,
			}
			networkConfig, nerr := rpcClient.GetManagedHostNetworkConfig(ctx, networkConfigReq)
			if nerr != nil {
				logger.Warn().Err(nerr).Str("DPU Machine ID", machine.Id.Id).Msg("Failed to retrieve network config for DPU machine, continuing without it")
				// Don't fail the entire request if network config is unavailable
			}

			logger.Debug().Str("DPU Machine ID", machine.Id.Id).Msg("Retrieved network config for DPU machine")
			dpuMachines = append(dpuMachines, &cwssaws.DpuMachine{
				Machine:          machine,
				DpuNetworkConfig: networkConfig,
			})
		}
	}

	logger.Info().Int("DPU Machine Count", len(dpuMachines)).Msg("Completed activity")

	return dpuMachines, nil
}

// NewManageMachine returns a new ManageMachine activity
func NewManageMachine(nicoCoreAtomicClient *cClient.NICoCoreAtomicClient) ManageMachine {
	return ManageMachine{
		nicoCoreAtomicClient: nicoCoreAtomicClient,
	}
}

// ManageMachineInventory is an activity wrapper for Machine inventory collection and publishing
type ManageMachineInventory struct {
	siteID                uuid.UUID
	nicoCoreAtomicClient  *cClient.NICoCoreAtomicClient
	temporalPublishClient tClient.Client
	temporalPublishQueue  string
	sitePageSize          int
	cloudPageSize         int
}

// CollectAndPublishMachineInventory is an activity to collect Machine inventory and publish to Temporal queue
func (mmi *ManageMachineInventory) CollectAndPublishMachineInventory(ctx context.Context) error {
	logger := log.With().Str("Activity", "CollectAndPublishMachineInventory").Logger()

	logger.Info().Msg("Starting activity")

	// Define workflow options
	workflowOptions := tClient.StartWorkflowOptions{
		ID:        "update-machine-inventory-" + mmi.siteID.String(),
		TaskQueue: mmi.temporalPublishQueue,
	}

	// Call Site Controller gRPC endpoint to get available Machine IDs
	nicoClient := mmi.nicoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return cClient.ErrClientNotConnected
	}
	rpcClient := nicoClient.NICo()

	machineIDList, err := rpcClient.FindMachineIds(ctx, &cwssaws.MachineSearchConfig{})
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to retreive available Machine IDs using Site Controller API")

		// Error encountered before we've published anything, report inventory collection error to Cloud
		inventory := &cwssaws.MachineInventory{
			Timestamp: &timestamppb.Timestamp{
				Seconds: time.Now().Unix(),
			},
			InventoryStatus: cwssaws.InventoryStatus_INVENTORY_STATUS_FAILED,
			StatusMsg:       err.Error(),
		}

		_, serr := mmi.temporalPublishClient.ExecuteWorkflow(context.Background(), workflowOptions, "UpdateMachineInventory", mmi.siteID, inventory)
		if serr != nil {
			logger.Error().Err(serr).Msg("Failed to publish Machine inventory error to Cloud")
			return serr
		}
		return err
	}

	// Paginate IDs and collect Machine inventory
	totalSiteCount := len(machineIDList.MachineIds)
	totalSitePages := len(machineIDList.MachineIds) / mmi.sitePageSize
	if totalSiteCount%mmi.sitePageSize > 0 {
		totalSitePages++
	}

	allMachineIDs := []*cwssaws.MachineId{}
	allMachineIDs = append(allMachineIDs, machineIDList.MachineIds...)

	if totalSitePages == 0 {
		inventoryPage := getPagedMachineInventory([]*cwssaws.Machine{}, allMachineIDs, totalSiteCount, 1, mmi.cloudPageSize, cwssaws.InventoryStatus_INVENTORY_STATUS_SUCCESS, "No Machines reported by SIte Controller")

		_, serr := mmi.temporalPublishClient.ExecuteWorkflow(context.Background(), workflowOptions, "UpdateMachineInventory", mmi.siteID, inventoryPage)
		if serr != nil {
			logger.Error().Err(serr).Msg("Failed to publish Machine inventory to Cloud")
			return serr
		}
	}

	// Iterate through all pages and publish Machine inventory
	effectiveCloudPage := 1
	for sitePage := 1; sitePage <= totalSitePages; sitePage++ {
		pagedMachineIDs := getPagedMachineIDs(machineIDList.MachineIds, sitePage, mmi.sitePageSize)

		// Call Site Controller gRPC endpoint to get Machines for the paged IDs
		pagedMachines, serr := rpcClient.FindMachinesByIds(ctx, &cwssaws.MachinesByIdsRequest{
			MachineIds: pagedMachineIDs,
		})
		if serr != nil {
			logger.Warn().Err(serr).Int("Site Page", sitePage).Msg("Failed to retreive Machines using Site Controller API")
			return serr
		}

		totalCloudCount := len(pagedMachines.Machines)
		totalCloudPages := len(pagedMachines.Machines) / mmi.cloudPageSize
		if totalCloudCount%mmi.cloudPageSize > 0 {
			totalCloudPages++
		}

		// Publish machine inventory to Cloud in separate chunks
		for cloudPage := 1; cloudPage <= totalCloudPages; cloudPage++ {
			startIndex := (cloudPage - 1) * mmi.cloudPageSize
			endIndex := startIndex + mmi.cloudPageSize
			if endIndex > totalCloudCount {
				endIndex = totalCloudCount
			}

			pagedWorkflowOptions := client.StartWorkflowOptions{
				ID:        fmt.Sprintf("%v-%v", workflowOptions.ID, effectiveCloudPage),
				TaskQueue: workflowOptions.TaskQueue,
			}

			// Create an inventory page with the subset of Machines
			inventoryPage := getPagedMachineInventory(pagedMachines.Machines[startIndex:endIndex], allMachineIDs, totalSiteCount, effectiveCloudPage, mmi.cloudPageSize, cwssaws.InventoryStatus_INVENTORY_STATUS_SUCCESS, "Successfully retrieved Machines from Site Controller")

			logger.Info().Msgf("Publishing Machine inventory page %d to Cloud", effectiveCloudPage)

			_, serr = mmi.temporalPublishClient.ExecuteWorkflow(context.Background(), pagedWorkflowOptions, "UpdateMachineInventory", mmi.siteID, inventoryPage)
			if serr != nil {
				logger.Error().Err(serr).Int("Cloud Page", effectiveCloudPage).Msg("Failed to publish Machine inventory to Cloud")
				return serr
			}

			effectiveCloudPage++
		}
	}

	return nil
}

// getPagedMachineIDs returns a slice of Machine IDs for a given page
func getPagedMachineIDs(machineIDs []*cwssaws.MachineId, page int, pageSize int) []*cwssaws.MachineId {
	totalCount := len(machineIDs)
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize
	if endIndex > totalCount {
		endIndex = totalCount
	}

	return machineIDs[startIndex:endIndex]
}

// getPagedMachineInventory returns a subset of MachineInventory for a given page
func getPagedMachineInventory(pagedMachines []*cwssaws.Machine, machineIDs []*cwssaws.MachineId, totalCount int, page int, pageSize int, status cwssaws.InventoryStatus, statusMessage string) *cwssaws.MachineInventory {
	totalPages := (totalCount / pageSize)
	if totalCount%pageSize > 0 {
		totalPages++
	}

	pagedMachineInfo := []*cwssaws.MachineInfo{}
	for _, machine := range pagedMachines {
		pagedMachineInfo = append(pagedMachineInfo, &cwssaws.MachineInfo{
			Machine: machine,
		})
	}

	itemIDs := []string{}
	for _, machineID := range machineIDs {
		itemIDs = append(itemIDs, machineID.Id)
	}

	// Create an inventory page with the subset of Machines
	inventoryPage := &cwssaws.MachineInventory{
		Machines: pagedMachineInfo,
		Timestamp: &timestamppb.Timestamp{
			Seconds: time.Now().Unix(),
		},
		InventoryStatus: status,
		StatusMsg:       statusMessage,
		InventoryPage: &cwssaws.InventoryPage{
			TotalPages:  int32(totalPages),
			CurrentPage: int32(page),
			PageSize:    int32(pageSize),
			TotalItems:  int32(totalCount),
			ItemIds:     itemIDs,
		},
	}

	return inventoryPage
}

// NewManageMachineInventory returns a new ManageMachineInventory activity
func NewManageMachineInventory(siteID uuid.UUID, nicoCoreAtomicClient *cClient.NICoCoreAtomicClient, temporalPublishClient tClient.Client, temporalPublishQueue string, sitePageSize int, cloudPageSize int) ManageMachineInventory {
	return ManageMachineInventory{
		siteID:                siteID,
		nicoCoreAtomicClient:  nicoCoreAtomicClient,
		temporalPublishClient: temporalPublishClient,
		temporalPublishQueue:  temporalPublishQueue,
		sitePageSize:          sitePageSize,
		cloudPageSize:         cloudPageSize,
	}
}
