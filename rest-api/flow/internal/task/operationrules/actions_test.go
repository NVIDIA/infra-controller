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

package operationrules

import (
	"testing"
	"time"
)

func TestActionConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ActionConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid Sleep action",
			config: ActionConfig{
				Name: ActionSleep,
				Parameters: map[string]any{
					ParamDuration: "30s",
				},
			},
			wantErr: false,
		},
		{
			name: "Sleep missing duration",
			config: ActionConfig{
				Name:       ActionSleep,
				Parameters: map[string]any{},
			},
			wantErr: true,
			errMsg:  "missing required parameter: duration",
		},
		{
			name: "valid VerifyPowerStatus action",
			config: ActionConfig{
				Name:         ActionVerifyPowerStatus,
				Timeout:      15 * time.Second,
				PollInterval: 5 * time.Second,
				Parameters: map[string]any{
					ParamExpectedStatus: "on",
				},
			},
			wantErr: false,
		},
		{
			name: "VerifyPowerStatus missing timeout",
			config: ActionConfig{
				Name:         ActionVerifyPowerStatus,
				PollInterval: 5 * time.Second,
				Parameters: map[string]any{
					ParamExpectedStatus: "on",
				},
			},
			wantErr: true,
			errMsg:  "requires timeout",
		},
		{
			name: "VerifyPowerStatus missing poll_interval",
			config: ActionConfig{
				Name:    ActionVerifyPowerStatus,
				Timeout: 15 * time.Second,
				Parameters: map[string]any{
					ParamExpectedStatus: "on",
				},
			},
			wantErr: true,
			errMsg:  "requires poll_interval",
		},
		{
			name: "VerifyPowerStatus invalid expected_status",
			config: ActionConfig{
				Name:         ActionVerifyPowerStatus,
				Timeout:      15 * time.Second,
				PollInterval: 5 * time.Second,
				Parameters: map[string]any{
					ParamExpectedStatus: "invalid",
				},
			},
			wantErr: true,
			errMsg:  "must be 'on' or 'off'",
		},
		{
			name: "valid VerifyReachability action",
			config: ActionConfig{
				Name:         ActionVerifyReachability,
				Timeout:      3 * time.Minute,
				PollInterval: 10 * time.Second,
				Parameters: map[string]any{
					ParamComponentTypes: []string{"compute", "nvlswitch"},
				},
			},
			wantErr: false,
		},
		{
			name: "VerifyReachability invalid component type",
			config: ActionConfig{
				Name:         ActionVerifyReachability,
				Timeout:      3 * time.Minute,
				PollInterval: 10 * time.Second,
				Parameters: map[string]any{
					ParamComponentTypes: []string{"invalid_type"},
				},
			},
			wantErr: true,
			errMsg:  "invalid component type",
		},
		{
			name: "valid PowerControl action",
			config: ActionConfig{
				Name: ActionPowerControl,
			},
			wantErr: false,
		},
		{
			name: "unknown action",
			config: ActionConfig{
				Name: "UnknownAction",
			},
			wantErr: true,
			errMsg:  "unknown action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ActionConfig.Validate() error = nil, wantErr %v",
						tt.wantErr)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf(
						"ActionConfig.Validate() error = %v, want error containing %q",
						err,
						tt.errMsg,
					)
				}
			} else {
				if err != nil {
					t.Errorf("ActionConfig.Validate() error = %v, wantErr %v",
						err, tt.wantErr)
				}
			}
		})
	}
}

func TestActionConfig_ValidateParameters(t *testing.T) {
	tests := []struct {
		name    string
		config  ActionConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "Sleep with time.Duration parameter",
			config: ActionConfig{
				Name: ActionSleep,
				Parameters: map[string]any{
					ParamDuration: 30 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "Sleep with numeric parameter",
			config: ActionConfig{
				Name: ActionSleep,
				Parameters: map[string]any{
					ParamDuration: 30.0,
				},
			},
			wantErr: false,
		},
		{
			name: "VerifyReachability with []any",
			config: ActionConfig{
				Name:         ActionVerifyReachability,
				Timeout:      3 * time.Minute,
				PollInterval: 10 * time.Second,
				Parameters: map[string]any{
					ParamComponentTypes: []any{"compute", "nvlswitch"},
				},
			},
			wantErr: false,
		},
		{
			name: "VerifyReachability component_types not array",
			config: ActionConfig{
				Name:         ActionVerifyReachability,
				Timeout:      3 * time.Minute,
				PollInterval: 10 * time.Second,
				Parameters: map[string]any{
					ParamComponentTypes: "compute",
				},
			},
			wantErr: true,
			errMsg:  "must be array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf(
						"ActionConfig.Validate() error = nil, wantErr %v",
						tt.wantErr,
					)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf(
						"ActionConfig.Validate() error = %v, want error containing %q", //nolint
						err,
						tt.errMsg,
					)
				}
			} else {
				if err != nil {
					t.Errorf(
						"ActionConfig.Validate() error = %v, wantErr %v",
						err,
						tt.wantErr,
					)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
