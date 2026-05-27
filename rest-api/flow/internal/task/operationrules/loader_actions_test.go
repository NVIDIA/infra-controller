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
	"os"
	"testing"
	"time"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/common"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

func TestYAMLRuleLoader_ActionBasedConfiguration(t *testing.T) {
	// Create temporary YAML file with action-based configuration
	// Note: Must include all required operations for validation to pass
	yamlContent := `version: v1
rules:
  - name: "Power On with Actions"
    description: "Power on with action-based configuration"
    operation_type: power_control
    operation: power_on
    steps:
      - component_type: powershelf
        stage: 1
        max_parallel: 1
        timeout: 15m
        retry:
          max_attempts: 3
          initial_interval: 5s
          backoff_coefficient: 2.0
        main_operation:
          name: PowerControl
        post_operation:
          - name: VerifyPowerStatus
            timeout: 15s
            poll_interval: 5s
            parameters:
              expected_status: "on"
          - name: VerifyReachability
            timeout: 3m
            poll_interval: 10s
            parameters:
              component_types: ["compute", "nvlswitch"]
          - name: Sleep
            parameters:
              duration: 30s

      - component_type: nvlswitch
        stage: 2
        max_parallel: 4
        timeout: 15m
        main_operation:
          name: PowerControl
        post_operation:
          - name: Sleep
            parameters:
              duration: 15s

      - component_type: compute
        stage: 3
        max_parallel: 8
        timeout: 20m
        pre_operation:
          - name: Sleep
            parameters:
              duration: 10s
        main_operation:
          name: PowerControl
        post_operation:
          - name: VerifyPowerStatus
            timeout: 15s
            poll_interval: 5s
            parameters:
              expected_status: "on"

  - name: "Power Off with Actions"
    operation_type: power_control
    operation: power_off
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        timeout: 10m
        main_operation:
          name: PowerControl

  - name: "Restart with Actions"
    operation_type: power_control
    operation: restart
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        timeout: 10m
        main_operation:
          name: PowerControl
`

	tmpfile, err := os.CreateTemp("", "test-action-rules-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Load rules
	loader, err := NewYAMLRuleLoader(tmpfile.Name())
	if err != nil {
		t.Fatalf("NewYAMLRuleLoader() error = %v", err)
	}

	rules, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify rules were loaded
	powerRules, ok := rules[common.TaskTypePowerControl]
	if !ok {
		t.Fatal("No power_control rules loaded")
	}

	rule, ok := powerRules[SequencePowerOn]
	if !ok {
		t.Fatal("No power_on rule loaded")
	}

	// Verify rule metadata
	if rule.Name != "Power On with Actions" {
		t.Errorf("rule.Name = %q, want %q", rule.Name, "Power On with Actions")
	}

	// Verify steps
	if len(rule.RuleDefinition.Steps) != 3 {
		t.Fatalf("len(steps) = %d, want 3", len(rule.RuleDefinition.Steps))
	}

	// Verify first step (powershelf)
	step0 := rule.RuleDefinition.Steps[0]
	if step0.ComponentType != devicetypes.ComponentTypePowerShelf {
		t.Errorf("step[0].ComponentType = %v, want ComponentTypePowerShelf",
			step0.ComponentType)
	}

	// Verify main operation
	if step0.MainOperation.Name != ActionPowerControl {
		t.Errorf("step[0].MainOperation.Name = %q, want %q",
			step0.MainOperation.Name, ActionPowerControl)
	}

	// Verify post-operation actions
	if len(step0.PostOperation) != 3 {
		t.Fatalf("len(step[0].PostOperation) = %d, want 3",
			len(step0.PostOperation))
	}

	// Verify VerifyPowerStatus action
	verifyAction := step0.PostOperation[0]
	if verifyAction.Name != ActionVerifyPowerStatus {
		t.Errorf("PostOperation[0].Name = %q, want %q",
			verifyAction.Name, ActionVerifyPowerStatus)
	}
	if verifyAction.Timeout != 15*time.Second {
		t.Errorf("PostOperation[0].Timeout = %v, want 15s",
			verifyAction.Timeout)
	}
	if verifyAction.PollInterval != 5*time.Second {
		t.Errorf("PostOperation[0].PollInterval = %v, want 5s",
			verifyAction.PollInterval)
	}
	if status, ok := verifyAction.Parameters[ParamExpectedStatus]; !ok ||
		status != "on" {
		t.Errorf("PostOperation[0].Parameters[expected_status] = %v, want 'on'",
			status)
	}

	// Verify VerifyReachability action
	reachAction := step0.PostOperation[1]
	if reachAction.Name != ActionVerifyReachability {
		t.Errorf("PostOperation[1].Name = %q, want %q",
			reachAction.Name, ActionVerifyReachability)
	}
	if reachAction.Timeout != 3*time.Minute {
		t.Errorf("PostOperation[1].Timeout = %v, want 3m",
			reachAction.Timeout)
	}
	if reachAction.PollInterval != 10*time.Second {
		t.Errorf("PostOperation[1].PollInterval = %v, want 10s",
			reachAction.PollInterval)
	}

	// Verify Sleep action
	sleepAction := step0.PostOperation[2]
	if sleepAction.Name != ActionSleep {
		t.Errorf("PostOperation[2].Name = %q, want %q",
			sleepAction.Name, ActionSleep)
	}
	durationParam := sleepAction.Parameters[ParamDuration]
	if durationParam == nil {
		t.Error("PostOperation[2].Parameters[duration] is nil")
	} else {
		// Should be parsed as time.Duration
		if dur, ok := durationParam.(time.Duration); !ok {
			t.Errorf("PostOperation[2].Parameters[duration] type = %T, want time.Duration", //nolint
				durationParam)
		} else if dur != 30*time.Second {
			t.Errorf("PostOperation[2].Parameters[duration] = %v, want 30s",
				dur)
		}
	}

	// Verify third step (compute) has pre-operation
	step2 := rule.RuleDefinition.Steps[2]
	if step2.ComponentType != devicetypes.ComponentTypeCompute {
		t.Errorf("step[2].ComponentType = %v, want ComponentTypeCompute",
			step2.ComponentType)
	}

	if len(step2.PreOperation) != 1 {
		t.Fatalf("len(step[2].PreOperation) = %d, want 1",
			len(step2.PreOperation))
	}

	preAction := step2.PreOperation[0]
	if preAction.Name != ActionSleep {
		t.Errorf("PreOperation[0].Name = %q, want %q",
			preAction.Name, ActionSleep)
	}
}

func TestYAMLRuleLoader_InvalidDurations(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid action timeout",
			yaml: `version: v1
rules:
  - name: "Invalid Rule"
    operation_type: power_control
    operation: power_on
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: VerifyPowerStatus
          timeout: "10xyz"
          poll_interval: 5s
          parameters:
            expected_status: "on"
  - name: "Power Off"
    operation_type: power_control
    operation: power_off
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: PowerControl
  - name: "Restart"
    operation_type: power_control
    operation: restart
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: PowerControl
`,
			wantErr: true,
			errMsg:  "invalid timeout ('10xyz')",
		},
		{
			name: "invalid action poll_interval",
			yaml: `version: v1
rules:
  - name: "Invalid Rule"
    operation_type: power_control
    operation: power_on
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: VerifyPowerStatus
          timeout: 15s
          poll_interval: "bad"
          parameters:
            expected_status: "on"
  - name: "Power Off"
    operation_type: power_control
    operation: power_off
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: PowerControl
  - name: "Restart"
    operation_type: power_control
    operation: restart
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: PowerControl
`,
			wantErr: true,
			errMsg:  "invalid poll_interval ('bad')",
		},
		{
			name: "invalid duration parameter",
			yaml: `version: v1
rules:
  - name: "Invalid Rule"
    operation_type: power_control
    operation: power_on
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: Sleep
          parameters:
            duration: "notaduration"
  - name: "Power Off"
    operation_type: power_control
    operation: power_off
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: PowerControl
  - name: "Restart"
    operation_type: power_control
    operation: restart
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: PowerControl
`,
			wantErr: true,
			errMsg:  "invalid duration parameter for action 'Sleep' ('notaduration')",
		},
		{
			name: "invalid step timeout",
			yaml: `version: v1
rules:
  - name: "Invalid Rule"
    operation_type: power_control
    operation: power_on
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        timeout: "invalid"
        main_operation:
          name: PowerControl
  - name: "Power Off"
    operation_type: power_control
    operation: power_off
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: PowerControl
  - name: "Restart"
    operation_type: power_control
    operation: restart
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: PowerControl
`,
			wantErr: true,
			errMsg:  "invalid timeout ('invalid')",
		},
		{
			name: "invalid retry initial_interval",
			yaml: `version: v1
rules:
  - name: "Invalid Rule"
    operation_type: power_control
    operation: power_on
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        retry:
          max_attempts: 3
          initial_interval: "bad"
          backoff_coefficient: 2.0
        main_operation:
          name: PowerControl
  - name: "Power Off"
    operation_type: power_control
    operation: power_off
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: PowerControl
  - name: "Restart"
    operation_type: power_control
    operation: restart
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: PowerControl
`,
			wantErr: true,
			errMsg:  "invalid initial_interval ('bad')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test-invalid-duration-*.yaml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.yaml)); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			loader, err := NewYAMLRuleLoader(tmpfile.Name())
			if err != nil {
				t.Fatalf("NewYAMLRuleLoader() error = %v", err)
			}

			_, err = loader.Load()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Load() error = %v, want error containing %q",
						err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestYAMLRuleLoader_ActionValidation(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing required parameter",
			yaml: `version: v1
rules:
  - name: "Invalid Rule"
    operation_type: power_control
    operation: power_on
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: VerifyPowerStatus
          parameters:
            # missing expected_status
`,
			wantErr: true,
			errMsg:  "missing required parameter",
		},
		{
			name: "missing timeout for action requiring it",
			yaml: `version: v1
rules:
  - name: "Invalid Rule"
    operation_type: power_control
    operation: power_on
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: VerifyPowerStatus
          poll_interval: 5s
          parameters:
            expected_status: "on"
          # missing timeout
`,
			wantErr: true,
			errMsg:  "requires timeout",
		},
		{
			name: "unknown action",
			yaml: `version: v1
rules:
  - name: "Invalid Rule"
    operation_type: power_control
    operation: power_on
    steps:
      - component_type: compute
        stage: 1
        max_parallel: 1
        main_operation:
          name: UnknownAction
`,
			wantErr: true,
			errMsg:  "unknown action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test-invalid-*.yaml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.yaml)); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			loader, err := NewYAMLRuleLoader(tmpfile.Name())
			if err != nil {
				t.Fatalf("NewYAMLRuleLoader() error = %v", err)
			}

			_, err = loader.Load()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Load() error = %v, want error containing %q",
						err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
