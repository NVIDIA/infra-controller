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

package managertypes

import (
	bootstraptypes "github.com/NVIDIA/infra-controller-rest/site-agent/pkg/datatypes/managertypes/bootstrap"
	flowtypes "github.com/NVIDIA/infra-controller-rest/site-agent/pkg/datatypes/managertypes/flow"
	nicotypes "github.com/NVIDIA/infra-controller-rest/site-agent/pkg/datatypes/managertypes/nico"
	workflowtypes "github.com/NVIDIA/infra-controller-rest/site-agent/pkg/datatypes/managertypes/workflow"
)

// Managers - manager ds
type Managers struct {
	Version string
	// All the datastructures of Managers below
	Workflow  *workflowtypes.Workflow
	NICo      *nicotypes.NICo
	Flow      *flowtypes.Flow
	Bootstrap *bootstraptypes.Bootstrap
}

// NewManagerType - get new type of all managers
func NewManagerType() *Managers {
	return &Managers{
		Version: "0.0.1",
		// All the managers below
		Workflow:  workflowtypes.NewWorkflowInstance(),
		NICo:      nicotypes.NewNICoInstance(),
		Flow:      flowtypes.NewFlowInstance(),
		Bootstrap: bootstraptypes.NewBootstrapInstance(),
	}
}
