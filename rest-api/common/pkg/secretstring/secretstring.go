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
package secretstring

import (
	"encoding/json"
	"strings"
)

// SecretString wraps sensitive string data and prevents accidental exposure
// in logs/JSON
type SecretString struct {
	Value string `json:"-"` // Never serialize the actual value
}

// New creates a new SecretString with the given string
func New(v string) SecretString {
	return SecretString{Value: v}
}

// String implements fmt.Stringer to hide the actual value in string
// representations
func (s SecretString) String() string {
	return "******"
}

// MarshalJSON implements json.Marshaler to hide the value during JSON
// serialization
func (s SecretString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// IsEmpty returns true if the secret string has no value
func (s SecretString) IsEmpty() bool {
	return strings.TrimSpace(s.Value) == ""
}

// IsEqual returns true if the give secret string is the same as the one.
func (s SecretString) IsEqual(n SecretString) bool {
	return s.Value == n.Value
}
