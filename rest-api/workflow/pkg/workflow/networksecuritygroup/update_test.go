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

package networksecuritygroup

import (
	"errors"
	"testing"

	networkSecurityGroupActivity "github.com/NVIDIA/infra-controller-rest/workflow/pkg/activity/networksecuritygroup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

type UpdateNetworkSecurityGroupTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *UpdateNetworkSecurityGroupTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UpdateNetworkSecurityGroupTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UpdateNetworkSecurityGroupTestSuite) Test_UpdateNetworkSecurityGroupInventory_Success() {

	var networkSecurityGroupManager networkSecurityGroupActivity.ManageNetworkSecurityGroup

	siteID := uuid.New()
	networkSecurityGroupInventory := &cwssaws.NetworkSecurityGroupInventory{
		NetworkSecurityGroups: []*cwssaws.NetworkSecurityGroup{
			{
				Id: uuid.NewString(),
			},
			{
				Id: uuid.NewString(),
			},
		},
	}

	// Mock UpdateNetworkSecurityGroupViaSiteAgent activity
	s.env.RegisterActivity(networkSecurityGroupManager.UpdateNetworkSecurityGroupsInDB)
	s.env.OnActivity(networkSecurityGroupManager.UpdateNetworkSecurityGroupsInDB, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// execute UpdateNetworkSecurityGroupInventory workflow
	s.env.ExecuteWorkflow(UpdateNetworkSecurityGroupInventory, siteID.String(), networkSecurityGroupInventory)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UpdateNetworkSecurityGroupTestSuite) Test_UpdateNetworkSecurityGroupInventory_ActivityFails() {

	var networkSecurityGroupManager networkSecurityGroupActivity.ManageNetworkSecurityGroup

	siteID := uuid.New()
	networkSecurityGroupInventory := &cwssaws.NetworkSecurityGroupInventory{
		NetworkSecurityGroups: []*cwssaws.NetworkSecurityGroup{
			{
				Id: uuid.NewString(),
			},
			{
				Id: uuid.NewString(),
			},
		},
	}

	// Mock UpdateNetworkSecurityGroupsViaSiteAgent activity failure
	s.env.RegisterActivity(networkSecurityGroupManager.UpdateNetworkSecurityGroupsInDB)
	s.env.OnActivity(networkSecurityGroupManager.UpdateNetworkSecurityGroupsInDB, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("UpdateNetworkSecurityGroupInventory Failure"))

	// execute UpdateNetworkSecurityGroupStatus workflow
	s.env.ExecuteWorkflow(UpdateNetworkSecurityGroupInventory, siteID.String(), networkSecurityGroupInventory)
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Error(err)

	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("UpdateNetworkSecurityGroupInventory Failure", applicationErr.Error())
}

func TestUpdateNetworkSecurityGroupSuite(t *testing.T) {
	suite.Run(t, new(UpdateNetworkSecurityGroupTestSuite))
}
