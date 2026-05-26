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

package instancetype

import (
	swa "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	sww "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/workflow"
)

// RegisterSubscriber registers InstanceType CRUD workflows and activities with Temporal
func (api *API) RegisterSubscriber() error {
	ManagerAccess.Data.EB.Log.Info().Msg("InstanceType: Registering CRUD workflows and activities")

	// Register workflows

	// Register CreateInstanceType workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateInstanceType)
	ManagerAccess.Data.EB.Log.Info().Msg("InstanceType: Successfully registered CreateInstanceType workflow")

	// Register UpdateInstanceType workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.UpdateInstanceType)
	ManagerAccess.Data.EB.Log.Info().Msg("InstanceType: Successfully registered UpdateInstanceType workflow")

	// Register DeleteInstanceType workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DeleteInstanceType)
	ManagerAccess.Data.EB.Log.Info().Msg("InstanceType: Successfully registered DeleteInstanceType workflow")

	// Register AssociateMachinesWithInstanceType workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.AssociateMachinesWithInstanceType)
	ManagerAccess.Data.EB.Log.Info().Msg("InstanceType: Successfully registered AssociateMachinesWithInstanceType workflow")

	// Register RemoveMachineInstanceTypeAssociation workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.RemoveMachineInstanceTypeAssociation)
	ManagerAccess.Data.EB.Log.Info().Msg("InstanceType: Successfully registered RemoveMachineInstanceTypeAssociation workflow")

	// Register activities
	instanceTypeManager := swa.NewManageInstanceType(ManagerAccess.Data.EB.Managers.NICo.Client)

	// Register CreateInstanceTypeOnSite
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(instanceTypeManager.CreateInstanceTypeOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("InstanceType: Successfully registered CreateInstanceTypeOnSite activity")

	// Register UpdateInstanceTypeOnSite
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(instanceTypeManager.UpdateInstanceTypeOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("InstanceType: Successfully registered UpdateInstanceTypeOnSite activity")

	// Register DeleteInstanceTypeOnSite
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(instanceTypeManager.DeleteInstanceTypeOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("InstanceType: Successfully registered DeleteInstanceTypeOnSite activity")

	// Register AssociateMachinesWithInstanceTypeOnSite
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(instanceTypeManager.AssociateMachinesWithInstanceTypeOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("InstanceType: Successfully registered AssociateMachinesWithInstanceTypeOnSite activity")

	// Register RemoveMachineInstanceTypeAssociationOnSite
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(instanceTypeManager.RemoveMachineInstanceTypeAssociationOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("InstanceType: Successfully registered RemoveMachineInstanceTypeAssociationOnSite activity")

	return nil
}
