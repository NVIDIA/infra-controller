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
	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/model/util"
	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
	"net/url"
	"time"
)

// APIAuditEntry is a data structure to capture audit log information
type APIAuditEntry struct {
	ID            string                 `json:"id"`
	Endpoint      string                 `json:"endpoint"`
	QueryParams   url.Values             `json:"queryParams"`
	Method        string                 `json:"method"`
	Body          map[string]interface{} `json:"body"`
	StatusCode    int                    `json:"statusCode"`
	StatusMessage string                 `json:"statusMessage"`
	ClientIP      string                 `json:"clientIP"`
	UserID        *string                `json:"userID"`
	User          *APIUser               `json:"user"`
	OrgName       string                 `json:"orgName"`
	ExtraData     map[string]interface{} `json:"extraData"`
	Timestamp     time.Time              `json:"timestamp"`
	DurationMs    int64                  `json:"durationMs"`
	APIVersion    string                 `json:"apiVersion"`
}

// NewAPIAuditEntry creates and returns a new APIAuditEntry object
func NewAPIAuditEntry(dbAuditEntry cdbm.AuditEntry, dbUser *cdbm.User) APIAuditEntry {
	apiAuditEntry := APIAuditEntry{
		ID:            dbAuditEntry.ID.String(),
		Endpoint:      dbAuditEntry.Endpoint,
		QueryParams:   dbAuditEntry.QueryParams,
		Method:        dbAuditEntry.Method,
		Body:          dbAuditEntry.Body,
		StatusCode:    dbAuditEntry.StatusCode,
		StatusMessage: dbAuditEntry.StatusMessage,
		ClientIP:      dbAuditEntry.ClientIP,
		OrgName:       dbAuditEntry.OrgName,
		ExtraData:     dbAuditEntry.ExtraData,
		Timestamp:     dbAuditEntry.Timestamp,
		DurationMs:    dbAuditEntry.Duration.Milliseconds(),
		APIVersion:    dbAuditEntry.APIVersion,
	}

	if dbAuditEntry.UserID != nil {
		apiAuditEntry.UserID = util.GetUUIDPtrToStrPtr(dbAuditEntry.UserID)
	}
	if dbUser != nil {
		apiAuditEntry.User = NewAPIUserFromDBUser(*dbUser)
	}

	return apiAuditEntry
}
