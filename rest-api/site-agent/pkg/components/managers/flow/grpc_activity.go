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

package flow

import (
	"context"

	"github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

// CreateGRPCClientActivity - Create GRPC client Activity
func (Flow *API) CreateGRPCClientActivity(ctx context.Context, ResourceID string) (client *client.FlowClient, err error) {
	ManagerAccess.Data.EB.Log.Info().Interface("Request", ResourceID).Msg("Flow: Starting the gRPC connection Activity")

	// Use temporal logger for temporal logs
	logger := activity.GetLogger(ctx)
	withLogger := log.With(logger, "Activity", "CreateGRPCClientActivity", "ResourceReq", ResourceID)
	withLogger.Info("Flow: Starting the gRPC connection Activity")

	// Create the client
	ManagerAccess.Data.EB.Log.Info().Interface("Request", ResourceID).Msg("Flow: Creating gRPC client")

	err = Flow.CreateGRPCClient()
	if err != nil {
		return nil, err
	}
	return Flow.GetGRPCClient(), nil
}

// RegisterGRPC - Register GRPC
func (Flow *API) RegisterGRPC() {
	// Register activity
	activityRegisterOptions := activity.RegisterOptions{
		Name: "CreateRlaGrpcClientActivity",
	}

	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivityWithOptions(
		ManagerAccess.API.Flow.CreateGRPCClientActivity, activityRegisterOptions,
	)
	ManagerAccess.Data.EB.Log.Info().Msg("Flow: successfully registered GRPC client activity")
}
