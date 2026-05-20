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
	"go.temporal.io/sdk/workflow"

	swa "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	sww "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/workflow"
)

// RegisterSubscriber registers Instance CRUD workflows and activities with Temporal
func (api *API) RegisterSubscriber() error {
	ManagerAccess.Data.EB.Log.Info().Msg("Instance: Registering CRUD workflows and activities")

	// Register workflows

	// Register CreateInstanceV2 workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateInstanceV2)
	ManagerAccess.Data.EB.Log.Info().Msg("Instance: Successfully registered CreateInstanceV2 workflow")

	// Register CreateInstances workflow (Batch)
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateInstances)
	ManagerAccess.Data.EB.Log.Info().Msg("Instance: Successfully registered CreateInstances workflow")

	// Register DeleteInstanceV2 workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DeleteInstanceV2)
	ManagerAccess.Data.EB.Log.Info().Msg("Instance: Successfully registered DeleteInstanceV2 workflow")

	// Register UpdateInstance workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.UpdateInstance)
	ManagerAccess.Data.EB.Log.Info().Msg("Instance: Successfully registered UpdateInstance workflow")

	// Register RebootInstanceV2 workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflowWithOptions(sww.RebootInstance, workflow.RegisterOptions{
		Name: "RebootInstanceV2",
	})
	ManagerAccess.Data.EB.Log.Info().Msg("Instance: Successfully registered RebootInstanceV2 workflow")

	// Register activities

	instanceManager := swa.NewManageInstance(ManagerAccess.Data.EB.Managers.NICo.Client)

	// Register CreateInstanceOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(instanceManager.CreateInstanceOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("Instance: Successfully registered CreateInstanceOnSite activity")

	// Register CreateInstancesOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(instanceManager.CreateInstancesOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("Instance: Successfully registered CreateInstancesOnSite activity")

	// Register DeleteInstanceOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(instanceManager.DeleteInstanceOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("Instance: Successfully registered DeleteInstanceOnSite activity")

	// Register UpdateInstanceOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(instanceManager.UpdateInstanceOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("Instance: Successfully registered UpdateInstanceOnSite activity")

	// Register RebootInstanceOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(instanceManager.RebootInstanceOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("Instance: Successfully registered RebootInstanceOnSite activity")

	return nil
}
