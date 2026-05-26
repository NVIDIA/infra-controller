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
	swa "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	sww "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/workflow"
)

// RegisterSubscriber registers VPC CRUD workflows and activities with Temporal
func (api *API) RegisterSubscriber() error {
	ManagerAccess.Data.EB.Log.Info().Msg("VPC: Registering CRUD workflows and activities")

	// Register workflows

	// Register CreateVPCV2 workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateVPCV2)
	ManagerAccess.Data.EB.Log.Info().Msg("VPC: Successfully registered CreateVPCV2 workflow")

	// Register UpdateVPC workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.UpdateVPC)
	ManagerAccess.Data.EB.Log.Info().Msg("VPC: Successfully registered UpdateVPC workflow")

	// Register UpdateVPCVirtualization workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.UpdateVPCVirtualization)
	ManagerAccess.Data.EB.Log.Info().Msg("VPC: Successfully registered UpdateVPCVirtualization workflow")

	// Register DeleteVPCV2 workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DeleteVPCV2)
	ManagerAccess.Data.EB.Log.Info().Msg("VPC: Successfully registered DeleteVPCV2 workflow")

	// Register activities
	vpcManager := swa.NewManageVPC(ManagerAccess.Data.EB.Managers.NICo.Client)

	// Register CreateVpcOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(vpcManager.CreateVpcOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("VPC: Successfully registered CreateVpcOnSite activity")

	// Register UpdateVpcOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(vpcManager.UpdateVpcOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("VPC: Successfully registered UpdateVpcOnSite activity")

	// Register UpdateVpcVirtualizationOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(vpcManager.UpdateVpcVirtualizationOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("VPC: Successfully registered UpdateVpcVirtualizationOnSite activity")

	// Register DeleteVpcOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(vpcManager.DeleteVpcOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("VPC: Successfully registered DeleteVpcOnSite activity")

	return nil
}
