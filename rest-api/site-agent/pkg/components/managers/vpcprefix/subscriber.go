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

package vpcprefix

import (
	swa "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	sww "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/workflow"
)

// RegisterSubscriber registers VpcPrefix CRUD workflows and activities with Temporal
func (api *API) RegisterSubscriber() error {
	ManagerAccess.Data.EB.Log.Info().Msg("VpcPrefix: Registering CRUD workflows and activities")

	// Register workflows

	// Register CreateVpcPrefix workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateVpcPrefix)
	ManagerAccess.Data.EB.Log.Info().Msg("VpcPrefix: Successfully registered CreateVpcPrefix workflow")

	// Register UpdateVpcPrefix workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.UpdateVpcPrefix)
	ManagerAccess.Data.EB.Log.Info().Msg("VpcPrefix: Successfully registered UpdateVpcPrefix workflow")

	// Register DeleteVpcPrefix workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DeleteVpcPrefix)
	ManagerAccess.Data.EB.Log.Info().Msg("VpcPrefix: Successfully registered DeleteVpcPrefix workflow")

	// Register activities
	vpcPrefixManager := swa.NewManageVpcPrefix(ManagerAccess.Data.EB.Managers.NICo.Client)

	// Register CreateVpcPrefixOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(vpcPrefixManager.CreateVpcPrefixOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("VpcPrefix: Successfully registered CreateVpcPrefixOnSite activity")

	// Register UpdateVpcPrefixOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(vpcPrefixManager.UpdateVpcPrefixOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("VpcPrefix: Successfully registered UpdateVpcPrefixOnSite activity")

	// Register DeleteVpcPrefixOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(vpcPrefixManager.DeleteVpcPrefixOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("VpcPrefix: Successfully registered DeleteVpcPrefixOnSite activity")

	return nil
}
