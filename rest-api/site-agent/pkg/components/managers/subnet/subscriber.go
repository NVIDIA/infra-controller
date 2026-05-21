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

package subnet

import (
	swa "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	sww "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/workflow"
)

// RegisterSubscriber registers Subnet CRUD workflows and activities with Temporal
func (api *API) RegisterSubscriber() error {
	ManagerAccess.Data.EB.Log.Info().Msg("Subnet: Registering CRUD workflows and activities")

	// Register workflows

	// Register CreateSubnetV2 workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateSubnetV2)
	ManagerAccess.Data.EB.Log.Info().Msg("Subnet: Successfully registered CreateSubnetV2 workflow")

	// Register DeleteSubnetV2 workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DeleteSubnetV2)
	ManagerAccess.Data.EB.Log.Info().Msg("Subnet: Successfully registered DeleteSubnetV2 workflow")

	// Register activities

	subnetManager := swa.NewManageSubnet(ManagerAccess.Data.EB.Managers.NICo.Client)

	// Register CreateSubnetOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(subnetManager.CreateSubnetOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("Subnet: Successfully registered CreateSubnetOnSite activity")

	// Register DeleteSubnetOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(subnetManager.DeleteSubnetOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("Subnet: Successfully registered DeleteSubnetOnSite activity")

	return nil
}
