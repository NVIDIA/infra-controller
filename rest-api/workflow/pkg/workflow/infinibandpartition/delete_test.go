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

package infinibandpartition

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	ibpActivity "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

type DeleteInfiniBandPartitionTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *DeleteInfiniBandPartitionTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *DeleteInfiniBandPartitionTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *DeleteInfiniBandPartitionTestSuite) Test_DeleteInfiniBandPartitionWorkflow_Success() {
	var ibpManager ibpActivity.ManageInfiniBandPartition

	ibpID := uuid.New()
	request := &cwssaws.IBPartitionDeletionRequest{
		Id: &cwssaws.IBPartitionId{Value: ibpID.String()},
	}

	// Mock DeleteInfiniBandPartitionOnSite activity
	s.env.RegisterActivity(ibpManager.DeleteInfiniBandPartitionOnSite)
	s.env.OnActivity(ibpManager.DeleteInfiniBandPartitionOnSite, mock.Anything, request).Return(nil)

	// Execute DeleteInfiniBandPartitionByID workflow
	s.env.ExecuteWorkflow(DeleteInfiniBandPartitionByID, ibpID)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *DeleteInfiniBandPartitionTestSuite) Test_DeleteInfiniBandPartitionWorkflow_ActivityFails() {
	var ibpManager ibpActivity.ManageInfiniBandPartition

	ibpID := uuid.New()
	request := &cwssaws.IBPartitionDeletionRequest{
		Id: &cwssaws.IBPartitionId{Value: ibpID.String()},
	}

	// Mock DeleteInfiniBandPartitionOnSite activity failure
	s.env.RegisterActivity(ibpManager.DeleteInfiniBandPartitionOnSite)
	s.env.OnActivity(ibpManager.DeleteInfiniBandPartitionOnSite, mock.Anything, request).Return(errors.New("DeleteInfiniBandPartitionOnSite Failure"))

	// Execute DeleteInfiniBandPartitionByID workflow
	s.env.ExecuteWorkflow(DeleteInfiniBandPartitionByID, ibpID)
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Error(err)

	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("DeleteInfiniBandPartitionOnSite Failure", applicationErr.Error())
}

func TestDeleteInfiniBandPartitionSuite(t *testing.T) {
	suite.Run(t, new(DeleteInfiniBandPartitionTestSuite))
}
