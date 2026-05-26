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
	"fmt"

	computils "github.com/NVIDIA/infra-controller-rest/site-agent/pkg/components/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricFlowStatus - Metric Flow Status
	MetricFlowStatus = "flow_health_status"
)

// Init - initialize Flow manager
func (f *API) Init() {
	// Check if Flow is enabled via environment variable
	if !ManagerAccess.Conf.EB.Flow.Enabled {
		ManagerAccess.Data.EB.Log.Info().Msg("Flow: Flow is disabled, skipping initialization")
		return
	}

	ManagerAccess.Data.EB.Log.Info().Msg("Flow: Initializing Flow manager")

	prometheus.MustRegister(
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace: "elektra_site_agent",
			Name:      MetricFlowStatus,
			Help:      "Flow gRPC health status",
		},
			func() float64 {
				return float64(ManagerAccess.Data.EB.Managers.Flow.State.HealthStatus.Load())
			}))
	ManagerAccess.Data.EB.Managers.Flow.State.HealthStatus.Store(uint64(computils.CompNotKnown))

	// initialize workflow metrics
	ManagerAccess.Data.EB.Managers.Flow.State.WflowMetrics = newWorkflowMetrics()
}

// Start - Start Flow manager
func (f *API) Start() {
	ManagerAccess.Data.EB.Log.Info().Msg("Flow: Starting Flow manager")

	// Check if Flow is enabled via environment variable
	if !ManagerAccess.Conf.EB.Flow.Enabled {
		ManagerAccess.Data.EB.Log.Info().Msg("Flow: Flow is disabled, skipping gRPC client initialization")
		return
	}

	// Create the client here
	// Each workflow will check and reinitialize the client if needed
	if err := f.CreateGRPCClient(); err != nil {
		ManagerAccess.Data.EB.Log.Error().Msgf("Flow: failed to create GRPC client: %v", err)
	}
}

// GetState Machine
func (f *API) GetState() []string {
	state := ManagerAccess.Data.EB.Managers.Flow.State
	var strs []string
	strs = append(strs, fmt.Sprintln(" GRPC Succeeded:", state.GrpcSucc.Load()))
	strs = append(strs, fmt.Sprintln(" GRPC Failed:", state.GrpcFail.Load()))
	strs = append(strs, fmt.Sprintln(" Flow Status:", computils.CompStatus(state.HealthStatus.Load())))
	strs = append(strs, fmt.Sprintln(" Flow Last Error:", state.Err))

	return strs
}

// GetGrpcClientVersion returns the current version of the GRPC client
func (f *API) GetGRPCClientVersion() int64 {
	return ManagerAccess.Data.EB.Managers.Flow.Client.Version()
}
