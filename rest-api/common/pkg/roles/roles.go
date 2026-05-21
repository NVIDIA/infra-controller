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

// Package roles defines the canonical authorization role suffix names
// used throughout the codebase. It deliberately has no dependencies so it
// can be imported from any package — including db model tests (which would
// otherwise create an import cycle through auth/pkg/authorization) and
// from production binaries that don't ship the auth module.
package roles

const (
	// ProviderAdminRole is the suffix for the provider admin authorization role.
	ProviderAdminRole = "PROVIDER_ADMIN"

	// ProviderViewerRole is the suffix for the provider viewer authorization role.
	ProviderViewerRole = "PROVIDER_VIEWER"

	// TenantAdminRole is the suffix for the tenant admin authorization role.
	TenantAdminRole = "TENANT_ADMIN"
)
