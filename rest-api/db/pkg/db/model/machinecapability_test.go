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
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"testing"

	"github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	"github.com/NVIDIA/infra-controller-rest/db/pkg/db/paginator"
	stracer "github.com/NVIDIA/infra-controller-rest/db/pkg/tracer"
	"github.com/NVIDIA/infra-controller-rest/db/pkg/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	otrace "go.opentelemetry.io/otel/trace"
)

func TestMachineCapabilitySQLDAO_Create(t *testing.T) {
	ctx := context.Background()

	dbSession := util.TestInitDB(t)
	defer dbSession.Close()

	TestSetupSchema(t, dbSession)

	ip, site, ins := testMachineInstanceTypeBuildInstanceType(t, dbSession, "sm.x86")
	mach := testMachineBuildMachine(t, dbSession, ip.ID, site.ID, &ins.ID, ins.ControllerMachineType)

	mcd := NewMachineCapabilityDAO(dbSession)

	// OTEL Spanner configuration
	_, _, ctx = testCommonTraceProviderSetup(t, ctx)

	tests := []struct {
		desc               string
		mcs                []MachineCapability
		expectError        bool
		verifyChildSpanner bool
	}{
		{
			desc: "create machine capability for CPU",
			mcs: []MachineCapability{
				{
					MachineID:        &mach.ID,
					InstanceTypeID:   nil,
					Type:             MachineCapabilityTypeCPU,
					Name:             "AMD Opteron Series x10",
					Frequency:        db.GetStrPtr("3.0 Ghz"),
					Count:            db.GetIntPtr(2),
					Cores:            db.GetIntPtr(128),
					Threads:          db.GetIntPtr(256),
					HardwareRevision: db.GetStrPtr("v.12345"),
					Index:            5,
					Info: map[string]interface{}{
						"Version": "2.0",
					},
				},
			},
			expectError:        false,
			verifyChildSpanner: true,
		},
		{
			desc: "create instance capability for Memory",
			mcs: []MachineCapability{
				{
					MachineID:      nil,
					InstanceTypeID: &ins.ID,
					Type:           MachineCapabilityTypeMemory,
					Name:           "Corsair Vengeance LPX",
					Frequency:      db.GetStrPtr("3200 Mhz"),
					Capacity:       db.GetStrPtr("128GB"),
					Count:          db.GetIntPtr(4),
				},
			},
			expectError: false,
		},
		{
			desc: "create instance capability for Network",
			mcs: []MachineCapability{
				{
					MachineID:      nil,
					InstanceTypeID: &ins.ID,
					Type:           MachineCapabilityTypeNetwork,
					Name:           "MT42822 BlueField-2 integrated ConnectX-6 Dx network controller",
					Capacity:       db.GetStrPtr("100GB"),
					Count:          db.GetIntPtr(2),
					DeviceType:     db.GetStrPtr("DPU"),
				},
			},
			expectError: false,
		},
		{
			desc: "create machine capability for InfiniBand",
			mcs: []MachineCapability{
				{
					MachineID:       &mach.ID,
					InstanceTypeID:  nil,
					Type:            MachineCapabilityTypeInfiniBand,
					Name:            "MT28908 Family [ConnectX-6]",
					Vendor:          db.GetStrPtr("Mellanox Technologies"),
					Count:           db.GetIntPtr(2),
					InactiveDevices: []int{2, 4},
				},
			},
			expectError: false,
		},
		{
			desc: "create with both machine and instance set to nil",
			mcs: []MachineCapability{
				{
					MachineID:      nil,
					InstanceTypeID: nil,
					Type:           MachineCapabilityTypeMemory,
					Name:           "Corsair Vengeance LPX",
					Frequency:      db.GetStrPtr("3200 Mhz"),
					Capacity:       db.GetStrPtr("128GB"),
					Count:          db.GetIntPtr(4),
					Info: map[string]interface{}{
						"DDR": "v4",
					},
				},
			},
			expectError: true,
		},
		{
			desc: "create with invalid capability type",
			mcs: []MachineCapability{
				{
					MachineID:      &mach.ID,
					InstanceTypeID: nil,
					Type:           "",
					Name:           "AMD Opteron Series x10",
					Frequency:      db.GetStrPtr("3.0 Ghz"),
					Count:          db.GetIntPtr(2),
				},
			},
			expectError: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			for _, i := range tc.mcs {

				mc, err := mcd.Create(ctx, nil, MachineCapabilityCreateInput{
					MachineID:        i.MachineID,
					InstanceTypeID:   i.InstanceTypeID,
					Type:             i.Type,
					Name:             i.Name,
					Frequency:        i.Frequency,
					Capacity:         i.Capacity,
					Vendor:           i.Vendor,
					Cores:            i.Cores,
					Threads:          i.Threads,
					HardwareRevision: i.HardwareRevision,
					Count:            i.Count,
					DeviceType:       i.DeviceType,
					Info:             i.Info,
					Index:            i.Index,
				})
				assert.Equal(t, tc.expectError, err != nil)
				if !tc.expectError {
					assert.NotNil(t, mc)
					assert.Nil(t, err)
				} else {
					assert.Nil(t, mc)
				}

				if err != nil {
					t.Logf("%s", err.Error())
					return
				}

				assert.Equal(t, i.Index, mc.Index)

				if tc.verifyChildSpanner {
					span := otrace.SpanFromContext(ctx)
					assert.True(t, span.SpanContext().IsValid())
					_, ok := ctx.Value(stracer.TracerKey).(otrace.Tracer)
					assert.True(t, ok)
				}
			}
		})
	}
}

func testMachineCapabilitySQLDAOCreateSlice(ctx context.Context, t *testing.T, dbSession *db.Session) []MachineCapability {
	ip, site, ins := testMachineInstanceTypeBuildInstanceType(t, dbSession, "sm.x86")
	mach := testMachineBuildMachine(t, dbSession, ip.ID, site.ID, &ins.ID, ins.ControllerMachineType)

	mcd := NewMachineCapabilityDAO(dbSession)

	mcsExp := []MachineCapability{
		{
			MachineID:      &mach.ID,
			InstanceTypeID: nil,
			Type:           MachineCapabilityTypeCPU,
			Name:           "AMD Opteron Series x10",
			Frequency:      db.GetStrPtr("3.0 Ghz"),
			Count:          db.GetIntPtr(2),
			Info: map[string]interface{}{
				"Version": "2.0",
			},
		},
		{
			MachineID:      &mach.ID,
			InstanceTypeID: nil,
			Type:           MachineCapabilityTypeCPU,
			Name:           "Corsair Vengeance LPX DDR4",
			Frequency:      db.GetStrPtr("3200 Mhz"),
			Capacity:       db.GetStrPtr("32 GB"),
			Count:          db.GetIntPtr(2),
		},
		{
			MachineID:      &mach.ID,
			InstanceTypeID: nil,
			Type:           MachineCapabilityTypeStorage,
			Name:           "Dell Ent NVMe CM6 RI 1.92TB",
			Capacity:       db.GetStrPtr("1.92TB"),
			Count:          db.GetIntPtr(4),
		},
		{
			MachineID:      &mach.ID,
			InstanceTypeID: nil,
			Type:           MachineCapabilityTypeStorage,
			Name:           "Dell Ent NVMe CM6 RI 1.92TB",
			Capacity:       db.GetStrPtr("1.92TB"),
			Count:          db.GetIntPtr(4),
		},
		{
			MachineID:      &mach.ID,
			InstanceTypeID: nil,
			Type:           MachineCapabilityTypeNetwork,
			Name:           "MT42822 BlueField-2 integrated ConnectX-6 Dx network controller",
			Capacity:       db.GetStrPtr("100GB"),
			Count:          db.GetIntPtr(2),
			DeviceType:     db.GetStrPtr(""),
		},
		{
			MachineID:      &mach.ID,
			InstanceTypeID: nil,
			Type:           MachineCapabilityTypeInfiniBand,
			Name:           "MT28908 Family [ConnectX-6]",
			Vendor:         db.GetStrPtr("Mellanox Technologies"),
			Count:          db.GetIntPtr(2),
		},
	}

	// MachineCapability created
	for i := 0; i < len(mcsExp); i++ {
		mcCre, _ := mcd.Create(ctx, nil,
			MachineCapabilityCreateInput{
				MachineID:        mcsExp[i].MachineID,
				InstanceTypeID:   mcsExp[i].InstanceTypeID,
				Type:             mcsExp[i].Type,
				Name:             mcsExp[i].Name,
				Frequency:        mcsExp[i].Frequency,
				Capacity:         mcsExp[i].Capacity,
				Vendor:           mcsExp[i].Vendor,
				Cores:            mcsExp[i].Cores,
				DeviceType:       mcsExp[i].DeviceType,
				Threads:          mcsExp[i].Threads,
				HardwareRevision: mcsExp[i].HardwareRevision,
				Count:            mcsExp[i].Count,
				InactiveDevices:  mcsExp[i].InactiveDevices,
				Info:             mcsExp[i].Info,
			},
		)
		assert.NotNil(t, mcCre)
		mcsExp[i].ID = mcCre.ID
	}

	return mcsExp
}

func TestMachineCapabilitySQLDAO_GetByID(t *testing.T) {
	ctx := context.Background()

	dbSession := util.TestInitDB(t)
	defer dbSession.Close()

	TestSetupSchema(t, dbSession)

	mcsExp := testMachineCapabilitySQLDAOCreateSlice(ctx, t, dbSession)
	mcd := NewMachineCapabilityDAO(dbSession)

	// OTEL Spanner configuration
	_, _, ctx = testCommonTraceProviderSetup(t, ctx)

	tests := []struct {
		desc               string
		mc                 MachineCapability
		expectError        bool
		expectedErrVal     error
		paramRelations     []string
		verifyChildSpanner bool
	}{
		{
			desc:               "GetById success when MachineCapability exists on [1]",
			mc:                 mcsExp[0],
			expectError:        false,
			paramRelations:     []string{InstanceTypeRelationName},
			verifyChildSpanner: true,
		},
		{
			desc:           "GetById success when MachineCapability exists on [2]",
			mc:             mcsExp[1],
			expectError:    false,
			paramRelations: []string{InstanceTypeRelationName},
		},
		{
			desc: "GetById success when MachineCapability not found",
			mc: MachineCapability{
				ID: uuid.New(),
			},
			paramRelations: []string{InstanceTypeRelationName},
			expectError:    true,
			expectedErrVal: db.ErrDoesNotExist,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			tmp, err := mcd.GetByID(ctx, nil, tc.mc.ID, tc.paramRelations)
			assert.Equal(t, tc.expectError, err != nil)
			if tc.expectError {
				assert.Equal(t, tc.expectedErrVal, err)
			}
			if err == nil {
				assert.EqualValues(t, tc.mc.ID, tmp.ID)
				assert.EqualValues(t, tc.mc.MachineID, tmp.MachineID)
				assert.EqualValues(t, tc.mc.InstanceTypeID, tmp.InstanceTypeID)
				assert.EqualValues(t, tc.mc.Type, tmp.Type)
				assert.EqualValues(t, tc.mc.Name, tmp.Name)
				assert.EqualValues(t, tc.mc.Frequency, tmp.Frequency)
				assert.EqualValues(t, tc.mc.Capacity, tmp.Capacity)
				assert.EqualValues(t, tc.mc.Vendor, tmp.Vendor)
				assert.EqualValues(t, tc.mc.Count, tmp.Count)
				assert.EqualValues(t, tc.mc.Info, tmp.Info)
			} else {
				t.Logf("%s", err.Error())
			}
			if tc.verifyChildSpanner {
				span := otrace.SpanFromContext(ctx)
				assert.True(t, span.SpanContext().IsValid())
				_, ok := ctx.Value(stracer.TracerKey).(otrace.Tracer)
				assert.True(t, ok)
			}
		})
	}
}

func TestMachineCapabilitySQLDAO_GetAll(t *testing.T) {
	ctx := context.Background()

	dbSession := util.TestInitDB(t)
	defer dbSession.Close()

	TestSetupSchema(t, dbSession)

	mcd := NewMachineCapabilityDAO(dbSession)

	itCapCount := 0

	ip, site, it := testMachineInstanceTypeBuildInstanceType(t, dbSession, "sm.x86")
	mcd.Create(ctx, nil, MachineCapabilityCreateInput{InstanceTypeID: &it.ID, Type: MachineCapabilityTypeCPU, Name: "Test Capability", Frequency: db.GetStrPtr("3.2 GHz"), Count: db.GetIntPtr(2)})
	mcd.Create(ctx, nil, MachineCapabilityCreateInput{InstanceTypeID: &it.ID, Type: MachineCapabilityTypeMemory, Name: "Test Capability", Capacity: db.GetStrPtr("32GB"), Count: db.GetIntPtr(4)})
	itCapCount += 2

	user := TestBuildUser(t, dbSession, "test-user", "test-org", []string{"test-role"})
	it2 := TestBuildInstanceType(t, dbSession, "lg.x86", ip, site, user)
	mcd.Create(ctx, nil, MachineCapabilityCreateInput{InstanceTypeID: &it2.ID, Type: MachineCapabilityTypeCPU, Name: "Test Capability", Frequency: db.GetStrPtr("4 GHz"), Count: db.GetIntPtr(1)})
	itCapCount++

	ms := []Machine{}
	mcs := []MachineCapability{}
	var lastEntry *MachineCapability
	totalMachineCount := 10

	lastCount := 0
	totalCount := totalMachineCount*len(MachineCapabilityTypeChoiceMap) + itCapCount

	for i := 0; i < totalMachineCount; i++ {
		m := testMachineBuildMachine(t, dbSession, ip.ID, site.ID, &it.ID, it.ControllerMachineType)
		ms = append(ms, *m)

		capTypes := make([]string, 0, len(MachineCapabilityTypeChoiceMap))
		for cap := range MachineCapabilityTypeChoiceMap {
			capTypes = append(capTypes, cap)
		}
		sort.Strings(capTypes)

		for _, cap := range capTypes {
			var vendor *string
			var deviceType *string
			if cap == MachineCapabilityTypeInfiniBand {
				vendor = db.GetStrPtr("Test Vendor")
			}
			if i == 0 && cap == MachineCapabilityTypeNetwork {
				deviceType = db.GetStrPtr("DPU")
			}

			var inactiveDevices []int
			if cap == MachineCapabilityTypeInfiniBand {
				inactiveDevices = []int{1, 3}
			}

			mc, err := mcd.Create(ctx, nil, MachineCapabilityCreateInput{
				MachineID: &m.ID, Type: cap,
				Name:            "Test Capability",
				Frequency:       db.GetStrPtr("3 GHz"),
				Capacity:        db.GetStrPtr("12 TB"),
				Vendor:          vendor,
				Count:           db.GetIntPtr(1),
				DeviceType:      deviceType,
				InactiveDevices: inactiveDevices,
			})
			assert.NoError(t, err)
			mcs = append(mcs, *mc)

			// Track last one
			lastCount = +1
			if lastCount == totalCount {
				lastEntry = mc
			}
		}
	}

	// OTEL Spanner configuration
	_, _, ctx = testCommonTraceProviderSetup(t, ctx)

	tests := []struct {
		desc               string
		MachineIDs         []string
		InstanceTypeIDs    []uuid.UUID
		Type               *string
		Name               *string
		Frequency          *string
		Capacity           *string
		Vendor             *string
		Count              *int
		DeviceType         *string
		InactiveDevices    []int
		offset             *int
		limit              *int
		orderBy            *paginator.OrderBy
		firstEntry         *MachineCapability
		expectedCount      int
		expectedTotal      *int
		paramRelations     []string
		verifyChildSpanner bool
	}{
		{
			desc:               "GetAll with no filters returns all objects",
			MachineIDs:         nil,
			InstanceTypeIDs:    nil,
			Type:               nil,
			Name:               nil,
			Frequency:          nil,
			Capacity:           nil,
			Vendor:             nil,
			Count:              nil,
			DeviceType:         nil,
			expectedCount:      paginator.DefaultLimit,
			expectedTotal:      &totalCount,
			verifyChildSpanner: true,
		},
		{
			desc:            "GetAll with relation returns all objects",
			MachineIDs:      nil,
			InstanceTypeIDs: nil,
			Type:            nil,
			Name:            nil,
			Frequency:       nil,
			Capacity:        nil,
			Vendor:          nil,
			Count:           nil,
			DeviceType:      nil,
			expectedCount:   paginator.DefaultLimit,
			expectedTotal:   &totalCount,
			paramRelations:  []string{InstanceTypeRelationName},
		},
		{
			desc:            "GetAll with machine id filter",
			MachineIDs:      []string{ms[0].ID, ms[1].ID},
			InstanceTypeIDs: nil,
			Type:            nil,
			Name:            nil,
			Frequency:       nil,
			Capacity:        nil,
			Vendor:          nil,
			Count:           nil,
			DeviceType:      nil,
			expectedCount:   len(MachineCapabilityTypeChoiceMap) * 2,
		},
		{
			desc:            "GetAll with non-existent machine id filter",
			MachineIDs:      []string{"test-id"},
			InstanceTypeIDs: nil,
			Type:            nil,
			Name:            nil,
			Frequency:       nil,
			Capacity:        nil,
			Vendor:          nil,
			Count:           nil,
			DeviceType:      nil,
			expectedCount:   0,
		},
		{
			desc:            "GetAll with instance id filter",
			MachineIDs:      nil,
			InstanceTypeIDs: []uuid.UUID{it.ID},
			Type:            nil,
			Name:            nil,
			Frequency:       nil,
			Capacity:        nil,
			Vendor:          nil,
			Count:           nil,
			DeviceType:      nil,
			expectedCount:   2,
		},
		{
			desc:            "GetAll with Type filter",
			MachineIDs:      nil,
			InstanceTypeIDs: nil,
			Type:            db.GetStrPtr(MachineCapabilityTypeGPU),
			Name:            nil,
			Frequency:       nil,
			Capacity:        nil,
			Vendor:          nil,
			Count:           nil,
			expectedCount:   totalMachineCount,
		},
		{
			desc:            "GetAll with Type and Name filter",
			MachineIDs:      nil,
			InstanceTypeIDs: nil,
			Type:            db.GetStrPtr(MachineCapabilityTypeGPU),
			Name:            &mcs[0].Name,
			Frequency:       nil,
			Capacity:        nil,
			Vendor:          nil,
			Count:           nil,
			DeviceType:      nil,
			expectedCount:   totalMachineCount,
		},
		{
			desc:            "GetAll with Type and Capacity filter",
			MachineIDs:      nil,
			InstanceTypeIDs: nil,
			Type:            &mcs[1].Type,
			Name:            nil,
			Frequency:       nil,
			Capacity:        mcs[1].Capacity,
			Vendor:          nil,
			Count:           nil,
			DeviceType:      nil,
			expectedCount:   totalMachineCount,
		},
		{
			desc:            "GetAll with Type and Frequency filter",
			MachineIDs:      nil,
			InstanceTypeIDs: nil,
			Type:            &mcs[1].Type,
			Name:            nil,
			Frequency:       mcs[1].Frequency,
			Vendor:          nil,
			Capacity:        nil,
			Count:           nil,
			DeviceType:      nil,
			expectedCount:   totalMachineCount,
		},
		{
			desc:            "GetAll with Type and Count filter",
			MachineIDs:      nil,
			InstanceTypeIDs: nil,
			Type:            db.GetStrPtr(MachineCapabilityTypeGPU),
			Name:            nil,
			Frequency:       nil,
			Capacity:        nil,
			Vendor:          nil,
			Count:           db.GetIntPtr(1),
			expectedCount:   totalMachineCount,
		},
		{
			desc:            "GetAll with Vendor filter",
			MachineIDs:      nil,
			InstanceTypeIDs: nil,
			Type:            nil,
			Name:            nil,
			Frequency:       nil,
			Capacity:        nil,
			Vendor:          db.GetStrPtr("Test Vendor"),
			Count:           nil,
			DeviceType:      nil,
			expectedCount:   totalMachineCount,
		},
		{
			desc:            "GetAll with DeviceType filter",
			MachineIDs:      nil,
			InstanceTypeIDs: nil,
			Type:            nil,
			Name:            nil,
			Frequency:       nil,
			Capacity:        nil,
			Vendor:          nil,
			Count:           nil,
			DeviceType:      db.GetStrPtr("DPU"),
			expectedCount:   1,
		},
		{
			desc:            "GetAll with InactiveDevices filter",
			InactiveDevices: []int{1, 3},
			expectedCount:   totalMachineCount,
		},
		{
			desc:            "GetAll with limit returns objects",
			MachineIDs:      nil,
			InstanceTypeIDs: nil,
			Type:            nil,
			offset:          db.GetIntPtr(0),
			limit:           db.GetIntPtr(5),
			expectedCount:   5,
			expectedTotal:   &totalCount,
		},
		{
			desc:            "GetAll with offset returns objects",
			MachineIDs:      nil,
			InstanceTypeIDs: nil,
			Type:            nil,
			offset:          db.GetIntPtr(5),
			expectedCount:   paginator.DefaultLimit,
			expectedTotal:   &totalCount,
		},
		{
			desc:            "GetAll with order by returns objects",
			MachineIDs:      nil,
			InstanceTypeIDs: nil,
			Type:            nil,
			orderBy: &paginator.OrderBy{
				Field: "type",
				Order: paginator.OrderDescending,
			},
			firstEntry:    lastEntry, // last one created entry which would be the first in descending sort
			expectedCount: paginator.DefaultLimit,
			expectedTotal: &totalCount,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, total, err := mcd.GetAll(ctx, nil, tc.MachineIDs, tc.InstanceTypeIDs, tc.Type, tc.Name, tc.Frequency, tc.Capacity, tc.Vendor, tc.Count, tc.DeviceType, tc.InactiveDevices, tc.paramRelations,
				tc.offset, tc.limit, tc.orderBy)
			if err != nil {
				t.Logf("%s", err.Error())
			}
			assert.Equal(t, tc.expectedCount, len(got))
			if len(tc.paramRelations) > 0 {
				assert.NotNil(t, got[0].InstanceType)
			}

			if tc.expectedTotal != nil {
				assert.Equal(t, *tc.expectedTotal, total)
			}

			if tc.firstEntry != nil {
				assert.Equal(t, tc.firstEntry.Type, got[0].Type)
			}
			if tc.verifyChildSpanner {
				span := otrace.SpanFromContext(ctx)
				assert.True(t, span.SpanContext().IsValid())
				_, ok := ctx.Value(stracer.TracerKey).(otrace.Tracer)
				assert.True(t, ok)
			}
		})
	}
}

func TestMachineCapabilitySQLDAO_GetAllDistinct(t *testing.T) {
	ctx := context.Background()

	dbSession := util.TestInitDB(t)
	defer dbSession.Close()

	TestSetupSchema(t, dbSession)

	mcd := NewMachineCapabilityDAO(dbSession)

	ip, site, it := testMachineInstanceTypeBuildInstanceType(t, dbSession, "sm.x86")

	ms := []Machine{}
	mcs := []MachineCapability{}

	totalMachineCount := 10
	for i := 0; i < totalMachineCount; i++ {
		m := testMachineBuildMachine(t, dbSession, ip.ID, site.ID, &it.ID, it.ControllerMachineType)
		ms = append(ms, *m)

		capTypes := make([]string, 0, len(MachineCapabilityTypeChoiceMap))
		for cap := range MachineCapabilityTypeChoiceMap {
			capTypes = append(capTypes, cap)
		}
		sort.Strings(capTypes)

		for _, cap := range capTypes {
			var deviceType *string
			if i == 0 && cap == MachineCapabilityTypeNetwork {
				deviceType = db.GetStrPtr("DPU")
			}
			var inactiveDevices []int
			if cap == MachineCapabilityTypeInfiniBand {
				inactiveDevices = []int{1, 3}
			}
			mc, err := mcd.Create(ctx, nil, MachineCapabilityCreateInput{
				MachineID:       &m.ID,
				InstanceTypeID:  &it.ID,
				Type:            cap,
				Name:            "Test Capability",
				Frequency:       db.GetStrPtr("3 GHz"),
				Capacity:        db.GetStrPtr("12 TB"),
				Vendor:          db.GetStrPtr("Test Vendor"),
				Count:           db.GetIntPtr(1),
				DeviceType:      deviceType,
				InactiveDevices: inactiveDevices,
			})
			assert.NoError(t, err)
			mcs = append(mcs, *mc)
		}
	}

	// totalCount := totalMachineCount * len(MachineCapabilityTypeChoiceMap)

	// OTEL Spanner configuration
	_, _, ctx = testCommonTraceProviderSetup(t, ctx)

	tests := []struct {
		desc               string
		MachineIDs         []string
		InstanceTypeID     *uuid.UUID
		Type               *string
		Name               *string
		Frequency          *string
		Capacity           *string
		Vendor             *string
		Count              *int
		DeviceType         *string
		InactiveDevices    []int
		expectedCount      int
		expectedTotal      *int
		verifyChildSpanner bool
	}{
		{
			desc:               "GetAll with no filters returns all objects",
			MachineIDs:         nil,
			InstanceTypeID:     nil,
			Type:               nil,
			Name:               nil,
			Frequency:          nil,
			Capacity:           nil,
			Vendor:             nil,
			Count:              nil,
			DeviceType:         nil,
			expectedCount:      len(MachineCapabilityTypeChoiceMap) + 1,
			verifyChildSpanner: true,
		},
		{
			desc:           "GetAll with machine id filter",
			MachineIDs:     []string{ms[0].ID, ms[1].ID},
			InstanceTypeID: nil,
			Type:           nil,
			Name:           nil,
			Frequency:      nil,
			Capacity:       nil,
			Vendor:         nil,
			Count:          nil,
			DeviceType:     nil,
			expectedCount:  len(MachineCapabilityTypeChoiceMap) + 1,
		},
		{
			desc:           "GetAll with non-existent machine id filter",
			MachineIDs:     []string{"test-id"},
			InstanceTypeID: nil,
			Type:           nil,
			Name:           nil,
			Frequency:      nil,
			Capacity:       nil,
			Vendor:         nil,
			Count:          nil,
			DeviceType:     nil,
			expectedCount:  0,
		},
		{
			desc:           "GetAll with instance id filter",
			MachineIDs:     nil,
			InstanceTypeID: &it.ID,
			Type:           nil,
			Name:           nil,
			Frequency:      nil,
			Capacity:       nil,
			Vendor:         nil,
			Count:          nil,
			DeviceType:     nil,
			expectedCount:  len(MachineCapabilityTypeChoiceMap) + 1,
		},
		{
			desc:           "GetAll with Type filter",
			MachineIDs:     nil,
			InstanceTypeID: nil,
			Type:           db.GetStrPtr(MachineCapabilityTypeCPU),
			Name:           nil,
			Frequency:      nil,
			Capacity:       nil,
			Vendor:         nil,
			Count:          nil,
			DeviceType:     nil,
			expectedCount:  1,
		},
		{
			desc:           "GetAll with Type and Name filter",
			MachineIDs:     nil,
			InstanceTypeID: nil,
			Type:           &mcs[0].Type,
			Name:           &mcs[0].Name,
			Frequency:      nil,
			Capacity:       nil,
			Vendor:         nil,
			Count:          nil,
			DeviceType:     nil,
			expectedCount:  1,
		},
		{
			desc:           "GetAll with Type and Capacity filter",
			MachineIDs:     nil,
			InstanceTypeID: nil,
			Type:           &mcs[1].Type,
			Name:           nil,
			Frequency:      nil,
			Capacity:       mcs[1].Capacity,
			Vendor:         nil,
			Count:          nil,
			DeviceType:     nil,
			expectedCount:  1,
		},
		{
			desc:           "GetAll with Type and Frequency filter",
			MachineIDs:     nil,
			InstanceTypeID: nil,
			Type:           &mcs[1].Type,
			Name:           nil,
			Frequency:      mcs[1].Frequency,
			Vendor:         nil,
			Capacity:       nil,
			Count:          nil,
			DeviceType:     nil,
			expectedCount:  1,
		},
		{
			desc:           "GetAll with Type and Count filter",
			MachineIDs:     nil,
			InstanceTypeID: nil,
			Type:           &mcs[1].Type,
			Name:           nil,
			Frequency:      nil,
			Capacity:       nil,
			Vendor:         nil,
			Count:          mcs[1].Count,
			DeviceType:     nil,
			expectedCount:  1,
		},
		{
			desc:           "GetAll with Vendor filter",
			MachineIDs:     nil,
			InstanceTypeID: nil,
			Type:           nil,
			Name:           nil,
			Frequency:      nil,
			Capacity:       nil,
			Vendor:         db.GetStrPtr("Test Vendor"),
			Count:          nil,
			DeviceType:     nil,
			expectedCount:  len(MachineCapabilityTypeChoiceMap) + 1,
		},
		{
			desc:           "GetAll with DeviceType filter",
			MachineIDs:     nil,
			InstanceTypeID: nil,
			Type:           nil,
			Name:           nil,
			Frequency:      nil,
			Capacity:       nil,
			Vendor:         nil,
			Count:          nil,
			DeviceType:     db.GetStrPtr("DPU"),
			expectedCount:  1,
		},
		{
			desc:            "GetAll with InactiveDevices filter",
			InactiveDevices: []int{1, 3},
			expectedCount:   1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, total, err := mcd.GetAllDistinct(ctx, nil, tc.MachineIDs, tc.InstanceTypeID, tc.Type, tc.Name, tc.Frequency, tc.Capacity, tc.Vendor, tc.Count, tc.DeviceType, tc.InactiveDevices, nil, db.GetIntPtr(paginator.TotalLimit), nil)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCount, len(got))

			if tc.expectedTotal != nil {
				assert.Equal(t, *tc.expectedTotal, total)
			}

			if tc.verifyChildSpanner {
				span := otrace.SpanFromContext(ctx)
				assert.True(t, span.SpanContext().IsValid())
				_, ok := ctx.Value(stracer.TracerKey).(otrace.Tracer)
				assert.True(t, ok)
			}
		})
	}
}

func TestMachineCapabilitySQLDAO_Update(t *testing.T) {
	ctx := context.Background()

	dbSession := util.TestInitDB(t)
	defer dbSession.Close()

	TestSetupSchema(t, dbSession)

	mcsExp := testMachineCapabilitySQLDAOCreateSlice(ctx, t, dbSession)
	mcd := NewMachineCapabilityDAO(dbSession)
	assert.NotNil(t, mcd)

	name := "Corsair Vengeance LPX DDR4"
	frequency := "3200 MHz"
	capacity := "32 GB"
	count := 4
	deviceType := "DPU"

	// OTEL Spanner configuration
	_, _, ctx = testCommonTraceProviderSetup(t, ctx)

	tests := []struct {
		desc               string
		ID                 uuid.UUID
		mc                 *MachineCapability
		MachineID          *string
		InstanceTypeID     *uuid.UUID
		Type               *string
		Name               *string
		Frequency          *string
		Capacity           *string
		Vendor             *string
		Count              *int
		DeviceType         *string
		Info               map[string]interface{}
		Index              *int
		Threads            *int
		Cores              *int
		HardwareRevision   *string
		InactiveDevices    []int
		expectedError      bool
		verifyChildSpanner bool
	}{
		{
			desc:               "Update instance id",
			ID:                 mcsExp[0].ID,
			mc:                 &mcsExp[0],
			MachineID:          nil,
			InstanceTypeID:     mcsExp[3].InstanceTypeID,
			Type:               nil,
			Name:               nil,
			Frequency:          nil,
			Capacity:           nil,
			Count:              nil,
			expectedError:      false,
			verifyChildSpanner: true,
		},
		{
			desc:           "Update machine ID",
			ID:             mcsExp[3].ID,
			mc:             &mcsExp[3],
			MachineID:      mcsExp[0].MachineID,
			InstanceTypeID: nil,
			Type:           nil,
			Name:           nil,
			Frequency:      nil,
			Capacity:       nil,
			Count:          nil,
			expectedError:  false,
		},
		{
			desc:             "Update Capability Type and Name, Frequency, Capacity, Cores, Threads, HardwareRevision, Count",
			ID:               mcsExp[2].ID,
			mc:               &mcsExp[2],
			MachineID:        nil,
			InstanceTypeID:   nil,
			Type:             db.GetStrPtr(MachineCapabilityTypeMemory),
			Name:             db.GetStrPtr(name),
			Frequency:        db.GetStrPtr(frequency),
			Capacity:         db.GetStrPtr(capacity),
			Count:            db.GetIntPtr(count),
			Cores:            db.GetIntPtr(128),
			Threads:          db.GetIntPtr(256),
			HardwareRevision: db.GetStrPtr("v.12345"),
			Index:            db.GetIntPtr(1000),
			InactiveDevices:  []int{2, 4},
			expectedError:    false,
		},
		{
			desc:           "Update Vendor",
			ID:             mcsExp[3].ID,
			mc:             &mcsExp[3],
			MachineID:      nil,
			InstanceTypeID: nil,
			Type:           nil,
			Name:           nil,
			Frequency:      nil,
			Capacity:       nil,
			Vendor:         db.GetStrPtr("Test Vendor"),
			Count:          nil,
			expectedError:  false,
		},
		{
			desc:           "Update DeviceType for Network",
			ID:             mcsExp[4].ID,
			mc:             &mcsExp[4],
			MachineID:      nil,
			InstanceTypeID: nil,
			Type:           nil,
			Name:           nil,
			Frequency:      nil,
			Capacity:       nil,
			Vendor:         nil,
			DeviceType:     db.GetStrPtr(deviceType),
			Count:          nil,
			expectedError:  false,
		},
		{
			desc:           "Update to invalid Capability Type",
			ID:             mcsExp[2].ID,
			mc:             &mcsExp[2],
			MachineID:      nil,
			InstanceTypeID: nil,
			Type:           db.GetStrPtr("invalid-type"),
			Name:           &name,
			Frequency:      &frequency,
			Capacity:       &capacity,
			Count:          &count,
			expectedError:  true,
		},
		{
			desc:           "Update info",
			ID:             mcsExp[2].ID,
			mc:             &mcsExp[2],
			MachineID:      nil,
			InstanceTypeID: nil,
			Type:           nil,
			Name:           nil,
			Frequency:      nil,
			Capacity:       nil,
			Count:          nil,
			Info: map[string]interface{}{
				"test": "test",
			},
			expectedError: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := mcd.Update(ctx, nil,
				MachineCapabilityUpdateInput{
					ID:               tc.ID,
					MachineID:        tc.MachineID,
					InstanceTypeID:   tc.InstanceTypeID,
					Type:             tc.Type,
					Name:             tc.Name,
					Frequency:        tc.Frequency,
					Capacity:         tc.Capacity,
					Vendor:           tc.Vendor,
					Count:            tc.Count,
					DeviceType:       tc.DeviceType,
					Info:             tc.Info,
					Index:            tc.Index,
					Threads:          tc.Threads,
					Cores:            tc.Cores,
					InactiveDevices:  tc.InactiveDevices,
					HardwareRevision: tc.HardwareRevision,
				},
			)

			if tc.expectedError {
				return
			} else {
				assert.NoError(t, err)
			}

			assert.Nil(t, err)
			assert.NotNil(t, got)

			if tc.MachineID != nil {
				assert.Equal(t, *tc.MachineID, *got.MachineID)
			}
			if tc.InstanceTypeID != nil {
				assert.Equal(t, *tc.InstanceTypeID, *got.InstanceTypeID)
			}
			if tc.Type != nil {
				assert.Equal(t, *tc.Type, got.Type)
			}
			if tc.Name != nil {
				assert.Equal(t, *tc.Name, got.Name)
			}
			if tc.Frequency != nil {
				assert.Equal(t, *tc.Frequency, *got.Frequency)
			}
			if tc.Capacity != nil {
				assert.Equal(t, *tc.Capacity, *got.Capacity)
			}
			if tc.Vendor != nil {
				assert.Equal(t, *tc.Vendor, *got.Vendor)
			}
			if tc.Count != nil {
				assert.Equal(t, *tc.Count, *got.Count)
			}
			if tc.DeviceType != nil {
				assert.Equal(t, *tc.DeviceType, *got.DeviceType)
			}
			if tc.Info != nil {
				assert.Equal(t, tc.Info, got.Info)
			}
			if tc.Threads != nil {
				assert.Equal(t, *tc.Threads, *got.Threads)
			}
			if tc.Cores != nil {
				assert.Equal(t, *tc.Cores, *got.Cores)
			}
			if tc.Index != nil {
				assert.Equal(t, *tc.Index, got.Index)
			}
			if tc.HardwareRevision != nil {
				assert.Equal(t, *tc.HardwareRevision, *got.HardwareRevision)
			}
			if tc.InactiveDevices != nil {
				assert.Equal(t, tc.InactiveDevices, got.InactiveDevices)
			}

			assert.NotEqual(t, got.Updated.String(), tc.mc.Updated.String())

			if tc.verifyChildSpanner {
				span := otrace.SpanFromContext(ctx)
				assert.True(t, span.SpanContext().IsValid())
				_, ok := ctx.Value(stracer.TracerKey).(otrace.Tracer)
				assert.True(t, ok)
			}
		})
	}
}

func TestMachineCapabilitySQLDAO_ClearFromParams(t *testing.T) {
	ctx := context.Background()

	dbSession := util.TestInitDB(t)
	defer dbSession.Close()

	TestSetupSchema(t, dbSession)

	mcsExp := testMachineCapabilitySQLDAOCreateSlice(ctx, t, dbSession)
	mcd := NewMachineCapabilityDAO(dbSession)
	assert.NotNil(t, mcd)

	// OTEL Spanner configuration
	_, _, ctx = testCommonTraceProviderSetup(t, ctx)

	tests := []struct {
		desc                string
		mc                  MachineCapability
		paramMachineID      bool
		paramInstanceTypeID bool
		paramFrequency      bool
		paramCapacity       bool
		paramVendor         bool
		paramInfo           bool
		expectError         bool
		expectUpdate        bool
		verifyChildSpanner  bool
	}{
		{
			desc:                "can clear MachineID",
			mc:                  mcsExp[0],
			paramMachineID:      true,
			paramInstanceTypeID: false,
			paramFrequency:      false,
			paramCapacity:       false,
			paramInfo:           false,
			expectError:         false,
			expectUpdate:        true,
			verifyChildSpanner:  true,
		},
		{
			desc:                "cannot clear both InstanceTypeID and MachineID",
			mc:                  mcsExp[0],
			paramMachineID:      true,
			paramInstanceTypeID: true,
			paramFrequency:      false,
			paramCapacity:       false,
			paramInfo:           false,
			expectError:         true,
			expectUpdate:        false,
		},
		{
			desc:                "nop when no cleared fields are specified",
			mc:                  mcsExp[1],
			paramMachineID:      false,
			paramInstanceTypeID: false,
			paramFrequency:      false,
			paramCapacity:       false,
			paramInfo:           false,
			expectError:         false,
			expectUpdate:        true,
		},
		{
			desc:                "can clear capacity, frequency and info",
			mc:                  mcsExp[1],
			paramMachineID:      false,
			paramInstanceTypeID: false,
			paramFrequency:      true,
			paramCapacity:       true,
			paramInfo:           true,
			expectUpdate:        true,
		},
		{
			desc:                "can clear capacity, frequency and info",
			mc:                  mcsExp[1],
			paramMachineID:      false,
			paramInstanceTypeID: false,
			paramFrequency:      false,
			paramCapacity:       false,
			paramVendor:         true,
			paramInfo:           false,
			expectUpdate:        true,
		},
		{
			desc:                "can clear InstanceTypeID",
			mc:                  mcsExp[3],
			paramMachineID:      false,
			paramInstanceTypeID: true,
			paramFrequency:      false,
			paramCapacity:       false,
			paramInfo:           false,
			expectError:         false,
			expectUpdate:        true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			tmp, err := mcd.ClearFromParams(ctx, nil, tc.mc.ID,
				tc.paramMachineID, tc.paramInstanceTypeID, tc.paramFrequency, tc.paramCapacity, tc.paramVendor, tc.paramInfo)

			if tc.expectError {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.NotNil(t, tmp)

			if tc.paramMachineID {
				assert.Nil(t, tmp.MachineID)
			}
			if tc.paramInstanceTypeID {
				assert.Nil(t, tmp.InstanceTypeID)
			}
			if tc.paramFrequency {
				assert.Nil(t, tmp.Frequency)
			}
			if tc.paramCapacity {
				assert.Nil(t, tmp.Capacity)
			}
			if tc.paramVendor {
				assert.Nil(t, tmp.Vendor)
			}
			if tc.paramInfo {
				assert.Nil(t, tmp.Info)
			}
			if tc.expectUpdate {
				assert.True(t, tmp.Updated.After(tc.mc.Updated))
			}
			if tc.verifyChildSpanner {
				span := otrace.SpanFromContext(ctx)
				assert.True(t, span.SpanContext().IsValid())
				_, ok := ctx.Value(stracer.TracerKey).(otrace.Tracer)
				assert.True(t, ok)
			}
		})
	}
}

func TestMachineCapabilitySQLDAO_DeleteByID(t *testing.T) {
	ctx := context.Background()

	dbSession := util.TestInitDB(t)
	defer dbSession.Close()

	TestSetupSchema(t, dbSession)

	mcsExp := testMachineCapabilitySQLDAOCreateSlice(ctx, t, dbSession)

	mcDAO := NewMachineCapabilityDAO(dbSession)

	// OTEL Spanner configuration
	_, _, ctx = testCommonTraceProviderSetup(t, ctx)

	tests := []struct {
		desc               string
		mcID               uuid.UUID
		purge              bool
		expectedError      bool
		verifyChildSpanner bool
	}{
		{
			desc:               "test deleting existing object success",
			mcID:               mcsExp[1].ID,
			expectedError:      false,
			verifyChildSpanner: true,
		},
		{
			desc:          "test deleting non-existent object success",
			mcID:          uuid.New(),
			expectedError: false,
		},
		{
			desc:          "test deleting existing object with purge success",
			mcID:          mcsExp[2].ID,
			purge:         true,
			expectedError: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := mcDAO.DeleteByID(ctx, nil, tc.mcID, tc.purge)

			if tc.expectedError {
				assert.Error(t, err)

				// Check that object was not deleted
				tmp, serr := mcDAO.GetByID(ctx, nil, tc.mcID, nil)
				assert.NoError(t, serr)
				assert.NotNil(t, tmp)
				return
			}

			var res MachineCapability

			if tc.purge {
				err = dbSession.DB.NewSelect().Model(&res).Where("mc.id = ?", tc.mcID).WhereAllWithDeleted().Scan(ctx)
			} else {
				err = dbSession.DB.NewSelect().Model(&res).Where("mc.id = ?", tc.mcID).Scan(ctx)
			}
			assert.ErrorIs(t, err, sql.ErrNoRows)

			if tc.verifyChildSpanner {
				span := otrace.SpanFromContext(ctx)
				assert.True(t, span.SpanContext().IsValid())
				_, ok := ctx.Value(stracer.TracerKey).(otrace.Tracer)
				assert.True(t, ok)
			}
		})
	}
}

func TestMachineCapability_GetStrInfo(t *testing.T) {
	type fields struct {
		Info map[string]interface{}
	}
	type args struct {
		name string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *string
	}{
		{
			name: "test existing key returns string value",
			fields: fields{
				Info: map[string]interface{}{
					"foo": "bar",
				},
			},
			args: args{
				name: "foo",
			},
			want: db.GetStrPtr("bar"),
		},
		{
			name: "test non-existent key returns nil",
			fields: fields{
				Info: map[string]interface{}{
					"foo": "bar",
				},
			},
			args: args{
				name: "baz",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MachineCapability{
				Info: tt.fields.Info,
			}

			val := mc.GetStrInfo(tt.args.name)
			if tt.want != nil {
				require.NotNil(t, val)
				assert.Equal(t, *tt.want, *val)
			} else {
				assert.Nil(t, val)
			}
		})
	}
}

func TestMachineCapability_GetIntInfo(t *testing.T) {
	type fields struct {
		Info map[string]interface{}
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *int
	}{
		{
			name: "test existing key from info read from DB returns int value",
			fields: fields{
				Info: map[string]interface{}{
					"foo": json.Number('5'),
				},
			},
			args: args{
				name: "foo",
			},
			want: db.GetIntPtr(5),
		},
		{
			name: "test existing key from info populated in struct returns int value",
			fields: fields{
				Info: map[string]interface{}{
					"foo": 5,
				},
			},
			args: args{
				name: "foo",
			},
			want: db.GetIntPtr(5),
		},
		{
			name: "test non-existent key returns nil",
			fields: fields{
				Info: map[string]interface{}{
					"foo": 5,
				},
			},
			args: args{
				name: "baz",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MachineCapability{
				Info: tt.fields.Info,
			}

			val := mc.GetIntInfo(tt.args.name)
			if tt.want != nil {
				require.NotNil(t, val)
				assert.Equal(t, *tt.want, *val)
			} else {
				assert.Nil(t, val)
			}
		})
	}
}
