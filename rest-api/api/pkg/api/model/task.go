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
	"fmt"
	"time"

	flowv1 "github.com/NVIDIA/infra-controller-rest/workflow-schema/flow/protobuf/v1"
)

var ProtoToAPIRackTaskStatusName = map[flowv1.TaskStatus]string{
	flowv1.TaskStatus_TASK_STATUS_UNKNOWN:    "Unknown",
	flowv1.TaskStatus_TASK_STATUS_PENDING:    "Pending",
	flowv1.TaskStatus_TASK_STATUS_RUNNING:    "Running",
	flowv1.TaskStatus_TASK_STATUS_COMPLETED:  "Succeeded",
	flowv1.TaskStatus_TASK_STATUS_FAILED:     "Failed",
	flowv1.TaskStatus_TASK_STATUS_TERMINATED: "Terminated",
	flowv1.TaskStatus_TASK_STATUS_WAITING:    "Waiting",
}

// APIRackTask is the API response model for a rack task (OpenAPI schema RackTask).
type APIRackTask struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	Description string     `json:"description"`
	Message     string     `json:"message"`
	Started     *time.Time `json:"started"`
	Finished    *time.Time `json:"finished"`
	Created     time.Time  `json:"created"`
	Updated     time.Time  `json:"updated"`
}

func (t *APIRackTask) FromProto(task *flowv1.Task) {
	if task == nil {
		return
	}
	if task.GetId() != nil {
		t.ID = task.GetId().GetId()
	}
	t.Status = enumOr(ProtoToAPIRackTaskStatusName, task.GetStatus(), "Unknown")
	t.Description = task.GetDescription()
	t.Message = task.GetMessage()
	if ts := task.GetStartedAt(); ts != nil {
		v := ts.AsTime().UTC()
		t.Started = &v
	}
	if ts := task.GetFinishedAt(); ts != nil {
		v := ts.AsTime().UTC()
		t.Finished = &v
	}
	t.Created = task.GetCreatedAt().AsTime().UTC()
	t.Updated = task.GetUpdatedAt().AsTime().UTC()
}

func NewAPIRackTask(task *flowv1.Task) *APIRackTask {
	t := &APIRackTask{}
	t.FromProto(task)
	return t
}

// APIGetTaskRequest captures query parameters for getting a task by ID.
type APIGetTaskRequest struct {
	SiteID string `query:"siteId"`
}

func (r *APIGetTaskRequest) Validate() error {
	if r.SiteID == "" {
		return fmt.Errorf("siteId query parameter is required")
	}
	return nil
}

// APICancelTaskRequest is the request body for cancelling a task by ID.
type APICancelTaskRequest struct {
	SiteID string `json:"siteId"`
}

// Validate validates the cancel task request
func (r *APICancelTaskRequest) Validate() error {
	if r.SiteID == "" {
		return fmt.Errorf("siteId is required")
	}
	return nil
}
