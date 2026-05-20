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

package executor

import (
	"context"
	"fmt"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/common"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/task"
)

// Executor is the engine-agnostic interface for executing tasks. Implementations
// hide all engine-specific details (Temporal, local, mock) from the task manager.
type Executor interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Type() common.ExecutorType
	// Execute dispatches a task to the appropriate handler based on
	// req.Info.OperationType. The operation payload must be pre-serialized
	// into req.Info.OperationInfo before calling.
	Execute(ctx context.Context, req *task.ExecutionRequest) (*task.ExecutionResponse, error)
	CheckStatus(ctx context.Context, executionID string) (common.TaskStatus, error)
	TerminateTask(ctx context.Context, executionID string, reason string) error
}

// ExecutorConfig is implemented by engine-specific configuration structs.
// Build is called once at startup to construct the live Executor.
type ExecutorConfig interface {
	Validate() error
	// Build constructs the Executor. updater receives task status transitions
	// from the execution engine back to the store. It must not be nil.
	Build(ctx context.Context, updater task.TaskStatusUpdater) (Executor, error)
}

// New validates the config and builds the Executor, wiring updater so the
// engine can report task status changes without importing store packages.
func New(
	ctx context.Context,
	executorConfig ExecutorConfig,
	updater task.TaskStatusUpdater,
) (Executor, error) {
	if executorConfig == nil {
		return nil, fmt.Errorf("executor config is required")
	}

	if err := executorConfig.Validate(); err != nil {
		return nil, err
	}

	if updater == nil {
		return nil, fmt.Errorf("task status updater is required")
	}

	return executorConfig.Build(ctx, updater)
}
