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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

func TestDescriptorNormalize(t *testing.T) {
	descriptor, err := Descriptor{
		Type:              devicetypes.ComponentTypeCompute,
		Implementation:    " custom ",
		RequiredProviders: []string{" beta ", "alpha", "beta"},
	}.Normalize()

	require.NoError(t, err)
	require.Equal(t, Descriptor{
		Type:              devicetypes.ComponentTypeCompute,
		Implementation:    "custom",
		RequiredProviders: []string{"alpha", "beta"},
	}, descriptor)
}

func TestDescriptorEqual(t *testing.T) {
	descriptor := Descriptor{
		Type:              devicetypes.ComponentTypeCompute,
		Implementation:    "custom",
		RequiredProviders: []string{"alpha", "beta"},
	}

	require.True(t, descriptor.Equal(Descriptor{
		Type:              devicetypes.ComponentTypeCompute,
		Implementation:    "custom",
		RequiredProviders: []string{"alpha", "beta"},
	}))
	require.False(t, descriptor.Equal(Descriptor{
		Type:              devicetypes.ComponentTypeNVLSwitch,
		Implementation:    "custom",
		RequiredProviders: []string{"alpha", "beta"},
	}))
	require.False(t, descriptor.Equal(Descriptor{
		Type:              devicetypes.ComponentTypeCompute,
		Implementation:    "other",
		RequiredProviders: []string{"alpha", "beta"},
	}))
	require.False(t, descriptor.Equal(Descriptor{
		Type:              devicetypes.ComponentTypeCompute,
		Implementation:    "custom",
		RequiredProviders: []string{"alpha"},
	}))
}

func TestDescriptorNormalizeRejectsInvalidDescriptor(t *testing.T) {
	tests := []struct {
		name       string
		descriptor Descriptor
		wantErr    error
	}{
		{
			name: "unknown component type",
			descriptor: Descriptor{
				Type:           devicetypes.ComponentTypeUnknown,
				Implementation: "custom",
			},
			wantErr: ErrUnknownComponentType,
		},
		{
			name: "empty implementation",
			descriptor: Descriptor{
				Type:           devicetypes.ComponentTypeCompute,
				Implementation: " ",
			},
			wantErr: ErrComponentManagerImplementationNameEmpty,
		},
		{
			name: "empty required provider",
			descriptor: Descriptor{
				Type:              devicetypes.ComponentTypeCompute,
				Implementation:    "custom",
				RequiredProviders: []string{"nico", " "},
			},
			wantErr: providerapi.ErrProviderNameEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.descriptor.Normalize()

			require.Error(t, err)
			require.True(t, errors.Is(err, tt.wantErr))
		})
	}
}

func TestNewIndexesNormalizedDescriptors(t *testing.T) {
	catalog, err := New([]Descriptor{
		{
			Type:              devicetypes.ComponentTypeCompute,
			Implementation:    " custom ",
			RequiredProviders: []string{" beta ", "alpha", "beta"},
		},
		{
			Type:           devicetypes.ComponentTypeCompute,
			Implementation: "builtin",
		},
		{
			Type:           devicetypes.ComponentTypePowerShelf,
			Implementation: "psm",
		},
	})
	require.NoError(t, err)

	descriptor, ok := catalog.Get(devicetypes.ComponentTypeCompute, "custom")
	require.True(t, ok)
	require.Equal(t, Descriptor{
		Type:              devicetypes.ComponentTypeCompute,
		Implementation:    "custom",
		RequiredProviders: []string{"alpha", "beta"},
	}, descriptor)

	_, ok = catalog.Get(devicetypes.ComponentTypeCompute, "missing")
	require.False(t, ok)
	_, ok = catalog.Get(devicetypes.ComponentTypeNVLSwitch, "custom")
	require.False(t, ok)

	require.Equal(
		t,
		[]string{"builtin", "custom"},
		catalog.Implementations(devicetypes.ComponentTypeCompute),
	)
	require.Empty(t, catalog.Implementations(devicetypes.ComponentTypeNVLSwitch))
	require.Equal(
		t,
		map[devicetypes.ComponentType][]string{
			devicetypes.ComponentTypeCompute:    {"builtin", "custom"},
			devicetypes.ComponentTypePowerShelf: {"psm"},
		},
		catalog.ListImplementations(),
	)
}

func TestGetReturnsDescriptorCopy(t *testing.T) {
	catalog, err := New([]Descriptor{
		{
			Type:              devicetypes.ComponentTypeCompute,
			Implementation:    "custom",
			RequiredProviders: []string{"alpha", "beta"},
		},
	})
	require.NoError(t, err)

	descriptor, ok := catalog.Get(devicetypes.ComponentTypeCompute, "custom")
	require.True(t, ok)
	descriptor.RequiredProviders = append(descriptor.RequiredProviders[:1], "mutated")

	descriptor, ok = catalog.Get(devicetypes.ComponentTypeCompute, "custom")
	require.True(t, ok)
	require.Equal(t, []string{"alpha", "beta"}, descriptor.RequiredProviders)

	descriptor.RequiredProviders[0] = "mutated"

	descriptor, ok = catalog.Get(devicetypes.ComponentTypeCompute, "custom")
	require.True(t, ok)
	require.Equal(t, []string{"alpha", "beta"}, descriptor.RequiredProviders)
}

func TestNewRejectsDuplicateDescriptor(t *testing.T) {
	_, err := New([]Descriptor{
		{
			Type:           devicetypes.ComponentTypeCompute,
			Implementation: " custom ",
		},
		{
			Type:           devicetypes.ComponentTypeCompute,
			Implementation: "custom",
		},
	})

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrDuplicateDescriptor))

	var duplicateErr DuplicateDescriptorError
	require.True(t, errors.As(err, &duplicateErr))
	require.Equal(t, devicetypes.ComponentTypeCompute, duplicateErr.ComponentType)
	require.Equal(t, "custom", duplicateErr.Implementation)
}

func TestSelectedDescriptors(t *testing.T) {
	catalog, err := New([]Descriptor{
		{
			Type:              devicetypes.ComponentTypeNVLSwitch,
			Implementation:    "mock",
			RequiredProviders: []string{},
		},
		{
			Type:           devicetypes.ComponentTypePowerShelf,
			Implementation: "multi-provider",
			RequiredProviders: []string{
				"beta",
				"alpha",
			},
		},
		{
			Type:           devicetypes.ComponentTypeCompute,
			Implementation: "custom",
			RequiredProviders: []string{
				"zeta",
				"alpha",
			},
		},
	})
	require.NoError(t, err)

	descriptors, err := catalog.SelectedDescriptors(map[devicetypes.ComponentType]string{
		devicetypes.ComponentTypePowerShelf: "multi-provider",
		devicetypes.ComponentTypeNVLSwitch:  "mock",
		devicetypes.ComponentTypeCompute:    "custom",
	})

	require.NoError(t, err)
	require.Equal(t, []Descriptor{
		{
			Type:           devicetypes.ComponentTypeCompute,
			Implementation: "custom",
			RequiredProviders: []string{
				"alpha",
				"zeta",
			},
		},
		{
			Type:              devicetypes.ComponentTypeNVLSwitch,
			Implementation:    "mock",
			RequiredProviders: []string{},
		},
		{
			Type:           devicetypes.ComponentTypePowerShelf,
			Implementation: "multi-provider",
			RequiredProviders: []string{
				"alpha",
				"beta",
			},
		},
	}, descriptors)
}

func TestSelectedDescriptorsReturnsDescriptorCopies(t *testing.T) {
	catalog, err := New([]Descriptor{
		{
			Type:              devicetypes.ComponentTypeCompute,
			Implementation:    "custom",
			RequiredProviders: []string{"alpha", "beta"},
		},
	})
	require.NoError(t, err)

	descriptors, err := catalog.SelectedDescriptors(map[devicetypes.ComponentType]string{
		devicetypes.ComponentTypeCompute: "custom",
	})
	require.NoError(t, err)
	require.Len(t, descriptors, 1)
	descriptors[0].RequiredProviders[0] = "mutated"

	descriptors, err = catalog.SelectedDescriptors(map[devicetypes.ComponentType]string{
		devicetypes.ComponentTypeCompute: "custom",
	})
	require.NoError(t, err)
	require.Equal(t, []Descriptor{
		{
			Type:              devicetypes.ComponentTypeCompute,
			Implementation:    "custom",
			RequiredProviders: []string{"alpha", "beta"},
		},
	}, descriptors)
}

func TestSelectedDescriptorsAllowsEmptySelection(t *testing.T) {
	catalog, err := New([]Descriptor{
		{
			Type:           devicetypes.ComponentTypeCompute,
			Implementation: "custom",
		},
	})
	require.NoError(t, err)

	descriptors, err := catalog.SelectedDescriptors(nil)

	require.NoError(t, err)
	require.Empty(t, descriptors)
}

func TestSelectedDescriptorsRejectsUnregisteredComponentType(t *testing.T) {
	catalog, err := New([]Descriptor{
		{
			Type:           devicetypes.ComponentTypeCompute,
			Implementation: "custom",
		},
	})
	require.NoError(t, err)

	descriptors, err := catalog.SelectedDescriptors(map[devicetypes.ComponentType]string{
		devicetypes.ComponentTypeUMS: "custom",
	})

	require.Nil(t, descriptors)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrComponentManagerFactoryNotRegistered))

	var factoryErr ComponentManagerFactoryNotRegisteredError
	require.True(t, errors.As(err, &factoryErr))
	require.Equal(t, devicetypes.ComponentTypeUMS, factoryErr.ComponentType)
}

func TestSelectedDescriptorsRejectsUnknownImplementation(t *testing.T) {
	catalog, err := New([]Descriptor{
		{
			Type:           devicetypes.ComponentTypeCompute,
			Implementation: "known",
		},
		{
			Type:           devicetypes.ComponentTypeCompute,
			Implementation: "alternate",
		},
		{
			Type:           devicetypes.ComponentTypeNVLSwitch,
			Implementation: "switch",
		},
		{
			Type:           devicetypes.ComponentTypePowerShelf,
			Implementation: "switch",
		},
	})
	require.NoError(t, err)

	descriptors, err := catalog.SelectedDescriptors(map[devicetypes.ComponentType]string{
		devicetypes.ComponentTypeCompute: "switch",
	})

	require.Nil(t, descriptors)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrUnknownComponentManagerImplementation))

	var implErr UnknownComponentManagerImplementationError
	require.True(t, errors.As(err, &implErr))
	require.Equal(t, devicetypes.ComponentTypeCompute, implErr.ComponentType)
	require.Equal(t, "switch", implErr.Implementation)
	require.Equal(t, []string{"alternate", "known"}, implErr.Available)
	require.Equal(t, []devicetypes.ComponentType{
		devicetypes.ComponentTypeNVLSwitch,
		devicetypes.ComponentTypePowerShelf,
	}, implErr.RegisteredFor)
}
