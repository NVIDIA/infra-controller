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
	"strings"

	"github.com/labstack/echo/v4"
)

// RequestHandler defines the Echo compatible interface all API route handlers
// should implement
type RequestHandler interface {
	Handle(c echo.Context) error
}

// Route defines the data structure to organize route information that can
// be used to initialize Echo routes
type Route struct {
	Path    string
	Method  string
	Handler RequestHandler
}

// MetricsURLSkipper ignores metrics for certain routes
func MetricsURLSkipper(c echo.Context) bool {
	// Allow v2 API paths to be tracked
	if strings.HasPrefix(c.Path(), "/v2/") {
		return false
	}

	if c.Path() == "/metrics" {
		return false
	}

	return true
}
