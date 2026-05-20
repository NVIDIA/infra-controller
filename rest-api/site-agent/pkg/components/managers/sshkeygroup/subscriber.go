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

package sshkeygroup

import (
	swa "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	sww "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/workflow"
)

// RegisterSubscriber registers SSHKeyGroup CRUD workflows and activities with Temporal
func (api *API) RegisterSubscriber() error {
	ManagerAccess.Data.EB.Log.Info().Msg("SSHKeyGroup: Registering CRUD workflows and activities")

	// Register workflows

	// Register CreateSSHKeyGroupV2 workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateSSHKeyGroupV2)
	ManagerAccess.Data.EB.Log.Info().Msg("SSHKeyGroup: Successfully registered CreateSSHKeyGroupV2 workflow")

	// Register UpdateSSHKeyGroupV2 workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.UpdateSSHKeyGroupV2)
	ManagerAccess.Data.EB.Log.Info().Msg("SSHKeyGroup: Successfully registered UpdateSSHKeyGroupV2 workflow")

	// Register DeleteSSHKeyGroupV2 workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DeleteSSHKeyGroupV2)
	ManagerAccess.Data.EB.Log.Info().Msg("SSHKeyGroup: Successfully registered DeleteSSHKeyGroupV2 workflow")

	// Register activities
	sshKeyGroupManager := swa.NewManageSSHKeyGroup(ManagerAccess.Data.EB.Managers.NICo.Client)

	// Register CreateSSHKeyGroupOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(sshKeyGroupManager.CreateSSHKeyGroupOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("SSHKeyGroup: Successfully registered CreateSSHKeyGroupOnSite activity")

	// Register UpdateSSHKeyGroupOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(sshKeyGroupManager.UpdateSSHKeyGroupOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("SSHKeyGroup: Successfully registered UpdateSSHKeyGroupOnSite activity")

	// Register DeleteSSHKeyGroupOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(sshKeyGroupManager.DeleteSSHKeyGroupOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("SSHKeyGroup: Successfully registered DeleteSSHKeyGroupOnSite activity")

	return nil
}
