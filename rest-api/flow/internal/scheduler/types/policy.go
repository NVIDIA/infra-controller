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

package types

// Policy controls what happens when a trigger fires while a job is running.
type Policy int

const (
	// Skip drops the new event if the worker is busy. Default.
	Skip Policy = iota
	// Queue keeps only the latest pending event; delivers it when the worker
	// is free. Earlier events are discarded.
	Queue
	// QueueAll buffers every event and delivers them in FIFO order.
	// Use this when each event carries unique data that must not be dropped.
	QueueAll
	// Replace cancels the current job and starts a new one.
	Replace
)
