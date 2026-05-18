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

package nicotypes

import (
	"github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"
	"go.uber.org/atomic"
)

// State - NICo state
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

// NICo represents the gRPC client for NICo and state
type NICo struct {
	Client *client.NICoCoreAtomicClient
	State  *State
}

// NewNICoInstance creates a new instance of NICo
func NewNICoInstance() *NICo {
	nico := &NICo{
		State:  &State{},
		Client: client.NewNICoCoreAtomicClient(&client.NICoCoreClientConfig{}),
	}

	return nico
}

// GetClient returns the NICo client
func (c *NICo) GetClient() *client.NICoCoreClient {
	return c.Client.GetClient()
}
