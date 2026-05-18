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

package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAPIDeprecation(t *testing.T) {
	oldValue := "blockSize"
	newValue := "prefixLength"

	preEffectTime := time.Now().Add(24 * time.Hour)
	postEffectTime := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name         string
		oldValue     string
		newValue     string
		fieldType    string
		takeActionBy time.Time
		expect       APIDeprecation
	}{
		{
			name:         "test new API deprecation - pre-deprecation notice",
			oldValue:     oldValue,
			newValue:     newValue,
			fieldType:    DeprecationTypeAttribute,
			takeActionBy: preEffectTime,
			expect: APIDeprecation{
				Notice:       fmt.Sprintf(deprecationPreTemplate, oldValue, fmt.Sprintf(" in favor of '%s'", newValue)),
				Attribute:    &oldValue,
				TakeActionBy: preEffectTime,
				ReplacedBy:   &newValue,
			},
		},
		{
			name:         "test new API deprecation - post-deprecation notice",
			oldValue:     oldValue,
			newValue:     newValue,
			fieldType:    DeprecationTypeAttribute,
			takeActionBy: postEffectTime,
			expect: APIDeprecation{
				Notice:       fmt.Sprintf(deprecationPostTemplate, oldValue, fmt.Sprintf(" in favor of '%s'", newValue)),
				Attribute:    &oldValue,
				TakeActionBy: postEffectTime,
				ReplacedBy:   &newValue,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAPIDeprecation(DeprecatedEntity{
				OldValue:     tt.oldValue,
				NewValue:     &tt.newValue,
				Type:         tt.fieldType,
				TakeActionBy: tt.takeActionBy,
			})
			assert.EqualValues(t, tt.expect, got)
		})
	}
}
