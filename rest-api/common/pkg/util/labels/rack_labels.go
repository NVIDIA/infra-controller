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

package labels

// Well-known label keys for Expected/Managed Rack metadata.
// These mirror the constants defined in Core's api-model crate so REST callers,
// the site-workflow, and Core stay aligned on rack chassis and location labels.

const (
	// Chassis identity labels — physically identifies the rack hardware.
	RackLabelChassisManufacturer = "chassis.manufacturer"
	RackLabelChassisSerialNumber = "chassis.serial-number"
	RackLabelChassisModel        = "chassis.model"

	// Physical location labels — identifies where the rack lives.
	RackLabelLocationRegion     = "location.region"
	RackLabelLocationDatacenter = "location.datacenter"
	RackLabelLocationRoom       = "location.room"
	RackLabelLocationPosition   = "location.position"
)
