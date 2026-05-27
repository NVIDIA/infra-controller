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

package expectedrack

import (
	swa "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/activity"
	sww "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/workflow"
)

// RegisterSubscriber registers ExpectedRack CRUD workflows and activities with Temporal
func (api *API) RegisterSubscriber() error {
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedRack: Registering CRUD workflows and activities")

	// Register workflows

	// Register CreateExpectedRack workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateExpectedRack)
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedRack: Successfully registered CreateExpectedRack workflow")

	// Register UpdateExpectedRack workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.UpdateExpectedRack)
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedRack: Successfully registered UpdateExpectedRack workflow")

	// Register DeleteExpectedRack workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DeleteExpectedRack)
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedRack: Successfully registered DeleteExpectedRack workflow")

	// Register ReplaceAllExpectedRacks workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.ReplaceAllExpectedRacks)
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedRack: Successfully registered ReplaceAllExpectedRacks workflow")

	// Register DeleteAllExpectedRacks workflow
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DeleteAllExpectedRacks)
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedRack: Successfully registered DeleteAllExpectedRacks workflow")

	// Register activities
	expectedRackManager := swa.NewManageExpectedRack(ManagerAccess.Data.EB.Managers.NICo.Client, ManagerAccess.Data.EB.Managers.Flow.Client)

	// Register CreateExpectedRackOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(expectedRackManager.CreateExpectedRackOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedRack: Successfully registered CreateExpectedRackOnSite activity")

	// Register CreateExpectedRackOnFlow activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(expectedRackManager.CreateExpectedRackOnFlow)
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedRack: Successfully registered CreateExpectedRackOnFlow activity")

	// Register UpdateExpectedRackOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(expectedRackManager.UpdateExpectedRackOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedRack: Successfully registered UpdateExpectedRackOnSite activity")

	// Register DeleteExpectedRackOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(expectedRackManager.DeleteExpectedRackOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedRack: Successfully registered DeleteExpectedRackOnSite activity")

	// Register ReplaceAllExpectedRacksOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(expectedRackManager.ReplaceAllExpectedRacksOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedRack: Successfully registered ReplaceAllExpectedRacksOnSite activity")

	// Register DeleteAllExpectedRacksOnSite activity
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(expectedRackManager.DeleteAllExpectedRacksOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("ExpectedRack: Successfully registered DeleteAllExpectedRacksOnSite activity")

	return nil
}
