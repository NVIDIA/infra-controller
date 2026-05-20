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

// Package builtin wires the component manager extensions compiled into the
// Flow binary.
package builtin

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager"
	cmconfig "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/config"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
)

// LoadConfig loads the component manager config for the Flow service.
// If path is empty, the embedded service defaults are used. If path is set, the
// YAML file is authoritative and must satisfy service config validation.
func LoadConfig(path string) (cmconfig.Config, error) {
	decoders, err := newProviderDecoderRegistry()
	if err != nil {
		return cmconfig.Config{}, fmt.Errorf(
			"initialize service provider config decoders: %w",
			err,
		)
	}

	catalog, err := newCatalog()
	if err != nil {
		return cmconfig.Config{}, err
	}

	var config cmconfig.Config
	if path != "" {
		config, err = cmconfig.LoadConfig(
			path,
			decoders,
			catalog,
		)
		if err != nil {
			return cmconfig.Config{}, fmt.Errorf("load config from file: %w", err)
		}
	} else {
		config, err = cmconfig.New(
			defaultServiceComponentManagers(),
			decoders,
			catalog,
		)
		if err != nil {
			return cmconfig.Config{}, fmt.Errorf("get default config: %w", err)
		}
	}

	if err := config.Validate(decoders, catalog); err != nil {
		return cmconfig.Config{}, fmt.Errorf("validate loaded config: %w", err)
	}

	return config, nil
}

// NewProviderRegistry creates and initializes the Flow service provider
// registry from decoded provider configs.
func NewProviderRegistry(
	ctx context.Context,
	config cmconfig.Config,
) (*providerapi.ProviderRegistry, error) {
	providerRegistry := providerapi.NewProviderRegistry()

	for name, providerConfig := range config.ProviderConfigs {
		// LoadConfig builds ProviderConfigs through service decoders, but keep
		// this defensive for Config values constructed in tests or by callers.
		if providerConfig == nil {
			return nil, providerapi.ProviderNotConfiguredError{Name: name}
		}
		configName := providerConfig.Name()
		if name != configName {
			return nil, providerapi.ProviderConfigNameMismatchError{
				Name:       name,
				ConfigName: configName,
			}
		}

		provider, err := providerConfig.NewProvider(ctx)
		if err != nil {
			return nil, fmt.Errorf("create provider %q: %w", name, err)
		}
		if provider == nil {
			return nil, providerapi.ProviderNotConfiguredError{Name: name}
		}

		providerName := provider.Name()
		if providerName != name {
			return nil, providerapi.ProviderNameMismatchError{
				Name:         name,
				ProviderName: providerName,
			}
		}
		if err := providerRegistry.Register(provider); err != nil {
			return nil, err
		}
		log.Info().
			Str("provider", name).
			Msg("Initialized provider")
	}

	registeredProviders := providerRegistry.List()
	log.Info().
		Strs("providers", registeredProviders).
		Msg("Provider registry initialized")

	return providerRegistry, nil
}

// NewComponentManagerRegistry creates the component manager registry for the
// Flow service using all component manager implementations compiled into the
// binary.
func NewComponentManagerRegistry(
	config cmconfig.Config,
	providers *providerapi.ProviderRegistry,
) (*componentmanager.Registry, error) {
	catalog, err := newCatalog()
	if err != nil {
		return nil, err
	}

	factorySpecs, err := serviceFactorySpecs(config)
	if err != nil {
		return nil, err
	}

	registry, err := componentmanager.NewRegistry(factorySpecs, config, providers)
	if err != nil {
		return nil, fmt.Errorf("initialize component managers: %w", err)
	}

	impls := catalog.ListImplementations()
	for compType, names := range impls {
		log.Debug().
			Str("component_type", compType.String()).
			Strs("implementations", names).
			Msg("Registered component manager implementations")
	}

	return registry, nil
}
