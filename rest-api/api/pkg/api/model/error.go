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

// common validation errors
const (
	validationErrorValueRequired                  = "a value is required"
	validationErrorInvalidUUID                    = "must be a valid UUID"
	validationErrorStringLength                   = "must be at least 2 characters and maximum 256 characters"
	validationErrorDescriptionStringLength        = "maximum 1024 characters are allowed in description"
	validationErrorMachineMaintenanceStringLength = "must be at least 5 characters and maximum 256 characters"
	validationErrorInvalidIPAddress               = "invalid IP address"
	validationErrorInvalidIPv4Address             = "invalid IPv4 address"
	validationErrorInvalidHostname                = "invalid hostname"
	validationErrorInvalidIPv6Address             = "invalid IPv6 address"
	validationErrorStringLength64                 = "must be at least 2 characters and maximum 64 characters"

	validationCommonErrorField = "__all__"
)
