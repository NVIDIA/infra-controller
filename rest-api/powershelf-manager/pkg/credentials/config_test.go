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
package credentials

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	testCases := map[string]struct {
		cfg         Config
		expectErr   bool
		errContains string
	}{
		"in-memory datastore returns nil": {
			cfg:       Config{DataStoreType: DatastoreTypeInMemory},
			expectErr: false,
		},
		"unknown/empty datastore returns nil": {
			cfg:       Config{}, // zero-value DataStoreType means no extra validation
			expectErr: false,
		},
		"vault datastore with nil VaultConfig returns error": {
			cfg: Config{
				DataStoreType: DatastoreTypeVault,
				VaultConfig:   nil,
			},
			expectErr:   true,
			errContains: "vault config needs to be specified",
		},
		"vault datastore with non-nil VaultConfig delegates to VaultConfig.Validate": {
			cfg: Config{
				DataStoreType: DatastoreTypeVault,
				VaultConfig:   &VaultConfig{Address: "http://127.0.0.1", Token: "x"},
			},
			expectErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.cfg.Validate()

			if tc.errContains != "" {
				assert.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.errContains))
				return
			}

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
