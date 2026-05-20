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

package util

import "time"

const (
	// InventoryReceiptInterval is the interval between 2 subsequent inventory receipts
	InventoryReceiptInterval = 3 * time.Minute
	// WorkflowExecutionTimeout is the timeout for a workflow execution
	WorkflowExecutionTimeout = time.Minute * 1
	// WorkflowContextTimeout is the timeout for a workflow context
	WorkflowContextTimeout = time.Second * 50
	// WorkflowContextNewAfterTimeout is the timeout for a new workflow context
	WorkflowContextNewAfterTimeout = time.Second * 5
)
