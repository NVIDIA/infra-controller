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
	"errors"
	"fmt"

	cmcatalog "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/catalog"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

var (
	// ErrConfigNotConfigured reports that a nil component manager config was
	// provided where a Config value is required.
	ErrConfigNotConfigured = errors.New("component manager config is not configured")

	// ErrUnknownComponentType reports an unrecognized component type in config.
	ErrUnknownComponentType = cmcatalog.ErrUnknownComponentType

	// ErrComponentManagerImplementationNameEmpty reports that a component type
	// was configured without an implementation name.
	ErrComponentManagerImplementationNameEmpty = cmcatalog.ErrComponentManagerImplementationNameEmpty

	// ErrComponentManagersNotConfigured reports that the service config has no
	// component manager entries.
	ErrComponentManagersNotConfigured = errors.New("component managers are not configured")

	// ErrDuplicateProviderConfig reports duplicate provider configuration after
	// provider names are normalized.
	ErrDuplicateProviderConfig = errors.New("duplicate provider config")

	// ErrProviderConfigDecoderNotRegistered reports that a provider is required
	// but no config decoder is registered for it.
	ErrProviderConfigDecoderNotRegistered = errors.New("provider config decoder is not registered")

	// ErrProviderConfigDecoderRegistryRequired reports that a config operation
	// requires a provider config decoder registry argument.
	ErrProviderConfigDecoderRegistryRequired = errors.New("provider config decoder registry is required")
)

// UnknownComponentTypeError includes the unrecognized component type string.
type UnknownComponentTypeError = cmcatalog.UnknownComponentTypeError

// ComponentManagerImplementationNameEmptyError includes the component type
// whose configured implementation name is empty.
type ComponentManagerImplementationNameEmptyError = cmcatalog.ComponentManagerImplementationNameEmptyError

// DuplicateProviderConfigError includes the normalized duplicate provider name.
type DuplicateProviderConfigError struct {
	// Name is the duplicate provider name after trimming whitespace.
	Name string
}

func (e DuplicateProviderConfigError) Error() string {
	return fmt.Sprintf("duplicate provider config for %q", e.Name)
}

func (e DuplicateProviderConfigError) Is(target error) bool {
	return target == ErrDuplicateProviderConfig
}

// ProviderConfigDecoderNotRegisteredError includes the provider name with no
// registered config decoder.
type ProviderConfigDecoderNotRegisteredError struct {
	// Name is the provider name that has no registered config decoder.
	Name string
	// ComponentType is the component manager type requiring the provider.
	ComponentType devicetypes.ComponentType
	// Implementation is the component manager implementation requiring the
	// provider.
	Implementation string
}

func (e ProviderConfigDecoderNotRegisteredError) Error() string {
	if e.ComponentType != devicetypes.ComponentTypeUnknown || e.Implementation != "" {
		return fmt.Sprintf(
			"provider config decoder %q required by component manager %s/%s is not registered",
			e.Name,
			devicetypes.ComponentTypeToString(e.ComponentType),
			e.Implementation,
		)
	}
	return fmt.Sprintf("provider config decoder %q is not registered", e.Name)
}

func (e ProviderConfigDecoderNotRegisteredError) Is(target error) bool {
	return target == ErrProviderConfigDecoderNotRegistered
}

// RequiredProviderNotConfiguredError includes the required provider and the
// component manager identity that requires it.
type RequiredProviderNotConfiguredError struct {
	// Provider is the provider name required by the component manager.
	Provider string
	// ComponentType is the component manager type requiring the provider.
	ComponentType devicetypes.ComponentType
	// Implementation is the component manager implementation requiring the
	// provider.
	Implementation string
}

func (e RequiredProviderNotConfiguredError) Error() string {
	return fmt.Sprintf(
		"provider %q required by component manager %s/%s is not configured",
		e.Provider,
		devicetypes.ComponentTypeToString(e.ComponentType),
		e.Implementation,
	)
}

func (e RequiredProviderNotConfiguredError) Is(target error) bool {
	return target == providerapi.ErrProviderNotConfigured
}
