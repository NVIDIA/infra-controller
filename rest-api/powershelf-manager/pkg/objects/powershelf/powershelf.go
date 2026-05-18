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
package powershelf

import (
	"github.com/NVIDIA/infra-controller-rest/powershelf-manager/pkg/objects/pmc"
	"github.com/NVIDIA/infra-controller-rest/powershelf-manager/pkg/objects/powersupply"

	gofish "github.com/stmcginnis/gofish/redfish"
)

// PowerShelf is a snapshot of a powershelf. Consists of the power-shelf's PMC and its Redfish-exposed chassis, manager, and power supplies.
type PowerShelf struct {
	PMC           *pmc.PMC
	Chassis       *gofish.Chassis
	Manager       *gofish.Manager
	PowerSupplies []*powersupply.PowerSupply
}

type Component string

const (
	PMC Component = "PMC"
	PSU Component = "PSU"
)

// FirmwareState represents the overall state of a firmware operation.
type FirmwareState string

const (
	FirmwareStateQueued    FirmwareState = "Queued"
	FirmwareStateVerifying FirmwareState = "Verifying"
	FirmwareStateCompleted FirmwareState = "Completed"
	FirmwareStateFailed    FirmwareState = "Failed"
)

type FirmwareUpdate struct {
	PmcMacAddress string
	Component     Component
	VersionFrom   string
	VersionTo     string
	State         FirmwareState
	JobID         string
	ErrorMessage  string
}
