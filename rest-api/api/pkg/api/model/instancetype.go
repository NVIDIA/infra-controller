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
	"errors"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	validationis "github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/model/util"
	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
)

// APIInstanceTypeCreateRequest is the data structure to capture user request to create a new InstanceType
type APIInstanceTypeCreateRequest struct {
	// Name is the name of the InstanceType
	Name string `json:"name"`
	// Description is the description of the Instance Type
	Description *string `json:"description"`
	// SiteID is the ID of the site
	SiteID string `json:"siteId"`
	// Labels is the labels of the Instance Type
	Labels map[string]string `json:"labels"`
	// ControllerMachineType is the Site Controller assigned Machine type
	ControllerMachineType *string `json:"controllerMachineType"`
	// MachineCapabilities is the list of Machine Capabilities to match
	MachineCapabilities []APIMachineCapability `json:"machineCapabilities"`
}

// Validate ensure the values passed in request are acceptable
func (itcr APIInstanceTypeCreateRequest) Validate() error {
	err := validation.ValidateStruct(&itcr,
		validation.Field(&itcr.Name,
			validation.Required.Error(validationErrorStringLength),
			validation.By(util.ValidateNameCharacters),
			validation.Length(2, 256).Error(validationErrorStringLength)),
		validation.Field(&itcr.ControllerMachineType,
			validation.When(itcr.ControllerMachineType != nil, validation.Length(2, 0).Error("not a valid value"))),
		validation.Field(&itcr.SiteID,
			validation.Required.Error(validationErrorValueRequired),
			validationis.UUID.Error(validationErrorInvalidUUID)),
	)

	if err != nil {
		return err
	}

	if err := util.ValidateLabels(itcr.Labels); err != nil {
		return err
	}

	if itcr.MachineCapabilities != nil {
		err = validation.Validate(itcr.MachineCapabilities)
		if err != nil {
			return err
		}

		mcNameMap := map[string]bool{}
		for _, mc := range itcr.MachineCapabilities {
			capKey := mc.Type + "-" + mc.Name

			if !cdbm.MachineCapabilityTypeChoiceMap[mc.Type] {
				return validation.Errors{
					"machineCapabilities": errors.New("requested Capability type is not valid: " + mc.Type),
				}
			}

			// Validate device type for network capabilities
			if mc.Type == cdbm.MachineCapabilityTypeNetwork && mc.DeviceType != nil {
				if !cdbm.MachineCapabilityDeviceTypeChoiceMap[*mc.DeviceType] {
					return validation.Errors{
						"machineCapabilities": errors.New("requested device type  `" + *mc.DeviceType + "`  for Capability type `" + mc.Type + "` is not valid"),
					}
				}
			}

			_, found := mcNameMap[capKey]
			if found {
				return validation.Errors{
					"machineCapabilities": fmt.Errorf("requested Capability type `%s` cannot contain duplicate Capability name: %s", mc.Type, mc.Name),
				}
			}
			mcNameMap[capKey] = true
		}

	}

	return nil
}

// APIInstanceTypeUpdateRequest is the data structure to capture user request to update an Instance Type
type APIInstanceTypeUpdateRequest struct {
	// Name is the name of the Instance Type
	Name *string `json:"name"`
	// Description is the description of the Instance Type
	Description *string `json:"description"`
	// Labels is the labels of the Instance Type
	Labels map[string]string `json:"labels"`
	// MachineCapabilities is the list of Machine Capabilities to match
	MachineCapabilities []APIMachineCapability `json:"machineCapabilities"`
}

// Validate ensure the values passed in request are acceptable
func (itur APIInstanceTypeUpdateRequest) Validate() error {
	err := validation.ValidateStruct(&itur,
		validation.Field(&itur.Name,
			// length validation rule accepts empty string as valid, hence, required is needed
			validation.When(itur.Name != nil, validation.Required.Error(validationErrorStringLength)),
			validation.When(itur.Name != nil, validation.By(util.ValidateNameCharacters)),
			validation.When(itur.Name != nil, validation.Length(2, 256).Error(validationErrorStringLength))),
	)
	if err != nil {
		return err
	}

	if err := util.ValidateLabels(itur.Labels); err != nil {
		return err
	}

	if itur.MachineCapabilities != nil {
		err = validation.Validate(itur.MachineCapabilities)
		if err != nil {
			return err
		}

		mcNameMap := map[string]bool{}
		for _, mc := range itur.MachineCapabilities {

			capKey := mc.Type + "-" + mc.Name

			if !cdbm.MachineCapabilityTypeChoiceMap[mc.Type] {
				return validation.Errors{
					"machineCapabilities": errors.New("requested Capability type is not valid: " + mc.Type),
				}
			}

			// Validate device type for network capabilities
			if mc.Type == cdbm.MachineCapabilityTypeNetwork && mc.DeviceType != nil {
				if !cdbm.MachineCapabilityDeviceTypeChoiceMap[*mc.DeviceType] {
					return validation.Errors{
						"machineCapabilities": errors.New("requested device type  `" + *mc.DeviceType + "`  for Capability type `" + mc.Type + "` is not valid"),
					}
				}
			}

			_, found := mcNameMap[capKey]
			if found {
				return validation.Errors{
					"machineCapabilities": fmt.Errorf("requested Capability type `%s` cannot contain duplicate Capability name: %s", mc.Type, mc.Name),
				}
			}
			mcNameMap[capKey] = true
		}
	}

	return nil
}

// APIInstanceType is the data structure to capture API representation of an Instance Type
type APIInstanceType struct {
	// ID is the unique UUID v4 identifier for the Instance Type
	ID string `json:"id"`
	// Name is the name of the Instance Type
	Name string `json:"name"`
	// Description is the description of the Instance Type
	Description *string `json:"description"`
	// ControllerMachineType is the Machine type assigned by Site Controller
	ControllerMachineType *string `json:"controllerMachineType"`
	// InfrastructureProviderID is the ID of the InfrastructureProvider that owns the Instance Type
	InfrastructureProviderID string `json:"infrastructureProviderId"`
	// InfrastructureProvider is the summary of the InfrastructureProvider
	InfrastructureProvider *APIInfrastructureProviderSummary `json:"infrastructureProvider,omitempty"`
	// SiteID is the ID of the Site that owns the Instance Type
	SiteID string `json:"siteId"`
	// Site is the summary of the Site
	Site *APISiteSummary `json:"site,omitempty"`
	// Labels is the labels of the Instance Type
	Labels map[string]string `json:"labels"`
	// MachineCapabilities is the list of capabilities that are supported by the Machine's of this Instance Type
	MachineCapabilities []APIMachineCapability `json:"machineCapabilities"`
	// MachineInstanceTypes is the list of machines that are associated to this Instance Type
	MachineInstanceTypes []APIMachineInstanceType `json:"machineInstanceTypes,omitempty"`
	// AllocationStats is the stats of allocation that are associated to this Instance Type
	AllocationStats *APIAllocationStats `json:"allocationStats,omitempty"`
	// Deprecations is the list of deprecation messages denoting fields which are being deprecated
	Deprecations []APIDeprecation `json:"deprecations,omitempty"`
	// Status is the status of the Instance Type
	Status string `json:"status"`
	// StatusHistory is the history of statuses for the Instance Type
	StatusHistory []APIStatusDetail `json:"statusHistory"`
	// Created is the date and time the entity was created
	Created time.Time `json:"created"`
	// Updated is the date and time the entity was last updated
	Updated time.Time `json:"updated"`
}

// NewAPIInstanceType accepts a DB layer Instance Type object returns an API layer object
func NewAPIInstanceType(dbit *cdbm.InstanceType, dbsds []cdbm.StatusDetail, mcs []cdbm.MachineCapability, mit []cdbm.MachineInstanceType, aas *APIAllocationStats) *APIInstanceType {
	if dbit == nil {
		return nil
	}

	apiit := &APIInstanceType{
		ID:                       dbit.ID.String(),
		Name:                     dbit.Name,
		Description:              dbit.Description,
		ControllerMachineType:    dbit.ControllerMachineType,
		InfrastructureProviderID: dbit.InfrastructureProviderID.String(),
		SiteID:                   dbit.SiteID.String(),
		Labels:                   dbit.Labels,
		Status:                   dbit.Status,
		Created:                  dbit.Created,
		Updated:                  dbit.Updated,
	}

	apiit.AllocationStats = aas

	if dbit.InfrastructureProvider != nil {
		apiit.InfrastructureProvider = NewAPIInfrastructureProviderSummary(dbit.InfrastructureProvider)
	}

	if dbit.Site != nil {
		apiit.Site = NewAPISiteSummary(dbit.Site)
	}

	apiit.StatusHistory = []APIStatusDetail{}
	for _, dbsd := range dbsds {
		apiit.StatusHistory = append(apiit.StatusHistory, NewAPIStatusDetail(dbsd))
	}

	apiit.MachineCapabilities = []APIMachineCapability{}
	for _, mc := range mcs {
		cmc := mc
		apiit.MachineCapabilities = append(apiit.MachineCapabilities, *NewAPIMachineCapability(&cmc))
	}

	apiit.MachineInstanceTypes = []APIMachineInstanceType{}
	for _, mi := range mit {
		cmi := mi
		apiit.MachineInstanceTypes = append(apiit.MachineInstanceTypes, *NewAPIMachineInstanceType(&cmi))
	}

	return apiit
}

// APIInstanceTypeSummary is the data structure to capture summary of an Instance Type
type APIInstanceTypeSummary struct {
	// ID of the Instance Type
	ID string `json:"id"`
	// Name of the InstanceType, only lowercase characters, digits, hyphens and cannot begin/end with hyphen
	Name string `json:"name"`
	// InfrastructureProviderID is the ID of the InfrastructureProvider that owns the Instance Type
	InfrastructureProviderID string `json:"infrastructureProviderId"`
	// SiteID is the ID of the Site that owns the Instance Type
	SiteID string `json:"siteId"`
	// Status is the status of the Instance Type
	Status string `json:"status"`
}

// NewAPIInstanceTypeSummary accepts a DB layer Instance object returns an API layer summary object
func NewAPIInstanceTypeSummary(dbist *cdbm.InstanceType) *APIInstanceTypeSummary {
	inst := APIInstanceTypeSummary{
		ID:                       dbist.ID.String(),
		Name:                     dbist.Name,
		InfrastructureProviderID: dbist.InfrastructureProviderID.String(),
		SiteID:                   dbist.SiteID.String(),
		Status:                   dbist.Status,
	}

	return &inst
}

// APIAllocationStats is the data structure to capture API representation of an InstanceType allocation stats
type APIAllocationStats struct {
	// Assigned is the total number of Machines assigned to this Instance Type
	Assigned int `json:"assigned"`
	// Total is the total number of Machines allocated to different Tenants for this Instance Type
	Total int `json:"total"`
	// Used is the total number of allocated Machines of this Instance Type currently being used by Tenants
	Used int `json:"used"`
	// Unused is the total number of allocated Machines of this Instance Type that is currently not being used by Tenants
	Unused int `json:"unused"`
	// UnusedUsable is the total number of allocated Machines of this Instance Type that is currently not in use
	// but in Ready state, therefore can be provisioned by Tenant
	UnusedUsable int `json:"unusedUsable"`
	// MaxAllocatable is the maximum number of Machines of this Instance Type that can be allocated to a Tenant
	MaxAllocatable *int `json:"maxAllocatable,omitempty"`
}
