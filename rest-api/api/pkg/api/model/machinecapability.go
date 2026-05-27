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
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/model/util"
	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
	cwma "github.com/NVIDIA/infra-controller-rest/workflow/pkg/activity/machine"
)

// APIMachineCapability is the datastructure to capture API representation of a MachineCapability
type APIMachineCapability struct {
	// Type is the type of the machine capability
	Type string `json:"type"`
	// Name describes the capability
	Name string `json:"name"`
	// Frequency describes the frequency of the capability
	Frequency *string `json:"frequency,omitempty"`
	// Cores describes the number of cores for the capability
	Cores *int `json:"cores,omitempty"`
	// Threads describes the number of threads for the capability
	Threads *int `json:"threads,omitempty"`
	// Capacity describes the capacity of the capability
	Capacity *string `json:"capacity,omitempty"`
	// Vendor describes the vendor of the capability
	Vendor *string `json:"vendor,omitempty"`
	// InactiveDevices describes a set of inactive devices
	InactiveDevices []int `json:"inactiveDevices,omitempty"`
	// HardwareRevision describes the hardware revision of the capability
	HardwareRevision *string `json:"hardwareRevision,omitempty"`
	// Count describes the number of items present for this capability
	Count *int `json:"count"`
	// DeviceType describes the type of the device
	DeviceType *string `json:"deviceType,omitempty"`
}

// Validate ensure the values passed in request are acceptable
func (mc APIMachineCapability) Validate() error {
	mctypes := make([]interface{}, 0)
	for mctype, _ := range cdbm.MachineCapabilityTypeChoiceMap {
		mctypes = append(mctypes, mctype)
	}

	err := validation.ValidateStruct(&mc,
		validation.Field(&mc.Type,
			validation.Required.Error("type must be specified for each Machine Capability"),
			validation.In(mctypes...).Error(fmt.Sprintf("invalid value: %v for Machine Capability type", mc.Type))),
		validation.Field(&mc.Name,
			validation.Required.Error("name must be specified for each Machine Capability"),
			validation.By(util.ValidateNameCharacters),
			validation.Length(2, 256).Error(validationErrorStringLength)),
	)

	return err
}

// NewAPIMachineCapability accepts a DB layer MachineCapability object and returns an API object
func NewAPIMachineCapability(dbmc *cdbm.MachineCapability) *APIMachineCapability {
	apimc := &APIMachineCapability{
		Type:            dbmc.Type,
		Name:            dbmc.Name,
		Frequency:       dbmc.Frequency,
		Capacity:        dbmc.Capacity,
		Vendor:          dbmc.Vendor,
		InactiveDevices: dbmc.InactiveDevices,
		Count:           dbmc.Count,
		DeviceType:      dbmc.DeviceType,
	}

	if dbmc.Type == cdbm.MachineCapabilityTypeCPU && dbmc.Info != nil {
		cores := dbmc.GetIntInfo(cwma.MachineCPUCoreCount)
		if cores != nil {
			apimc.Cores = cores
		}
		threads := dbmc.GetIntInfo(cwma.MachineCPUThreadCount)
		if threads != nil {
			apimc.Threads = threads
		}
	}

	return apimc
}
