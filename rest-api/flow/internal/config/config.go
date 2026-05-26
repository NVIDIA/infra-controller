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

// Package config defines a configuration file that allows tweaking values on a per environment basis.
// Default values are used if the file is missing or for individual values that were not specified.

package config

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const defaultLocation = "/etc/flow/flowconfig.yaml"

// Config is a set of configuration operations that will be read from an environment specific config file
type Config struct {
	InventoryRunFrequency time.Duration `yaml:"inventory_run_frequency"`
	GRPCTimeout           time.Duration `yaml:"grpc_timeout"`
	DisableInventory      bool          `yaml:"disable_inventory"`
	LeakDetectionInterval time.Duration `yaml:"leak_detection_interval"`
	DisableLeakDetection  bool          `yaml:"disable_leak_detection"`
}

// defaultConfig sets up the default values used when something is not specified
func defaultConfig() Config {
	return Config{InventoryRunFrequency: time.Minute,
		GRPCTimeout:           time.Minute,
		LeakDetectionInterval: time.Minute,
		DisableLeakDetection:  false,
	}
}

// ReadConfig reads a configuration file if present and returns a Config with the details.  A config file with
// invalid syntax will result in a fatal error.
func ReadConfig() (config Config) {
	filename := os.Getenv("FLOW_CONFIG_FILE")
	if filename == "" {
		filename = defaultLocation
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Warn().Msgf("Config file %s not found, using defaults.", filename)
		return defaultConfig()
	}
	defer file.Close()

	// Will use default values for anything not specified
	config = defaultConfig()
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		log.Fatal().Msgf("Invalid configuration file %s: %v", filename, err)
	}

	return config
}

// UnitTestConfig returns a Config that is suitable for running unit tests.
func UnitTestConfig() Config {
	cfg := defaultConfig()
	cfg.InventoryRunFrequency = time.Millisecond

	return cfg
}
