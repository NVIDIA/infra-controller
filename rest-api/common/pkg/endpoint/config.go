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

package endpoint

import (
	"errors"
	"fmt"

	"github.com/NVIDIA/infra-controller-rest/common/pkg/credential"
)

// Config represents a network endpoint with optional authentication and TLS.
type Config struct {
	Host              string
	Port              int
	Credential        *credential.Credential
	CACertificatePath string
}

// Validate checks if the Config fields are set correctly.
func (c *Config) Validate() error {
	if c.Host == "" {
		return errors.New("host is required")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("port must be between (0, 65535]")
	}

	if c.Credential != nil && !c.Credential.IsValid() {
		return errors.New("valid credential is required")
	}

	return nil
}

// Target returns the host:port connection string.
func (c *Config) Target() string {
	return fmt.Sprintf("%s:%v", c.Host, c.Port)
}
