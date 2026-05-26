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

	swe "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/error"
	"github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"
	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"
)

// ManageMachineValidation is an activity wrapper for Machine Validation management
type ManageMachineValidation struct {
	NICoCoreAtomicClient *client.NICoCoreAtomicClient
}

// NewManageMachineValidation returns a new ManageMachineValidation client
func NewManageMachineValidation(nicoClient *client.NICoCoreAtomicClient) ManageMachineValidation {
	return ManageMachineValidation{
		NICoCoreAtomicClient: nicoClient,
	}
}

func (mmv *ManageMachineValidation) EnableDisableMachineValidationTestOnSite(ctx context.Context, request *cwssaws.MachineValidationTestEnableDisableTestRequest) error {
	logger := log.With().Str("Activity", "EnableDisableMachineValidationTestOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty enable/disable machine validation test request")
	} else if request.TestId == "" {
		err = errors.New("received enable/disable machine validation test request missing TestId")
	} else if request.Version == "" {
		err = errors.New("received enable/disable machine validation test request missing Version")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mmv.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return client.ErrClientNotConnected
	}

	_, err = nicoClient.NICo().MachineValidationTestEnableDisableTest(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to enable/disable machine validation test using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

func (mmv *ManageMachineValidation) PersistValidationResultOnSite(ctx context.Context, request *cwssaws.MachineValidationResultPostRequest) error {
	logger := log.With().Str("Activity", "PersistValidationResultOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty persist validation results request")
	} else if request.Result == nil {
		err = errors.New("received persist validation results request missing Result")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mmv.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return client.ErrClientNotConnected
	}

	_, err = nicoClient.NICo().PersistValidationResult(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to persist validation results using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

func (mmv *ManageMachineValidation) GetMachineValidationResultsFromSite(ctx context.Context, request *cwssaws.MachineValidationGetRequest) (*cwssaws.MachineValidationResultList, error) {
	logger := log.With().Str("Activity", "GetMachineValidationResultsFromSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty get machine validation results request")
	}

	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mmv.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return nil, client.ErrClientNotConnected
	}

	result, err := nicoClient.NICo().GetMachineValidationResults(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to get machine validation results using Site Controller API")
		return nil, swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return result, nil
}

func (mmv *ManageMachineValidation) GetMachineValidationRunsFromSite(ctx context.Context, request *cwssaws.MachineValidationRunListGetRequest) (*cwssaws.MachineValidationRunList, error) {
	logger := log.With().Str("Activity", "GetMachineValidationRunsFromSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty get machine validation runs request")
	} else if request.MachineId == nil {
		err = errors.New("received get machine validation runs request missing MachineId")
	}

	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mmv.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return nil, client.ErrClientNotConnected
	}

	result, err := nicoClient.NICo().GetMachineValidationRuns(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to get machine validation runs using Site Controller API")
		return nil, err
	}

	logger.Info().Msg("Completed activity")

	return result, nil
}

func (mmv *ManageMachineValidation) GetMachineValidationTestsFromSite(ctx context.Context, request *cwssaws.MachineValidationTestsGetRequest) (*cwssaws.MachineValidationTestsGetResponse, error) {
	logger := log.With().Str("Activity", "GetMachineValidationTestsFromSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty get machine validation tests request")
	}

	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mmv.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return nil, client.ErrClientNotConnected
	}

	result, err := nicoClient.NICo().GetMachineValidationTests(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to get machine validation tests using Site Controller API")
		return nil, swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return result, nil
}

func (mmv *ManageMachineValidation) AddMachineValidationTestOnSite(ctx context.Context, request *cwssaws.MachineValidationTestAddRequest) (*cwssaws.MachineValidationTestAddUpdateResponse, error) {
	logger := log.With().Str("Activity", "AddMachineValidationTestOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty add machine validation test request")
	} else if request.Name == "" {
		err = errors.New("received add machine validation test request missing Name")
	} else if request.Command == "" {
		err = errors.New("received add machine validation test request missing Command")
	} else if request.Args == "" {
		err = errors.New("received add machine validation test request missing Args")
	}

	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mmv.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return nil, client.ErrClientNotConnected
	}

	response, err := nicoClient.NICo().AddMachineValidationTest(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to add machine validation test using Site Controller API")
		return nil, swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return response, nil
}

func (mmv *ManageMachineValidation) UpdateMachineValidationTestOnSite(ctx context.Context, request *cwssaws.MachineValidationTestUpdateRequest) error {
	logger := log.With().Str("Activity", "UpdateMachineValidationTestOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty update machine validation test request")
	} else if request.TestId == "" {
		err = errors.New("received update machine validation test request missing TestId")
	} else if request.Version == "" {
		err = errors.New("received update machine validation test request missing Version")
	} else if request.Payload == nil {
		err = errors.New("received update machine validation test request missing Payload")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mmv.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return client.ErrClientNotConnected
	}

	_, err = nicoClient.NICo().UpdateMachineValidationTest(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to update machine validation test using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

func (mmv *ManageMachineValidation) GetMachineValidationExternalConfigsFromSite(ctx context.Context, request *cwssaws.GetMachineValidationExternalConfigsRequest) (*cwssaws.GetMachineValidationExternalConfigsResponse, error) {
	logger := log.With().Str("Activity", "GetMachineValidationExternalConfigsFromSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty get machine validation external configs request")
	}

	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mmv.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return nil, client.ErrClientNotConnected
	}

	result, err := nicoClient.NICo().GetMachineValidationExternalConfigs(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to get machine validation external configs using Site Controller API")
		return nil, swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return result, nil
}

func (mmv *ManageMachineValidation) AddUpdateMachineValidationExternalConfigOnSite(ctx context.Context, request *cwssaws.AddUpdateMachineValidationExternalConfigRequest) error {
	logger := log.With().Str("Activity", "AddUpdateMachineValidationExternalConfigOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty add/update machine validation external config request")
	} else if request.Name == "" {
		err = errors.New("received add/update machine validation external config request missing Name")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mmv.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return client.ErrClientNotConnected
	}

	_, err = nicoClient.NICo().AddUpdateMachineValidationExternalConfig(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to add/update machine validation external config using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}

func (mmv *ManageMachineValidation) RemoveMachineValidationExternalConfigOnSite(ctx context.Context, request *cwssaws.RemoveMachineValidationExternalConfigRequest) error {
	logger := log.With().Str("Activity", "RemoveMachineValidationExternalConfigOnSite").Logger()

	logger.Info().Msg("Starting activity")

	var err error

	// Validate request
	if request == nil {
		err = errors.New("received empty remove machine validation external config request")
	} else if request.Name == "" {
		err = errors.New("received remove machine validation external config request missing Name")
	}

	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), swe.ErrTypeInvalidRequest, err)
	}

	// Call Site Controller gRPC endpoint
	nicoClient := mmv.NICoCoreAtomicClient.GetClient()
	if nicoClient == nil {
		return client.ErrClientNotConnected
	}

	_, err = nicoClient.NICo().RemoveMachineValidationExternalConfig(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to remove machine validation external config using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")

	return nil
}
