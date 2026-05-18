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

// Package otelecho provides OpenTelemetry instrumentation for the labstack/echo package.
//
// This package wraps the upstream OpenTelemetry contrib package
// (go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho)
// and adds custom functionality:
//   - Zerolog logging of trace IDs
//   - Setting X-Ngc-Trace-Id header
//   - Storing tracer in context for use by other packages
//
// The upstream package is used directly as a dependency instead of copying the code
// to avoid Apache license issues and to benefit from upstream updates and fixes.
//
// Note: The upstream package's go.mod contains a replace directive for the b3
// propagator, but this is resolved automatically by Go's module system when using
// the package as a dependency. The replace directive is ignored for dependencies
// and only affects builds within the upstream repository itself.
package otelecho
