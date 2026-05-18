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

package networksecuritygroup

import (
	swa "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	sww "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/workflow"
)

// RegisterSubscriber registers NetworkSecurityGroup CRUD workflows and activities with Temporal
func (api *API) RegisterSubscriber() error {
	ManagerAccess.Data.EB.Log.Info().Msg("NetworkSecurityGroup: Registering CRUD workflows and activities")

	// Register workflows

	// Register CreateNetworkSecurityGroup workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateNetworkSecurityGroup)
	ManagerAccess.Data.EB.Log.Info().Msg("NetworkSecurityGroup: Successfully registered CreateNetworkSecurityGroup workflow")

	// Register UpdateNetworkSecurityGroup workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.UpdateNetworkSecurityGroup)
	ManagerAccess.Data.EB.Log.Info().Msg("NetworkSecurityGroup: Successfully registered UpdateNetworkSecurityGroup workflow")

	// Register DeleteNetworkSecurityGroup workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DeleteNetworkSecurityGroup)
	ManagerAccess.Data.EB.Log.Info().Msg("NetworkSecurityGroup: Successfully registered DeleteNetworkSecurityGroup workflow")

	// Register activities
	networkSecurityGroupManager := swa.NewManageNetworkSecurityGroup(ManagerAccess.Data.EB.Managers.NICo.Client)

	// Register CreateNetworkSecurityGroupOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(networkSecurityGroupManager.CreateNetworkSecurityGroupOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("NetworkSecurityGroup: Successfully registered CreateNetworkSecurityGroupOnSite activity")

	// Register UpdateNetworkSecurityGroupOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(networkSecurityGroupManager.UpdateNetworkSecurityGroupOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("NetworkSecurityGroup: Successfully registered UpdateNetworkSecurityGroupOnSite activity")

	// Register DeleteNetworkSecurityGroupOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(networkSecurityGroupManager.DeleteNetworkSecurityGroupOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("NetworkSecurityGroup: Successfully registered DeleteNetworkSecurityGroupOnSite activity")

	return nil
}
