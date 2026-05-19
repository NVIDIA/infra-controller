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

package nicotypes

import "time"

type WorkflowStatus string

const (
	// WorkflowStatusSuccess workflow has completed successfully
	WorkflowStatusSuccess WorkflowStatus = "Success"
	// WorkflowStatusActFailed workflow activity execution has failed
	WorkflowStatusActivityFailed WorkflowStatus = "ActivityFailed"
	// WorkflowStatusPubFailed workflow status publish failed
	WorkflowStatusPublishFailed WorkflowStatus = "PublishFailed"
	// WorkflowStatusActPubFailed both workflow activity execution and status publish failed
	WorkflowStatusActivityPublishFailed WorkflowStatus = "ActivityPublishFailed"
)

// WorkflowMetrics defines interface to be used for workflow metrics
type WorkflowMetrics interface {
	// RecordLatency function to record latency for a workflow
	RecordLatency(activity string, status WorkflowStatus, duration time.Duration)
}
