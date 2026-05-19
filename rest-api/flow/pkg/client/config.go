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

package client

import (
	"errors"
	"fmt"

	pkgcerts "github.com/NVIDIA/infra-controller-rest/flow/pkg/certs"
)

// Config represents the configuration needed to create a new Flow service
// gRPC client.
type Config struct {
	Host       string
	Port       int
	ServerName string // overrides the server name used for TLS SNI and certificate verification

	// CertConfig holds certificate file paths for mTLS. Either all three
	// fields must be set (mTLS enabled) or all must be empty (insecure).
	// Providing only some is a validation error.
	CertConfig pkgcerts.Config
}

// Validate checks if the config fields are set correctly.
func (c *Config) Validate() error {
	if c.Host == "" {
		return errors.New("host is required")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("port must be within (0, 65535]")
	}

	return c.CertConfig.Validate()
}

// Target builds the target string for connecting to Flow gRPC server.
func (c *Config) Target() string {
	return fmt.Sprintf("%s:%v", c.Host, c.Port)
}
