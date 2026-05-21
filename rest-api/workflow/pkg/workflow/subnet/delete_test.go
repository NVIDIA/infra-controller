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

package subnet

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	subnetActivity "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

type DeleteSubnetTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *DeleteSubnetTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *DeleteSubnetTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *DeleteSubnetTestSuite) Test_DeleteSubnetWorkflow_Success() {
	var subnetManager subnetActivity.ManageSubnet

	subnetID := uuid.New()
	request := &cwssaws.NetworkSegmentDeletionRequest{
		Id: &cwssaws.NetworkSegmentId{Value: subnetID.String()},
	}

	// Mock DeleteSubnetOnSite activity
	s.env.RegisterActivity(subnetManager.DeleteSubnetOnSite)
	s.env.OnActivity(subnetManager.DeleteSubnetOnSite, mock.Anything, request).Return(nil)

	// execute DeleteSubnetByID workflow
	s.env.ExecuteWorkflow(DeleteSubnetByID, subnetID)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *DeleteSubnetTestSuite) Test_DeleteSubnetWorkflow_ActivityFails() {
	var subnetManager subnetActivity.ManageSubnet

	subnetID := uuid.New()

	request := &cwssaws.NetworkSegmentDeletionRequest{
		Id: &cwssaws.NetworkSegmentId{Value: subnetID.String()},
	}

	// Mock DeleteSubnetOnSite activity failure
	s.env.RegisterActivity(subnetManager.DeleteSubnetOnSite)
	s.env.OnActivity(subnetManager.DeleteSubnetOnSite, mock.Anything, request).Return(errors.New("DeleteSubnetOnSite Failure"))

	// execute DeleteSubnetByID workflow
	s.env.ExecuteWorkflow(DeleteSubnetByID, subnetID)
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Error(err)

	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("DeleteSubnetOnSite Failure", applicationErr.Error())
}

func TestDeleteSubnetSuite(t *testing.T) {
	suite.Run(t, new(DeleteSubnetTestSuite))
}
