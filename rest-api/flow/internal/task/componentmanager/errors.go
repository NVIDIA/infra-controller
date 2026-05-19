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
	"fmt"

	cmcatalog "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/catalog"
	cmconfig "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/config"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

var (
	// ErrRegistryNotConfigured reports that the component manager registry is
	// not available.
	ErrRegistryNotConfigured = errors.New("component manager registry is not configured")

	// ErrManagerNotConfigured reports that no active manager is configured for
	// the requested component type.
	ErrManagerNotConfigured = errors.New("component manager is not configured")

	// ErrComponentManagerFactoryNotRegistered reports that no factories were
	// registered for a component type.
	ErrComponentManagerFactoryNotRegistered = cmcatalog.ErrComponentManagerFactoryNotRegistered

	// ErrComponentManagerFactoryNotConfigured reports that a factory spec was
	// created without a factory.
	ErrComponentManagerFactoryNotConfigured = errors.New("component manager factory is not configured")

	// ErrDuplicateDescriptor reports duplicate descriptor metadata for the same
	// component type and implementation.
	ErrDuplicateDescriptor = cmcatalog.ErrDuplicateDescriptor

	// ErrUnknownComponentManagerImplementation reports that the configured
	// implementation name is not registered for a component type.
	ErrUnknownComponentManagerImplementation = cmcatalog.ErrUnknownComponentManagerImplementation

	// ErrManagerCreationFailed reports that a registered manager factory failed.
	ErrManagerCreationFailed = errors.New("component manager creation failed")

	// ErrManagerNotCreated reports that a manager factory returned nil without
	// an error.
	ErrManagerNotCreated = errors.New("component manager was not created")

	// ErrManagerDescriptorMismatch reports that a manager factory returned a
	// manager with different descriptor metadata than its factory spec.
	ErrManagerDescriptorMismatch = errors.New("component manager descriptor mismatch")

	// ErrConfigNotConfigured reports that a nil component manager config was
	// provided where a config value is required.
	ErrConfigNotConfigured = cmconfig.ErrConfigNotConfigured

	// ErrUnknownComponentType reports an unrecognized component type in config.
	ErrUnknownComponentType = cmcatalog.ErrUnknownComponentType

	// ErrComponentManagerImplementationNameEmpty reports that a component type
	// was configured without an implementation name.
	ErrComponentManagerImplementationNameEmpty = cmcatalog.ErrComponentManagerImplementationNameEmpty

	// ErrComponentManagersNotConfigured reports that the service config has no
	// component manager entries.
	ErrComponentManagersNotConfigured = cmconfig.ErrComponentManagersNotConfigured

	// ErrProviderRegistryNotConfigured reports that the provider registry is not
	// available.
	ErrProviderRegistryNotConfigured = providerapi.ErrProviderRegistryNotConfigured

	// ErrProviderNotConfigured reports that a provider or provider config is not
	// available.
	ErrProviderNotConfigured = providerapi.ErrProviderNotConfigured

	// ErrUnknownProvider reports that a provider name is not known in the
	// current provider context.
	ErrUnknownProvider = providerapi.ErrUnknownProvider

	// ErrProviderTypeMismatch reports that a provider exists but has a different
	// concrete type than the caller requested.
	ErrProviderTypeMismatch = providerapi.ErrProviderTypeMismatch

	// ErrProviderNameEmpty reports an empty provider name.
	ErrProviderNameEmpty = providerapi.ErrProviderNameEmpty

	// ErrDuplicateProvider reports that a provider is already registered.
	ErrDuplicateProvider = providerapi.ErrDuplicateProvider

	// ErrProviderConfigNameMismatch reports that a provider config's name does
	// not match the name it was registered under.
	ErrProviderConfigNameMismatch = providerapi.ErrProviderConfigNameMismatch

	// ErrProviderNameMismatch reports that a created provider's name does not
	// match the provider config name.
	ErrProviderNameMismatch = providerapi.ErrProviderNameMismatch

	// ErrDuplicateProviderConfig reports duplicate provider configuration after
	// provider names are normalized.
	ErrDuplicateProviderConfig = cmconfig.ErrDuplicateProviderConfig

	// ErrProviderConfigDecoderNotRegistered reports that a provider is required
	// but no config decoder is registered for it.
	ErrProviderConfigDecoderNotRegistered = cmconfig.ErrProviderConfigDecoderNotRegistered

	// ErrProviderConfigTypeMismatch reports that a provider config has a
	// different concrete type than the caller expected.
	ErrProviderConfigTypeMismatch = errors.New("provider config type mismatch")
)

// ManagerNotConfiguredError includes the component type that has no active
// manager.
type ManagerNotConfiguredError struct {
	ComponentType devicetypes.ComponentType
}

func (e ManagerNotConfiguredError) Error() string {
	return fmt.Sprintf(
		"no active component manager configured for component type %s",
		devicetypes.ComponentTypeToString(e.ComponentType),
	)
}

func (e ManagerNotConfiguredError) Is(target error) bool {
	return target == ErrManagerNotConfigured
}

// ComponentManagerFactoryNotRegisteredError includes the component type that
// has no registered factories.
type ComponentManagerFactoryNotRegisteredError = cmcatalog.ComponentManagerFactoryNotRegisteredError

// ComponentManagerFactoryNotConfiguredError includes the descriptor identity
// that has no factory.
type ComponentManagerFactoryNotConfiguredError struct {
	ComponentType  devicetypes.ComponentType
	Implementation string
}

func (e ComponentManagerFactoryNotConfiguredError) Error() string {
	return fmt.Sprintf(
		"factory is not configured for component type %s with implementation %q",
		devicetypes.ComponentTypeToString(e.ComponentType),
		e.Implementation,
	)
}

func (e ComponentManagerFactoryNotConfiguredError) Is(target error) bool {
	return target == ErrComponentManagerFactoryNotConfigured
}

// DuplicateDescriptorError includes the duplicate descriptor identity.
type DuplicateDescriptorError = cmcatalog.DuplicateDescriptorError

// UnknownComponentManagerImplementationError includes the implementation name
// that was requested and the implementations that were available.
type UnknownComponentManagerImplementationError = cmcatalog.UnknownComponentManagerImplementationError

// ManagerCreationError includes the configured manager identity and wraps the
// factory error.
type ManagerCreationError struct {
	ComponentType  devicetypes.ComponentType
	Implementation string
	Err            error
}

func (e ManagerCreationError) Error() string {
	msg := fmt.Sprintf(
		"failed to create manager for component type %s with implementation '%s'",
		devicetypes.ComponentTypeToString(e.ComponentType),
		e.Implementation,
	)
	if e.Err == nil {
		return msg
	}
	return fmt.Sprintf("%s: %v", msg, e.Err)
}

func (e ManagerCreationError) Unwrap() error {
	return e.Err
}

func (e ManagerCreationError) Is(target error) bool {
	return target == ErrManagerCreationFailed
}

// ManagerNotCreatedError includes the descriptor identity whose factory
// returned no manager.
type ManagerNotCreatedError struct {
	ComponentType  devicetypes.ComponentType
	Implementation string
}

func (e ManagerNotCreatedError) Error() string {
	return fmt.Sprintf(
		"factory returned nil manager for component type %s with implementation '%s'",
		devicetypes.ComponentTypeToString(e.ComponentType),
		e.Implementation,
	)
}

func (e ManagerNotCreatedError) Is(target error) bool {
	return target == ErrManagerNotCreated
}

// ManagerDescriptorMismatchError includes the expected factory spec descriptor
// and the descriptor reported by the created manager.
type ManagerDescriptorMismatchError struct {
	Expected cmcatalog.Descriptor
	Actual   cmcatalog.Descriptor
}

func (e ManagerDescriptorMismatchError) Error() string {
	return fmt.Sprintf(
		"manager descriptor mismatch: expected %s/%s providers %v, got %s/%s providers %v",
		devicetypes.ComponentTypeToString(e.Expected.Type),
		e.Expected.Implementation,
		e.Expected.RequiredProviders,
		devicetypes.ComponentTypeToString(e.Actual.Type),
		e.Actual.Implementation,
		e.Actual.RequiredProviders,
	)
}

func (e ManagerDescriptorMismatchError) Is(target error) bool {
	return target == ErrManagerDescriptorMismatch
}

// UnknownComponentTypeError includes the unrecognized component type string.
type UnknownComponentTypeError = cmcatalog.UnknownComponentTypeError

// ComponentManagerImplementationNameEmptyError includes the component type
// whose configured implementation name is empty.
type ComponentManagerImplementationNameEmptyError = cmcatalog.ComponentManagerImplementationNameEmptyError

// UnknownProviderError includes the unknown provider name.
type UnknownProviderError = providerapi.UnknownProviderError

// ProviderNotConfiguredError includes the provider name that is required but
// not configured.
type ProviderNotConfiguredError = providerapi.ProviderNotConfiguredError

// ProviderTypeMismatchError includes the provider name with the unexpected
// concrete type.
type ProviderTypeMismatchError = providerapi.ProviderTypeMismatchError

// DuplicateProviderError includes the duplicate provider name.
type DuplicateProviderError = providerapi.DuplicateProviderError

// ProviderConfigNameMismatchError includes the provider config map key and the
// name returned by the config.
type ProviderConfigNameMismatchError = providerapi.ProviderConfigNameMismatchError

// ProviderNameMismatchError includes the expected provider name and the name
// returned by the created provider.
type ProviderNameMismatchError = providerapi.ProviderNameMismatchError

// DuplicateProviderConfigError includes the normalized duplicate provider name.
type DuplicateProviderConfigError = cmconfig.DuplicateProviderConfigError

// ProviderConfigDecoderNotRegisteredError includes the provider name with no
// registered config decoder.
type ProviderConfigDecoderNotRegisteredError = cmconfig.ProviderConfigDecoderNotRegisteredError

// RequiredProviderNotConfiguredError includes the required provider and the
// manager identity that requires it.
type RequiredProviderNotConfiguredError = cmconfig.RequiredProviderNotConfiguredError

// ProviderConfigTypeMismatchError includes the provider config type and the
// type expected by the caller.
type ProviderConfigTypeMismatchError struct {
	Name string
	Got  any
	Want string
}

func (e ProviderConfigTypeMismatchError) Error() string {
	return fmt.Sprintf(
		"provider %q returned config type %T, want %s",
		e.Name,
		e.Got,
		e.Want,
	)
}

func (e ProviderConfigTypeMismatchError) Is(target error) bool {
	return target == ErrProviderConfigTypeMismatch
}
