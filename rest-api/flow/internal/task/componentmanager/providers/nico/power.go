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

package nico

import (
	pb "github.com/NVIDIA/infra-controller-rest/flow/internal/nicoapi/gen"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/operations"
)

// ExtractPowerState derives an operations.PowerStatus from the first
// ComputerSystem in a site exploration report.  Returns PowerStatusUnknown
// when the report is nil or contains no systems.
func ExtractPowerState(report *pb.EndpointExplorationReport) operations.PowerStatus {
	if report == nil {
		return operations.PowerStatusUnknown
	}
	systems := report.GetSystems()
	if len(systems) == 0 {
		return operations.PowerStatusUnknown
	}
	switch systems[0].GetPowerState() {
	case pb.ComputerSystemPowerState_On:
		return operations.PowerStatusOn
	case pb.ComputerSystemPowerState_Off:
		return operations.PowerStatusOff
	default:
		return operations.PowerStatusUnknown
	}
}
