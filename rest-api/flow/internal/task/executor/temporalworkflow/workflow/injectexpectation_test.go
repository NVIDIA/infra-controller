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

package workflow

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"

	activitypkg "github.com/NVIDIA/infra-controller-rest/flow/internal/task/executor/temporalworkflow/activity"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/executor/temporalworkflow/common"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/operations"
	taskdef "github.com/NVIDIA/infra-controller-rest/flow/internal/task/task"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/inventoryobjects/component"
)

func mockInjectExpectation(
	ctx context.Context,
	target common.Target,
	info operations.InjectExpectationTaskInfo,
) error {
	return nil
}

func TestInjectExpectationWorkflow(t *testing.T) {
	computeID1 := uuid.New()
	computeID2 := uuid.New()
	powershelfID := uuid.New()
	nvlswitchID := uuid.New()

	fullComponents := []*component.Component{
		newTestComponent(powershelfID, "powershelf-1", "ext-powershelf-1", devicetypes.ComponentTypePowerShelf),
		newTestComponent(nvlswitchID, "nvlswitch-1", "ext-nvlswitch-1", devicetypes.ComponentTypeNVLSwitch),
		newTestComponent(computeID1, "compute-1", "ext-compute-1", devicetypes.ComponentTypeCompute),
		newTestComponent(computeID2, "compute-2", "ext-compute-2", devicetypes.ComponentTypeCompute),
	}

	computeOnly := []*component.Component{
		newTestComponent(computeID1, "compute-1", "ext-compute-1", devicetypes.ComponentTypeCompute),
		newTestComponent(computeID2, "compute-2", "ext-compute-2", devicetypes.ComponentTypeCompute),
	}

	powershelfOnly := []*component.Component{
		newTestComponent(powershelfID, "powershelf-1", "ext-powershelf-1", devicetypes.ComponentTypePowerShelf),
	}

	testCases := map[string]struct {
		components    []*component.Component
		activityError error
		expectError   bool
		errorContains string
	}{
		"full components success": {
			components:  fullComponents,
			expectError: false,
		},
		"compute only success": {
			components:  computeOnly,
			expectError: false,
		},
		"powershelf only success": {
			components:  powershelfOnly,
			expectError: false,
		},
		"activity failure returns error": {
			components:    computeOnly,
			activityError: errors.New("component manager service unavailable"),
			expectError:   true,
			errorContains: "component manager service unavailable",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			env.RegisterActivityWithOptions(mockInjectExpectation, activity.RegisterOptions{
				Name: activitypkg.NameInjectExpectation,
			})
			env.RegisterActivityWithOptions(mockUpdateTaskStatus, activity.RegisterOptions{
				Name: activitypkg.NameUpdateTaskStatus,
			})

			env.OnActivity(mockInjectExpectation, mock.Anything, mock.Anything, mock.Anything).Return(tc.activityError)
			env.OnActivity(mockUpdateTaskStatus, mock.Anything, mock.Anything).Return(nil)

			info := &operations.InjectExpectationTaskInfo{}
			reqInfo := taskdef.ExecutionInfo{
				TaskID:     uuid.New(),
				Components: toWorkflowComponents(tc.components),
			}

			env.ExecuteWorkflow(injectExpectation, reqInfo, info)

			assert.True(t, env.IsWorkflowCompleted())

			wfErr := env.GetWorkflowError()
			if tc.expectError {
				assert.Error(t, wfErr)
				if tc.errorContains != "" && wfErr != nil {
					assert.Contains(t, wfErr.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, wfErr)
			}
		})
	}
}
