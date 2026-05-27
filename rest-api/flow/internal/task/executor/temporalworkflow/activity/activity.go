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
	"fmt"

	"github.com/google/uuid"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/executor/temporalworkflow/common"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/operations"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/task"
)

// Canonical Temporal activity names. These constants are the single source of
// truth: used in All() for worker registration and when scheduling via
// workflow.ExecuteActivity.
const (
	NameInjectExpectation         = "InjectExpectation"
	NamePowerControl              = "PowerControl"
	NameGetPowerStatus            = "GetPowerStatus"
	NameUpdateTaskStatus          = "UpdateTaskStatus"
	NameFirmwareControl           = "FirmwareControl"
	NameGetFirmwareStatus         = "GetFirmwareStatus"
	NameBringUpControl            = "BringUpControl"
	NameGetBringUpStatus          = "GetBringUpStatus"
	NameVerifyFirmwareConsistency = "VerifyFirmwareConsistency"
)

// InjectExpectation is a Temporal activity that registers expected component
// configurations with the appropriate component manager service.
func (a *Activities) InjectExpectation(
	ctx context.Context,
	target common.Target,
	info operations.InjectExpectationTaskInfo,
) error {
	cm, err := a.validAndGetComponentManager(target)
	if err != nil {
		return err
	}

	return cm.InjectExpectation(ctx, target, info)
}

// PowerControl is a Temporal activity that applies a power state transition
// to the target components via the appropriate component manager.
func (a *Activities) PowerControl(
	ctx context.Context,
	target common.Target,
	info operations.PowerControlTaskInfo,
) error {
	cm, err := a.validAndGetComponentManager(target)
	if err != nil {
		return err
	}

	return cm.PowerControl(ctx, target, info)
}

// GetPowerStatus is a Temporal activity that queries current power states for
// all components in the target. Returns a map of component ID to PowerStatus.
func (a *Activities) GetPowerStatus(
	ctx context.Context,
	target common.Target,
) (map[string]operations.PowerStatus, error) {
	cm, err := a.validAndGetComponentManager(target)
	if err != nil {
		return nil, err
	}

	return cm.GetPowerStatus(ctx, target)
}

// UpdateTaskStatus is a Temporal activity that updates task status by ID.
func (a *Activities) UpdateTaskStatus(
	ctx context.Context,
	arg *task.TaskStatusUpdate,
) error {
	if a.updater == nil {
		return fmt.Errorf("task status updater is not configured")
	}

	if arg == nil || arg.ID == uuid.Nil {
		return fmt.Errorf("invalid task identifier")
	}

	return a.updater.UpdateTaskStatus(ctx, arg)
}

// FirmwareControl initiates firmware update without waiting for completion.
// This activity returns immediately after the update request is accepted.
func (a *Activities) FirmwareControl(
	ctx context.Context,
	target common.Target,
	info operations.FirmwareControlTaskInfo,
) error {
	cm, err := a.validAndGetComponentManager(target)
	if err != nil {
		return err
	}

	return cm.FirmwareControl(ctx, target, info)
}

// GetFirmwareStatusResult is the result of GetFirmwareStatus activity.
type GetFirmwareStatusResult struct {
	// Statuses maps each component ID to its current firmware update state.
	Statuses map[string]operations.FirmwareUpdateStatus
}

// GetFirmwareStatus returns the current status of firmware updates.
// This activity is designed to be called repeatedly in a polling loop.
func (a *Activities) GetFirmwareStatus(
	ctx context.Context,
	target common.Target,
) (*GetFirmwareStatusResult, error) {
	cm, err := a.validAndGetComponentManager(target)
	if err != nil {
		return nil, err
	}

	statuses, err := cm.GetFirmwareStatus(ctx, target)
	if err != nil {
		return nil, err
	}

	return &GetFirmwareStatusResult{Statuses: statuses}, nil
}

// BringUpControl opens the power-on gate for the target components.
func (a *Activities) BringUpControl(
	ctx context.Context,
	target common.Target,
) error {
	cm, err := a.validAndGetComponentManager(target)
	if err != nil {
		return err
	}

	buc, ok := cm.(componentmanager.BringUpController)
	if !ok {
		return fmt.Errorf("component manager for %s does not support BringUpControl", target.Type)
	}

	return buc.BringUpControl(ctx, target)
}

// GetBringUpStatusResult is the result of GetBringUpStatus activity.
type GetBringUpStatusResult struct {
	// States maps each component ID to its current bring-up state.
	States map[string]operations.MachineBringUpState
}

// GetBringUpStatus returns the bring-up state for target components.
func (a *Activities) GetBringUpStatus(
	ctx context.Context,
	target common.Target,
) (*GetBringUpStatusResult, error) {
	cm, err := a.validAndGetComponentManager(target)
	if err != nil {
		return nil, err
	}

	buc, ok := cm.(componentmanager.BringUpController)
	if !ok {
		return nil, fmt.Errorf("component manager for %s does not support GetBringUpStatus", target.Type)
	}

	states, err := buc.GetBringUpStatus(ctx, target)
	if err != nil {
		return nil, err
	}

	return &GetBringUpStatusResult{States: states}, nil
}

// VerifyFirmwareConsistency checks that all target components report the
// same firmware version set. Only supported by component managers that
// implement FirmwareConsistencyChecker.
func (a *Activities) VerifyFirmwareConsistency(
	ctx context.Context,
	target common.Target,
) error {
	cm, err := a.validAndGetComponentManager(target)
	if err != nil {
		return err
	}

	checker, ok := cm.(componentmanager.FirmwareConsistencyChecker)
	if !ok {
		return fmt.Errorf("component manager for %s does not support firmware consistency check",
			target.Type)
	}

	return checker.VerifyFirmwareConsistency(ctx, target)
}

// validAndGetComponentManager validates the target and returns the component
// manager registered for its type. Returns an error if the target is invalid
// or no manager is found.
func (a *Activities) validAndGetComponentManager(
	target common.Target,
) (componentmanager.ComponentManager, error) {
	if err := target.Validate(); err != nil {
		return nil, fmt.Errorf("target is invalid: %w", err)
	}

	return a.registry.GetManager(target.Type)
}
