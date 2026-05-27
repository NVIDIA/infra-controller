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

package vpcpeering

import (
	swa "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	sww "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/workflow"
)

// RegisterSubscriber registers VPC Peering CRUD workflows and activities with Temporal
func (api *API) RegisterSubscriber() error {
	ManagerAccess.Data.EB.Log.Info().Msg("VpcPeering: Registering CRUD workflows and activities")

	// Register workflows
	// Register CreateVpcPeering workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateVpcPeering)
	ManagerAccess.Data.EB.Log.Info().Msg("VpcPeering: Successfully registered CreateVpcPeering workflow")

	// Register DeleteVpcPeering workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DeleteVpcPeering)
	ManagerAccess.Data.EB.Log.Info().Msg("VpcPeering: Successfully registered DeleteVpcPeering workflow")

	// Register activities
	vpcPeeringManager := swa.NewManageVpcPeering(ManagerAccess.Data.EB.Managers.NICo.Client)

	// Register CreateVpcPeeringOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(vpcPeeringManager.CreateVpcPeeringOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("VpcPeering: Successfully registered CreateVpcPeeringOnSite activity")

	// Register DeleteVpcPeeringOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(vpcPeeringManager.DeleteVpcPeeringOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("VpcPeering: Successfully registered DeleteVpcPeeringOnSite activity")

	return nil
}
