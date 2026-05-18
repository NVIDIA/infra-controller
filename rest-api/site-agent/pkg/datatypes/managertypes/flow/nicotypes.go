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

package flowtypes

import (
	"github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"
	"go.uber.org/atomic"
)

// State - Flow state
type State struct {
	// GrpcFail the number of times the rpc has failed
	GrpcFail atomic.Uint64
	// GrpcSucc the number of times the rpc has succeeded
	GrpcSucc atomic.Uint64
	// HealthStatus current health state
	HealthStatus atomic.Uint64
	// Err is error message
	Err string
	// WflowMetrics workflow metrics
	WflowMetrics WorkflowMetrics
}

// Flow represents the gRPC client for Flow and state
type Flow struct {
	Client *client.FlowAtomicClient
	State  *State
}

// NewFlowInstance creates a new instance of Flow
func NewFlowInstance() *Flow {
	f := &Flow{
		State:  &State{},
		Client: client.NewFlowAtomicClient(&client.FlowClientConfig{}),
	}

	return f
}

// GetClient returns the Flow client
func (c *Flow) GetClient() *client.FlowClient {
	return c.Client.GetClient()
}
