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
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/nicoapi"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
)

const (
	// ProviderName is the unique identifier for the NICo provider.
	ProviderName = "nico"

	// DefaultTimeout is the default timeout for NICo gRPC calls.
	DefaultTimeout = time.Minute

	// DefaultComputePowerDelay is the default delay between sequential
	// power control calls for compute trays. A small stagger avoids
	// overwhelming the power delivery system.
	DefaultComputePowerDelay = 2 * time.Second
)

// Config holds configuration for the NICo provider.
type Config struct {
	// Timeout is the gRPC call timeout for NICo operations.
	Timeout time.Duration

	// ComputePowerDelay is the delay inserted between sequential power
	// control calls when commanding multiple compute trays.
	// 0 means no delay.
	ComputePowerDelay time.Duration
}

type rawConfig struct {
	Timeout           string `yaml:"timeout"`
	ComputePowerDelay string `yaml:"compute_power_delay"`
}

// Name returns the provider name for this config.
func (*Config) Name() string {
	return ProviderName
}

// NewProvider creates a NICo provider from this config.
func (c *Config) NewProvider(ctx context.Context) (providerapi.Provider, error) {
	// TODO: Thread ctx into nicoapi client creation if provider construction
	// starts performing cancellable work.
	_ = ctx
	return New(*c)
}

// ConfigDecoder owns NICo provider config defaults and YAML decoding.
type ConfigDecoder struct{}

// Name returns the provider name handled by this decoder.
func (ConfigDecoder) Name() string {
	return ProviderName
}

// DefaultConfig returns the default NICo provider config.
func (ConfigDecoder) DefaultConfig() providerapi.ProviderConfig {
	return &Config{
		Timeout:           DefaultTimeout,
		ComputePowerDelay: DefaultComputePowerDelay,
	}
}

// DecodeYAML decodes NICo provider YAML into a typed config.
func (d ConfigDecoder) DecodeYAML(raw yaml.Node) (providerapi.ProviderConfig, error) {
	config := d.DefaultConfig().(*Config)

	var parsed rawConfig
	if err := providerapi.DecodeYAMLStrict(raw, &parsed); err != nil {
		return nil, providerapi.InvalidProviderConfigError{
			Provider: ProviderName,
			Err:      err,
		}
	}

	if parsed.Timeout != "" {
		timeout, err := time.ParseDuration(parsed.Timeout)
		if err != nil {
			return nil, providerapi.InvalidProviderConfigFieldError{
				Provider: ProviderName,
				Field:    "timeout",
				Err:      err,
			}
		}
		config.Timeout = timeout
	}

	if parsed.ComputePowerDelay != "" {
		delay, err := time.ParseDuration(parsed.ComputePowerDelay)
		if err != nil {
			return nil, providerapi.InvalidProviderConfigFieldError{
				Provider: ProviderName,
				Field:    "compute_power_delay",
				Err:      err,
			}
		}
		config.ComputePowerDelay = delay
	}

	return config, nil
}

// Provider wraps a nicoapi.Client and provides it to component manager
// implementations.
type Provider struct {
	client nicoapi.Client
}

// New creates a new Provider using the provided configuration.
func New(config Config) (*Provider, error) {
	client, err := nicoapi.NewClient(config.Timeout)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create NICo client")
		return nil, err
	}
	log.Info().Msg("Successfully created NICo client")
	return &Provider{client: client}, nil
}

// NewWithDefault creates a new Provider with the default configuration.
func NewWithDefault() (*Provider, error) {
	cfg := ConfigDecoder{}.DefaultConfig().(*Config)
	return New(*cfg)
}

// NewFromClient creates a Provider from an existing client.
// This is primarily useful for testing with mock clients.
func NewFromClient(client nicoapi.Client) *Provider {
	return &Provider{client: client}
}

// Name returns the unique identifier for this provider type.
func (p *Provider) Name() string {
	return ProviderName
}

// Client returns the underlying nicoapi.Client.
func (p *Provider) Client() nicoapi.Client {
	return p.client
}
