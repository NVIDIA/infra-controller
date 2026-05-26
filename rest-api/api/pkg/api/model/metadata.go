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

package model

import (
	"github.com/NVIDIA/infra-controller-rest/api/pkg/metadata"
)

// APIMetadata is a data structure to capture NICo API system information
type APIMetadata struct {
	// Version contains the API version
	Version string `json:"version"`
	// BuildTime contains the time the binary was built
	BuildTime string `json:"buildTime"`
}

// NewAPIMetadata creates and returns a new APISystemInfo object
func NewAPIMetadata() *APIMetadata {
	amd := &APIMetadata{
		Version:   metadata.Version,
		BuildTime: metadata.BuildTime,
	}

	return amd
}
