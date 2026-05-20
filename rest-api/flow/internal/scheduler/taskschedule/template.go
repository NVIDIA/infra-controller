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

package taskschedule

import (
	"encoding/json"
	"fmt"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/operation"
	taskcommon "github.com/NVIDIA/infra-controller-rest/flow/internal/task/common"
)

// TaskTemplate is the JSON-serialized operation stored in task_schedule.operation_template.
// It carries the operation type, code, and info needed to reconstruct an operation.Request
// at fire time. The target is resolved from task_schedule_scope rows at fire time,
// so it is not stored here.
type TaskTemplate struct {
	Type             taskcommon.TaskType `json:"type"`
	Code             string              `json:"code"`
	Info             json.RawMessage     `json:"information"`
	ConflictStrategy int                 `json:"conflict_strategy,omitempty"`
	QueueTimeoutSecs int64               `json:"queue_timeout_secs,omitempty"`
	RuleID           string              `json:"rule_id,omitempty"`
}

// TemplateOptions holds the scheduling-policy fields stored alongside the
// operation in a TaskTemplate. They are passed in at creation time and restored
// at fire time to reconstruct the full operation.Request.
type TemplateOptions struct {
	// ConflictStrategy is operation.ConflictStrategy stored as its underlying
	// int value. 0 = ConflictStrategyReject (default), 1 = ConflictStrategyQueue.
	ConflictStrategy int
	// QueueTimeoutSecs is operation.Request.QueueTimeout expressed in seconds.
	// Zero means use the server default.
	QueueTimeoutSecs int64
	// RuleID is the override rule UUID as a string. Empty string means no override.
	RuleID string
}

// WrapperFromTemplate unmarshals an operation_template JSON blob and returns
// the corresponding operation.Wrapper.
func WrapperFromTemplate(raw json.RawMessage) (operation.Wrapper, error) {
	var tmpl TaskTemplate
	if err := json.Unmarshal(raw, &tmpl); err != nil {
		return operation.Wrapper{}, fmt.Errorf("unmarshal operation_template: %w", err)
	}

	return operation.Wrapper{
		Type: tmpl.Type,
		Code: tmpl.Code,
		Info: tmpl.Info,
	}, nil
}

// MarshalTemplate serializes an operation into a TaskTemplate JSON blob
// suitable for storing in task_schedule.operation_template.
func MarshalTemplate(
	opType taskcommon.TaskType,
	code string,
	info json.RawMessage,
	opts TemplateOptions,
) (json.RawMessage, error) {
	tmpl := TaskTemplate{
		Type:             opType,
		Code:             code,
		Info:             info,
		ConflictStrategy: opts.ConflictStrategy,
		QueueTimeoutSecs: opts.QueueTimeoutSecs,
		RuleID:           opts.RuleID,
	}

	return json.Marshal(tmpl)
}

// OptionsFromTemplate extracts the scheduling-policy fields from a stored
// TaskTemplate JSON blob. These are restored at fire time to reconstruct the
// full operation.Request (conflict strategy, queue timeout, rule override).
func OptionsFromTemplate(raw json.RawMessage) (TemplateOptions, error) {
	var tmpl TaskTemplate
	if err := json.Unmarshal(raw, &tmpl); err != nil {
		return TemplateOptions{}, fmt.Errorf("unmarshal operation_template: %w", err)
	}

	return TemplateOptions{
		ConflictStrategy: tmpl.ConflictStrategy,
		QueueTimeoutSecs: tmpl.QueueTimeoutSecs,
		RuleID:           tmpl.RuleID,
	}, nil
}

// SummaryFromTemplate derives the operation_type and human-readable description
// that are surfaced on a TaskSchedule response. Both values are derived entirely
// from the stored operation_template so no live operation object is needed.
//
// opType is a stable SCREAMING_SNAKE_CASE string suitable for client filtering
// (e.g. "POWER_ON", "BRING_UP"). description is a short English phrase for
// display (e.g. "Power Off (forced)", "Upgrade Firmware to v2.1.0").
func SummaryFromTemplate(templateJSON json.RawMessage) (opType, description string, err error) {
	var tmpl TaskTemplate
	if err = json.Unmarshal(templateJSON, &tmpl); err != nil {
		return "", "", fmt.Errorf("unmarshal operation_template: %w", err)
	}

	switch tmpl.Type {
	case taskcommon.TaskTypePowerControl:
		switch tmpl.Code {
		case taskcommon.OpCodePowerControlPowerOn, taskcommon.OpCodePowerControlForcePowerOn:
			return "POWER_ON", "Power On", nil
		case taskcommon.OpCodePowerControlPowerOff:
			return "POWER_OFF", "Power Off", nil
		case taskcommon.OpCodePowerControlForcePowerOff:
			return "POWER_OFF", "Power Off (forced)", nil
		case taskcommon.OpCodePowerControlRestart, taskcommon.OpCodePowerControlWarmReset:
			return "POWER_RESET", "Power Reset", nil
		case taskcommon.OpCodePowerControlForceRestart, taskcommon.OpCodePowerControlColdReset:
			return "POWER_RESET", "Power Reset (forced)", nil
		default:
			return "POWER_CONTROL", tmpl.Code, nil
		}

	case taskcommon.TaskTypeBringUp:
		if tmpl.Code == taskcommon.OpCodeIngest {
			return "INGEST", "Ingest", nil
		}
		return "BRING_UP", "Bring Up", nil

	case taskcommon.TaskTypeFirmwareControl:
		var info struct {
			TargetVersion string `json:"target_version"`
		}
		if len(tmpl.Info) > 0 {
			if err = json.Unmarshal(tmpl.Info, &info); err != nil {
				return "", "", fmt.Errorf("unmarshal firmware info: %w", err)
			}
		}
		switch tmpl.Code {
		case taskcommon.OpCodeFirmwareControlUpgrade:
			if info.TargetVersion != "" {
				return "UPGRADE_FIRMWARE", "Upgrade Firmware to " + info.TargetVersion, nil
			}
			return "UPGRADE_FIRMWARE", "Upgrade Firmware", nil
		case taskcommon.OpCodeFirmwareControlDowngrade:
			if info.TargetVersion != "" {
				return "DOWNGRADE_FIRMWARE", "Downgrade Firmware to " + info.TargetVersion, nil
			}
			return "DOWNGRADE_FIRMWARE", "Downgrade Firmware", nil
		case taskcommon.OpCodeFirmwareControlRollback:
			return "ROLLBACK_FIRMWARE", "Rollback Firmware", nil
		default: // unrecognized
			return "FIRMWARE_CONTROL", tmpl.Code, nil
		}

	default:
		return string(tmpl.Type), tmpl.Code, nil
	}
}
