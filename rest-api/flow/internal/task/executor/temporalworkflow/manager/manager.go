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

package manager

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	temporalactivity "go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	temporalworkflow "go.temporal.io/sdk/workflow"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/clients/temporal"
	taskcommon "github.com/NVIDIA/infra-controller-rest/flow/internal/task/common"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/executor"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/executor/temporalworkflow/activity"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/executor/temporalworkflow/common"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/executor/temporalworkflow/workflow"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/task"
)

const (
	WorkflowQueue = "flow-tasks"
)

// Config holds all configuration required to build a Temporal-backed executor.
// WorkerOptions maps Temporal task-queue names to per-queue worker settings;
// each key results in a dedicated worker started by Build.
type Config struct {
	ClientConf    temporal.Config
	WorkerOptions map[string]worker.Options

	// ComponentManagerRegistry is the registry containing initialized component managers.
	ComponentManagerRegistry *componentmanager.Registry
}

// Validate checks that the configuration is complete and consistent.
// It validates the Temporal client config.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("configuration for Temporal Executor is nil")
	}

	if err := c.ClientConf.Validate(); err != nil {
		return err
	}

	if len(c.WorkerOptions) == 0 {
		return errors.New("at least one Temporal worker queue must be configured")
	}

	if _, ok := c.WorkerOptions[WorkflowQueue]; !ok {
		return fmt.Errorf(
			"worker options must include workflow queue %q",
			WorkflowQueue,
		)
	}

	return nil
}

// Manager is the Temporal implementation of executor.Executor. It owns two
// Temporal clients (publisher for workflow submission and status queries,
// subscriber for worker registration) and one worker per configured task queue.
type Manager struct {
	conf             Config
	publisherClient  *temporal.Client
	subscriberClient *temporal.Client
	workers          map[string]worker.Worker
}

// Build creates the Temporal executor: it wires the status updater and component
// manager registry into the activity layer, then starts the Temporal clients and
// workers for each configured queue. On success the caller must eventually call
// Stop() to release the Temporal client connections — typically via defer,
// regardless of whether Start() succeeds.
func (c *Config) Build(
	ctx context.Context,
	updater task.TaskStatusUpdater,
) (executor.Executor, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	if updater == nil {
		return nil, errors.New("task status updater is required")
	}

	if c.ComponentManagerRegistry == nil {
		return nil, errors.New("component manager registry is required")
	}

	// Bind dependencies into an Activities instance so each manager has its
	// own isolated copy — no shared mutable globals between managers.
	acts := activity.New(updater, c.ComponentManagerRegistry)

	publisherClient, err := temporal.New(c.ClientConf)
	if err != nil {
		return nil, err
	}

	subscriberClient, err := temporal.New(c.ClientConf)
	if err != nil {
		publisherClient.Client().Close()
		return nil, err
	}

	allActivities := acts.All()
	allWorkflows := workflow.GetAllWorkflows()
	workers := make(map[string]worker.Worker)
	for queue, options := range c.WorkerOptions {
		worker := worker.New(subscriberClient.Client(), queue, options)
		for name, fn := range allActivities {
			worker.RegisterActivityWithOptions(
				fn,
				temporalactivity.RegisterOptions{Name: name},
			)
		}

		for _, wf := range allWorkflows {
			worker.RegisterWorkflowWithOptions(
				wf.WorkflowFunc,
				temporalworkflow.RegisterOptions{Name: wf.WorkflowName},
			)
		}

		workers[queue] = worker
	}

	return &Manager{
		conf:             *c,
		publisherClient:  publisherClient,
		subscriberClient: subscriberClient,
		workers:          workers,
	}, nil
}

// Start begins polling for workflow and activity tasks on all configured queues.
func (m *Manager) Start(ctx context.Context) error {
	started := make([]worker.Worker, 0, len(m.workers))
	for queue, worker := range m.workers {
		log.Info().Msgf("Starting temporal worker for queue %s", queue)
		if err := worker.Start(); err != nil {
			for i := len(started) - 1; i >= 0; i-- {
				started[i].Stop()
			}
			// Do not close publisherClient/subscriberClient here: they are
			// owned by the Manager (created in Build, not in Start) and must
			// remain open until Stop() is called. The caller is expected to
			// defer Stop() immediately after a successful Build(), so Stop()
			// will run even when Start() returns an error.
			return fmt.Errorf("failed to start temporal worker: %w", err)
		}
		started = append(started, worker)
		log.Info().Msgf("Temporal worker started for queue %s", queue)
	}

	return nil
}

// Stop is the full teardown for a Manager created by Build: it stops all
// workers (safe to call even if Start was never called or failed partway) and
// closes the Temporal client connections.
func (m *Manager) Stop(ctx context.Context) error {
	for queue, worker := range m.workers {
		log.Info().Msgf("Stopping temporal worker for queue %s", queue)
		worker.Stop()
		log.Info().Msgf("Temporal worker stopped for queue %s", queue)
	}

	m.publisherClient.Client().Close()
	m.subscriberClient.Client().Close()

	return nil
}

// Type returns ExecutorTypeTemporal, identifying this executor implementation.
func (m *Manager) Type() taskcommon.ExecutorType {
	return taskcommon.ExecutorTypeTemporal
}

// CheckStatus decodes the execution ID and queries Temporal for the current
// workflow execution status, mapping it to a TaskStatus.
func (m *Manager) CheckStatus(
	ctx context.Context,
	encodedExecutionID string,
) (taskcommon.TaskStatus, error) {
	executionID, err := common.NewFromEncoded(encodedExecutionID)
	if err != nil {
		return taskcommon.TaskStatusUnknown, err
	}

	// Use empty runID to get the latest execution.
	resp, err := m.publisherClient.Client().DescribeWorkflowExecution(
		ctx,
		executionID.WorkflowID,
		"",
	)
	if err != nil {
		return taskcommon.TaskStatusUnknown, fmt.Errorf(
			"failed to describe temporal workflow execution %s: %v",
			executionID.String(),
			err,
		)
	}

	return taskStatusFromTemporalWorkflowStatus(
		resp.GetWorkflowExecutionInfo().GetStatus(),
	), nil
}

// TerminateTask terminates the Temporal workflow backing the given execution ID.
func (m *Manager) TerminateTask(
	ctx context.Context,
	encodedExecutionID string,
	reason string,
) error {
	executionID, err := common.NewFromEncoded(encodedExecutionID)
	if err != nil {
		return fmt.Errorf("invalid execution ID %q: %w", encodedExecutionID, err)
	}

	// Empty runID targets the latest run.
	// ignoreNotFound: workflow already completed/terminated before this call.
	return ignoreNotFound(m.publisherClient.Client().TerminateWorkflow(
		ctx,
		executionID.WorkflowID,
		"",
		reason,
	))
}

// Execute dispatches the task to the Temporal workflow registered for its
// OperationType. All Temporal mechanics (client, options, workflow submission)
// are contained here — nothing engine-specific crosses the Executor boundary.
func (m *Manager) Execute(
	ctx context.Context,
	req *task.ExecutionRequest,
) (*task.ExecutionResponse, error) {
	if req == nil {
		return nil, errors.New("execution request is nil")
	}

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid execution request: %w", err)
	}

	desc, ok := workflow.Get(req.Info.OperationType)
	if !ok {
		return nil, fmt.Errorf(
			"no workflow registered for task type %q (registered types: %v) — "+
				"ensure the workflow package is imported and its init() runs",
			req.Info.OperationType,
			workflow.RegisteredTaskTypes(),
		)
	}

	return executeWorkflow(ctx, m.publisherClient.Client(), desc, req)
}
