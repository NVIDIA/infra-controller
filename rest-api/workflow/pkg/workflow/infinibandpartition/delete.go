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

package infinibandpartition

import (
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"

	ibpActivity "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
)

// DeleteInfiniBandPartitionByID is a helper Temporal workflow to delete an existing InfiniBand Partition by ID
// This workflow is useful for invoking from Temporal CLI because it does not require us to create a proto request object
func DeleteInfiniBandPartitionByID(ctx workflow.Context, ibpID uuid.UUID) error {
	logger := log.With().Str("Workflow", "InfiniBandPartition").Str("Action", "Delete").Str("InfiniBand Partition ID", ibpID.String()).Logger()

	logger.Info().Msg("starting workflow")

	// RetryPolicy specifies how to automatically handle retries if an Activity fails.
	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    2 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    2 * time.Minute,
		MaximumAttempts:    10,
	}

	options := workflow.ActivityOptions{
		// Timeout options specify when to automatically timeout Activity functions.
		StartToCloseTimeout: 3 * time.Minute,
		// Optionally provide a customized RetryPolicy.
		RetryPolicy: retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var ibPartitionManager ibpActivity.ManageInfiniBandPartition

	request := &cwssaws.IBPartitionDeletionRequest{
		Id: &cwssaws.IBPartitionId{Value: ibpID.String()},
	}

	err := workflow.ExecuteActivity(ctx, ibPartitionManager.DeleteInfiniBandPartitionOnSite, request).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "DeleteInfiniBandPartitionOnSite").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("completing workflow")

	return nil
}
