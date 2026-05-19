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

package vpc

import (
	"errors"
	"testing"

	cwm "github.com/NVIDIA/infra-controller-rest/workflow/internal/metrics"
	vpcActivity "github.com/NVIDIA/infra-controller-rest/workflow/pkg/activity/vpc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

type UpdateVpcTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *UpdateVpcTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UpdateVpcTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UpdateVpcTestSuite) Test_UpdateVpcInventory_Success() {
	var vpcManager vpcActivity.ManageVpc
	var lifecycleMetricsManager vpcActivity.ManageVpcLifecycleMetrics
	var inventoryMetricsManager cwm.ManageInventoryMetrics

	siteID := uuid.New()
	vpcInventory := &cwssaws.VPCInventory{
		Vpcs: []*cwssaws.Vpc{
			{
				Id: &cwssaws.VpcId{Value: uuid.NewString()},
			},
			{
				Id: &cwssaws.VpcId{Value: uuid.NewString()},
			},
		},
	}

	// Mock UpdateVpcsInDB activity
	s.env.RegisterActivity(vpcManager.UpdateVpcsInDB)
	s.env.OnActivity(vpcManager.UpdateVpcsInDB, mock.Anything, siteID, mock.Anything).Return([]cwm.InventoryObjectLifecycleEvent{}, nil)

	// Mock RecordVpcStatusTransitionMetrics activity
	s.env.RegisterActivity(lifecycleMetricsManager.RecordVpcStatusTransitionMetrics)
	s.env.OnActivity(lifecycleMetricsManager.RecordVpcStatusTransitionMetrics, mock.Anything, siteID, mock.Anything).Return(nil)

	// Mock RecordLatency activity
	s.env.RegisterActivity(inventoryMetricsManager.RecordLatency)
	s.env.OnActivity(inventoryMetricsManager.RecordLatency, mock.Anything, siteID, "UpdateVpcInventory", false, mock.Anything).Return(nil)

	// execute UpdateVpcInventory workflow
	s.env.ExecuteWorkflow(UpdateVpcInventory, siteID.String(), vpcInventory)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UpdateVpcTestSuite) Test_UpdateVpcInventory_ActivityFails() {
	var vpcManager vpcActivity.ManageVpc

	siteID := uuid.New()
	vpcInventory := &cwssaws.VPCInventory{
		Vpcs: []*cwssaws.Vpc{
			{
				Id: &cwssaws.VpcId{Value: uuid.NewString()},
			},
			{
				Id: &cwssaws.VpcId{Value: uuid.NewString()},
			},
		},
	}

	// Mock UpdateVpcsInDB activity failure
	s.env.RegisterActivity(vpcManager.UpdateVpcsInDB)
	s.env.OnActivity(vpcManager.UpdateVpcsInDB, mock.Anything, siteID, mock.Anything).Return([]cwm.InventoryObjectLifecycleEvent{}, errors.New("UpdateVpcInventory Failure"))

	// execute UpdateVPCStatus workflow
	s.env.ExecuteWorkflow(UpdateVpcInventory, siteID.String(), vpcInventory)
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Error(err)

	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("UpdateVpcInventory Failure", applicationErr.Error())
}

func TestUpdateVpcSuite(t *testing.T) {
	suite.Run(t, new(UpdateVpcTestSuite))
}
