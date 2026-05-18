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

package nico

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
)

func TestConfigName(t *testing.T) {
	assert.Equal(t, ProviderName, (&Config{}).Name())
}

func TestConfigDecoderDecodeYAML(t *testing.T) {
	decoder := ConfigDecoder{}

	decoded, err := decoder.DecodeYAML(yaml.Node{})
	require.NoError(t, err)
	config := decoded.(*Config)
	assert.Equal(t, DefaultTimeout, config.Timeout)
	assert.Equal(t, DefaultComputePowerDelay, config.ComputePowerDelay)

	decoded, err = decoder.DecodeYAML(providerYAMLNode(t, `
timeout: 15s
compute_power_delay: 0s
`))
	require.NoError(t, err)
	config = decoded.(*Config)
	assert.Equal(t, 15*time.Second, config.Timeout)
	assert.Equal(t, 0*time.Second, config.ComputePowerDelay)

	_, err = decoder.DecodeYAML(providerYAMLNode(t, `timeout: nope`))
	require.Error(t, err)
	assert.True(t, errors.Is(err, providerapi.ErrInvalidProviderConfigField))
	assertInvalidConfigField(t, err, "timeout")

	_, err = decoder.DecodeYAML(providerYAMLNode(t, `compute_power_delay: nope`))
	require.Error(t, err)
	assert.True(t, errors.Is(err, providerapi.ErrInvalidProviderConfigField))
	assertInvalidConfigField(t, err, "compute_power_delay")

	_, err = decoder.DecodeYAML(providerYAMLNode(t, `timout: 15s`))
	require.Error(t, err)
	assert.True(t, errors.Is(err, providerapi.ErrInvalidProviderConfig))

	var configErr providerapi.InvalidProviderConfigError
	require.True(t, errors.As(err, &configErr))
	assert.Equal(t, ProviderName, configErr.Provider)
}

func assertInvalidConfigField(t *testing.T, err error, field string) {
	t.Helper()

	var fieldErr providerapi.InvalidProviderConfigFieldError
	require.True(t, errors.As(err, &fieldErr))
	assert.Equal(t, ProviderName, fieldErr.Provider)
	assert.Equal(t, field, fieldErr.Field)
}

func providerYAMLNode(t *testing.T, data string) yaml.Node {
	t.Helper()

	var node yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(data), &node))
	require.NotEmpty(t, node.Content)
	return *node.Content[0]
}
