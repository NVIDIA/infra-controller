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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cmcatalog "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/catalog"
	cmconfig "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/config"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

func TestRegistryGetManager(t *testing.T) {
	t.Run("nil registry", func(t *testing.T) {
		var registry *Registry

		manager, err := registry.GetManager(devicetypes.ComponentTypeCompute)

		require.Nil(t, manager)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrRegistryNotConfigured))
	})

	t.Run("missing active manager", func(t *testing.T) {
		registry, err := NewRegistry(
			nil,
			cmconfig.Config{},
			providerapi.NewProviderRegistry(),
		)
		require.NoError(t, err)

		manager, err := registry.GetManager(devicetypes.ComponentTypeCompute)

		require.Nil(t, manager)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrManagerNotConfigured))

		var managerErr ManagerNotConfiguredError
		require.True(t, errors.As(err, &managerErr))
		require.Equal(t, devicetypes.ComponentTypeCompute, managerErr.ComponentType)
	})
}

func TestRegistryGetDescriptor(t *testing.T) {
	registry, err := NewRegistry(
		[]FactorySpec{
			testFactorySpec(
				devicetypes.ComponentTypeCompute,
				"custom",
				managerFactory(devicetypes.ComponentTypeCompute, "custom"),
			),
		},
		cmconfig.Config{
			ComponentManagers: map[devicetypes.ComponentType]string{
				devicetypes.ComponentTypeCompute: "custom",
			},
		},
		providerapi.NewProviderRegistry(),
	)
	require.NoError(t, err)

	descriptor, err := registry.GetDescriptor(devicetypes.ComponentTypeCompute)

	require.NoError(t, err)
	require.Equal(t, devicetypes.ComponentTypeCompute, descriptor.Type)
	require.Equal(t, "custom", descriptor.Implementation)
}

func TestRegistryGetDescriptorErrors(t *testing.T) {
	t.Run("nil registry", func(t *testing.T) {
		var registry *Registry

		descriptor, err := registry.GetDescriptor(devicetypes.ComponentTypeCompute)

		require.Equal(t, cmcatalog.Descriptor{}, descriptor)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrRegistryNotConfigured))
	})

	t.Run("missing active manager", func(t *testing.T) {
		registry, err := NewRegistry(
			nil,
			cmconfig.Config{},
			providerapi.NewProviderRegistry(),
		)
		require.NoError(t, err)

		descriptor, err := registry.GetDescriptor(devicetypes.ComponentTypeCompute)

		require.Equal(t, cmcatalog.Descriptor{}, descriptor)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrManagerNotConfigured))
	})
}

func TestNewRegistryErrors(t *testing.T) {
	t.Run("factory not registered", func(t *testing.T) {
		_, err := NewRegistry(
			nil,
			cmconfig.Config{
				ComponentManagers: map[devicetypes.ComponentType]string{
					devicetypes.ComponentTypeCompute: "mock",
				},
			},
			providerapi.NewProviderRegistry(),
		)

		require.Error(t, err)
		require.True(t, errors.Is(err, ErrComponentManagerFactoryNotRegistered))

		var factoryErr ComponentManagerFactoryNotRegisteredError
		require.True(t, errors.As(err, &factoryErr))
		require.Equal(t, devicetypes.ComponentTypeCompute, factoryErr.ComponentType)
	})

	t.Run("factory not configured", func(t *testing.T) {
		_, err := NewRegistry(
			[]FactorySpec{
				{
					Descriptor: testDescriptor(
						devicetypes.ComponentTypeCompute,
						"custom",
					),
				},
			},
			cmconfig.Config{
				ComponentManagers: map[devicetypes.ComponentType]string{
					devicetypes.ComponentTypeCompute: "custom",
				},
			},
			providerapi.NewProviderRegistry(),
		)

		require.Error(t, err)
		require.True(t, errors.Is(err, ErrComponentManagerFactoryNotConfigured))

		var factoryErr ComponentManagerFactoryNotConfiguredError
		require.True(t, errors.As(err, &factoryErr))
		require.Equal(t, devicetypes.ComponentTypeCompute, factoryErr.ComponentType)
		require.Equal(t, "custom", factoryErr.Implementation)
	})

	t.Run("unknown implementation", func(t *testing.T) {
		_, err := NewRegistry(
			[]FactorySpec{
				testFactorySpec(
					devicetypes.ComponentTypeCompute,
					"known",
					managerFactory(devicetypes.ComponentTypeCompute, "known"),
				),
			},
			cmconfig.Config{
				ComponentManagers: map[devicetypes.ComponentType]string{
					devicetypes.ComponentTypeCompute: "missing",
				},
			},
			providerapi.NewProviderRegistry(),
		)

		require.Error(t, err)
		require.True(t, errors.Is(err, ErrUnknownComponentManagerImplementation))

		var implErr UnknownComponentManagerImplementationError
		require.True(t, errors.As(err, &implErr))
		require.Equal(t, devicetypes.ComponentTypeCompute, implErr.ComponentType)
		require.Equal(t, "missing", implErr.Implementation)
		require.ElementsMatch(t, []string{"known"}, implErr.Available)
	})

	t.Run("implementation registered for another type", func(t *testing.T) {
		_, err := NewRegistry(
			[]FactorySpec{
				testFactorySpec(
					devicetypes.ComponentTypeCompute,
					"nico",
					managerFactory(devicetypes.ComponentTypeCompute, "nico"),
				),
				testFactorySpec(
					devicetypes.ComponentTypeNVLSwitch,
					"nvswitchmanager",
					managerFactory(devicetypes.ComponentTypeNVLSwitch, "nvswitchmanager"),
				),
			},
			cmconfig.Config{
				ComponentManagers: map[devicetypes.ComponentType]string{
					devicetypes.ComponentTypeCompute: "nvswitchmanager",
				},
			},
			providerapi.NewProviderRegistry(),
		)

		require.Error(t, err)
		require.True(t, errors.Is(err, ErrUnknownComponentManagerImplementation))

		var implErr UnknownComponentManagerImplementationError
		require.True(t, errors.As(err, &implErr))
		require.Equal(t, devicetypes.ComponentTypeCompute, implErr.ComponentType)
		require.Equal(t, "nvswitchmanager", implErr.Implementation)
		require.Equal(t, []string{"nico"}, implErr.Available)
		require.Equal(t, []devicetypes.ComponentType{
			devicetypes.ComponentTypeNVLSwitch,
		}, implErr.RegisteredFor)
	})

	t.Run("manager creation failed", func(t *testing.T) {
		rootErr := errors.New("boom")

		_, err := NewRegistry(
			[]FactorySpec{
				testFactorySpec(
					devicetypes.ComponentTypeCompute,
					"broken",
					func(*providerapi.ProviderRegistry) (ComponentManager, error) {
						return nil, rootErr
					},
				),
			},
			cmconfig.Config{
				ComponentManagers: map[devicetypes.ComponentType]string{
					devicetypes.ComponentTypeCompute: "broken",
				},
			},
			providerapi.NewProviderRegistry(),
		)

		require.Error(t, err)
		require.True(t, errors.Is(err, ErrManagerCreationFailed))
		require.True(t, errors.Is(err, rootErr))

		var creationErr ManagerCreationError
		require.True(t, errors.As(err, &creationErr))
		require.Equal(t, devicetypes.ComponentTypeCompute, creationErr.ComponentType)
		require.Equal(t, "broken", creationErr.Implementation)
	})

	t.Run("manager not created", func(t *testing.T) {
		_, err := NewRegistry(
			[]FactorySpec{
				testFactorySpec(
					devicetypes.ComponentTypeCompute,
					"nil-manager",
					func(*providerapi.ProviderRegistry) (ComponentManager, error) {
						return nil, nil
					},
				),
			},
			cmconfig.Config{
				ComponentManagers: map[devicetypes.ComponentType]string{
					devicetypes.ComponentTypeCompute: "nil-manager",
				},
			},
			providerapi.NewProviderRegistry(),
		)

		require.Error(t, err)
		require.True(t, errors.Is(err, ErrManagerNotCreated))

		var nilErr ManagerNotCreatedError
		require.True(t, errors.As(err, &nilErr))
		require.Equal(t, devicetypes.ComponentTypeCompute, nilErr.ComponentType)
		require.Equal(t, "nil-manager", nilErr.Implementation)
	})

	t.Run("manager descriptor mismatch", func(t *testing.T) {
		_, err := NewRegistry(
			[]FactorySpec{
				testFactorySpec(
					devicetypes.ComponentTypeCompute,
					"wrong-type",
					managerFactory(devicetypes.ComponentTypeNVLSwitch, "wrong-type"),
				),
			},
			cmconfig.Config{
				ComponentManagers: map[devicetypes.ComponentType]string{
					devicetypes.ComponentTypeCompute: "wrong-type",
				},
			},
			providerapi.NewProviderRegistry(),
		)

		require.Error(t, err)
		require.True(t, errors.Is(err, ErrManagerDescriptorMismatch))

		var mismatchErr ManagerDescriptorMismatchError
		require.True(t, errors.As(err, &mismatchErr))
		require.Equal(t, devicetypes.ComponentTypeCompute, mismatchErr.Expected.Type)
		require.Equal(t, "wrong-type", mismatchErr.Expected.Implementation)
		require.Equal(t, devicetypes.ComponentTypeNVLSwitch, mismatchErr.Actual.Type)
		require.Equal(t, "wrong-type", mismatchErr.Actual.Implementation)
	})
}

func TestCreateManagerRejectsDescriptorMismatch(t *testing.T) {
	tests := []struct {
		name       string
		expected   cmcatalog.Descriptor
		factory    ManagerFactory
		wantActual cmcatalog.Descriptor
	}{
		{
			name: "type mismatch",
			expected: testDescriptor(
				devicetypes.ComponentTypeCompute,
				"custom",
			),
			factory: managerFactory(
				devicetypes.ComponentTypeNVLSwitch,
				"custom",
			),
			wantActual: testDescriptor(
				devicetypes.ComponentTypeNVLSwitch,
				"custom",
			),
		},
		{
			name: "implementation mismatch",
			expected: testDescriptor(
				devicetypes.ComponentTypeCompute,
				"custom",
			),
			factory: managerFactory(
				devicetypes.ComponentTypeCompute,
				"other",
			),
			wantActual: testDescriptor(
				devicetypes.ComponentTypeCompute,
				"other",
			),
		},
		{
			name: "required providers mismatch",
			expected: testDescriptor(
				devicetypes.ComponentTypeCompute,
				"custom",
				"alpha",
			),
			factory: managerFactory(
				devicetypes.ComponentTypeCompute,
				"custom",
				"beta",
			),
			wantActual: testDescriptor(
				devicetypes.ComponentTypeCompute,
				"custom",
				"beta",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected, err := tt.expected.Normalize()
			require.NoError(t, err)
			wantActual, err := tt.wantActual.Normalize()
			require.NoError(t, err)

			manager, err := createManager(
				expected,
				tt.factory,
				providerapi.NewProviderRegistry(),
			)

			require.Nil(t, manager)
			require.Error(t, err)
			require.True(t, errors.Is(err, ErrManagerDescriptorMismatch))

			var mismatchErr ManagerDescriptorMismatchError
			require.True(t, errors.As(err, &mismatchErr))
			require.Equal(t, expected, mismatchErr.Expected)
			require.Equal(t, wantActual, mismatchErr.Actual)
		})
	}
}

func TestNewRegistryReturnsNilWhenManagerValidationFails(t *testing.T) {
	registry, err := NewRegistry(
		[]FactorySpec{
			testFactorySpec(
				devicetypes.ComponentTypeCompute,
				"compute",
				managerFactory(devicetypes.ComponentTypeCompute, "compute"),
			),
			testFactorySpec(
				devicetypes.ComponentTypeNVLSwitch,
				"wrong-type",
				managerFactory(devicetypes.ComponentTypePowerShelf, "wrong-type"),
			),
		},
		cmconfig.Config{
			ComponentManagers: map[devicetypes.ComponentType]string{
				devicetypes.ComponentTypeCompute:   "compute",
				devicetypes.ComponentTypeNVLSwitch: "wrong-type",
			},
		},
		providerapi.NewProviderRegistry(),
	)

	require.Nil(t, registry)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrManagerDescriptorMismatch))
}

func TestRegistryFindManager(t *testing.T) {
	t.Run("nil registry", func(t *testing.T) {
		var registry *Registry

		manager := registry.FindManager(devicetypes.ComponentTypeCompute)

		require.Nil(t, manager)
	})

	t.Run("missing active manager", func(t *testing.T) {
		registry, err := NewRegistry(
			nil,
			cmconfig.Config{},
			providerapi.NewProviderRegistry(),
		)
		require.NoError(t, err)

		manager := registry.FindManager(devicetypes.ComponentTypeCompute)

		require.Nil(t, manager)
	})
}

func TestRegistryGetAllManagers(t *testing.T) {
	t.Run("nil registry", func(t *testing.T) {
		var registry *Registry

		managers := registry.GetAllManagers()

		require.Nil(t, managers)
	})

	registry, err := NewRegistry(
		[]FactorySpec{
			testFactorySpec(
				devicetypes.ComponentTypeCompute,
				"compute",
				managerFactory(devicetypes.ComponentTypeCompute, "compute"),
			),
			testFactorySpec(
				devicetypes.ComponentTypeNVLSwitch,
				"switch",
				managerFactory(devicetypes.ComponentTypeNVLSwitch, "switch"),
			),
		},
		cmconfig.Config{
			ComponentManagers: map[devicetypes.ComponentType]string{
				devicetypes.ComponentTypeCompute:   "compute",
				devicetypes.ComponentTypeNVLSwitch: "switch",
			},
		},
		providerapi.NewProviderRegistry(),
	)
	require.NoError(t, err)

	managers := registry.GetAllManagers()

	require.Len(t, managers, 2)
	descriptors := make([]cmcatalog.Descriptor, 0, len(managers))
	for _, manager := range managers {
		descriptors = append(descriptors, manager.Descriptor())
	}
	require.ElementsMatch(t, []cmcatalog.Descriptor{
		testDescriptor(devicetypes.ComponentTypeCompute, "compute"),
		testDescriptor(devicetypes.ComponentTypeNVLSwitch, "switch"),
	}, descriptors)
}
