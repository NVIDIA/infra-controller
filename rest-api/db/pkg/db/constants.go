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

package db

const (
	// DefaultPageSize is the size for query results to request from DB
	DefaultPageSize = 20

	// MaxBatchItems limits the maximum number of items allowed in a single batch operation
	// to prevent performance degradation and potential timeouts from overly large batches.
	MaxBatchItems = 100

	// MaxBatchItemsToTrace limits the number of items traced in detail for batch operations
	// to avoid producing overly-large spans and reduce the risk of hitting tracing backend limits.
	// Items beyond this limit will still be processed but won't have their individual field values traced.
	MaxBatchItemsToTrace = 20
)
