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

package scheduler

import (
	"context"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/scheduler/types"
)

// maxQueueSize is the default upper limit for the relay's in-memory queue.
// When the queue is full the oldest event is dropped to make room for the new
// one, preventing unbounded memory growth if the worker falls far behind.
const maxQueueSize = 2048

// entry pairs a Job with its Trigger, Policy, and the channels that wire the
// pipeline together. Entries are created internally by Scheduler.Schedule and
// hidden from the callers.
type entry struct {
	job     types.Job
	trigger types.Trigger
	policy  types.Policy
	eventCh chan types.Event // Trigger → relay  (capacity 1)
	workCh  chan workItem    // relay → worker   (unbuffered)
	relay   *relay           // created and assigned in Scheduler.Start
}

// workItem is the unit of work dispatched from relay to worker.
// ctx is a cancellable context for the job, derived from relay.forceCtx so
// that relay.forceStop cancels all in-flight jobs atomically by cancelling
// the parent context — no per-job registration race is possible.
type workItem struct {
	ctx    context.Context
	cancel context.CancelFunc
	ev     types.Event
}
