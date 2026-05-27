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

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/conflict"
	taskstore "github.com/NVIDIA/infra-controller-rest/flow/internal/task/store"
	taskdef "github.com/NVIDIA/infra-controller-rest/flow/internal/task/task"
)

// notifyingTaskStore wraps a taskstore.Store and fires Promoter notifications
// whenever a task transitions to a finished state. This bridges task completion
// events to the Promoter without coupling workflow activities to Manager
// internals.
type notifyingTaskStore struct {
	taskstore.Store
	promoter *conflict.Promoter
}

func newNotifyingTaskStore(
	store taskstore.Store,
	promoter *conflict.Promoter,
) *notifyingTaskStore {
	return &notifyingTaskStore{
		Store:    store,
		promoter: promoter,
	}
}

// UpdateTaskStatus delegates to the underlying store and notifies the Promoter
// when the task reaches a finished state. The task is fetched from the DB to
// obtain its rack ID, making this correct across service restarts.
func (s *notifyingTaskStore) UpdateTaskStatus(
	ctx context.Context,
	u *taskdef.TaskStatusUpdate,
) error {
	if err := s.Store.UpdateTaskStatus(ctx, u); err != nil {
		return err
	}
	if !u.Status.IsFinished() {
		return nil
	}

	tasks, err := s.Store.GetTasks(ctx, []uuid.UUID{u.ID})
	if err != nil || len(tasks) == 0 {
		log.Warn().
			Str("task_id", u.ID.String()).
			Msg("notifying store: could not look up task for promotion")
		return nil
	}

	s.promoter.Notify(tasks[0].RackID)
	return nil
}

// Ensure notifyingTaskStore satisfies taskdef.TaskStatusUpdater so it can
// be passed to activity.SetTaskStatusUpdater.
var _ taskdef.TaskStatusUpdater = (*notifyingTaskStore)(nil) //nolint
