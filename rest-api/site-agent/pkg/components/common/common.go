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
	"errors"
)

// Error Type
var (
	// ErrResourceStale Requested update of stale object to DB
	ErrResourceStale = errors.New("requested update of stale object to DB")
)

// Resource Type
var (
	// ResourceTypeVpc is VPC
	ResourceTypeVpc = "VPC"
	// ResourceTypeSubnet is Subnet
	ResourceTypeSubnet = "Subnet"
	// ResourceTypeInstance is Instance
	ResourceTypeInstance = "Instance"
	// ResourceTypeSSHKeyGroup is SSHKeyGroup
	ResourceTypeSSHKeyGroup = "SSHKeyGroup"
	// ResourceTypeInfiniBandPartition is InfiniBandPartition
	ResourceTypeInfiniBandPartition = "InfiniBandPartition"
	// ResourceTypeExpectedMachine is ExpectedMachine
	ResourceTypeExpectedMachine = "ExpectedMachine"
	// ResourceTypeSKU is SKU
	ResourceTypeSKU = "SKU"
	// ResourceTypeDpuExtensionService is DpuExtensionService
	ResourceTypeDpuExtensionService = "DpuExtensionService"
	// ResourceTypeNVLinkLogicalPartition is NVLinkLogicalPartition
	ResourceTypeNVLinkLogicalPartition = "NVLinkLogicalPartition"
)

// OpType is type of operation
type OpType int

const (
	// OpCreate is create operation
	OpCreate OpType = iota
	// OpUpdate is update request operation
	OpUpdate
	// OpDelete is delete operation
	OpDelete
	// No op
	OpNone
)
