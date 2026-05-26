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

package devicetypes

import "strings"

type BMCType int

const (
	BMCTypeUnknown BMCType = iota
	BMCTypeHost
	BMCTypeDPU
)

var (
	bmcTypeStrings = map[BMCType]string{
		BMCTypeUnknown: "Unknown",
		BMCTypeHost:    "Host",
		BMCTypeDPU:     "DPU",
	}

	bmcTypeStringMaxLen int
)

func init() {
	for _, str := range bmcTypeStrings {
		if len(str) > bmcTypeStringMaxLen {
			bmcTypeStringMaxLen = len(str)
		}
	}
}

// BMCTypes returns all the supported BMC types
func BMCTypes() []BMCType {
	return []BMCType{
		BMCTypeUnknown,
		BMCTypeHost,
		BMCTypeDPU,
	}
}

// BMCTypeFromString returns the BMC type from the given string.
func BMCTypeFromString(str string) BMCType {
	for bt, bmcTypeStr := range bmcTypeStrings {
		if strings.EqualFold(str, bmcTypeStr) {
			return bt
		}
	}
	return BMCTypeUnknown
}

// BMCTypeToString returns the string representation for the given BMC type.
func BMCTypeToString(bt BMCType) string {
	return bmcTypeStrings[bt]
}

func IsValidBMCTypeString(str string) bool {
	return BMCTypeFromString(str) != BMCTypeUnknown
}

// String return the aligned string representation for the given BMC type
func (bt BMCType) String() string {
	spaces := bmcTypeStringMaxLen - len(BMCTypeToString(bt))
	return strings.Repeat(" ", spaces) + BMCTypeToString(bt)
}
