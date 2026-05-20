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

package vpcprefix

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"

	vpcPrefixActivity "github.com/NVIDIA/infra-controller-rest/workflow/pkg/activity/vpcprefix"
)

type UpdateVpcPrefixTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *UpdateVpcPrefixTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UpdateVpcPrefixTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UpdateVpcPrefixTestSuite) Test_UpdateVpcPrefixInventory_Success() {

	var VpcPrefixManager vpcPrefixActivity.ManageVpcPrefix

	siteID := uuid.New()
	VpcPrefixInventory := &cwssaws.VpcPrefixInventory{
		VpcPrefixes: []*cwssaws.VpcPrefix{
			{
				Id: &cwssaws.VpcPrefixId{Value: uuid.NewString()},
			},
			{
				Id: &cwssaws.VpcPrefixId{Value: uuid.NewString()},
			},
		},
	}

	// Mock UpdateVpcPrefixViaSiteAgent activity
	s.env.RegisterActivity(VpcPrefixManager.UpdateVpcPrefixesInDB)
	s.env.OnActivity(VpcPrefixManager.UpdateVpcPrefixesInDB, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// execute UpdateVpcPrefixInventory workflow
	s.env.ExecuteWorkflow(UpdateVpcPrefixInventory, siteID.String(), VpcPrefixInventory)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UpdateVpcPrefixTestSuite) Test_UpdateVpcPrefixInventory_ActivityFails() {

	var VpcPrefixManager vpcPrefixActivity.ManageVpcPrefix

	siteID := uuid.New()
	VpcPrefixInventory := &cwssaws.VpcPrefixInventory{
		VpcPrefixes: []*cwssaws.VpcPrefix{
			{
				Id: &cwssaws.VpcPrefixId{Value: uuid.NewString()},
			},
			{
				Id: &cwssaws.VpcPrefixId{Value: uuid.NewString()},
			},
		},
	}

	// Mock UpdateVpcPrefixesInDB activity failure
	s.env.RegisterActivity(VpcPrefixManager.UpdateVpcPrefixesInDB)
	s.env.OnActivity(VpcPrefixManager.UpdateVpcPrefixesInDB, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("UpdateVpcPrefixInventory Failure"))

	// execute UpdateVpcPrefixStatus workflow
	s.env.ExecuteWorkflow(UpdateVpcPrefixInventory, siteID.String(), VpcPrefixInventory)
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Error(err)

	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("UpdateVpcPrefixInventory Failure", applicationErr.Error())
}

func TestUpdateVpcPrefixSuite(t *testing.T) {
	suite.Run(t, new(UpdateVpcPrefixTestSuite))
}
