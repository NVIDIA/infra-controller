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

package client

import (
	"github.com/google/uuid"

	"github.com/NVIDIA/infra-controller-rest/flow/pkg/types"
)

// UpgradeFirmwareResult represents the result of a firmware upgrade operation.
type UpgradeFirmwareResult struct {
	TaskIDs []uuid.UUID // Multiple task IDs (1 task per rack)
}

// PowerControlResult represents the result of a power control operation.
type PowerControlResult struct {
	TaskIDs []uuid.UUID // Multiple task IDs (1 task per rack)
}

// GetExpectedComponentsResult contains the result of GetExpectedComponents operation.
type GetExpectedComponentsResult struct {
	Components []*types.Component
	Total      int
}

// ValidateComponentsResult represents the result of ValidateComponents call.
type ValidateComponentsResult struct {
	Diffs           []*types.ComponentDiff
	TotalDiffs      int
	MissingCount    int
	UnexpectedCount int
	DriftCount      int
	MatchCount      int
}

// IngestRackResult represents the result of an IngestRack operation.
type IngestRackResult struct {
	TaskIDs []uuid.UUID
}

// ListTasksResult represents the result of ListTasks call.
type ListTasksResult struct {
	Tasks []*types.Task
	Total int
}
