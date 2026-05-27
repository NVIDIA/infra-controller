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

package instance

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	instanceActivity "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

type RebootInstanceTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *RebootInstanceTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *RebootInstanceTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *RebootInstanceTestSuite) Test_RebootInstanceByIDWorkflow_Success() {
	var instanceManager instanceActivity.ManageInstance

	instanceID := uuid.New()
	request := &cwssaws.InstancePowerRequest{
		MachineId: &cwssaws.MachineId{
			Id: instanceID.String(),
		},
		BootWithCustomIpxe:   true,
		ApplyUpdatesOnReboot: true,
	}

	// Mock RebootInstanceOnSite activity
	s.env.RegisterActivity(instanceManager.RebootInstanceOnSite)
	s.env.OnActivity(instanceManager.RebootInstanceOnSite, mock.Anything, request).Return(nil)

	// execute RebootInstanceByID workflow
	s.env.ExecuteWorkflow(RebootInstanceByID, instanceID, true, true)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *RebootInstanceTestSuite) Test_RebootInstanceByIDWorkflow_ActivityFailsErrorActivityFails() {
	var instanceManager instanceActivity.ManageInstance

	instanceID := uuid.New()
	request := &cwssaws.InstancePowerRequest{
		MachineId: &cwssaws.MachineId{
			Id: instanceID.String(),
		},
		BootWithCustomIpxe:   true,
		ApplyUpdatesOnReboot: true,
	}

	// Mock RebootInstanceViaSiteAgent activity failure
	s.env.RegisterActivity(instanceManager.RebootInstanceOnSite)
	s.env.OnActivity(instanceManager.RebootInstanceOnSite, mock.Anything, request).Return(errors.New("RebootInstanceOnSite Failure"))

	// execute RebootInstanceByID workflow
	s.env.ExecuteWorkflow(RebootInstanceByID, instanceID, true, true)
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Error(err)

	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("RebootInstanceOnSite Failure", applicationErr.Error())
}

func TestRebootInstanceSuite(t *testing.T) {
	suite.Run(t, new(RebootInstanceTestSuite))
}
