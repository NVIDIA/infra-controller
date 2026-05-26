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
package credentials

import (
	"context"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/NVIDIA/infra-controller-rest/common/pkg/credential"
)

// CredentialManager defines a key-value store for PMC credentials keyed by MAC address.
type CredentialManager interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Get(ctx context.Context, mac net.HardwareAddr) (*credential.Credential, error)
	Put(ctx context.Context, mac net.HardwareAddr, credentials *credential.Credential) error
	Patch(ctx context.Context, mac net.HardwareAddr, credentials *credential.Credential) error
	Delete(ctx context.Context, mac net.HardwareAddr) error
	Keys(ctx context.Context) ([]net.HardwareAddr, error)
}

// New creates a new Credential Manager based on the given configuration.
func New(ctx context.Context, config *Config) (CredentialManager, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	switch config.DataStoreType {
	case DatastoreTypeVault:
		log.Printf("Initializing CredentialManager with vault datastore (config: %s)", config.VaultConfig)
		return config.VaultConfig.NewManager()
	case DatastoreTypeInMemory:
		log.Printf("Initializing CredentialManager with in-memory datastore")
		return NewInMemoryCredentialManager(), nil
	}

	return nil, fmt.Errorf("unsupported datastore type %s", config.DataStoreType)
}
