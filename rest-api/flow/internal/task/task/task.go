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

package task

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/operation"
	taskcommon "github.com/NVIDIA/infra-controller-rest/flow/internal/task/common"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/operationrules"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

// Task defines the details of a task. It includes:
// -- ID: The unique identifier of the task.
// -- Operation: The operation to be performed by the task.
// -- RackID: The rack this task operates on (1 task = 1 rack).
// -- Attributes: Flexible metadata including targeted components by type.
// -- Description: The description of the task provided by the user.
// -- ExecutorType: The type of executor to be used for the task.
// -- ExecutionID: The identifier of the execution of the task.
// -- Status: The status of the task.
// -- Message: Status message or error details.
// -- AppliedRuleID: The ID of the operation rule that was applied (if any).
type Task struct {
	ID            uuid.UUID
	Operation     operation.Wrapper
	RackID        uuid.UUID // The rack this task operates on (1 task = 1 rack)
	Attributes    taskcommon.TaskAttributes
	Description   string
	ExecutorType  taskcommon.ExecutorType
	ExecutionID   string
	Status        taskcommon.TaskStatus
	Message       string
	AppliedRuleID *uuid.UUID // The ID of the operation rule that was applied
	CreatedAt     time.Time
	UpdatedAt     time.Time
	StartedAt     *time.Time
	FinishedAt    *time.Time

	// QueueExpiresAt is the deadline for a waiting task to be promoted.
	// After this time the Promoter terminates the task automatically.
	// Nil for non-waiting tasks.
	QueueExpiresAt *time.Time
}

// WorkflowComponent holds the minimal component data needed to execute
// a workflow. All fields are plain JSON-safe types.
type WorkflowComponent struct {
	Type        devicetypes.ComponentType `json:"type"`
	ComponentID string                    `json:"component_id"`
}

// ExecutionInfo contains the information needed to execute a task.
type ExecutionInfo struct {
	TaskID     uuid.UUID
	Components []WorkflowComponent

	// RuleDefinition is the resolved operation rule, determined at task
	// creation time and carried through to the workflow unchanged.
	RuleDefinition *operationrules.RuleDefinition

	// OperationType identifies which workflow to dispatch to. The executor
	// looks up the registered WorkflowDescriptor by this value and submits
	// the workflow by its stable Temporal name.
	OperationType taskcommon.TaskType

	// OperationInfo is the serialized operation-specific payload. The
	// executor passes it to the WorkflowDescriptor's Unmarshal function,
	// which deserializes and validates it before the workflow starts.
	OperationInfo json.RawMessage
}

// ExecutionRequest holds the parameters for submitting a task for execution.
type ExecutionRequest struct {
	Info  ExecutionInfo
	Async bool
}

// ExecutionResponse holds the result of a task execution submission.
type ExecutionResponse struct {
	ExecutionID string
}

// Validate returns an error if the ExecutionRequest is missing required fields.
func (r *ExecutionRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("request is nil")
	}

	if r.Info.TaskID == uuid.Nil {
		return fmt.Errorf("task ID is nil")
	}

	if !r.Info.OperationType.IsValid() {
		return fmt.Errorf("operation type is invalid or not set")
	}

	if len(r.Info.OperationInfo) == 0 {
		return fmt.Errorf("operation info is empty")
	}

	if !json.Valid(r.Info.OperationInfo) {
		return fmt.Errorf("operation info is not valid JSON")
	}

	if len(r.Info.Components) == 0 {
		return fmt.Errorf("components list is empty")
	}

	return nil
}

// IsValid reports whether the ExecutionResponse contains a non-empty execution ID.
func (r *ExecutionResponse) IsValid() bool {
	if r == nil {
		return false
	}

	if r.ExecutionID == "" {
		return false
	}

	return true
}

// TaskStatusUpdate carries the fields needed to update a task's status.
type TaskStatusUpdate struct {
	ID      uuid.UUID
	Status  taskcommon.TaskStatus
	Message string
}

// TaskStatusUpdater is implemented by any store that can persist task status changes.
type TaskStatusUpdater interface {
	// UpdateTaskStatus persists the status change described by arg.
	UpdateTaskStatus(ctx context.Context, arg *TaskStatusUpdate) error
}
