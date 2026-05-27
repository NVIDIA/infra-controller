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

package infinibandpartition

import (
	swa "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	sww "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/workflow"
)

// RegisterSubscriber registers InfiniBandPartition CRUD workflows and activities with Temporal
func (api *API) RegisterSubscriber() error {
	ManagerAccess.Data.EB.Log.Info().Msg("InfiniBandPartition: Registering CRUD workflows and activities")

	// Register workflows

	// Register CreateInfiniBandPartitionV2 workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateInfiniBandPartitionV2)
	ManagerAccess.Data.EB.Log.Info().Msg("InfiniBandPartition: Successfully registered CreateInfiniBandPartitionV2 workflow")

	// UpdateInfiniBandPartition
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.UpdateInfiniBandPartition)
	ManagerAccess.Data.EB.Log.Info().Msg("InfiniBandPartition: successfully registered UpdateInfiniBandPartition workflow")

	// Register DeleteInfiniBandPartitionV2 workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DeleteInfiniBandPartitionV2)
	ManagerAccess.Data.EB.Log.Info().Msg("InfiniBandPartition: Successfully registered DeleteInfiniBandPartitionV2 workflow")

	// Register activities
	ibpManager := swa.NewManageInfiniBandPartition(ManagerAccess.Data.EB.Managers.NICo.Client)

	// Register CreateInfiniBandPartitionOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(ibpManager.CreateInfiniBandPartitionOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("InfiniBandPartition: Successfully registered CreateInfiniBandPartitionOnSite activity")

	// Register UpdateInfiniBandPartitionOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(ibpManager.UpdateInfiniBandPartitionOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("InfiniBandPartition: Successfully registered UpdateInfiniBandPartitionOnSite activity")

	// Register DeleteInfiniBandPartitionOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(ibpManager.DeleteInfiniBandPartitionOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("InfiniBandPartition: Successfully registered DeleteInfiniBandPartitionOnSite activity")

	return nil
}
