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

package tenant

import (
	"errors"
	"testing"

	tenantActivity "github.com/NVIDIA/infra-controller-rest/workflow/pkg/activity/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"google.golang.org/protobuf/types/known/timestamppb"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

type UpdateTenantTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *UpdateTenantTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UpdateTenantTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UpdateTenantTestSuite) Test_UpdateTenantInventory_Success() {
	var tenantManager tenantActivity.ManageTenant

	siteID := uuid.New()

	tenantInventory := &cwssaws.TenantInventory{
		Tenants:   []*cwssaws.Tenant{},
		Timestamp: timestamppb.Now(),
	}

	// Mock UpdateTenantsInDB activity
	s.env.RegisterActivity(tenantManager.UpdateTenantsInDB)
	s.env.OnActivity(tenantManager.UpdateTenantsInDB, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// execute UpdateTenantInventory workflow
	s.env.ExecuteWorkflow(UpdateTenantInventory, siteID.String(), tenantInventory)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UpdateTenantTestSuite) Test_UpdateTenantInventory_ActivityFails() {
	var tenantManager tenantActivity.ManageTenant

	siteID := uuid.New()

	tenantInventory := &cwssaws.TenantInventory{
		Tenants:   []*cwssaws.Tenant{},
		Timestamp: timestamppb.Now(),
	}

	// Mock UpdateTenantsInDB activity
	s.env.RegisterActivity(tenantManager.UpdateTenantsInDB)
	s.env.OnActivity(tenantManager.UpdateTenantsInDB, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("UpdateTenantInventory Failure"))

	// execute UpdateTenantInventory workflow
	s.env.ExecuteWorkflow(UpdateTenantInventory, siteID.String(), tenantInventory)
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Error(err)

	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("UpdateTenantInventory Failure", applicationErr.Error())
}

func TestUpdateTenantInfoSuite(t *testing.T) {
	suite.Run(t, new(UpdateTenantTestSuite))
}
