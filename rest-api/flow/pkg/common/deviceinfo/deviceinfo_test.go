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

package deviceinfo

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDeviceInfo_InfoMsg(t *testing.T) {
	di := DeviceInfo{
		ID:           uuid.New(),
		Name:         "Device1",
		Manufacturer: "NVIDIA",
		Model:        "ModelX",
		SerialNumber: "12345",
		Description:  "A test device",
	}

	typ := "TestDevice"

	testCases := map[string]struct {
		byID     bool
		expected string
	}{
		"info message based on ID": {
			byID:     true,
			expected: fmt.Sprintf("%s [id: %s]", typ, di.ID.String()),
		},
		"inf message based on serial information": {
			byID:     false,
			expected: fmt.Sprintf("%s [manufacturer: %s, serial: %s]", typ, di.Manufacturer, di.SerialNumber), //nolint
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, di.InfoMsg(typ, testCase.byID))
		})
	}
}

func TestDeviceInfo_NewRandom(t *testing.T) {
	di := NewRandom("for-testing", 12)
	assert.Equal(t, "for-testing", di.Name)
	assert.Equal(t, 12, len(di.SerialNumber))
}
