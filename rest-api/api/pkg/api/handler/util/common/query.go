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

package common

import (
	cdb "github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	"github.com/labstack/echo/v4"
)

// QueryOverride provides values that override query params when delegating from
// path-scoped endpoints (e.g. instance/{instanceId}/interface, instance/{instanceId}/nvlink-interface) to more general endpoints.
// When set, error messages in general endpoints will be modulated
type QueryOverride struct {
	InstanceIDs   []string
	ValueFromPath bool
}

// GetSearchQuery returns a trimmed search query or nil when the query is blank.
func GetSearchQuery(c echo.Context) *string {
	rawQuery := c.QueryParams().Get("query")

	searchQuery, ok := cdb.TrimSearchQuery(rawQuery)
	if !ok {
		return nil
	}

	return &searchQuery
}
