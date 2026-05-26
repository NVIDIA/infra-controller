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
	"reflect"
	"testing"
)

func TestNewAPIHealthCheck(t *testing.T) {
	type args struct {
		isHealthy    bool
		errorMessage *string
	}
	tests := []struct {
		name string
		args args
		want *APIHealthCheck
	}{
		{
			name: "test initializing API model for HealthCheck",
			args: args{
				isHealthy:    true,
				errorMessage: nil,
			},
			want: &APIHealthCheck{
				IsHealthy: true,
				Error:     nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAPIHealthCheck(tt.args.isHealthy, tt.args.errorMessage); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAPIHealthCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}
