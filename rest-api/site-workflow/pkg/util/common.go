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

import (
	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

// GetStrPtr returns a pointer to the string passed in
func GetStrPtr(s string) *string {
	return &s
}

func ProtobufUUIDListToStringList(ids []*cwssaws.UUID) []string {
	s := make([]string, len(ids))

	for i, u := range ids {
		if u == nil {
			s[i] = ""
		} else {
			s[i] = u.Value
		}
	}

	return s
}

func StringsToProtobufUUIDList(ids []string) []*cwssaws.UUID {
	s := make([]*cwssaws.UUID, len(ids))

	for i, u := range ids {
		s[i] = &cwssaws.UUID{Value: u}
	}

	return s
}
