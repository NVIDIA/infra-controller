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

package componentmanager

import (
	"context"

	cmcatalog "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/catalog"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/executor/temporalworkflow/common"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/operations"
)

// ComponentManager defines the interface for managing various types of
// components. Implementations handle component-specific operations like
// power control, firmware management, and status monitoring.
type ComponentManager interface {
	// Descriptor returns the component manager metadata for this manager.
	Descriptor() cmcatalog.Descriptor

	// InjectExpectation registers expected component configurations with the
	// component manager service for the target components.
	InjectExpectation(ctx context.Context, target common.Target, info operations.InjectExpectationTaskInfo) error //nolint

	// PowerControl applies a power state transition to the target components.
	PowerControl(ctx context.Context, target common.Target, info operations.PowerControlTaskInfo) error //nolint

	// GetPowerStatus queries the current power state of each component in the
	// target. Returns a map of component ID to PowerStatus.
	GetPowerStatus(ctx context.Context, target common.Target) (map[string]operations.PowerStatus, error) //nolint

	// FirmwareControl initiates a firmware update without waiting for completion.
	// Returns immediately after the update request is accepted.
	FirmwareControl(ctx context.Context, target common.Target, info operations.FirmwareControlTaskInfo) error //nolint

	// GetFirmwareStatus returns the current firmware update state for each
	// component in the target. Returns a map of component ID to FirmwareUpdateStatus.
	GetFirmwareStatus(ctx context.Context, target common.Target) (map[string]operations.FirmwareUpdateStatus, error) //nolint
}

// BringUpController is an optional interface for component managers that support
// bring-up operations.
type BringUpController interface {
	// BringUpControl opens the power-on gate for the target components, allowing
	// them to proceed through the bring-up sequence.
	BringUpControl(ctx context.Context, target common.Target) error

	// GetBringUpStatus returns the current bring-up state for each component in
	// the target. Returns a map of component ID to MachineBringUpState.
	GetBringUpStatus(ctx context.Context, target common.Target) (map[string]operations.MachineBringUpState, error)
}

// FirmwareConsistencyChecker is an optional interface for component managers
// that can verify firmware version consistency across a set of components.
type FirmwareConsistencyChecker interface {
	// VerifyFirmwareConsistency checks that all target components report the same
	// firmware version set. Returns an error if versions diverge.
	VerifyFirmwareConsistency(ctx context.Context, target common.Target) error
}
