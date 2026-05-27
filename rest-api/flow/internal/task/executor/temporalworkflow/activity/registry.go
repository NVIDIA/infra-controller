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
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/task"
)

// Activities holds the per-manager-instance dependencies for all Temporal
// activities. Construct one via New and pass its methods to each Temporal
// worker via RegisterActivityWithOptions. Because each Activities instance is
// independent, multiple managers can coexist in the same process without
// sharing mutable state.
type Activities struct {
	updater  task.TaskStatusUpdater
	registry *componentmanager.Registry
}

// New creates an Activities instance bound to the given status updater and
// component manager registry. Either argument may be nil; activity calls that
// require the missing dependency will return an error at invocation time.
func New(
	updater task.TaskStatusUpdater,
	registry *componentmanager.Registry,
) *Activities {
	return &Activities{
		updater:  updater,
		registry: registry,
	}
}

// All returns a map of Temporal activity name to bound method for worker
// registration via RegisterActivityWithOptions. Each entry is a bound method
// that captures this Activities instance, so its dependencies are isolated
// from other Activities instances.
func (a *Activities) All() map[string]any {
	return map[string]any{
		NameInjectExpectation:         a.InjectExpectation,
		NamePowerControl:              a.PowerControl,
		NameGetPowerStatus:            a.GetPowerStatus,
		NameUpdateTaskStatus:          a.UpdateTaskStatus,
		NameFirmwareControl:           a.FirmwareControl,
		NameGetFirmwareStatus:         a.GetFirmwareStatus,
		NameBringUpControl:            a.BringUpControl,
		NameGetBringUpStatus:          a.GetBringUpStatus,
		NameVerifyFirmwareConsistency: a.VerifyFirmwareConsistency,
	}
}
