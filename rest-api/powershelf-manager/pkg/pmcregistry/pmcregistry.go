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
package pmcregistry

import (
	"context"
	"fmt"
	"github.com/NVIDIA/infra-controller-rest/powershelf-manager/pkg/objects/pmc"
	"net"

	log "github.com/sirupsen/logrus"
)

// PmcRegistry stores non-sensitive PMC identity (MAC, IP, vendor) and supports CRUD queries.
type PmcRegistry interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	RegisterPmc(ctx context.Context, pmc *pmc.PMC) error
	IsPmcRegistered(ctx context.Context, mac net.HardwareAddr) (bool, error)
	GetPmc(ctx context.Context, mac net.HardwareAddr) (*pmc.PMC, error)
	GetAllPmcs(ctx context.Context) ([]*pmc.PMC, error)
}

// New creates a new instance of DataStore based on the given configuration.
func New(ctx context.Context, c *Config) (PmcRegistry, error) {
	switch c.DSType {
	case RegisterTypePostgres:
		if err := c.DSConf.Validate(); err != nil {
			return nil, err
		}

		log.Printf("Initializing PostGres PMC Register")
		return newPostgresRegistry(ctx, c.DSConf)
	case RegisterTypeInMemory:
		log.Printf("Initializing In-Memory PMC Register")
		return NewMemRegistry(), nil
	}

	return nil, fmt.Errorf("unsupported datastore type %s", c.DSType)
}
