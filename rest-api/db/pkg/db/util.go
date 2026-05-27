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

import (
	"hash/fnv"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GetStrPtr returns a pointer for the provided string
func GetStrPtr(s string) *string {
	sp := s
	return &sp
}

// GetBoolPtr returns a pointer for the provided bool
func GetBoolPtr(b bool) *bool {
	bp := b
	return &bp
}

// GetUUIDPtr returns a pointer for the provided UUID
func GetUUIDPtr(u uuid.UUID) *uuid.UUID {
	up := u
	return &up
}

// GetIntPtr returns a pointer for the provided int
func GetIntPtr(i int) *int {
	ip := i
	return &ip
}

// GetTimePtr returns a pointer for the provided time
func GetTimePtr(t time.Time) *time.Time {
	tp := t
	return &tp
}

// GetCurTime returns the current time
func GetCurTime() time.Time {
	// Standardize time to match Postgres resolution
	return time.Now().UTC().Round(time.Microsecond)
}

// IsStrInSlice returns true if the provided string is in the provided slice
func IsStrInSlice(s string, sl []string) bool {
	for _, v := range sl {
		if v == s {
			return true
		}
	}
	return false
}

// GetStringToUint64Hash returns a uint64 hash of the input string
// this is used for advisory lock ids
func GetStringToUint64Hash(id string) uint64 {
	h := fnv.New64()
	h.Write([]byte(id))
	return h.Sum64()
}

// GetStringToTsQuery returns a string into a to_tsquery format from the input string
func GetStringToTsQuery(inputQuery string) string {
	inputQuery, ok := TrimSearchQuery(inputQuery)
	if !ok {
		return ""
	}

	tokens := strings.Fields(inputQuery)
	if len(tokens) == 0 {
		return ""
	}

	hasOperator := false
	for _, token := range tokens {
		switch token {
		case "|", "&":
			hasOperator = true
		case "!":
			return ""
		default:
			if strings.ContainsAny(token, "|&!") {
				return ""
			}
		}
	}
	if !hasOperator {
		return strings.Join(tokens, " | ")
	}

	expectTerm := true
	for _, token := range tokens {
		switch token {
		case "|", "&":
			if expectTerm {
				return ""
			}
			expectTerm = true
		default:
			if !expectTerm {
				return ""
			}
			expectTerm = false
		}
	}
	if expectTerm {
		return ""
	}

	return strings.Join(tokens, " ")
}

// normalizeSearchQuery normalizes a search query by trimming it and converting it to a to_tsquery format
func NormalizeSearchQuery(input *string) (string, *string, bool) {
	if input == nil {
		return "", nil, false
	}

	searchQuery, ok := TrimSearchQuery(*input)
	if !ok {
		return "", nil, false
	}

	tsQuery := GetStringToTsQuery(searchQuery)
	if tsQuery == "" {
		return "", nil, false
	}

	return searchQuery, GetStrPtr(tsQuery), true
}

// TrimmedSearchQuery trims a search query and reports whether it is non-blank
func TrimSearchQuery(input string) (string, bool) {
	trimmed := strings.TrimSpace(input)
	return trimmed, trimmed != ""
}

// CompareStringSlicesIgnoreOrder compares two string slices ignoring order
func CompareStringSlicesIgnoreOrder(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	// Create sorted copies to compare
	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))
	copy(aCopy, a)
	copy(bCopy, b)
	slices.Sort(aCopy)
	slices.Sort(bCopy)
	return slices.Equal(aCopy, bCopy)
}
