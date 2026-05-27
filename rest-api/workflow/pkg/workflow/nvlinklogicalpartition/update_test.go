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

package nvlinklogicalpartition

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	nvlinklogicalpartitionActivity "github.com/NVIDIA/infra-controller-rest/workflow/pkg/activity/nvlinklogicalpartition"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

type UpdateNVLinkLogicalPartitionTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *UpdateNVLinkLogicalPartitionTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UpdateNVLinkLogicalPartitionTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UpdateNVLinkLogicalPartitionTestSuite) Test_UpdateNVLinkLogicalPartitionInventory_Success() {
	var nvlinklogicalpartitionManager nvlinklogicalpartitionActivity.ManageNVLinkLogicalPartition

	siteID := uuid.New()

	inv := &cwssaws.NVLinkLogicalPartitionInventory{
		Partitions: []*cwssaws.NVLinkLogicalPartition{},
	}

	s.env.RegisterActivity(nvlinklogicalpartitionManager.UpdateNVLinkLogicalPartitionsInDB)
	s.env.OnActivity(nvlinklogicalpartitionManager.UpdateNVLinkLogicalPartitionsInDB, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(UpdateNVLinkLogicalPartitionInventory, siteID.String(), inv)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UpdateNVLinkLogicalPartitionTestSuite) Test_UpdateNVLinkLogicalPartitionInventory_ActivityFails() {
	var nvlinklogicalpartitionManager nvlinklogicalpartitionActivity.ManageNVLinkLogicalPartition

	siteID := uuid.New()

	inv := &cwssaws.NVLinkLogicalPartitionInventory{
		Partitions: []*cwssaws.NVLinkLogicalPartition{},
	}

	s.env.RegisterActivity(nvlinklogicalpartitionManager.UpdateNVLinkLogicalPartitionsInDB)
	s.env.OnActivity(nvlinklogicalpartitionManager.UpdateNVLinkLogicalPartitionsInDB, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("UpdateNVLinkLogicalPartitionInventory Failure"))

	s.env.ExecuteWorkflow(UpdateNVLinkLogicalPartitionInventory, siteID.String(), inv)
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.NotNil(err)

	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("UpdateNVLinkLogicalPartitionInventory Failure", applicationErr.Error())
}

func TestUpdateNVLinkLogicalPartitionTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateNVLinkLogicalPartitionTestSuite))
}
