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

package catalog

import (
	"cmp"
	"slices"
	"strings"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

// Descriptor describes a component manager implementation registered in this
// process. The descriptor identity is Type plus Implementation; provider names
// stay separate because one manager can require multiple providers and one
// provider can serve multiple component manager implementations.
type Descriptor struct {
	Type              devicetypes.ComponentType
	Implementation    string
	RequiredProviders []string
}

// Catalog contains the component manager implementations supported by a
// particular binary. Service-specific packages such as builtin own the list of
// descriptors that goes into a catalog.
type Catalog struct {
	descriptors map[devicetypes.ComponentType]map[string]Descriptor // type -> impl_name -> descriptor
}

// New validates descriptors and indexes them by component type and
// implementation.
func New(descriptors []Descriptor) (Catalog, error) {
	catalog := Catalog{
		descriptors: make(map[devicetypes.ComponentType]map[string]Descriptor),
	}

	for _, descriptor := range descriptors {
		d, err := descriptor.Normalize()
		if err != nil {
			return Catalog{}, err
		}

		if _, ok := catalog.descriptors[d.Type]; !ok {
			catalog.descriptors[d.Type] = make(map[string]Descriptor)
		}

		if _, exists := catalog.descriptors[d.Type][d.Implementation]; exists {
			return Catalog{}, DuplicateDescriptorError{
				ComponentType:  d.Type,
				Implementation: d.Implementation,
			}
		}

		catalog.descriptors[d.Type][d.Implementation] = d
	}

	return catalog, nil
}

// Get returns the descriptor for a component type and implementation.
func (c Catalog) Get(
	componentType devicetypes.ComponentType,
	implementation string,
) (Descriptor, bool) {
	descriptors := c.descriptors[componentType]
	if descriptors == nil {
		return Descriptor{}, false
	}

	descriptor, ok := descriptors[implementation]
	if !ok {
		return Descriptor{}, false
	}

	return descriptor.Clone(), true
}

// Implementations returns the implementations registered for a component type.
func (c Catalog) Implementations(
	componentType devicetypes.ComponentType,
) []string {
	return descriptorImplementationNames(c.descriptors[componentType])
}

// ListImplementations returns all registered implementation names by component
// type.
func (c Catalog) ListImplementations() map[devicetypes.ComponentType][]string {
	result := make(map[devicetypes.ComponentType][]string)
	for componentType, descriptors := range c.descriptors {
		result[componentType] = descriptorImplementationNames(descriptors)
	}
	return result
}

// SelectedDescriptors returns descriptors for the component managers selected
// by config.
func (c Catalog) SelectedDescriptors(
	componentManagers map[devicetypes.ComponentType]string,
) ([]Descriptor, error) {
	descriptors := make([]Descriptor, 0, len(componentManagers))
	for componentType, implName := range componentManagers {
		descriptor, ok := c.Get(componentType, implName)
		if !ok {
			available := c.Implementations(componentType)
			if len(available) == 0 {
				return nil, ComponentManagerFactoryNotRegisteredError{
					ComponentType: componentType,
				}
			}

			return nil, UnknownComponentManagerImplementationError{
				ComponentType:  componentType,
				Implementation: implName,
				Available:      available,
				RegisteredFor:  c.componentTypesForImplementation(implName),
			}
		}

		descriptors = append(descriptors, descriptor)
	}

	sortDescriptors(descriptors)
	return descriptors, nil
}

func (c Catalog) componentTypesForImplementation(
	implementation string,
) []devicetypes.ComponentType {
	types := make([]devicetypes.ComponentType, 0)
	for componentType, descriptors := range c.descriptors {
		if _, ok := descriptors[implementation]; ok {
			types = append(types, componentType)
		}
	}
	slices.Sort(types)
	return types
}

// Normalize validates a descriptor and returns its normalized value.
func (d Descriptor) Normalize() (Descriptor, error) {
	if d.Type == devicetypes.ComponentTypeUnknown {
		return Descriptor{}, UnknownComponentTypeError{
			Name: devicetypes.ComponentTypeToString(d.Type),
		}
	}

	d.Implementation = strings.TrimSpace(d.Implementation)
	if d.Implementation == "" {
		return Descriptor{}, ComponentManagerImplementationNameEmptyError{
			ComponentType: d.Type,
		}
	}

	requiredProviders := make([]string, 0, len(d.RequiredProviders))
	seen := make(map[string]struct{}, len(d.RequiredProviders))
	for _, name := range d.RequiredProviders {
		name = strings.TrimSpace(name)
		if name == "" {
			return Descriptor{}, providerapi.ErrProviderNameEmpty
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		requiredProviders = append(requiredProviders, name)
	}
	slices.Sort(requiredProviders)
	d.RequiredProviders = requiredProviders

	return d, nil
}

// Clone returns a descriptor copy whose mutable fields do not share storage
// with the source descriptor.
func (d Descriptor) Clone() Descriptor {
	d.RequiredProviders = slices.Clone(d.RequiredProviders)
	return d
}

// Equal reports whether two normalized descriptors describe the same component
// manager implementation and provider requirements.
func (d Descriptor) Equal(other Descriptor) bool {
	return d.Type == other.Type &&
		d.Implementation == other.Implementation &&
		slices.Equal(d.RequiredProviders, other.RequiredProviders)
}

func sortDescriptors(descriptors []Descriptor) {
	slices.SortFunc(descriptors, func(a, b Descriptor) int {
		if n := cmp.Compare(a.Type, b.Type); n != 0 {
			return n
		}
		return cmp.Compare(a.Implementation, b.Implementation)
	})
}

func descriptorImplementationNames(
	descriptors map[string]Descriptor,
) []string {
	names := make([]string, 0, len(descriptors))
	for name := range descriptors {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
