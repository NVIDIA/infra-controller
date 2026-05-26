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

// CreateInstanceType is a workflow to create new InstanceTypes using the CreateInstanceTypeOnSite activity
// to speak to nico directly.
func CreateInstanceType(ctx workflow.Context, request *cwssaws.CreateInstanceTypeRequest) error {
	logger := log.With().Str("Workflow", "InstanceType").Str("Action", "Create").Str("InstanceType ID", request.GetId()).Logger()

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

	var instanceTypeManager activity.ManageInstanceType

	err := workflow.ExecuteActivity(ctx, instanceTypeManager.CreateInstanceTypeOnSite, request).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "CreateInstanceTypeOnSite").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("Completing workflow")

	return nil
}

// UpdateInstanceType is a workflow to update InstanceType data using then UpdateInstanceTypeOnSite activity
func UpdateInstanceType(ctx workflow.Context, updateRequest *cwssaws.UpdateInstanceTypeRequest) error {
	logger := log.With().Str("Workflow", "InstanceType").Str("Action", "Update").Str("InstanceType ID", updateRequest.GetId()).Logger()

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

	var instanceTypeManager activity.ManageInstanceType

	err := workflow.ExecuteActivity(ctx, instanceTypeManager.UpdateInstanceTypeOnSite, updateRequest).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "UpdateInstanceTypeOnSite").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("Completing workflow")

	return nil
}

// DeleteInstanceType is a workflow to delete new InstanceTypes using the DeleteInstanceTypeOnSite activity
func DeleteInstanceType(ctx workflow.Context, request *cwssaws.DeleteInstanceTypeRequest) error {

	logger := log.With().Str("Workflow", "InstanceType").Str("Action", "Delete").Str("Request", request.String()).Logger()

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

	var instanceTypeManager activity.ManageInstanceType

	err := workflow.ExecuteActivity(ctx, instanceTypeManager.DeleteInstanceTypeOnSite, request).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "DeleteInstanceTypeOnSite").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("Completing workflow")

	return nil
}

// AssociateMachinesWithInstanceType is a workflow to associate machines with an InstanceType
func AssociateMachinesWithInstanceType(ctx workflow.Context, request *cwssaws.AssociateMachinesWithInstanceTypeRequest) error {

	logger := log.With().Str("Workflow", "InstanceType").Str("Action", "AssociateMachinesWithInstanceType").Str("Request", request.String()).Logger()

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

	var instanceTypeManager activity.ManageInstanceType

	err := workflow.ExecuteActivity(ctx, instanceTypeManager.AssociateMachinesWithInstanceTypeOnSite, request).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "AssociateMachinesWithInstanceTypeOnSite").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("Completing workflow")

	return nil
}

// RemoveMachineInstanceTypeAssociation is a workflow to remove the relationship between a Machine and InstanceType
func RemoveMachineInstanceTypeAssociation(ctx workflow.Context, request *cwssaws.RemoveMachineInstanceTypeAssociationRequest) error {

	logger := log.With().Str("Workflow", "InstanceType").Str("Action", "RemoveMachineInstanceTypeAssociation").Str("Request", request.String()).Logger()

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

	var instanceTypeManager activity.ManageInstanceType

	err := workflow.ExecuteActivity(ctx, instanceTypeManager.RemoveMachineInstanceTypeAssociationOnSite, request).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "RemoveMachineInstanceTypeAssociation").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("Completing workflow")

	return nil
}

func DiscoverInstanceTypeInventory(ctx workflow.Context) error {
	logger := log.With().Str("Workflow", "DiscoverInstanceTypeInventory").Logger()

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

	// Invoke DiscoverInstanceTypeInventory activity
	var instanceTypeInventoryManager activity.ManageInstanceTypeInventory

	err := workflow.ExecuteActivity(ctx, instanceTypeInventoryManager.DiscoverInstanceTypeInventory).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "DiscoverInstanceTypeInventory").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("Completing workflow")

	return nil
}
