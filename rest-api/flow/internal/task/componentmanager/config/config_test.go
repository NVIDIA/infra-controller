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

package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	cmcatalog "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/catalog"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providers/nico"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providers/nvswitchmanager"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providers/psm"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

type customProviderConfig struct {
	name string
}

func (c customProviderConfig) Name() string {
	return c.name
}

func (c customProviderConfig) NewProvider(context.Context) (providerapi.Provider, error) {
	return nil, nil
}

type customProviderConfigDecoder struct {
	name string
}

func (d customProviderConfigDecoder) Name() string {
	return d.name
}

func (d customProviderConfigDecoder) DefaultConfig() providerapi.ProviderConfig {
	return customProviderConfig{name: d.name}
}

func (d customProviderConfigDecoder) DecodeYAML(raw yaml.Node) (providerapi.ProviderConfig, error) {
	return d.DefaultConfig(), nil
}

func TestParseConfigWithExplicitProviders(t *testing.T) {
	config, err := parseConfigWithBuiltins(t, `
component_managers:
  compute: nico
  nvlswitch: nvswitchmanager
  powershelf: psm
providers:
  nico:
    timeout: 45s
    compute_power_delay: 0s
  psm:
    timeout: 20s
  nvswitchmanager:
    timeout: 90s
`)
	require.NoError(t, err)

	assert.Equal(t, nico.ProviderName, config.ComponentManagers[devicetypes.ComponentTypeCompute])
	assert.Equal(t, nvswitchmanager.ProviderName, config.ComponentManagers[devicetypes.ComponentTypeNVLSwitch])
	assert.Equal(t, psm.ProviderName, config.ComponentManagers[devicetypes.ComponentTypePowerShelf])

	nicoConfig, ok := config.ProviderConfigs[nico.ProviderName].(*nico.Config)
	require.True(t, ok)
	assert.Equal(t, 45*time.Second, nicoConfig.Timeout)
	assert.Equal(t, 0*time.Second, nicoConfig.ComputePowerDelay)

	psmConfig, ok := config.ProviderConfigs[psm.ProviderName].(*psm.Config)
	require.True(t, ok)
	assert.Equal(t, 20*time.Second, psmConfig.Timeout)

	nsmConfig, ok := config.ProviderConfigs[nvswitchmanager.ProviderName].(*nvswitchmanager.Config)
	require.True(t, ok)
	assert.Equal(t, 90*time.Second, nsmConfig.Timeout)
}

func TestParseConfigDerivesProviders(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		wantEnabled []string
	}{
		{
			name: "mock managers do not need providers",
			configYAML: `
component_managers:
  compute: mock
  nvlswitch: mock
  powershelf: mock
`,
			wantEnabled: nil,
		},
		{
			name: "nico",
			configYAML: `
component_managers:
  compute: nico
`,
			wantEnabled: []string{nico.ProviderName},
		},
		{
			name: "psm",
			configYAML: `
component_managers:
  powershelf: psm
`,
			wantEnabled: []string{psm.ProviderName},
		},
		{
			name: "nvswitchmanager",
			configYAML: `
component_managers:
  nvlswitch: nvswitchmanager
`,
			wantEnabled: []string{nvswitchmanager.ProviderName},
		},
		{
			name: "deduplicates providers",
			configYAML: `
component_managers:
  compute: nico
  nvlswitch: nico
  powershelf: psm
`,
			wantEnabled: []string{nico.ProviderName, psm.ProviderName},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config, err := parseConfigWithBuiltins(t, tc.configYAML)
			require.NoError(t, err)
			assert.ElementsMatch(t, tc.wantEnabled, providerConfigNames(config))
		})
	}
}

func TestParseConfigCompletesMissingExplicitProviders(t *testing.T) {
	config, err := parseConfigWithBuiltins(t, `
component_managers:
  compute: nico
  powershelf: psm
providers:
  psm:
    timeout: 20s
`)
	require.NoError(t, err)

	assert.True(t, config.HasProvider(nico.ProviderName))
	assert.True(t, config.HasProvider(psm.ProviderName))

	nicoConfig, ok := config.ProviderConfigs[nico.ProviderName].(*nico.Config)
	require.True(t, ok)
	assert.Equal(t, nico.DefaultTimeout, nicoConfig.Timeout)
	assert.Equal(
		t,
		nico.DefaultComputePowerDelay,
		nicoConfig.ComputePowerDelay,
	)

	psmConfig, ok := config.ProviderConfigs[psm.ProviderName].(*psm.Config)
	require.True(t, ok)
	assert.Equal(t, 20*time.Second, psmConfig.Timeout)
}

func TestParseConfigCompletesEmptyProviders(t *testing.T) {
	config, err := parseConfigWithBuiltins(t, `
component_managers:
  compute: nico
providers: {}
`)
	require.NoError(t, err)

	assert.True(t, config.HasProvider(nico.ProviderName))

	err = config.Validate(serviceProviderConfigDecoderRegistry(t), testCatalog(t))
	require.NoError(t, err)
}

func TestParseConfigErrors(t *testing.T) {
	tests := []struct {
		name       string
		configYAML string
		wantErr    error
		checkErr   func(*testing.T, error)
	}{
		{
			name: "unknown provider",
			configYAML: `
component_managers:
  compute: mock
providers:
  madeup: {}
`,
			wantErr: providerapi.ErrUnknownProvider,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				var providerErr providerapi.UnknownProviderError
				require.True(t, errors.As(err, &providerErr))
				assert.Equal(t, "madeup", providerErr.Name)
			},
		},
		{
			name: "unknown component type",
			configYAML: `
component_managers:
  madeup: mock
`,
			wantErr: ErrUnknownComponentType,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				var ctErr UnknownComponentTypeError
				require.True(t, errors.As(err, &ctErr))
				assert.Equal(t, "madeup", ctErr.Name)
			},
		},
		{
			name: "empty implementation name",
			configYAML: `
component_managers:
  compute: " "
`,
			wantErr: ErrComponentManagerImplementationNameEmpty,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				var nameErr ComponentManagerImplementationNameEmptyError
				require.True(t, errors.As(err, &nameErr))
				assert.Equal(t, devicetypes.ComponentTypeCompute, nameErr.ComponentType)
			},
		},
		{
			name: "duplicate provider after trimming name",
			configYAML: `
component_managers:
  compute: mock
providers:
  nico:
    timeout: 30s
  " nico ":
    timeout: 45s
`,
			wantErr: ErrDuplicateProviderConfig,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				var duplicateErr DuplicateProviderConfigError
				require.True(t, errors.As(err, &duplicateErr))
				assert.Equal(t, nico.ProviderName, duplicateErr.Name)
			},
		},
		{
			name: "invalid nico timeout",
			configYAML: `
component_managers:
  compute: mock
providers:
  nico:
    timeout: nope
`,
			wantErr: providerapi.ErrInvalidProviderConfigField,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assertInvalidProviderConfigField(t, err, nico.ProviderName, "timeout")
			},
		},
		{
			name: "invalid psm timeout",
			configYAML: `
component_managers:
  compute: mock
providers:
  psm:
    timeout: nope
`,
			wantErr: providerapi.ErrInvalidProviderConfigField,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assertInvalidProviderConfigField(t, err, psm.ProviderName, "timeout")
			},
		},
		{
			name: "invalid nvswitchmanager timeout",
			configYAML: `
component_managers:
  compute: mock
providers:
  nvswitchmanager:
    timeout: nope
`,
			wantErr: providerapi.ErrInvalidProviderConfigField,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assertInvalidProviderConfigField(t, err, nvswitchmanager.ProviderName, "timeout")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseConfigWithBuiltins(t, tc.configYAML)
			require.Error(t, err)
			assert.True(t, errors.Is(err, tc.wantErr))
			if tc.checkErr != nil {
				tc.checkErr(t, err)
			}
		})
	}
}

func TestParseConfigAllowsCustomProviderDecoderRegistry(t *testing.T) {
	registry := providerapi.NewProviderConfigDecoderRegistry()
	require.NoError(t, registry.Register(customProviderConfigDecoder{name: "custom"}))

	config, err := ParseConfig([]byte(`
component_managers:
  compute: mock
providers:
  custom: {}
`), registry, testCatalog(t))
	require.NoError(t, err)

	assert.True(t, config.HasProvider("custom"))
	assert.Equal(t, customProviderConfig{name: "custom"}, config.ProviderConfigs["custom"])
}

func TestParseConfigDerivesProviderForDifferentImplementationName(t *testing.T) {
	registry := serviceProviderConfigDecoderRegistry(t)
	require.NoError(t, registry.Register(customProviderConfigDecoder{name: "custom"}))

	config, err := ParseConfig([]byte(`
component_managers:
  compute: custom-manager
`), registry, testCatalog(t))
	require.NoError(t, err)

	assert.True(t, config.HasProvider("custom"))
}

func TestParseConfigDerivesMultipleProvidersForManager(t *testing.T) {
	registry := serviceProviderConfigDecoderRegistry(t)
	require.NoError(t, registry.Register(customProviderConfigDecoder{name: "custom"}))

	config, err := ParseConfig([]byte(`
component_managers:
  compute: multi-provider
`), registry, testCatalog(t))
	require.NoError(t, err)

	assert.True(t, config.HasProvider("custom"))
	assert.True(t, config.HasProvider(nico.ProviderName))
}

func TestParseConfigRequiresDecoderRegistry(t *testing.T) {
	_, err := ParseConfig([]byte(`component_managers: {}`), nil, cmcatalog.Catalog{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrProviderConfigDecoderRegistryRequired))
}

func TestNewConfigRequiresDecoderRegistry(t *testing.T) {
	_, err := New(map[devicetypes.ComponentType]string{}, nil, cmcatalog.Catalog{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrProviderConfigDecoderRegistryRequired))
}

func TestValidateRequiresConfig(t *testing.T) {
	var config *Config

	err := config.Validate(serviceProviderConfigDecoderRegistry(t), cmcatalog.Catalog{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrConfigNotConfigured))
}

func TestValidateReportsRequiredProviderManagerIdentity(t *testing.T) {
	config := Config{
		ComponentManagers: map[devicetypes.ComponentType]string{
			devicetypes.ComponentTypeCompute: nico.ProviderName,
		},
		ProviderConfigs: map[string]providerapi.ProviderConfig{},
	}

	err := config.Validate(
		serviceProviderConfigDecoderRegistry(t),
		testCatalog(t),
	)

	require.Error(t, err)
	assert.True(t, errors.Is(err, providerapi.ErrProviderNotConfigured))

	var requiredErr RequiredProviderNotConfiguredError
	require.True(t, errors.As(err, &requiredErr))
	assert.Equal(t, nico.ProviderName, requiredErr.Provider)
	assert.Equal(t, devicetypes.ComponentTypeCompute, requiredErr.ComponentType)
	assert.Equal(t, nico.ProviderName, requiredErr.Implementation)
}

func TestCompleteProviderConfigsReportsRequiredProviderDecoderManagerIdentity(t *testing.T) {
	config := Config{
		ComponentManagers: map[devicetypes.ComponentType]string{
			devicetypes.ComponentTypeCompute: "custom-manager",
		},
		ProviderConfigs: map[string]providerapi.ProviderConfig{},
	}

	err := config.completeProviderConfigs(
		providerapi.NewProviderConfigDecoderRegistry(),
		testCatalog(t),
	)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrProviderConfigDecoderNotRegistered))

	var decoderErr ProviderConfigDecoderNotRegisteredError
	require.True(t, errors.As(err, &decoderErr))
	assert.Equal(t, "custom", decoderErr.Name)
	assert.Equal(t, devicetypes.ComponentTypeCompute, decoderErr.ComponentType)
	assert.Equal(t, "custom-manager", decoderErr.Implementation)
}

func assertInvalidProviderConfigField(
	t *testing.T,
	err error,
	provider string,
	field string,
) {
	t.Helper()

	var fieldErr providerapi.InvalidProviderConfigFieldError
	require.True(t, errors.As(err, &fieldErr))
	assert.Equal(t, provider, fieldErr.Provider)
	assert.Equal(t, field, fieldErr.Field)
}

func TestHasProviderUsesProviderConfigs(t *testing.T) {
	config := Config{
		ProviderConfigs: map[string]providerapi.ProviderConfig{
			nico.ProviderName: &nico.Config{},
		},
	}

	assert.True(t, config.HasProvider(nico.ProviderName))
	assert.False(t, config.HasProvider(psm.ProviderName))
}

func TestNewConfigDerivesDefaultProviderConfigs(t *testing.T) {
	config, err := New(
		map[devicetypes.ComponentType]string{
			devicetypes.ComponentTypeCompute: nico.ProviderName,
		},
		serviceProviderConfigDecoderRegistry(t),
		testCatalog(t),
	)
	require.NoError(t, err)

	assert.True(t, config.HasProvider(nico.ProviderName))
	nicoConfig, ok := config.ProviderConfigs[nico.ProviderName].(*nico.Config)
	require.True(t, ok)
	assert.Equal(t, nico.DefaultTimeout, nicoConfig.Timeout)
	assert.Equal(
		t,
		nico.DefaultComputePowerDelay,
		nicoConfig.ComputePowerDelay,
	)
}

func TestLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "componentmanager.yaml")
	err := os.WriteFile(path, []byte(`
component_managers:
  compute: nico
`), 0o600)
	require.NoError(t, err)

	config, err := LoadConfig(
		path,
		serviceProviderConfigDecoderRegistry(t),
		testCatalog(t),
	)
	require.NoError(t, err)
	assert.True(t, config.HasProvider(nico.ProviderName))
}

func providerConfigNames(config Config) []string {
	names := make([]string, 0, len(config.ProviderConfigs))
	for name := range config.ProviderConfigs {
		names = append(names, name)
	}
	return names
}

func parseConfigWithBuiltins(t *testing.T, data string) (Config, error) {
	t.Helper()
	return ParseConfig(
		[]byte(data),
		serviceProviderConfigDecoderRegistry(t),
		testCatalog(t),
	)
}

func testCatalog(t *testing.T) cmcatalog.Catalog {
	t.Helper()

	catalog, err := cmcatalog.New([]cmcatalog.Descriptor{
		{
			Type:           devicetypes.ComponentTypeCompute,
			Implementation: "mock",
		},
		{
			Type:           devicetypes.ComponentTypeNVLSwitch,
			Implementation: "mock",
		},
		{
			Type:           devicetypes.ComponentTypePowerShelf,
			Implementation: "mock",
		},
		{
			Type:              devicetypes.ComponentTypeCompute,
			Implementation:    nico.ProviderName,
			RequiredProviders: []string{nico.ProviderName},
		},
		{
			Type:              devicetypes.ComponentTypeNVLSwitch,
			Implementation:    nico.ProviderName,
			RequiredProviders: []string{nico.ProviderName},
		},
		{
			Type:              devicetypes.ComponentTypePowerShelf,
			Implementation:    nico.ProviderName,
			RequiredProviders: []string{nico.ProviderName},
		},
		{
			Type:              devicetypes.ComponentTypePowerShelf,
			Implementation:    psm.ProviderName,
			RequiredProviders: []string{psm.ProviderName},
		},
		{
			Type:              devicetypes.ComponentTypeNVLSwitch,
			Implementation:    nvswitchmanager.ProviderName,
			RequiredProviders: []string{nvswitchmanager.ProviderName},
		},
		{
			Type:              devicetypes.ComponentTypeCompute,
			Implementation:    "custom-manager",
			RequiredProviders: []string{"custom"},
		},
		{
			Type:           devicetypes.ComponentTypeCompute,
			Implementation: "multi-provider",
			RequiredProviders: []string{
				"custom",
				nico.ProviderName,
			},
		},
	})
	require.NoError(t, err)

	return catalog
}

func serviceProviderConfigDecoderRegistry(t *testing.T) *providerapi.ProviderConfigDecoderRegistry {
	t.Helper()
	registry := providerapi.NewProviderConfigDecoderRegistry()
	require.NoError(t, registry.Register(nico.ConfigDecoder{}))
	require.NoError(t, registry.Register(psm.ConfigDecoder{}))
	require.NoError(t, registry.Register(nvswitchmanager.ConfigDecoder{}))
	return registry
}
