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

	"github.com/rs/zerolog/log"

	swe "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/error"
	cClient "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"
	flowv1 "github.com/NVIDIA/infra-controller-rest/workflow-schema/flow/protobuf/v1"
	"go.temporal.io/sdk/temporal"
)

// ManageTray is an activity wrapper for Tray management via Flow
type ManageTray struct {
	FlowAtomicClient *cClient.FlowAtomicClient
}

// GetTray retrieves a tray by its UUID from Flow
func (mt *ManageTray) GetTray(ctx context.Context, request *flowv1.GetComponentInfoByIDRequest) (*flowv1.GetComponentInfoResponse, error) {
	logger := log.With().Str("Activity", "GetTray").Logger()
	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	switch {
	case request == nil:
		err = errors.New("received empty get tray request")
	case request.Id == nil || request.Id.Id == "":
		err = errors.New("received get tray request missing tray ID")
	}

	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Flow gRPC endpoint
	flow, err := mt.FlowAtomicClient.GetFlowClient()
	if err != nil {
		return nil, err
	}

	response, err := flow.GetComponentInfoByID(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to get tray by ID using Flow API")
		return nil, swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return response, nil
}

// GetTrays retrieves a list of trays from Flow with optional filters.
func (mt *ManageTray) GetTrays(ctx context.Context, request *flowv1.GetComponentsRequest) (*flowv1.GetComponentsResponse, error) {
	logger := log.With().Str("Activity", "GetTrays").Logger()
	logger.Info().Msg("Starting activity")

	// Request can be nil or empty for getting all trays
	if request == nil {
		request = &flowv1.GetComponentsRequest{}
	}

	// Call Flow gRPC endpoint
	flow, err := mt.FlowAtomicClient.GetFlowClient()
	if err != nil {
		return nil, err
	}

	response, err := flow.GetComponents(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to get list of trays using Flow API")
		return nil, swe.WrapErr(err)
	}

	logger.Info().Int32("Total", response.GetTotal()).Msg("Completed activity")

	return response, nil
}

// NewManageTray returns a new ManageTray client
func NewManageTray(flowClient *cClient.FlowAtomicClient) ManageTray {
	return ManageTray{
		FlowAtomicClient: flowClient,
	}
}
