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

package model

import (
	"time"

	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/model/util"
	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
)

// APIMachineInterface is the data structure to capture API representation of a MachineInterface
type APIMachineInterface struct {
	// ID is the unique UUID v4 identifier for the Machine Interface
	ID string `json:"id"`
	// MachineID is the ID of the Machine
	MachineID string `json:"machineId"`
	// ControllerInterfaceID is the ID of the interface in Site Controller
	ControllerInterfaceID *string `json:"controllerInterfaceId"`
	// ControllerSegmentID is the ID of the network segment in Site Controller
	ControllerSegmentID *string `json:"controllerSegmentId"`
	// AttachedDpuID is the ID of the DPU that is attached to this interface in Site Controller
	AttachedDPUMachineID *string `json:"attachedDpuMachineID"`
	// SubnetID is the ID of the Subnet
	SubnetID *string `json:"subnetId"`
	// Hostname is the hostname of the Machine Interface
	Hostname *string `json:"hostname"`
	// IsPrimary is a boolean which indicates if the Machine Interface is primary
	IsPrimary bool `json:"isPrimary"`
	// MacAddress is the mac address of the Machine Interface
	MacAddress *string `json:"macAddress"`
	// IPAddresses is the list of ip addresses of the Machine Interface
	IPAddresses []string `json:"ipAddresses"`
	// CreatedAt indicates the ISO datetime string for when the entity was created
	Created time.Time `json:"created"`
	// UpdatedAt indicates the ISO datetime string for when the entity was last updated
	Updated time.Time `json:"updated"`
}

// NewAPIMachineInterface accepts a DB layer MachineInterface object and returns an API object
func NewAPIMachineInterface(mi *cdbm.MachineInterface, isProviderOrPrivilegedTenant bool) *APIMachineInterface {

	apimi := &APIMachineInterface{
		ID:                    mi.ID.String(),
		MachineID:             mi.MachineID,
		ControllerInterfaceID: util.GetUUIDPtrToStrPtr(mi.ControllerInterfaceID),
		ControllerSegmentID:   util.GetUUIDPtrToStrPtr(mi.ControllerSegmentID),
		SubnetID:              util.GetUUIDPtrToStrPtr(mi.SubnetID),
		Hostname:              mi.Hostname,
		IsPrimary:             mi.IsPrimary,
		MacAddress:            mi.MacAddress,
		IPAddresses:           mi.IPAddresses,
		Created:               mi.Created,
		Updated:               mi.Updated,
	}

	// Only provider admin can access it
	if isProviderOrPrivilegedTenant {
		apimi.AttachedDPUMachineID = mi.AttachedDPUMachineID
	}

	return apimi
}
