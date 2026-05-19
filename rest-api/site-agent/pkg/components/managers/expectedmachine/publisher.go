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

package expectedmachine

import (
	"github.com/google/uuid"

	swa "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	sww "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/workflow"
)

// RegisterPublisher registers ExpectedMachine inventory workflow and activity with Temporal
func (api *API) RegisterPublisher() error {
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedMachine: Registering inventory workflow and activity")

	// Register DiscoverExpectedMachineInventory workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DiscoverExpectedMachineInventory)
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedMachine: Successfully registered DiscoverExpectedMachineInventory workflow")

	// Register DiscoverExpectedMachineInventory activity
	inventoryManager := swa.NewManageExpectedMachineInventory(
		uuid.MustParse(ManagerAccess.Conf.EB.Temporal.ClusterID),
		ManagerAccess.Data.EB.Managers.NICo.Client,
		ManagerAccess.Data.EB.Managers.Workflow.Temporal.Publisher,
		ManagerAccess.Conf.EB.Temporal.TemporalPublishQueue,
		InventoryCarbidePageSize,
	)

	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(inventoryManager.DiscoverExpectedMachineInventory)
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedMachine: Successfully registered DiscoverExpectedMachineInventory activity")

	api.RegisterCron()

	return nil
}
