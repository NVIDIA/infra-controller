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

package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

func TestExtractComponentTargetVersion(t *testing.T) {
	tests := map[string]struct {
		rawVersion    string
		componentType devicetypes.ComponentType
		expected      string
	}{
		"empty string returns empty": {
			rawVersion:    "",
			componentType: devicetypes.ComponentTypeCompute,
			expected:      "",
		},
		"layered JSON — compute section extracted": {
			rawVersion:    `{"compute":{"bmc":"7.10.30","uefi":"2.22.1"},"nvlswitch":{"nvos":"1.2.3"}}`,
			componentType: devicetypes.ComponentTypeCompute,
			expected:      `{"bmc":"7.10.30","uefi":"2.22.1"}`,
		},
		"layered JSON — nvlswitch section extracted": {
			rawVersion:    `{"compute":{"bmc":"7.10.30"},"nvlswitch":{"nvos":"1.2.3","cpld":"4.5.6"}}`,
			componentType: devicetypes.ComponentTypeNVLSwitch,
			expected:      `{"nvos":"1.2.3","cpld":"4.5.6"}`,
		},
		"layered JSON — powershelf section extracted": {
			rawVersion:    `{"compute":{"bmc":"7.10.30"},"powershelf":{"firmware":"1.0.0"}}`,
			componentType: devicetypes.ComponentTypePowerShelf,
			expected:      `{"firmware":"1.0.0"}`,
		},
		"layered JSON — missing key returns empty (component omitted)": {
			rawVersion:    `{"compute":{"bmc":"7.10.30"},"nvlswitch":{"nvos":"1.2.3"}}`,
			componentType: devicetypes.ComponentTypePowerShelf,
			expected:      "",
		},
		"layered JSON — string scalar value is unquoted": {
			rawVersion:    `{"compute":{"bmc":"7.10.30"},"nvlswitch":"2.0.0"}`,
			componentType: devicetypes.ComponentTypeNVLSwitch,
			expected:      "2.0.0",
		},
		"layered JSON — string scalar with escapes is unquoted": {
			rawVersion:    `{"nvlswitch":"r1.3.9-alpha"}`,
			componentType: devicetypes.ComponentTypeNVLSwitch,
			expected:      "r1.3.9-alpha",
		},
		"old flat JSON — no known keys, returns as-is for backward compat": {
			rawVersion:    `{"bmc":"7.10.30","uefi":"2.22.1"}`,
			componentType: devicetypes.ComponentTypeCompute,
			expected:      `{"bmc":"7.10.30","uefi":"2.22.1"}`,
		},
		"non-JSON string — returns as-is": {
			rawVersion:    "2.0.0",
			componentType: devicetypes.ComponentTypeNVLSwitch,
			expected:      "2.0.0",
		},
		"only one component type present — other types get empty": {
			rawVersion:    `{"compute":{"bmc":"7.10.30"}}`,
			componentType: devicetypes.ComponentTypeNVLSwitch,
			expected:      "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := extractComponentTargetVersion(tc.rawVersion, tc.componentType)
			assert.Equal(t, tc.expected, result)
		})
	}
}
