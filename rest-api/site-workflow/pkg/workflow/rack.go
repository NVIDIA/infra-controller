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

	"github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	flowv1 "github.com/NVIDIA/infra-controller-rest/workflow-schema/flow/protobuf/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GetRack is a workflow to get a rack by its UUID from Flow
func GetRack(ctx workflow.Context, request *flowv1.GetRackInfoByIDRequest) (*flowv1.GetRackInfoResponse, error) {
	logger := log.With().Str("Workflow", "Rack").Str("Action", "Get").Logger()
	if request != nil && request.Id != nil {
		logger = log.With().Str("Workflow", "Rack").Str("Action", "Get").Str("RackID", request.Id.Id).Logger()
	}

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

	var rackManager activity.ManageRack
	var response flowv1.GetRackInfoResponse

	err := workflow.ExecuteActivity(ctx, rackManager.GetRack, request).Get(ctx, &response)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "GetRack").Msg("Failed to execute activity from workflow")
		return nil, err
	}

	logger.Info().Msg("Completing workflow")

	return &response, nil
}

// GetRacks is a workflow to get a list of racks from Flow with optional filters
func GetRacks(ctx workflow.Context, request *flowv1.GetListOfRacksRequest) (*flowv1.GetListOfRacksResponse, error) {
	logger := log.With().Str("Workflow", "Rack").Str("Action", "GetAll").Logger()

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

	var rackManager activity.ManageRack
	var response flowv1.GetListOfRacksResponse

	err := workflow.ExecuteActivity(ctx, rackManager.GetRacks, request).Get(ctx, &response)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "GetRacks").Msg("Failed to execute activity from workflow")
		return nil, err
	}

	logger.Info().Int32("Total", response.GetTotal()).Msg("Completing workflow")

	return &response, nil
}

// ValidateRackComponents is a workflow to validate rack components by comparing expected vs actual state via Flow.
// Supports validating a single rack, multiple racks with filters, or all racks in a site.
func ValidateRackComponents(ctx workflow.Context, request *flowv1.ValidateComponentsRequest) (*flowv1.ValidateComponentsResponse, error) {
	logger := log.With().Str("Workflow", "Rack").Str("Action", "ValidateRackComponents").Logger()

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

	var rackManager activity.ManageRack
	var response flowv1.ValidateComponentsResponse

	err := workflow.ExecuteActivity(ctx, rackManager.ValidateRackComponents, request).Get(ctx, &response)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "ValidateRackComponents").Msg("Failed to execute activity from workflow")
		return nil, err
	}

	logger.Info().Int32("TotalDiffs", response.GetTotalDiffs()).Msg("Completing workflow")

	return &response, nil
}

// PowerOnRack is a workflow to power on a rack or its specified components via Flow
func PowerOnRack(ctx workflow.Context, request *flowv1.PowerOnRackRequest) (*flowv1.SubmitTaskResponse, error) {
	logger := log.With().Str("Workflow", "Rack").Str("Action", "PowerOnRack").Logger()

	logger.Info().Msg("Starting workflow")

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var rackManager activity.ManageRack
	var response flowv1.SubmitTaskResponse

	err := workflow.ExecuteActivity(ctx, rackManager.PowerOnRack, request).Get(ctx, &response)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "PowerOnRack").Msg("Failed to execute activity from workflow")
		return nil, err
	}

	logger.Info().Int("TaskCount", len(response.GetTaskIds())).Msg("Completing workflow")

	return &response, nil
}

// PowerOffRack is a workflow to power off a rack or its specified components via Flow
func PowerOffRack(ctx workflow.Context, request *flowv1.PowerOffRackRequest) (*flowv1.SubmitTaskResponse, error) {
	logger := log.With().Str("Workflow", "Rack").Str("Action", "PowerOffRack").Logger()

	logger.Info().Msg("Starting workflow")

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var rackManager activity.ManageRack
	var response flowv1.SubmitTaskResponse

	err := workflow.ExecuteActivity(ctx, rackManager.PowerOffRack, request).Get(ctx, &response)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "PowerOffRack").Msg("Failed to execute activity from workflow")
		return nil, err
	}

	logger.Info().Int("TaskCount", len(response.GetTaskIds())).Msg("Completing workflow")

	return &response, nil
}

// PowerResetRack is a workflow to reset (power cycle) a rack or its specified components via Flow
func PowerResetRack(ctx workflow.Context, request *flowv1.PowerResetRackRequest) (*flowv1.SubmitTaskResponse, error) {
	logger := log.With().Str("Workflow", "Rack").Str("Action", "PowerResetRack").Logger()

	logger.Info().Msg("Starting workflow")

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var rackManager activity.ManageRack
	var response flowv1.SubmitTaskResponse

	err := workflow.ExecuteActivity(ctx, rackManager.PowerResetRack, request).Get(ctx, &response)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "PowerResetRack").Msg("Failed to execute activity from workflow")
		return nil, err
	}

	logger.Info().Int("TaskCount", len(response.GetTaskIds())).Msg("Completing workflow")

	return &response, nil
}

// BringUpRack is a workflow to bring up a rack or its specified components via Flow
func BringUpRack(ctx workflow.Context, request *flowv1.BringUpRackRequest) (*flowv1.SubmitTaskResponse, error) {
	logger := log.With().Str("Workflow", "Rack").Str("Action", "BringUpRack").Logger()

	logger.Info().Msg("Starting workflow")

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy:         retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var rackManager activity.ManageRack
	var response flowv1.SubmitTaskResponse

	err := workflow.ExecuteActivity(ctx, rackManager.BringUpRack, request).Get(ctx, &response)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "BringUpRack").Msg("Failed to execute activity from workflow")
		return nil, err
	}

	logger.Info().Int("TaskCount", len(response.GetTaskIds())).Msg("Completing workflow")

	return &response, nil
}

// GetRackTask is a workflow to get a rack task by its UUID
func GetRackTask(ctx workflow.Context, request *flowv1.GetTasksByIDsRequest) (*flowv1.GetTasksByIDsResponse, error) {
	logger := log.With().Str("Workflow", "Rack").Str("Action", "GetRackTask").Logger()

	logger.Info().Msg("Starting workflow")

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var rackManager activity.ManageRack
	var response flowv1.GetTasksByIDsResponse

	err := workflow.ExecuteActivity(ctx, rackManager.GetTaskByID, request).Get(ctx, &response)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "GetTaskByID").Msg("Failed to execute activity from workflow")
		return nil, err
	}

	logger.Info().Int("TaskCount", len(response.GetTasks())).Msg("Completing workflow")

	return &response, nil
}

// CancelRackTask is a workflow to cancel a rack task by its UUID via Flow.
//
// Cancel is best-effort: Flow marks the task Terminated and terminates the
// underlying Temporal workflow if one was scheduled. The returned task carries
// the post-cancel status reported by Flow.
func CancelRackTask(ctx workflow.Context, request *flowv1.CancelTaskRequest) (*flowv1.CancelTaskResponse, error) {
	logger := log.With().Str("Workflow", "Rack").Str("Action", "CancelRackTask").Logger()

	logger.Info().Msg("Starting workflow")

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var rackManager activity.ManageRack
	var response flowv1.CancelTaskResponse

	err := workflow.ExecuteActivity(ctx, rackManager.CancelTask, request).Get(ctx, &response)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "CancelTask").Msg("Failed to execute activity from workflow")
		return nil, err
	}

	logger.Info().Msg("Completing workflow")

	return &response, nil
}

// UpgradeFirmware is a workflow to upgrade firmware on racks or components via Flow
func UpgradeFirmware(ctx workflow.Context, request *flowv1.UpgradeFirmwareRequest) (*flowv1.SubmitTaskResponse, error) {
	logger := log.With().Str("Workflow", "Rack").Str("Action", "UpgradeFirmware").Logger()

	logger.Info().Msg("Starting workflow")

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy:         retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var rackManager activity.ManageRack
	var response flowv1.SubmitTaskResponse

	err := workflow.ExecuteActivity(ctx, rackManager.UpgradeFirmware, request).Get(ctx, &response)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "UpgradeFirmware").Msg("Failed to execute activity from workflow")
		return nil, err
	}

	logger.Info().Int("TaskCount", len(response.GetTaskIds())).Msg("Completing workflow")

	return &response, nil
}
