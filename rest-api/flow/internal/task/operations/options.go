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

package operations

import (
	"time"

	taskcommon "github.com/NVIDIA/infra-controller-rest/flow/internal/task/common"
)

type OperationOptions struct {
	Timeout time.Duration
}

var (
	defaultOperationOptions = map[taskcommon.TaskType]OperationOptions{
		taskcommon.TaskTypePowerControl: {
			Timeout: 60 * time.Minute,
		},
		taskcommon.TaskTypeFirmwareControl: {
			Timeout: 60 * time.Minute,
		},
		taskcommon.TaskTypeInjectExpectation: {
			Timeout: 60 * time.Minute,
		},
		taskcommon.TaskTypeBringUp: {
			Timeout: 120 * time.Minute,
		},
	}
)

func GetOperationOptions(typ taskcommon.TaskType) OperationOptions {
	if opt, ok := defaultOperationOptions[typ]; ok {
		return opt
	}
	return OperationOptions{
		Timeout: 0,
	}
}
