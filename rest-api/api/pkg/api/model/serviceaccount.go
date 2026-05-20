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
	cdb "github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
)

// type APIServiceAccount is the data structure to capture API representation of a Service Account
type APIServiceAccount struct {
	// Enabled is a flag to indicate if the Service Account is enabled
	Enabled bool `json:"enabled"`
	// InfrastructureProviderID is the ID of the InfrastructureProvider
	InfrastructureProviderID *string `json:"infrastructureProviderId"`
	// ID is the unique UUID v4 identifier for the Service Account
	TenantID *string `json:"tenantId"`
}

// NewAPIServiceAccount accepts a DB layer ServiceAccount object and returns an API object
func NewAPIServiceAccount(serviceAccountEnabled bool, dbProvider *cdbm.InfrastructureProvider, dbTenant *cdbm.Tenant) *APIServiceAccount {
	apiServiceAccount := APIServiceAccount{
		Enabled: serviceAccountEnabled,
	}

	if dbProvider != nil {
		apiServiceAccount.InfrastructureProviderID = cdb.GetStrPtr(dbProvider.ID.String())
	}
	if dbTenant != nil {
		apiServiceAccount.TenantID = cdb.GetStrPtr(dbTenant.ID.String())
	}

	return &apiServiceAccount
}
