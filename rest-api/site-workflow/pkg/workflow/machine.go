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

package workflow

import (
	"time"

	"github.com/rs/zerolog/log"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"

	"github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
)

// SetMachineMaintenance is a workflow to set Machine maintenance mode using SetMaintenanceOnSite activity
func SetMachineMaintenance(ctx workflow.Context, request *cwssaws.MaintenanceRequest) error {
	logger := log.With().Str("Workflow", "SetMachineMaintenance").Logger()

	logger.Info().Msg("Starting workflow")

	// RetryPolicy specifies how to automatically handle retries if an Activity fails.
	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		// Timeout options specify when to automatically timeout Activity functions.
		StartToCloseTimeout: 2 * time.Minute,
		// Optionally provide a customized RetryPolicy.
		RetryPolicy: retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	// Invoke SetMachineMaintenanceOnSite activity
	var machineManager activity.ManageMachine

	err := workflow.ExecuteActivity(ctx, machineManager.SetMachineMaintenanceOnSite, request).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "SetMachineMaintenanceOnSite").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("Completing workflow")

	return nil
}

func UpdateMachineMetadata(ctx workflow.Context, request *cwssaws.MachineMetadataUpdateRequest) error {
	logger := log.With().Str("Workflow", "UpdateMachineMetadata").Logger()

	logger.Info().Msg("Starting workflow")

	// RetryPolicy specifies how to automatically handle retries if an Activity fails.
	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		// Timeout options specify when to automatically timeout Activity functions.
		StartToCloseTimeout: 2 * time.Minute,
		// Optionally provide a customized RetryPolicy.
		RetryPolicy: retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	// Invoke UpdateMachineMetadataOnSite activity
	var machineManager activity.ManageMachine

	err := workflow.ExecuteActivity(ctx, machineManager.UpdateMachineMetadataOnSite, request).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "UpdateMachineMetadataOnSite").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("Completing workflow")

	return nil
}

func CollectAndPublishMachineInventory(ctx workflow.Context) error {
	logger := log.With().Str("Workflow", "CollectAndPublishMachineInventory").Logger()

	logger.Info().Msg("Starting workflow")

	// RetryPolicy specifies how to automatically handle retries if an Activity fails.
	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    2 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		// This is executed every 3 minutes, so we don't want too many retry attempts
		MaximumAttempts: 2,
	}
	options := workflow.ActivityOptions{
		// Timeout options specify when to automatically timeout Activity functions.
		StartToCloseTimeout: 2 * time.Minute,
		// Optionally provide a customized RetryPolicy.
		RetryPolicy: retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	// Invoke CollectAndPublishMachineInventory activity
	var machineInventoryManager activity.ManageMachineInventory

	err := workflow.ExecuteActivity(ctx, machineInventoryManager.CollectAndPublishMachineInventory).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "CollectAndPublishMachineInventory").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("Completing workflow")

	return nil
}

// GetDpuMachines is a workflow to retrieve DPU Machines by IDs with network configuration
func GetDpuMachines(ctx workflow.Context, dpuMachineIDs []string) ([]*cwssaws.DpuMachine, error) {
	logger := log.With().Str("Workflow", "GetDpuMachines").Logger()

	logger.Info().Msg("Starting workflow")

	// RetryPolicy specifies how to automatically handle retries if an Activity fails.
	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		// Timeout options specify when to automatically timeout Activity functions.
		StartToCloseTimeout: 2 * time.Minute,
		// Optionally provide a customized RetryPolicy.
		RetryPolicy: retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	// Invoke GetDpuMachinesByIDs activity
	var machineManager activity.ManageMachine

	var result []*cwssaws.DpuMachine
	err := workflow.ExecuteActivity(ctx, machineManager.GetDpuMachinesByIDs, dpuMachineIDs).Get(ctx, &result)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "GetDpuMachinesByIDs").Msg("Failed to execute activity from workflow")
		return nil, err
	}

	logger.Info().Msg("Completing workflow")

	return result, nil
}
