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
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	temporalworkflow "go.temporal.io/sdk/workflow"

	activitypkg "github.com/NVIDIA/infra-controller-rest/flow/internal/task/executor/temporalworkflow/activity"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/executor/temporalworkflow/common"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/operationrules"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/operations"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/task"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

func mockUpdateTaskStatusForBringUp(ctx context.Context, arg *task.TaskStatusUpdate) error {
	return nil
}

func mockBringUpControl(ctx context.Context, target common.Target) error {
	return nil
}

func mockGetBringUpStatus(ctx context.Context, target common.Target) (*activitypkg.GetBringUpStatusResult, error) {
	return &activitypkg.GetBringUpStatusResult{
		States: map[string]operations.MachineBringUpState{},
	}, nil
}

func createBringUpTestRuleDef() *operationrules.RuleDefinition {
	return &operationrules.RuleDefinition{
		Version: "v1",
		Steps: []operationrules.SequenceStep{
			{
				ComponentType: devicetypes.ComponentTypePowerShelf,
				Stage:         1,
				MaxParallel:   0,
				Timeout:       10 * time.Minute,
				PreOperation: []operationrules.ActionConfig{
					{
						Name:         operationrules.ActionVerifyReachability,
						Timeout:      5 * time.Second,
						PollInterval: 1 * time.Second,
						Parameters: map[string]any{
							operationrules.ParamComponentTypes: []string{"powershelf"},
							operationrules.ParamRequireAll:     true,
						},
					},
				},
				MainOperation: operationrules.ActionConfig{
					Name: operationrules.ActionPowerControl,
					Parameters: map[string]any{
						operationrules.ParamOperation: "power_on",
					},
				},
				PostOperation: []operationrules.ActionConfig{
					{
						Name:         operationrules.ActionVerifyPowerStatus,
						Timeout:      5 * time.Second,
						PollInterval: 1 * time.Second,
						Parameters: map[string]any{
							operationrules.ParamExpectedStatus: "on",
						},
					},
				},
			},
			{
				ComponentType: devicetypes.ComponentTypeCompute,
				Stage:         2,
				MaxParallel:   0,
				Timeout:       10 * time.Minute,
				MainOperation: operationrules.ActionConfig{
					Name: operationrules.ActionBringUpControl,
				},
				PostOperation: []operationrules.ActionConfig{
					{
						Name:         operationrules.ActionWaitBringUp,
						Timeout:      5 * time.Second,
						PollInterval: 1 * time.Second,
					},
				},
			},
		},
	}
}

func createBringUpTestComponents() []task.WorkflowComponent {
	return []task.WorkflowComponent{
		{
			ComponentID: "ps-1",
			Type:        devicetypes.ComponentTypePowerShelf,
		},
		{
			ComponentID: "compute-1",
			Type:        devicetypes.ComponentTypeCompute,
		},
	}
}

func registerBringUpActivities(env *testsuite.TestWorkflowEnvironment) {
	env.RegisterWorkflowWithOptions(bringUp, temporalworkflow.RegisterOptions{Name: "BringUp"})
	env.RegisterWorkflowWithOptions(genericComponentStepWorkflow, temporalworkflow.RegisterOptions{Name: nameGenericComponentStepWorkflow})
	env.RegisterActivityWithOptions(mockUpdateTaskStatusForBringUp,
		activity.RegisterOptions{Name: activitypkg.NameUpdateTaskStatus})
	env.RegisterActivityWithOptions(mockPowerControl,
		activity.RegisterOptions{Name: activitypkg.NamePowerControl})
	env.RegisterActivityWithOptions(mockGetPowerStatus,
		activity.RegisterOptions{Name: activitypkg.NameGetPowerStatus})
	env.RegisterActivityWithOptions(mockBringUpControl,
		activity.RegisterOptions{Name: activitypkg.NameBringUpControl})
	env.RegisterActivityWithOptions(mockGetBringUpStatus,
		activity.RegisterOptions{Name: activitypkg.NameGetBringUpStatus})
}

func TestBringUpWorkflow(t *testing.T) {
	testCases := map[string]struct {
		setupMocks  func(env *testsuite.TestWorkflowEnvironment)
		expectError bool
	}{
		"success": {
			setupMocks: func(env *testsuite.TestWorkflowEnvironment) {
				env.OnActivity(activitypkg.NameUpdateTaskStatus, mock.Anything, mock.Anything).Return(nil)
				env.OnActivity(activitypkg.NamePowerControl, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				env.OnActivity(activitypkg.NameGetPowerStatus, mock.Anything, mock.Anything).Return(
					map[string]operations.PowerStatus{
						"ps-1": operations.PowerStatusOn,
					}, nil)
				env.OnActivity(activitypkg.NameBringUpControl, mock.Anything, mock.Anything).Return(nil)
				env.OnActivity(activitypkg.NameGetBringUpStatus, mock.Anything, mock.Anything).Return(
					&activitypkg.GetBringUpStatusResult{
						States: map[string]operations.MachineBringUpState{
							"compute-1": operations.MachineBringUpStateMachineCreated,
						},
					}, nil)
			},
			expectError: false,
		},
		"power control failure": {
			setupMocks: func(env *testsuite.TestWorkflowEnvironment) {
				env.OnActivity(activitypkg.NameUpdateTaskStatus, mock.Anything, mock.Anything).Return(nil)
				env.OnActivity(activitypkg.NameGetPowerStatus, mock.Anything, mock.Anything).Return(
					map[string]operations.PowerStatus{
						"ps-1": operations.PowerStatusOff,
					}, nil)
				env.OnActivity(activitypkg.NamePowerControl, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("BMC unreachable"))
			},
			expectError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()
			registerBringUpActivities(env)
			tc.setupMocks(env)

			reqInfo := task.ExecutionInfo{
				TaskID:         uuid.New(),
				Components:     createBringUpTestComponents(),
				RuleDefinition: createBringUpTestRuleDef(),
			}
			info := &operations.BringUpTaskInfo{}

			env.ExecuteWorkflow("BringUp", reqInfo, info)

			assert.True(t, env.IsWorkflowCompleted())
			if tc.expectError {
				assert.Error(t, env.GetWorkflowError())
			} else {
				assert.NoError(t, env.GetWorkflowError())
			}
		})
	}
}

// TestBringUpWorkflowWithIngestion tests the BringUp workflow when executed
// with an ingestion-only rule (as triggered by IngestRack API). All component
// types run InjectExpectation in parallel within a single stage.
func TestBringUpWorkflowWithIngestion(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	mockInjectExpectation := func(
		ctx context.Context,
		target common.Target,
		info operations.InjectExpectationTaskInfo,
	) error {
		return nil
	}

	env.RegisterWorkflowWithOptions(bringUp, temporalworkflow.RegisterOptions{Name: "BringUp"})
	env.RegisterWorkflowWithOptions(genericComponentStepWorkflow, temporalworkflow.RegisterOptions{Name: nameGenericComponentStepWorkflow})
	env.RegisterActivityWithOptions(mockUpdateTaskStatusForBringUp,
		activity.RegisterOptions{Name: activitypkg.NameUpdateTaskStatus})
	env.RegisterActivityWithOptions(mockInjectExpectation,
		activity.RegisterOptions{Name: activitypkg.NameInjectExpectation})

	env.OnActivity(activitypkg.NameUpdateTaskStatus, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(activitypkg.NameInjectExpectation, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	testComponents := []task.WorkflowComponent{
		{ComponentID: "ps-1", Type: devicetypes.ComponentTypePowerShelf},
		{ComponentID: "compute-1", Type: devicetypes.ComponentTypeCompute},
		{ComponentID: "switch-1", Type: devicetypes.ComponentTypeNVLSwitch},
	}

	ingestRule := &operationrules.RuleDefinition{
		Version: "v1",
		Steps: []operationrules.SequenceStep{
			{
				ComponentType: devicetypes.ComponentTypePowerShelf,
				Stage:         1,
				MaxParallel:   0,
				Timeout:       10 * time.Minute,
				MainOperation: operationrules.ActionConfig{
					Name: operationrules.ActionInjectExpectation,
				},
			},
			{
				ComponentType: devicetypes.ComponentTypeCompute,
				Stage:         1,
				MaxParallel:   0,
				Timeout:       10 * time.Minute,
				MainOperation: operationrules.ActionConfig{
					Name: operationrules.ActionInjectExpectation,
				},
			},
			{
				ComponentType: devicetypes.ComponentTypeNVLSwitch,
				Stage:         1,
				MaxParallel:   0,
				Timeout:       10 * time.Minute,
				MainOperation: operationrules.ActionConfig{
					Name: operationrules.ActionInjectExpectation,
				},
			},
		},
	}

	reqInfo := task.ExecutionInfo{
		TaskID:         uuid.New(),
		Components:     testComponents,
		RuleDefinition: ingestRule,
	}
	info := &operations.BringUpTaskInfo{}

	env.ExecuteWorkflow("BringUp", reqInfo, info)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
}

func TestBringUpWorkflowEmptyRack(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflowWithOptions(bringUp, temporalworkflow.RegisterOptions{Name: "BringUp"})

	reqInfo := task.ExecutionInfo{
		TaskID:         uuid.New(),
		Components:     []task.WorkflowComponent{},
		RuleDefinition: createBringUpTestRuleDef(),
	}
	info := &operations.BringUpTaskInfo{}

	env.ExecuteWorkflow("BringUp", reqInfo, info)

	assert.True(t, env.IsWorkflowCompleted())
	assert.Error(t, env.GetWorkflowError())
}
