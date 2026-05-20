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

package site

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	siteActivity "github.com/NVIDIA/infra-controller-rest/workflow/pkg/activity/site"
	tmocks "go.temporal.io/sdk/mocks"
)

type DeleteSiteComponentsTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *DeleteSiteComponentsTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *DeleteSiteComponentsTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *DeleteSiteComponentsTestSuite) Test_DeleteSiteComponentsWorkflow_Success() {
	var siteManager siteActivity.ManageSite

	siteID := uuid.New()
	infrastructureProviderID := uuid.New()
	purgeMachines := true

	// Mock DeleteSiteComponentsFromDB activity success
	s.env.RegisterActivity(siteManager.DeleteSiteComponentsFromDB)
	s.env.OnActivity(siteManager.DeleteSiteComponentsFromDB, mock.Anything, siteID, infrastructureProviderID, purgeMachines).Return(nil)

	// Execute DeleteSiteComponents workflow
	s.env.ExecuteWorkflow(DeleteSiteComponents, siteID, infrastructureProviderID, purgeMachines)
	s.env.AssertCalled(s.T(), "DeleteSiteComponentsFromDB", mock.Anything, siteID, infrastructureProviderID, purgeMachines)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *DeleteSiteComponentsTestSuite) Test_DeleteSiteComponentsWorkflow_ActivityFails() {
	var siteManager siteActivity.ManageSite

	siteID := uuid.New()
	infrastructureProviderID := uuid.New()
	purgeMachines := false

	// Mock DeleteSiteComponentFromDB activity failure
	s.env.RegisterActivity(siteManager.DeleteSiteComponentsFromDB)
	s.env.OnActivity(siteManager.DeleteSiteComponentsFromDB, mock.Anything, siteID, infrastructureProviderID, purgeMachines).Return(errors.New("DeleteSiteComponentsFromDB Failure"))

	// Execute DeleteSiteComponents workflow
	s.env.ExecuteWorkflow(DeleteSiteComponents, siteID, infrastructureProviderID, purgeMachines)
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Error(err)

	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("DeleteSiteComponentsFromDB Failure", applicationErr.Error())
}

func (s *DeleteSiteComponentsTestSuite) Test_ExecuteDeleteSiteComponentsWorkflow_Success() {
	ctx := context.Background()
	siteID := uuid.New()
	infrastructureProviderID := uuid.New()
	purgeMachines := true

	wid := "test-workflow-id"

	wrun := &tmocks.WorkflowRun{}
	wrun.On("GetID").Return(wid)

	tc := &tmocks.Client{}

	tc.Mock.On("ExecuteWorkflow", context.Background(), mock.AnythingOfType("internal.StartWorkflowOptions"),
		mock.Anything, siteID, infrastructureProviderID, purgeMachines).Return(wrun, nil)

	rwid, err := ExecuteDeleteSiteComponentsWorkflow(ctx, tc, siteID, infrastructureProviderID, purgeMachines)
	s.NoError(err)
	s.Equal(wid, *rwid)
}

func TestDeleteSiteComponentsSuite(t *testing.T) {
	suite.Run(t, new(DeleteSiteComponentsTestSuite))
}
