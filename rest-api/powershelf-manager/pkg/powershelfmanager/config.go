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
package powershelfmanager

import (
	"github.com/NVIDIA/infra-controller-rest/powershelf-manager/pkg/credentials"
	"github.com/NVIDIA/infra-controller-rest/powershelf-manager/pkg/pmcregistry"
)

// DataStoreType selects between Persistent (Postgres+Vault) and InMemory backends.
type DataStoreType string

const (
	DatastoreTypePersistent DataStoreType = "Persistent"
	DatastoreTypeInMemory   DataStoreType = "InMemory"
)

// Config contains the orchestrator’s datastore mode and concrete backends for the PMC registry and the credential manager.
type Config struct {
	DSType          DataStoreType
	PmcRegistryConf pmcregistry.Config
	CredentialConf  credentials.Config
	FirmwareDir     string
}

// StringToDSType converts a string to a DataStoreType, returning false if unsupported.
func StringToDSType(s string) (DataStoreType, bool) {
	switch s {
	case string(DatastoreTypePersistent):
		return DatastoreTypePersistent, true
	case string(DatastoreTypeInMemory):
		return DatastoreTypeInMemory, true
	}

	return "", false
}
