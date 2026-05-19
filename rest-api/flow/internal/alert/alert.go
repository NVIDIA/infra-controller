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

// Package alert provides an abstraction for sending alerts/notifications
// from Flow workflows and activities. The concrete implementation can be
// replaced later (Slack, PagerDuty, etc.).
package alert

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
)

// Severity represents the urgency level of an alert.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Alert represents a single alert to be sent through the alerting system.
type Alert struct {
	Severity  Severity          `json:"severity"`
	Message   string            `json:"message"`
	Component string            `json:"component,omitempty"`
	Operation string            `json:"operation,omitempty"`
	TaskID    string            `json:"task_id,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
}

func (a Alert) String() string {
	return fmt.Sprintf("[%s] %s (component=%s, operation=%s, task=%s)",
		a.Severity, a.Message, a.Component, a.Operation, a.TaskID)
}

// Send delivers an alert. Currently just logs it.
// TODO: Replace with real alerting backend (Slack, PagerDuty, etc.) when ready.
func Send(_ context.Context, a Alert) error {
	log.Warn().
		Str("severity", string(a.Severity)).
		Str("component", a.Component).
		Str("operation", a.Operation).
		Str("task_id", a.TaskID).
		Msg("ALERT: " + a.Message)
	return nil
}
