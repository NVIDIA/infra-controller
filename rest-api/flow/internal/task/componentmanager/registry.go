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

package componentmanager

import (
	"sync"

	cmcatalog "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/catalog"
	cmconfig "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/config"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

// Registry maintains the active component managers selected from factory specs.
type Registry struct {
	mu     sync.RWMutex
	active map[devicetypes.ComponentType]ComponentManager
}

// NewRegistry creates and initializes a Registry from the supplied factory
// specs and component manager configuration.
func NewRegistry(
	factorySpecs []FactorySpec,
	config cmconfig.Config,
	providers *providerapi.ProviderRegistry,
) (*Registry, error) {
	descriptors, factories, err := selectFactorySpecs(
		factorySpecs,
		config.ComponentManagers,
	)
	if err != nil {
		return nil, err
	}

	registry := Registry{
		active: make(
			map[devicetypes.ComponentType]ComponentManager,
			len(config.ComponentManagers),
		),
	}

	for _, descriptor := range descriptors {
		factory, ok := factories[descriptorKeyOf(descriptor)]
		if !ok {
			return nil, ComponentManagerFactoryNotRegisteredError{
				ComponentType: descriptor.Type,
			}
		}

		manager, err := createManager(descriptor, factory, providers)
		if err != nil {
			return nil, err
		}

		registry.active[descriptor.Type] = manager
	}

	return &registry, nil
}

func createManager(
	descriptor cmcatalog.Descriptor,
	factory ManagerFactory,
	providers *providerapi.ProviderRegistry,
) (ComponentManager, error) {
	manager, err := factory(providers)
	if err != nil {
		return nil, ManagerCreationError{
			ComponentType:  descriptor.Type,
			Implementation: descriptor.Implementation,
			Err:            err,
		}
	}
	if manager == nil {
		return nil, ManagerNotCreatedError{
			ComponentType:  descriptor.Type,
			Implementation: descriptor.Implementation,
		}
	}

	managerDescriptor, err := manager.Descriptor().Normalize()
	if err != nil {
		return nil, err
	}

	if !descriptor.Equal(managerDescriptor) {
		return nil, ManagerDescriptorMismatchError{
			Expected: descriptor,
			Actual:   managerDescriptor,
		}
	}

	return manager, nil
}

// FindManager returns the active manager for the specified component type.
// It returns nil when the registry is nil or when no manager is active for the
// type. Use GetManager when the caller needs a descriptive configuration error.
func (r *Registry) FindManager(
	componentType devicetypes.ComponentType,
) ComponentManager {
	if r == nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.active[componentType]
}

// GetManager returns the active manager for the specified component type.
// It returns a descriptive error when the registry is nil or when no manager is
// active for the type.
func (r *Registry) GetManager(
	componentType devicetypes.ComponentType,
) (ComponentManager, error) {
	if r == nil {
		return nil, ErrRegistryNotConfigured
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	manager := r.active[componentType]
	if manager == nil {
		return nil, ManagerNotConfiguredError{ComponentType: componentType}
	}

	return manager, nil
}

// GetDescriptor returns the descriptor reported by the active manager for the
// specified component type.
func (r *Registry) GetDescriptor(
	componentType devicetypes.ComponentType,
) (cmcatalog.Descriptor, error) {
	if r == nil {
		return cmcatalog.Descriptor{}, ErrRegistryNotConfigured
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	manager := r.active[componentType]
	if manager == nil {
		return cmcatalog.Descriptor{}, ManagerNotConfiguredError{ComponentType: componentType}
	}

	return manager.Descriptor().Normalize()
}

// GetAllManagers returns all active managers.
func (r *Registry) GetAllManagers() []ComponentManager {
	if r == nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	managers := make([]ComponentManager, 0, len(r.active))
	for _, manager := range r.active {
		managers = append(managers, manager)
	}
	return managers
}
