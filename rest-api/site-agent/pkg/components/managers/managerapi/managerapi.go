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

package managerapi

import (
	"github.com/NVIDIA/infra-controller-rest/site-agent/pkg/conftypes"
	"github.com/NVIDIA/infra-controller-rest/site-agent/pkg/datatypes/elektratypes"
)

// ManagerHdl - local handle to be assigned
var ManagerHdl ManagerAPI

// ManagerAccess - access to all APIs/data/conf
// nolint
type ManagerAccess struct {
	API  *ManagerAPI
	Data *ManagerData
	Conf *ManagerConf
}

// ManagerData - super struct
type ManagerData struct {
	EB *elektratypes.Elektra
}

// ManagerAPI struct to hold all mgr interface
type ManagerAPI struct {
	// Add all the manager interfaces here
	Bootstrap              BootstrapInterface
	VPC                    VPCInterface
	VpcPrefix              VpcPrefixInterface
	VpcPeering             VpcPeeringInterface
	Subnet                 SubnetInterface
	Instance               InstanceInterface
	Machine                MachineInterface
	Orchestrator           OrchestratorInterface
	NICo                   NICoInterface
	SSHKeyGroup            SSHKeyGroupInterface
	InfiniBandPartition    InfiniBandPartitionInterface
	Tenant                 TenantInterface
	OperatingSystem        OperatingSystemInterface
	MachineValidation      MachineValidationInterface
	InstanceType           InstanceTypeInterface
	NetworkSecurityGroup   NetworkSecurityGroupInterface
	ExpectedMachine        ExpectedMachineInterface
	ExpectedPowerShelf     ExpectedPowerShelfInterface
	ExpectedRack           ExpectedRackInterface
	ExpectedSwitch         ExpectedSwitchInterface
	SKU                    SKUInterface
	DpuExtensionService    DpuExtensionServiceInterface
	NVLinkLogicalPartition NVLinkLogicalPartitionInterface
	Flow                   FlowInterface
}

// ManagerConf - Conf struct
type ManagerConf struct {
	EB *conftypes.Config
}
