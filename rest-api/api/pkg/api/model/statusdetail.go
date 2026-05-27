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
	"time"

	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
)

// APIStatusDetail captures API representation of a status detail DB object
type APIStatusDetail struct {
	// Status denotes the state of the associated entity at a particular time
	Status string `json:"status"`
	// Message contains the description of the state and cause/remedy in case of error
	Message *string `json:"message"`
	// Created indicates the ISO datetime string for when the associated entity assumed the status
	Created time.Time `json:"created"`
	// Updated indicates the ISO datetime string for when the associated entity was last found to have this status
	Updated time.Time `json:"updated"`
}

// NewAPIStatusDetail creates an API status detail object from status detail DB entry
func NewAPIStatusDetail(dbsd cdbm.StatusDetail) APIStatusDetail {
	apiStatusDetail := APIStatusDetail{
		Status:  dbsd.Status,
		Message: dbsd.Message,
		Created: dbsd.Created,
		Updated: dbsd.Updated,
	}

	return apiStatusDetail
}
