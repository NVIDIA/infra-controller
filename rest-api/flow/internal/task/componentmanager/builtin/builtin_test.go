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

package builtin

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager"
	cmcatalog "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/catalog"
	computenico "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/compute/nico"
	cmconfig "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/config"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/mock"
	nvlswitchnico "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/nvlswitch/nico"
	nvlswitchnsm "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/nvlswitch/nvswitchmanager"
	powershelfnico "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/powershelf/nico"
	powershelfpsm "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/powershelf/psm"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
	nicoprovider "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providers/nico"
	nsmprovider "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providers/nvswitchmanager"
	psmprovider "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providers/psm"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

type testProviderConfig struct {
	name string
}

func (c testProviderConfig) Name() string {
	return c.name
}

func (c testProviderConfig) NewProvider(context.Context) (providerapi.Provider, error) {
	return nil, nil
}

type testServiceProvider struct {
	name string
}

func (p testServiceProvider) Name() string {
	return p.name
}

type testServiceProviderConfig struct {
	name         string
	providerName string
	err          error
	nilProvider  bool
}

func (c testServiceProviderConfig) Name() string {
	return c.name
}

func (c testServiceProviderConfig) NewProvider(context.Context) (providerapi.Provider, error) {
	if c.err != nil {
		return nil, c.err
	}
	if c.nilProvider {
		return nil, nil
	}

	name := c.providerName
	if name == "" {
		name = c.name
	}
	return testServiceProvider{name: name}, nil
}

func TestDefaultServiceComponentManagers(t *testing.T) {
	componentManagers := defaultServiceComponentManagers()

	assert.Equal(t, computenico.ImplementationName, componentManagers[devicetypes.ComponentTypeCompute])
	assert.Equal(t, nvlswitchnico.ImplementationName, componentManagers[devicetypes.ComponentTypeNVLSwitch])
	assert.Equal(t, powershelfnico.ImplementationName, componentManagers[devicetypes.ComponentTypePowerShelf])

	componentManagers[devicetypes.ComponentTypeCompute] = "mutated"
	assert.Equal(
		t,
		computenico.ImplementationName,
		defaultServiceComponentManagers()[devicetypes.ComponentTypeCompute],
	)
}

func TestLoadConfigUsesDefaultsWithoutPath(t *testing.T) {
	config, err := LoadConfig("")
	require.NoError(t, err)

	assert.Equal(
		t,
		defaultServiceComponentManagers(),
		config.ComponentManagers,
	)
	assert.True(t, config.HasProvider(nicoprovider.ProviderName))
	assert.False(t, config.HasProvider(psmprovider.ProviderName))

	nicoConfig, ok := config.ProviderConfigs[nicoprovider.ProviderName].(*nicoprovider.Config)
	require.True(t, ok)
	assert.Equal(t, nicoprovider.DefaultTimeout, nicoConfig.Timeout)
	assert.Equal(
		t,
		nicoprovider.DefaultComputePowerDelay,
		nicoConfig.ComputePowerDelay,
	)
}

func TestLoadConfigUsesAuthoritativeFile(t *testing.T) {
	path := writeServiceConfig(t, `
component_managers:
  compute: mock
providers: {}
`)

	config, err := LoadConfig(path)
	require.NoError(t, err)

	assert.Equal(t, "mock", config.ComponentManagers[devicetypes.ComponentTypeCompute])
	assert.Empty(t, config.ProviderConfigs)
	assert.False(t, config.HasProvider(nicoprovider.ProviderName))
}

func TestLoadConfigRequiresComponentManagers(t *testing.T) {
	path := writeServiceConfig(t, `
providers: {}
`)

	config, err := LoadConfig(path)

	require.Empty(t, config.ComponentManagers)
	require.Error(t, err)
	assert.True(t, errors.Is(err, cmconfig.ErrComponentManagersNotConfigured))
}

func TestLoadConfigCompletesMissingProviders(t *testing.T) {
	path := writeServiceConfig(t, `
component_managers:
  compute: nico
providers: {}
`)

	config, err := LoadConfig(path)

	require.NoError(t, err)
	assert.Equal(t, computenico.ImplementationName, config.ComponentManagers[devicetypes.ComponentTypeCompute])
	assert.True(t, config.HasProvider(nicoprovider.ProviderName))
}

func TestNewProviderRegistry(t *testing.T) {
	registry, err := NewProviderRegistry(
		context.Background(),
		cmconfig.Config{
			ProviderConfigs: map[string]providerapi.ProviderConfig{
				"alpha": testServiceProviderConfig{name: "alpha"},
				"beta":  testServiceProviderConfig{name: "beta"},
			},
		},
	)

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"alpha", "beta"}, registry.List())
	assert.True(t, registry.Has("alpha"))
	assert.True(t, registry.Has("beta"))
}

func TestNewProviderRegistryErrors(t *testing.T) {
	rootErr := errors.New("boom")

	tests := []struct {
		name      string
		config    cmconfig.Config
		wantErr   error
		checkFunc func(*testing.T, error)
	}{
		{
			name: "nil provider config",
			config: cmconfig.Config{
				ProviderConfigs: map[string]providerapi.ProviderConfig{
					"alpha": nil,
				},
			},
			wantErr: providerapi.ErrProviderNotConfigured,
			checkFunc: func(t *testing.T, err error) {
				t.Helper()
				var providerErr providerapi.ProviderNotConfiguredError
				require.True(t, errors.As(err, &providerErr))
				assert.Equal(t, "alpha", providerErr.Name)
			},
		},
		{
			name: "config name mismatch",
			config: cmconfig.Config{
				ProviderConfigs: map[string]providerapi.ProviderConfig{
					"alpha": testServiceProviderConfig{name: "other"},
				},
			},
			wantErr: providerapi.ErrProviderConfigNameMismatch,
		},
		{
			name: "provider creation failed",
			config: cmconfig.Config{
				ProviderConfigs: map[string]providerapi.ProviderConfig{
					"alpha": testServiceProviderConfig{name: "alpha", err: rootErr},
				},
			},
			wantErr: rootErr,
		},
		{
			name: "nil provider",
			config: cmconfig.Config{
				ProviderConfigs: map[string]providerapi.ProviderConfig{
					"alpha": testServiceProviderConfig{
						name:        "alpha",
						nilProvider: true,
					},
				},
			},
			wantErr: providerapi.ErrProviderNotConfigured,
		},
		{
			name: "provider name mismatch",
			config: cmconfig.Config{
				ProviderConfigs: map[string]providerapi.ProviderConfig{
					"alpha": testServiceProviderConfig{
						name:         "alpha",
						providerName: "other",
					},
				},
			},
			wantErr: providerapi.ErrProviderNameMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, err := NewProviderRegistry(context.Background(), tt.config)

			require.Nil(t, registry)
			require.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
			if tt.checkFunc != nil {
				tt.checkFunc(t, err)
			}
		})
	}
}

func TestNewComponentManagerRegistryInitializesBuiltInMockManagers(t *testing.T) {
	config := cmconfig.Config{
		ComponentManagers: map[devicetypes.ComponentType]string{
			devicetypes.ComponentTypeCompute:    mock.ImplementationName,
			devicetypes.ComponentTypeNVLSwitch:  mock.ImplementationName,
			devicetypes.ComponentTypePowerShelf: mock.ImplementationName,
		},
	}

	registry, err := NewComponentManagerRegistry(
		config,
		providerapi.NewProviderRegistry(),
	)

	require.NoError(t, err)
	require.NotNil(t, registry)

	for componentType := range config.ComponentManagers {
		manager, err := registry.GetManager(componentType)
		require.NoError(t, err)
		assert.Equal(t, componentType, manager.Descriptor().Type)
	}
}

func TestNewComponentManagerRegistryRejectsImplementationForWrongType(t *testing.T) {
	config := cmconfig.Config{
		ComponentManagers: map[devicetypes.ComponentType]string{
			devicetypes.ComponentTypeCompute: nvlswitchnsm.ImplementationName,
		},
	}

	registry, err := NewComponentManagerRegistry(
		config,
		providerapi.NewProviderRegistry(),
	)

	require.Nil(t, registry)
	require.Error(t, err)
	require.True(t, errors.Is(err, componentmanager.ErrUnknownComponentManagerImplementation))

	var implErr componentmanager.UnknownComponentManagerImplementationError
	require.True(t, errors.As(err, &implErr))
	assert.Equal(t, devicetypes.ComponentTypeCompute, implErr.ComponentType)
	assert.Equal(t, nvlswitchnsm.ImplementationName, implErr.Implementation)
	assert.ElementsMatch(
		t,
		[]string{computenico.ImplementationName, mock.ImplementationName},
		implErr.Available,
	)
	assert.Equal(
		t,
		[]devicetypes.ComponentType{devicetypes.ComponentTypeNVLSwitch},
		implErr.RegisteredFor,
	)
}

func TestServiceProviderConfigDecoderRegistry(t *testing.T) {
	registry, err := newProviderDecoderRegistry()
	require.NoError(t, err)

	assert.ElementsMatch(
		t,
		[]string{
			nicoprovider.ProviderName,
			psmprovider.ProviderName,
			nsmprovider.ProviderName,
		},
		registry.List(),
	)

	_, ok := registry.Get(nicoprovider.ProviderName)
	assert.True(t, ok)

	_, ok = registry.Get(psmprovider.ProviderName)
	assert.True(t, ok)

	_, ok = registry.Get(nsmprovider.ProviderName)
	assert.True(t, ok)
}

func TestServiceCatalog(t *testing.T) {
	catalog, err := newCatalog()

	require.NoError(t, err)

	implementations := catalog.ListImplementations()
	assert.Equal(
		t,
		[]string{mock.ImplementationName, computenico.ImplementationName},
		implementations[devicetypes.ComponentTypeCompute],
	)
	assert.Equal(
		t,
		[]string{
			mock.ImplementationName,
			nvlswitchnico.ImplementationName,
			nvlswitchnsm.ImplementationName,
		},
		implementations[devicetypes.ComponentTypeNVLSwitch],
	)
	assert.Equal(
		t,
		[]string{
			mock.ImplementationName,
			powershelfnico.ImplementationName,
			powershelfpsm.ImplementationName,
		},
		implementations[devicetypes.ComponentTypePowerShelf],
	)

	assert.Equal(
		t,
		[]string{nicoprovider.ProviderName},
		requireDescriptor(
			t,
			catalog,
			devicetypes.ComponentTypeCompute,
			computenico.ImplementationName,
		).RequiredProviders,
	)
	assert.Equal(
		t,
		[]string{nicoprovider.ProviderName},
		requireDescriptor(
			t,
			catalog,
			devicetypes.ComponentTypeNVLSwitch,
			nvlswitchnico.ImplementationName,
		).RequiredProviders,
	)
	assert.Equal(
		t,
		[]string{nsmprovider.ProviderName},
		requireDescriptor(
			t,
			catalog,
			devicetypes.ComponentTypeNVLSwitch,
			nvlswitchnsm.ImplementationName,
		).RequiredProviders,
	)
	assert.Equal(
		t,
		[]string{nicoprovider.ProviderName},
		requireDescriptor(
			t,
			catalog,
			devicetypes.ComponentTypePowerShelf,
			powershelfnico.ImplementationName,
		).RequiredProviders,
	)
	assert.Equal(
		t,
		[]string{psmprovider.ProviderName},
		requireDescriptor(
			t,
			catalog,
			devicetypes.ComponentTypePowerShelf,
			powershelfpsm.ImplementationName,
		).RequiredProviders,
	)
	assert.Empty(t, requireDescriptor(
		t,
		catalog,
		devicetypes.ComponentTypeCompute,
		mock.ImplementationName,
	).RequiredProviders)
	assert.Empty(t, requireDescriptor(
		t,
		catalog,
		devicetypes.ComponentTypeNVLSwitch,
		mock.ImplementationName,
	).RequiredProviders)
	assert.Empty(t, requireDescriptor(
		t,
		catalog,
		devicetypes.ComponentTypePowerShelf,
		mock.ImplementationName,
	).RequiredProviders)
}

func TestNicoComputePowerDelayUsesProviderConfig(t *testing.T) {
	delay := 7 * time.Second
	config := cmconfig.Config{
		ProviderConfigs: map[string]providerapi.ProviderConfig{
			nicoprovider.ProviderName: &nicoprovider.Config{
				ComputePowerDelay: delay,
			},
		},
	}

	got, err := nicoComputePowerDelay(config)

	require.NoError(t, err)
	assert.Equal(t, delay, got)
}

func TestNicoComputePowerDelayDefaultsWhenProviderConfigMissing(t *testing.T) {
	got, err := nicoComputePowerDelay(cmconfig.Config{})

	require.NoError(t, err)
	assert.Equal(t, time.Duration(0), got)
}

func TestNicoComputePowerDelayRejectsUnexpectedConfigType(t *testing.T) {
	config := cmconfig.Config{
		ProviderConfigs: map[string]providerapi.ProviderConfig{
			nicoprovider.ProviderName: testProviderConfig{
				name: nicoprovider.ProviderName,
			},
		},
	}

	got, err := nicoComputePowerDelay(config)

	assert.Equal(t, time.Duration(0), got)
	require.Error(t, err)
	assert.True(t, errors.Is(err, componentmanager.ErrProviderConfigTypeMismatch))

	var mismatch componentmanager.ProviderConfigTypeMismatchError
	require.True(t, errors.As(err, &mismatch))
	assert.Equal(t, nicoprovider.ProviderName, mismatch.Name)
}

func writeServiceConfig(t *testing.T, data string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "componentmanager.yaml")
	err := os.WriteFile(path, []byte(data), 0o600)
	require.NoError(t, err)
	return path
}

func requireDescriptor(
	t *testing.T,
	catalog cmcatalog.Catalog,
	componentType devicetypes.ComponentType,
	implementation string,
) cmcatalog.Descriptor {
	t.Helper()

	descriptor, ok := catalog.Get(componentType, implementation)
	require.True(t, ok)
	return descriptor
}
