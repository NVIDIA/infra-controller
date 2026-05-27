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
	"fmt"
	"time"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager"
	cmcatalog "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/catalog"
	cmconfig "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/config"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
	nicoprovider "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providers/nico"
)

// newProviderDecoderRegistry creates the provider config decoder registry used
// by the Flow service.
func newProviderDecoderRegistry() (*providerapi.ProviderConfigDecoderRegistry, error) {
	registry := providerapi.NewProviderConfigDecoderRegistry()

	for _, decoder := range serviceProviderConfigDecoders() {
		if err := registry.Register(decoder); err != nil {
			return nil, fmt.Errorf(
				"register service provider config decoder %q: %w",
				decoder.Name(),
				err,
			)
		}
	}

	return registry, nil
}

// newCatalog builds the component manager catalog for the Flow service.
// The catalog contains the descriptors for all the built-in component managers
// supported by the Flow service.
func newCatalog() (cmcatalog.Catalog, error) {
	catalog, err := cmcatalog.New(serviceDescriptors())
	if err != nil {
		return cmcatalog.Catalog{}, fmt.Errorf(
			"build component manager catalog: %w",
			err,
		)
	}

	return catalog, nil
}

func nicoComputePowerDelay(config cmconfig.Config) (time.Duration, error) {
	providerConfig, ok := config.ProviderConfigs[nicoprovider.ProviderName]
	if !ok {
		return 0, nil
	}
	if providerConfig == nil {
		return 0, providerapi.ProviderNotConfiguredError{Name: nicoprovider.ProviderName}
	}

	nicoConfig, ok := providerConfig.(*nicoprovider.Config)
	if !ok {
		return 0, componentmanager.ProviderConfigTypeMismatchError{
			Name: nicoprovider.ProviderName,
			Got:  providerConfig,
			Want: "*nico.Config",
		}
	}
	return nicoConfig.ComputePowerDelay, nil
}
