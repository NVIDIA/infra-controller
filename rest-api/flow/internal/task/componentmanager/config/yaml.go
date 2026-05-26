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
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	cmcatalog "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/catalog"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

// LoadConfig loads the component manager configuration from a YAML file using
// the supplied provider config decoders and component manager catalog.
func LoadConfig(
	path string,
	decoders *providerapi.ProviderConfigDecoderRegistry,
	managerCatalog cmcatalog.Catalog,
) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	return ParseConfig(data, decoders, managerCatalog)
}

// ParseConfig parses the component manager configuration from YAML data using
// the supplied provider config decoders and component manager catalog.
func ParseConfig(
	data []byte,
	decoders *providerapi.ProviderConfigDecoderRegistry,
	managerCatalog cmcatalog.Catalog,
) (Config, error) {
	if decoders == nil {
		return Config{}, ErrProviderConfigDecoderRegistryRequired
	}

	rawComponentManagers, rawProviders, err := parseConfigYAML(data)
	if err != nil {
		return Config{}, err
	}

	config := newConfig()

	if err := setYAMLComponentManagers(&config, rawComponentManagers); err != nil {
		return Config{}, err
	}

	if err := setYAMLProviderConfigs(&config, rawProviders, decoders); err != nil {
		return Config{}, err
	}

	if err := config.completeProviderConfigs(decoders, managerCatalog); err != nil {
		return Config{}, err
	}

	return config, nil
}

// parseConfigYAML decodes only the generic YAML envelope. Provider-specific
// YAML remains as raw nodes so each provider decoder owns its own schema.
func parseConfigYAML(data []byte) (map[string]string, map[string]yaml.Node, error) {
	var raw struct {
		ComponentManagers map[string]string    `yaml:"component_managers"`
		Providers         map[string]yaml.Node `yaml:"providers"`
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil {
		return nil, nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return raw.ComponentManagers, raw.Providers, nil
}

// setYAMLComponentManagers converts YAML component type keys to typed
// component types before adding them to the config.
func setYAMLComponentManagers(
	config *Config,
	rawComponentManagers map[string]string,
) error {
	for typeStr, implName := range rawComponentManagers {
		ct := devicetypes.ComponentTypeFromString(typeStr)
		if ct == devicetypes.ComponentTypeUnknown {
			return UnknownComponentTypeError{Name: typeStr}
		}

		if err := config.addComponentManager(ct, implName); err != nil {
			return err
		}
	}
	return nil
}

// setYAMLProviderConfigs decodes explicitly configured provider overrides.
// Missing required providers are intentionally handled later by
// completeProviderConfigs.
func setYAMLProviderConfigs(
	config *Config,
	rawProviders map[string]yaml.Node,
	decoders *providerapi.ProviderConfigDecoderRegistry,
) error {
	for rawName, rawNode := range rawProviders {
		name, decoder, err := config.prepareProviderConfigForAdd(rawName, decoders)
		if err != nil {
			return err
		}

		if config.HasProvider(name) {
			return DuplicateProviderConfigError{Name: name}
		}

		providerConfig, err := decoder.DecodeYAML(rawNode)
		if err != nil {
			return fmt.Errorf("provider %q: %w", name, err)
		}

		config.ProviderConfigs[name] = providerConfig
	}

	return nil
}
