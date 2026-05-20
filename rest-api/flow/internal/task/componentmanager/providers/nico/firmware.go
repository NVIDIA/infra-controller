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

// MapFirmwareState converts a NICo protobuf FirmwareUpdateState into the
// corresponding operations.FirmwareUpdateState.
func MapFirmwareState(state pb.FirmwareUpdateState) operations.FirmwareUpdateState {
	switch state {
	case pb.FirmwareUpdateState_FW_STATE_QUEUED:
		return operations.FirmwareUpdateStateQueued
	case pb.FirmwareUpdateState_FW_STATE_IN_PROGRESS:
		return operations.FirmwareUpdateStateQueued // closest available state
	case pb.FirmwareUpdateState_FW_STATE_VERIFYING:
		return operations.FirmwareUpdateStateVerifying
	case pb.FirmwareUpdateState_FW_STATE_COMPLETED:
		return operations.FirmwareUpdateStateCompleted
	case pb.FirmwareUpdateState_FW_STATE_FAILED, pb.FirmwareUpdateState_FW_STATE_CANCELLED:
		return operations.FirmwareUpdateStateFailed
	default:
		return operations.FirmwareUpdateStateUnknown
	}
}
