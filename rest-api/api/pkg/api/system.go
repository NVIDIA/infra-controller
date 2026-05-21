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

package api

import (
	"net/http"

	apiHandler "github.com/NVIDIA/infra-controller-rest/api/pkg/api/handler"
)

// NewSystemAPIRoutes returns API routes that provide system level  functions
func NewSystemAPIRoutes() []Route {
	apiRoutes := []Route{
		// Health check endpoints
		{
			Path:    "/healthz",
			Method:  http.MethodGet,
			Handler: apiHandler.NewHealthCheckHandler(),
		},
		{
			Path:    "/readyz",
			Method:  http.MethodGet,
			Handler: apiHandler.NewHealthCheckHandler(),
		},
	}

	return apiRoutes
}

// IsSystemRoute returns true for a path registered as SystemAPIRoute
func IsSystemRoute(p string) bool {
	routes := NewSystemAPIRoutes()
	for _, r := range routes {
		if r.Path == p {
			return true
		}
	}

	return false
}
