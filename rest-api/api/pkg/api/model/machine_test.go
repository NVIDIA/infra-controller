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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	cdb "github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/model/util"
	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

func TestMachine_NewAPIMachine(t *testing.T) {
	mID := uuid.NewString()

	machineInfo1 := &cwssaws.MachineInfo{
		Machine: &cwssaws.Machine{
			Id:    &cwssaws.MachineId{Id: mID},
			State: "Ready",
			DiscoveryInfo: &cwssaws.DiscoveryInfo{
				Cpus: []*cwssaws.Cpu{
					{
						Vendor:    "GenuineIntel",
						Model:     "Intel(R) Xeon(R) Gold 6354 CPU @ 3.00GHz",
						Frequency: "1571.080",
						Number:    0,
						Core:      0,
						Socket:    0,
					},
					{
						Vendor:    "GenuineIntel",
						Model:     "Intel(R) Xeon(R) Gold 6354 CPU @ 3.00GHz",
						Frequency: "1571.080",
						Number:    1,
						Core:      0,
						Socket:    0,
					},
					{
						Vendor:    "GenuineIntel",
						Model:     "Intel(R) Xeon(R) Gold 6354 CPU @ 3.00GHz",
						Frequency: "3371.751",
						Number:    2,
						Core:      0,
						Socket:    1,
					},
					{
						Vendor:    "GenuineIntel",
						Model:     "Intel(R) Xeon(R) Gold 6354 CPU @ 3.00GHz",
						Frequency: "3017.142",
						Number:    3,
						Core:      0,
						Socket:    1,
					},
					{
						Vendor:    "GenuineIntel",
						Model:     "Intel(R) Xeon(R) Gold 6354 CPU @ 3.00GHz",
						Frequency: "3507.275",
						Number:    4,
						Core:      1,
						Socket:    0,
					},
					{
						Vendor:    "GenuineIntel",
						Model:     "Intel(R) Xeon(R) Gold 6354 CPU @ 3.00GHz",
						Frequency: "3255.853",
						Number:    5,
						Core:      1,
						Socket:    0,
					},
					{
						Vendor:    "GenuineIntel",
						Model:     "Intel(R) Xeon(R) Gold 6354 CPU @ 3.00GHz",
						Frequency: "3530.777",
						Number:    6,
						Core:      1,
						Socket:    1,
					},
				},
				NetworkInterfaces: []*cwssaws.NetworkInterface{
					{
						PciProperties: &cwssaws.PciDeviceProperties{
							Vendor:      "0x14e4",
							Device:      "0x165f",
							Path:        "/devices/pci0000:00/0000:00:1c.5/0000:04:00.0/net/eno8303",
							Description: cdb.GetStrPtr("NetXtreme BCM5720 2-port Gigabit Ethernet PCIe (PowerEdge Rx5xx LOM Board)"),
						},
					},
					{
						PciProperties: &cwssaws.PciDeviceProperties{
							Vendor:      "0x14e4",
							Device:      "0x165f",
							Path:        "/devices/pci0000:00/0000:00:1c.5/0000:04:00.1/net/eno8403",
							Description: cdb.GetStrPtr("NetXtreme BCM5720 2-port Gigabit Ethernet PCIe (PowerEdge Rx5xx LOM Board)"),
						},
					},
					{
						PciProperties: &cwssaws.PciDeviceProperties{
							Vendor:      "0x14e4",
							Device:      "0x16d7",
							Path:        "/devices/pci0000:30/0000:30:04.0/0000:31:00.0/net/eno12399np0",
							Description: cdb.GetStrPtr("BCM57414 NetXtreme-E 10Gb/25Gb RDMA Ethernet Controller"),
						},
					},
					{
						PciProperties: &cwssaws.PciDeviceProperties{
							Vendor:      "0x14e4",
							Device:      "0x16d7",
							Path:        "/devices/pci0000:30/0000:30:04.0/0000:31:00.1/net/eno12409np1",
							Description: cdb.GetStrPtr("BCM57414 NetXtreme-E 10Gb/25Gb RDMA Ethernet Controller"),
						},
					},
					{
						PciProperties: &cwssaws.PciDeviceProperties{
							Vendor:      "0x15b3",
							Device:      "0xa2d6",
							Path:        "/devices/pci0000:b0/0000:b0:02.0/0000:b1:00.0/net/enp177s0f0np0",
							NumaNode:    1,
							Description: cdb.GetStrPtr("MT42822 BlueField-2 integrated ConnectX-6 Dx network controller"),
						},
					},
					{
						PciProperties: &cwssaws.PciDeviceProperties{
							Vendor:      "0x15b3",
							Device:      "0xa2d6",
							Path:        "/devices/pci0000:b0/0000:b0:02.0/0000:b1:00.1/net/enp177s0f1np1",
							NumaNode:    1,
							Description: cdb.GetStrPtr("MT42822 BlueField-2 integrated ConnectX-6 Dx network controller"),
						},
					},
				},
				BlockDevices: []*cwssaws.BlockDevice{
					{
						Model:    "NO_MODEL",
						Revision: "NO_REVISION",
					},
					{
						Model:    "LOGICAL_VOLUME",
						Revision: "3.53",
						Serial:   "600508b1001cb4d1a278bf3ee7a72228",
					},
					{
						Model:    "Dell Ent NVMe CM6 RI 1.92TB",
						Revision: "2.1.3",
					},
					{
						Model:    "SSDPF2KE016T9L",
						Revision: "2CV1L028",
					},
					{
						Model:    "DELLBOSS_VD",
						Revision: "MV.R00-0",
					},
				},
				DmiData: &cwssaws.DmiData{
					BoardName:     "7Z23CTOLWW",
					BoardVersion:  "06",
					BiosVersion:   "U8E122J-1.51",
					ProductSerial: "J1050ACR",
					BoardSerial:   ".C1KS2CS001G.",
					ChassisSerial: "J1050ACR",
					BiosDate:      "03/30/2023",
					ProductName:   "ThinkSystem SR670 V2",
					SysVendor:     "Lenovo",
				},
				NvmeDevices: []*cwssaws.NvmeDevice{
					{
						Model:       "Dell Ent NVMe CM6 RI 1.92TB",
						FirmwareRev: "2.1.3",
					},
					{
						Model:       "Dell Ent NVMe CM6 RI 1.92TB",
						FirmwareRev: "2.1.3",
					},
					{
						Model:       "Dell Ent NVMe CM6 RI 1.92TB",
						FirmwareRev: "2.1.3",
					},
				},
				Gpus: []*cwssaws.Gpu{
					{
						Name:           "NVIDIA H100 PCIe",
						Serial:         "1654422005434",
						DriverVersion:  "530.30.02",
						VbiosVersion:   "96.00.30.00.01",
						InforomVersion: "1010.0200.00.02",
						TotalMemory:    "81559 MiB",
						Frequency:      "1755 MHz",
						PciBusId:       "00000000:17:00.0",
					},
				},
				InfinibandInterfaces: []*cwssaws.InfinibandInterface{
					{
						PciProperties: &cwssaws.PciDeviceProperties{
							Vendor:      "Mellanox Technologies",
							Device:      "MT28908 Family [ConnectX-6]",
							Path:        "/devices/pci0000:c9/0000:c9:02.0/0000:ca:00.0/infiniband/rocep202s0f0",
							NumaNode:    1,
							Description: cdb.GetStrPtr("MT28908 Family [ConnectX-6]"),
							Slot:        cdb.GetStrPtr("0000:ca:00.0"),
						},
						Guid: "1070fd0300bd43ac",
					},
					{
						PciProperties: &cwssaws.PciDeviceProperties{
							Vendor:      "Mellanox Technologies",
							Device:      "MT28908 Family [ConnectX-6]",
							Path:        "/devices/pci0000:c9/0000:c9:02.0/0000:ca:00.1/infiniband/rocep202s0f1",
							NumaNode:    1,
							Description: cdb.GetStrPtr("MT28908 Family [ConnectX-6]"),
							Slot:        cdb.GetStrPtr("0000:ca:00.1"),
						},
						Guid: "1070fd0300bd43ad",
					},
				},
			},
			BmcInfo: &cwssaws.BmcInfo{
				Ip:  cdb.GetStrPtr("10.100.1.1"),
				Mac: cdb.GetStrPtr("00-B0-D0-63-C2-26"),
			},
			Health: &cwssaws.HealthReport{
				Source: "aggregate-host-health",
				Successes: []*cwssaws.HealthProbeSuccess{
					{
						Id:     "BgpDaemonEnabled",
						Target: nil,
					},
					{
						Id:     "BgpStats",
						Target: nil,
					},
					{
						Id:     "ContainerExists",
						Target: nil,
					},
					{
						Id:     "DhcpServer",
						Target: nil,
					},
					{
						Id:     "FileExists",
						Target: cdb.GetStrPtr("/var/lib/hbn/etc/frr/daemons"),
					},
					{
						Id:     "FileExists",
						Target: cdb.GetStrPtr("/var/lib/hbn/etc/frr/frr.conf"),
					},
					{
						Id:     "FileExists",
						Target: cdb.GetStrPtr("/var/lib/hbn/etc/network/interfaces"),
					},
					{
						Id:     "FileExists",
						Target: cdb.GetStrPtr("/var/lib/hbn/etc/supervisor/conf.d/default-nico-dhcp-server.conf"),
					},
					{
						Id:     "FileExists",
						Target: cdb.GetStrPtr("/var/lib/hbn/etc/supervisor/conf.d/default-isc-dhcp-relay.conf"),
					},
					{
						Id:     "FileIsValid",
						Target: cdb.GetStrPtr("etc/frr/daemons"),
					},
					{
						Id:     "FileIsValid",
						Target: cdb.GetStrPtr("etc/frr/frr.conf"),
					},
					{
						Id:     "FileIsValid",
						Target: cdb.GetStrPtr("etc/network/interfaces"),
					},
					{
						Id:     "FileIsValid",
						Target: cdb.GetStrPtr("etc/supervisor/conf.d/default-nico-dhcp-server.conf"),
					},
					{
						Id:     "FileIsValid",
						Target: cdb.GetStrPtr("etc/supervisor/conf.d/default-isc-dhcp-relay.conf"),
					},
					{
						Id:     "Ifreload",
						Target: nil,
					},
					{
						Id:     "RestrictedMode",
						Target: nil,
					},
					{
						Id:     "ServiceRunning",
						Target: cdb.GetStrPtr("frr"),
					},
					{
						Id:     "ServiceRunning",
						Target: cdb.GetStrPtr("nl2doca"),
					},
					{
						Id:     "ServiceRunning",
						Target: cdb.GetStrPtr("rsyslog"),
					},
					{
						Id:     "SupervisorctlStatus",
						Target: nil,
					},
				},
				Alerts: []*cwssaws.HealthProbeAlert{
					{
						Id:            "HeartbeatTimeout",
						Target:        cdb.GetStrPtr("hardware-health"),
						InAlertSince:  nil,
						Message:       "",
						TenantMessage: nil,
						Classifications: []string{
							"PreventAllocations",
							"PreventHostStateChanges",
						},
					},
				},
			},
		},
	}

	// Convert Machine Health info data into health report interface
	var machineHealth map[string]interface{}
	machineHealthJSON, _ := json.Marshal(machineInfo1.Machine.Health)
	_ = json.Unmarshal(machineHealthJSON, &machineHealth)

	dbm := &cdbm.Machine{
		ID:                       mID,
		InfrastructureProviderID: uuid.New(),
		SiteID:                   uuid.New(),
		InstanceTypeID:           cdb.GetUUIDPtr(uuid.New()),
		ControllerMachineID:      mID,
		ControllerMachineType:    cdb.GetStrPtr("someType"),
		HwSkuDeviceType:          cdb.GetStrPtr("someHwSkuDeviceType"),
		Vendor:                   cdb.GetStrPtr("someVendor"),
		ProductName:              cdb.GetStrPtr("someProductName"),
		SerialNumber:             cdb.GetStrPtr(uuid.NewString()),
		Metadata:                 &cdbm.SiteControllerMachine{Machine: machineInfo1.Machine},
		Health:                   machineHealth,
		DefaultMacAddress:        cdb.GetStrPtr("00:00:00:00:00:00"),
		Hostname:                 cdb.GetStrPtr("test.com"),
		IsInMaintenance:          true,
		IsUsableByTenant:         true,
		MaintenanceMessage:       cdb.GetStrPtr("Scheduled maintenance"),
		Labels:                   map[string]string{"test": "test"},
		Status:                   cdbm.MachineStatusMaintenance,
		Created:                  cdb.GetCurTime(),
		Updated:                  cdb.GetCurTime(),
	}

	dbmcs := []cdbm.MachineCapability{
		{
			ID:             uuid.New(),
			MachineID:      cdb.GetStrPtr(dbm.ID),
			InstanceTypeID: cdb.GetUUIDPtr(uuid.New()),
			Type:           cdbm.MachineCapabilityTypeCPU,
			Name:           "AMD Opteron Series x10",
			Capacity:       cdb.GetStrPtr("3.0GHz"),
			Count:          cdb.GetIntPtr(2),
			Created:        cdb.GetCurTime(),
			Updated:        cdb.GetCurTime(),
		},
		{
			ID:             uuid.New(),
			MachineID:      cdb.GetStrPtr(dbm.ID),
			InstanceTypeID: cdb.GetUUIDPtr(uuid.New()),
			Type:           cdbm.MachineCapabilityTypeMemory,
			Name:           "Corsair Vengeance LPX",
			Capacity:       cdb.GetStrPtr("128GB"),
			Count:          cdb.GetIntPtr(2),
			Created:        cdb.GetCurTime(),
			Updated:        cdb.GetCurTime(),
		},
	}
	dbmis := []cdbm.MachineInterface{
		{
			ID:                    uuid.New(),
			MachineID:             uuid.NewString(),
			ControllerInterfaceID: cdb.GetUUIDPtr(uuid.New()),
			ControllerSegmentID:   cdb.GetUUIDPtr(uuid.New()),
			Hostname:              cdb.GetStrPtr("test.com"),
			IsPrimary:             true,
			SubnetID:              cdb.GetUUIDPtr(uuid.New()),
			MacAddress:            cdb.GetStrPtr("00:00:00:00:00:00"),
			IPAddresses:           []string{"192.168.0.1, 172.168.0.1"},
			Created:               cdb.GetCurTime(),
			Updated:               cdb.GetCurTime(),
		},
	}
	dbsds := []cdbm.StatusDetail{
		{
			ID:       uuid.New(),
			EntityID: dbm.ID,
			Status:   dbm.Status,
			Created:  cdb.GetCurTime(),
			Updated:  cdb.GetCurTime(),
		},
	}

	dbm.Site = &cdbm.Site{
		ID:                       dbm.SiteID,
		Name:                     "test-site",
		Description:              cdb.GetStrPtr("Test Description"),
		InfrastructureProviderID: dbm.InfrastructureProviderID,
		Status:                   cdbm.SiteStatusRegistered,
		Created:                  cdb.GetCurTime(),
		Updated:                  cdb.GetCurTime(),
		CreatedBy:                uuid.New(),
	}

	dbm.InstanceType = &cdbm.InstanceType{
		ID:                       uuid.New(),
		Name:                     "test",
		DisplayName:              cdb.GetStrPtr("Test"),
		Description:              cdb.GetStrPtr("Test Description"),
		InfrastructureProviderID: dbm.InfrastructureProviderID,
		SiteID:                   &dbm.Site.ID,
		Status:                   cdbm.InstanceTypeStatusReady,
		Created:                  cdb.GetCurTime(),
		Updated:                  cdb.GetCurTime(),
		CreatedBy:                dbm.Site.CreatedBy,
	}

	apimi := NewAPIMachine(dbm, dbmcs, dbmis, dbsds, nil, true, true)
	assert.NotNil(t, apimi)

	assert.Equal(t, apimi.ID, dbm.ID)
	assert.Equal(t, apimi.InfrastructureProviderID, dbm.InfrastructureProviderID.String())
	assert.Equal(t, apimi.SiteID, dbm.SiteID.String())
	assert.Equal(t, *apimi.InstanceTypeID, dbm.InstanceTypeID.String())
	assert.Equal(t, *apimi.ControllerMachineType, *dbm.ControllerMachineType)
	assert.Equal(t, *apimi.Vendor, *dbm.Vendor)
	assert.Equal(t, *apimi.ProductName, *dbm.ProductName)
	assert.Equal(t, *apimi.SerialNumber, *dbm.SerialNumber)

	assert.Equal(t, len(apimi.MachineCapabilities), len(dbmcs))

	assert.NotNil(t, apimi.Site)
	assert.Equal(t, apimi.Site.Name, dbm.Site.Name)

	assert.NotNil(t, apimi.InstanceType)
	assert.Equal(t, apimi.InstanceType.Name, dbm.InstanceType.Name)

	assert.Equal(t, *apimi.Hostname, *dbm.Hostname)
	assert.Equal(t, *apimi.MaintenanceMessage, *dbm.MaintenanceMessage)

	for i, v := range dbmcs {
		assert.Equal(t, apimi.MachineCapabilities[i].Type, v.Type)
	}
	for i, v := range dbmis {
		assert.Equal(t, apimi.MachineInterfaces[i].ID, v.ID.String())
	}
	for i, v := range dbsds {
		assert.Equal(t, apimi.StatusHistory[i].Status, v.Status)
	}

	if apimi.Metadata != nil {
		if apimi.Metadata.BMCInfo != nil {
			assert.Equal(t, *apimi.Metadata.BMCInfo.IP, *machineInfo1.Machine.BmcInfo.Ip)
			assert.Equal(t, *apimi.Metadata.BMCInfo.Mac, *machineInfo1.Machine.BmcInfo.Mac)
		}

		if apimi.Metadata.DMIData != nil {
			assert.Equal(t, *apimi.Metadata.DMIData.BoardName, machineInfo1.Machine.DiscoveryInfo.DmiData.BoardName)
			assert.Equal(t, *apimi.Metadata.DMIData.BoardVersion, machineInfo1.Machine.DiscoveryInfo.DmiData.BoardVersion)
			assert.Equal(t, *apimi.Metadata.DMIData.BiosDate, machineInfo1.Machine.DiscoveryInfo.DmiData.BiosDate)
			assert.Equal(t, *apimi.Metadata.DMIData.BiosVersion, machineInfo1.Machine.DiscoveryInfo.DmiData.BiosVersion)
			assert.Equal(t, *apimi.Metadata.DMIData.ProductSerial, machineInfo1.Machine.DiscoveryInfo.DmiData.ProductSerial)
			assert.Equal(t, *apimi.Metadata.DMIData.BoardSerial, machineInfo1.Machine.DiscoveryInfo.DmiData.BoardSerial)
			assert.Equal(t, *apimi.Metadata.DMIData.ChassisSerial, machineInfo1.Machine.DiscoveryInfo.DmiData.ChassisSerial)
			assert.Equal(t, *apimi.Metadata.DMIData.SysVendor, machineInfo1.Machine.DiscoveryInfo.DmiData.SysVendor)
		}

		if apimi.Metadata.GPUs != nil {
			assert.Equal(t, *apimi.Metadata.GPUs[0].Name, machineInfo1.Machine.DiscoveryInfo.Gpus[0].Name)
			assert.Equal(t, *apimi.Metadata.GPUs[0].Serial, machineInfo1.Machine.DiscoveryInfo.Gpus[0].Serial)
			assert.Equal(t, *apimi.Metadata.GPUs[0].DriverVersion, machineInfo1.Machine.DiscoveryInfo.Gpus[0].DriverVersion)
			assert.Equal(t, *apimi.Metadata.GPUs[0].VbiosVersion, machineInfo1.Machine.DiscoveryInfo.Gpus[0].VbiosVersion)
			assert.Equal(t, *apimi.Metadata.GPUs[0].InforomVersion, machineInfo1.Machine.DiscoveryInfo.Gpus[0].InforomVersion)
			assert.Equal(t, *apimi.Metadata.GPUs[0].TotalMemory, machineInfo1.Machine.DiscoveryInfo.Gpus[0].TotalMemory)
			assert.Equal(t, *apimi.Metadata.GPUs[0].Frequency, machineInfo1.Machine.DiscoveryInfo.Gpus[0].Frequency)
			assert.Equal(t, *apimi.Metadata.GPUs[0].PciBusId, machineInfo1.Machine.DiscoveryInfo.Gpus[0].PciBusId)
		}

		if apimi.Metadata.NetworkInterfaces != nil {
			assert.Equal(t, len(apimi.Metadata.NetworkInterfaces), len(machineInfo1.Machine.DiscoveryInfo.NetworkInterfaces))
		}

		if apimi.Metadata.InfiniBandInterfaces != nil {
			assert.Equal(t, len(apimi.Metadata.InfiniBandInterfaces), len(machineInfo1.Machine.DiscoveryInfo.InfinibandInterfaces))
		}
	}

	if apimi.Health != nil {
		assert.Equal(t, apimi.Health.Source, machineInfo1.Machine.Health.Source)
		if apimi.Health.ObservedAt != nil {
			assert.Equal(t, apimi.Health.ObservedAt, machineInfo1.Machine.Health.ObservedAt)
		}
		assert.Equal(t, len(apimi.Health.Successes), len(machineInfo1.Machine.Health.Successes))
		if apimi.Health.Alerts != nil {
			assert.Equal(t, apimi.Health.Alerts[0].ID, machineInfo1.Machine.Health.Alerts[0].Id)
			if apimi.Health.Alerts[0].Target != nil {
				assert.Equal(t, *apimi.Health.Alerts[0].Target, *machineInfo1.Machine.Health.Alerts[0].Target)
			}
			if apimi.Health.Alerts[0].TenantMessage != nil {
				assert.Equal(t, *apimi.Health.Alerts[0].TenantMessage, *machineInfo1.Machine.Health.Alerts[0].TenantMessage)
			}
			assert.Equal(t, apimi.Health.Alerts[0].Message, machineInfo1.Machine.Health.Alerts[0].Message)
			assert.Equal(t, len(apimi.Health.Alerts[0].Classifications), len(machineInfo1.Machine.Health.Alerts[0].Classifications))
		}
	}

	assert.Equal(t, apimi.Labels, dbm.Labels)
	assert.Equal(t, dbm.HwSkuDeviceType, apimi.HwSkuDeviceType)
	assert.Equal(t, dbm.IsUsableByTenant, apimi.IsUsableByTenant)

	if apimi.Deprecations != nil {
		assert.Equal(t, len(apimi.Deprecations), len(machineHealthAttributeDeprecations))
	}
}

func TestMachine_NewAPIMachineSummary(t *testing.T) {
	mID := uuid.NewString()
	dbm := &cdbm.Machine{
		ID:                       mID,
		InfrastructureProviderID: uuid.New(),
		SiteID:                   uuid.New(),
		InstanceTypeID:           cdb.GetUUIDPtr(uuid.New()),
		ControllerMachineID:      mID,
		ControllerMachineType:    cdb.GetStrPtr("someType"),
		HwSkuDeviceType:          cdb.GetStrPtr("someHwSkuDeviceType"),
		Vendor:                   cdb.GetStrPtr("someVendor"),
		ProductName:              cdb.GetStrPtr("someProductName"),
		SerialNumber:             cdb.GetStrPtr(uuid.NewString()),
		Metadata:                 nil,
		DefaultMacAddress:        cdb.GetStrPtr("00:00:00:00:00:00"),
		IsInMaintenance:          true,
		MaintenanceMessage:       cdb.GetStrPtr("Scheduled maintenance"),
		Status:                   cdbm.MachineStatusMaintenance,
		Created:                  cdb.GetCurTime(),
		Updated:                  cdb.GetCurTime(),
	}

	apims := NewAPIMachineSummary(dbm)
	assert.NotNil(t, apims)

	assert.Equal(t, dbm.ControllerMachineID, apims.ControllerMachineID)
	assert.Equal(t, dbm.ControllerMachineType, apims.ControllerMachineType)
	assert.Equal(t, dbm.HwSkuDeviceType, apims.HwSkuDeviceType)
	assert.Equal(t, dbm.Vendor, apims.Vendor)
	assert.Equal(t, dbm.ProductName, apims.ProductName)
	assert.Equal(t, *dbm.MaintenanceMessage, *apims.MaintenanceMessage)
	assert.Equal(t, dbm.Status, apims.Status)
}

func TestAPIMachineUpdateRequest_Validate(t *testing.T) {
	type fields struct {
		InstanceTypeID     *string
		ClearInstanceType  *bool
		SetMaintenanceMode *bool
		MaintenanceMessage *string
		Labels             map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test valid Machine update request with Instance Type ID",
			fields: fields{
				InstanceTypeID: cdb.GetStrPtr(uuid.NewString()),
			},
			wantErr: false,
		},
		{
			name: "test invalid Machine update request with Instance Type ID",
			fields: fields{
				InstanceTypeID: cdb.GetStrPtr("1234"),
			},
			wantErr: true,
		},
		{
			name: "test valid Machine update request to clear Instance Type",
			fields: fields{
				ClearInstanceType: cdb.GetBoolPtr(true),
			},
			wantErr: false,
		},
		{
			name: "test invalid Machine update request when both parameters are set",
			fields: fields{
				InstanceTypeID:    cdb.GetStrPtr(uuid.NewString()),
				ClearInstanceType: cdb.GetBoolPtr(true),
			},
			wantErr: true,
		},
		{
			name: "test invalid Machine update request when clearInstanceType is set to false",
			fields: fields{
				ClearInstanceType: cdb.GetBoolPtr(false),
			},
			wantErr: true,
		},
		{
			name: "test valid Machine update request with maintenance mode enabled and message",
			fields: fields{
				SetMaintenanceMode: cdb.GetBoolPtr(true),
				MaintenanceMessage: cdb.GetStrPtr("Scheduled maintenance"),
			},
			wantErr: false,
		},
		{
			name: "test invalid Machine update request when too many options are set",
			fields: fields{
				ClearInstanceType:  cdb.GetBoolPtr(true),
				SetMaintenanceMode: cdb.GetBoolPtr(true),
				InstanceTypeID:     cdb.GetStrPtr("a_uuid"),
			},
			wantErr: true,
		},
		{
			name: "test invalid Machine update request with maintenance mode enabled but no message",
			fields: fields{
				SetMaintenanceMode: cdb.GetBoolPtr(true),
			},
			wantErr: true,
		},
		{
			name: "test invalid Machine update request with maintenance mode enabled but maintenance message is empty",
			fields: fields{
				SetMaintenanceMode: cdb.GetBoolPtr(true),
				MaintenanceMessage: cdb.GetStrPtr(""),
			},
			wantErr: true,
		},
		{
			name: "test invalid Machine update request with maintenance mode enabled but all whitespace message",
			fields: fields{
				SetMaintenanceMode: cdb.GetBoolPtr(true),
				MaintenanceMessage: cdb.GetStrPtr("  \t\n "),
			},
			wantErr: true,
		},
		{
			name: "test invalid Machine update request with maintenance message but mode not set",
			fields: fields{
				MaintenanceMessage: cdb.GetStrPtr("Scheduled maintenance"),
			},
			wantErr: true,
		},
		{
			name: "test valid Machine update request with maintenance mode disabled",
			fields: fields{
				SetMaintenanceMode: cdb.GetBoolPtr(false),
			},
			wantErr: false,
		},
		{
			name:    "test invalid Machine update request when no parameters are set",
			fields:  fields{},
			wantErr: true,
		},
		{
			name: "test invalid Machine update request with maintenance mode enabled but message is less than 5 char",
			fields: fields{
				SetMaintenanceMode: cdb.GetBoolPtr(true),
				MaintenanceMessage: cdb.GetStrPtr("aa"),
			},
			wantErr: true,
		},
		{
			name: "test valid Machine update request with labels",
			fields: fields{
				Labels: map[string]string{"key": "value"},
			},
		},
		{
			name: "test invalid Machine update request with labels when key is empty",
			fields: fields{
				Labels: map[string]string{"": "value"},
			},
			wantErr: true,
		},
		{
			name: "test invalid Machine update request with labels when key is too long",
			fields: fields{
				Labels: map[string]string{
					util.GenerateRandomString(util.LabelKeyMaxLength+1, util.CharsetAlphaNumeric): "value",
				},
			},
			wantErr: true,
		},
		{
			name: "test invalid Machine update request with labels when value is too long",
			fields: fields{
				Labels: map[string]string{
					"key": util.GenerateRandomString(util.LabelValueMaxLength+1, util.CharsetAlphaNumeric),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mur := APIMachineUpdateRequest{
				InstanceTypeID:     tt.fields.InstanceTypeID,
				ClearInstanceType:  tt.fields.ClearInstanceType,
				SetMaintenanceMode: tt.fields.SetMaintenanceMode,
				MaintenanceMessage: tt.fields.MaintenanceMessage,
				Labels:             tt.fields.Labels,
			}
			err := mur.Validate()
			require.Equal(t, tt.wantErr, err != nil, "error: %v", err)
		})
	}
}
