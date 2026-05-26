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
	"sync"

	log "github.com/sirupsen/logrus"
)

// MemRegistry is an in-memory implementation of PmcRegistry, safe for concurrent access.
type MemRegistry struct {
	// PMC mac to IP mapping
	registry map[string]*pmc.PMC
	mu       sync.RWMutex
}

// NewMemRegistry creates a new in-memory PMC registry.
func NewMemRegistry() *MemRegistry {
	return &MemRegistry{
		registry: make(map[string]*pmc.PMC),
	}
}

// Start MemRegistry (NO-OP)
func (ms *MemRegistry) Start(ctx context.Context) error {
	log.Printf("Starting InMem datastore")
	return nil
}

// Stop MemRegistry (NO-OP)
func (ms *MemRegistry) Stop(ctx context.Context) error {
	log.Printf("Stopping InMem datastore")
	return nil
}

// RegisterPmc creates or updates a PMC entry keyed by MAC.
func (ms *MemRegistry) RegisterPmc(ctx context.Context, pmc *pmc.PMC) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if pmc == nil {
		return fmt.Errorf("cannot register nil PMC")
	}

	ms.registry[pmc.GetMac().String()] = pmc
	return nil
}

// IsPmcRegistered checks if a PMC has been registered with the specified MAC.
func (ms *MemRegistry) IsPmcRegistered(ctx context.Context, mac net.HardwareAddr) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	_, exists := ms.registry[mac.String()]
	return exists, nil

}

// GetAllPmcs returns all registered PMCs.
func (ms *MemRegistry) GetAllPmcs(ctx context.Context) ([]*pmc.PMC, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	pmcs := make([]*pmc.PMC, 0, len(ms.registry))
	for _, pmc := range ms.registry {
		pmcs = append(pmcs, pmc)
	}

	return pmcs, nil
}

// GetPmc returns the PMC by MAC or an error if not found.
func (ms *MemRegistry) GetPmc(ctx context.Context, mac net.HardwareAddr) (*pmc.PMC, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	macStr := mac.String()
	pmc, exists := ms.registry[macStr]
	if !exists {
		return nil, fmt.Errorf("PMC (%s) is not registered", macStr)
	}

	return pmc, nil
}
