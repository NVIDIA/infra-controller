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

package sku

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	skuActivity "github.com/NVIDIA/infra-controller-rest/workflow/pkg/activity/sku"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

type UpdateSkuTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *UpdateSkuTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UpdateSkuTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UpdateSkuTestSuite) Test_UpdateSkuInventory_Success() {
	var skuManager skuActivity.ManageSku

	siteID := uuid.New()

	inv := &cwssaws.SkuInventory{Skus: []*cwssaws.Sku{}}

	s.env.RegisterActivity(skuManager.UpdateSkusInDB)
	s.env.OnActivity(skuManager.UpdateSkusInDB, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(UpdateSkuInventory, siteID.String(), inv)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UpdateSkuTestSuite) Test_UpdateSkuInventory_ActivityFails() {
	var skuManager skuActivity.ManageSku

	siteID := uuid.New()

	inv := &cwssaws.SkuInventory{Skus: []*cwssaws.Sku{}}

	s.env.RegisterActivity(skuManager.UpdateSkusInDB)
	s.env.OnActivity(skuManager.UpdateSkusInDB, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("UpdateSkuInventory Failure"))

	s.env.ExecuteWorkflow(UpdateSkuInventory, siteID.String(), inv)
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.NotNil(err)

	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("UpdateSkuInventory Failure", applicationErr.Error())
}

func TestUpdateSkuTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateSkuTestSuite))
}
