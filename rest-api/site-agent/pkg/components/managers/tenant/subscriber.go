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
	swa "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	sww "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/workflow"
)

// RegisterSubscriber registers Tenant CRUD workflows and activities with Temporal
func (api *API) RegisterSubscriber() error {
	ManagerAccess.Data.EB.Log.Info().Msg("Tenant: Registering CRUD workflows and activities")

	// Register workflows

	// Register CreateTenant workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateTenant)
	ManagerAccess.Data.EB.Log.Info().Msg("Tenant: Successfully registered CreateTenant workflow")

	// Register UpdateTenant workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.UpdateTenant)
	ManagerAccess.Data.EB.Log.Info().Msg("Tenant: Successfully registered UpdateTenant workflow")

	// Register activities
	tenantManager := swa.NewManageTenant(ManagerAccess.Data.EB.Managers.NICo.Client)

	// Register CreateTenantOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(tenantManager.CreateTenantOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("Tenant: Successfully registered CreateTenantOnSite activity")

	// Register UpdateTenantOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(tenantManager.UpdateTenantOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("Tenant: Successfully registered UpdateTenantOnSite activity")

	return nil
}
