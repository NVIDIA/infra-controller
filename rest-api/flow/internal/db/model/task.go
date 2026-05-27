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
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	dbquery "github.com/NVIDIA/infra-controller-rest/flow/internal/db/query"
	taskcommon "github.com/NVIDIA/infra-controller-rest/flow/internal/task/common"
)

var defaultTaskPagination = dbquery.Pagination{
	Offset: 0,
	Limit:  100,
	Total:  0,
}

// Task models the persisted task metadata managed by Flow.
type Task struct {
	bun.BaseModel `bun:"table:task,alias:t"`

	ID            uuid.UUID                 `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Type          taskcommon.TaskType       `bun:"type,type:varchar(64),notnull"`
	ExecutorType  taskcommon.ExecutorType   `bun:"executor_type,type:varchar(64),nullzero"`
	Information   json.RawMessage           `bun:"information,type:jsonb,json_use_number"`
	Description   string                    `bun:"description,nullzero"`
	RackID        uuid.UUID                 `bun:"rack_id,type:uuid,notnull"` // The rack this task operates on
	Attributes    taskcommon.TaskAttributes `bun:"attributes,type:jsonb"`
	ExecutionID   string                    `bun:"execution_id,notnull"`
	Status        taskcommon.TaskStatus     `bun:"status,type:varchar(32),notnull"`
	Message       string                    `bun:"message,nullzero"`
	AppliedRuleID *uuid.UUID                `bun:"applied_rule_id,type:uuid"` // Which operation rule was applied
	CreatedAt     time.Time                 `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt     time.Time                 `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	StartedAt     *time.Time                `bun:"started_at"`
	FinishedAt    *time.Time                `bun:"finished_at"`

	// QueueExpiresAt is set only for waiting tasks. After this time, the
	// Promoter will discard the task instead of promoting it.
	QueueExpiresAt *time.Time `bun:"queue_expires_at"`
}

// Create inserts the task record into the backing store.
func (t *Task) Create(ctx context.Context, idb bun.IDB) error {
	_, err := idb.NewInsert().Model(t).Exec(ctx)
	return err
}

// UpdateScheduledTask updates the scheduled task information.
func (t *Task) UpdateScheduledTask(
	ctx context.Context,
	idb bun.IDB,
) error {
	if t.ExecutionID == "" {
		return fmt.Errorf("execution ID is not set")
	}

	if t.ExecutorType == taskcommon.ExecutorTypeUnknown {
		return fmt.Errorf("executor type is not set")
	}

	t.UpdatedAt = time.Now().UTC()

	_, err := idb.NewUpdate().
		Model(t).
		Column("execution_id", "executor_type", "updated_at").
		Where("id = ?", t.ID).
		Exec(ctx)

	return err
}

// UpdateTaskStatus updates the status of the task.
func (t *Task) UpdateTaskStatus(
	ctx context.Context,
	idb bun.IDB,
	status taskcommon.TaskStatus,
	message string,
) error {
	t.Status = status
	t.Message = message
	t.UpdatedAt = time.Now().UTC()

	columns := []string{"status", "message", "updated_at", "finished_at"}

	if status == taskcommon.TaskStatusRunning && t.StartedAt == nil {
		t.StartedAt = &t.UpdatedAt
		columns = append(columns, "started_at")
	}
	if status.IsFinished() {
		t.FinishedAt = &t.UpdatedAt
	} else {
		t.FinishedAt = nil
	}

	_, err := idb.NewUpdate().
		Model(t).
		Column(columns...).
		Where("id = ?", t.ID).
		Exec(ctx)

	return err
}

func taskListOptionsToFilterable(
	options *taskcommon.TaskListOptions,
) dbquery.Filterable {
	if options == nil {
		return nil
	}

	filters := make([]dbquery.Filter, 0, 3)

	// Filter by rack_id directly
	if options.RackID != uuid.Nil {
		filters = append(
			filters,
			dbquery.Filter{
				Column:   "rack_id",
				Operator: dbquery.OperatorEqual,
				Value:    options.RackID,
			},
		)
	}

	if options.TaskType != taskcommon.TaskTypeUnknown {
		filters = append(filters, dbquery.Filter{
			Column:   "type",
			Operator: dbquery.OperatorEqual,
			Value:    options.TaskType,
		})
	}

	if options.ActiveOnly {
		filters = append(filters, dbquery.Filter{
			Column:   "status",
			Operator: dbquery.OperatorIn,
			Value: []taskcommon.TaskStatus{
				taskcommon.TaskStatusWaiting,
				taskcommon.TaskStatusPending,
				taskcommon.TaskStatusRunning,
			},
		})
	}

	return &dbquery.FilterGroup{
		Filters:   filters,
		Connector: dbquery.ConnectorAND,
	}
}

// GetTask retrieves the task by its UUID.
func GetTask(ctx context.Context, idb bun.IDB, id uuid.UUID) (*Task, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("task UUID is required")
	}

	var task Task
	if err := idb.NewSelect().
		Model(&task).
		Where("id = ?", id).
		Scan(ctx); err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasksForRackByStatus returns tasks for a rack matching any of the given
// statuses, ordered oldest-first.
func ListTasksForRackByStatus(
	ctx context.Context,
	idb bun.IDB,
	rackID uuid.UUID,
	statuses []taskcommon.TaskStatus,
) ([]Task, error) {
	var tasks []Task
	err := idb.NewSelect().
		Model(&tasks).
		Where("rack_id = ?", rackID).
		Where("status IN (?)", bun.In(statuses)).
		OrderExpr("created_at ASC").
		Scan(ctx)
	return tasks, err
}

// ListRacksWithWaitingTasks returns the distinct rack IDs that have at least
// one task in the waiting state.
func ListRacksWithWaitingTasks(
	ctx context.Context,
	idb bun.IDB,
) ([]uuid.UUID, error) {
	var rackIDs []uuid.UUID
	err := idb.NewSelect().
		TableExpr("task").
		ColumnExpr("DISTINCT rack_id").
		Where("status = ?", taskcommon.TaskStatusWaiting).
		Scan(ctx, &rackIDs)
	return rackIDs, err
}

// CountWaitingTasksForRack returns the count of waiting tasks for a rack.
func CountWaitingTasksForRack(
	ctx context.Context,
	idb bun.IDB,
	rackID uuid.UUID,
) (int, error) {
	return idb.NewSelect().
		TableExpr("task").
		Where("rack_id = ?", rackID).
		Where("status = ?", taskcommon.TaskStatusWaiting).
		Count(ctx)
}

// ListTasks returns all tasks that match the given criteria.
func ListTasks(
	ctx context.Context,
	idb bun.IDB,
	options *taskcommon.TaskListOptions,
	pagination *dbquery.Pagination,
) ([]Task, int32, error) {
	var tasks []Task
	conf := &dbquery.Config{
		IDB:   idb,
		Model: &tasks,
	}

	if pagination != nil {
		conf.Pagination = pagination
	} else {
		conf.Pagination = &defaultTaskPagination
	}

	if filterable := taskListOptionsToFilterable(options); filterable != nil {
		conf.Filterables = []dbquery.Filterable{filterable}
	}

	q, err := dbquery.New(ctx, conf)
	if err != nil {
		return nil, 0, err
	}

	if err := q.Scan(ctx); err != nil {
		return nil, 0, err
	}

	return tasks, int32(q.TotalCount()), nil
}
