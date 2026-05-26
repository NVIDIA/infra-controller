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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	"github.com/NVIDIA/infra-controller-rest/db/pkg/db/paginator"
	"github.com/google/uuid"

	"github.com/uptrace/bun"

	stracer "github.com/NVIDIA/infra-controller-rest/db/pkg/tracer"
)

// MachineCapabilityType
const (
	MachineCapabilityTypeCPU           = "CPU"
	MachineCapabilityTypeMemory        = "Memory"
	MachineCapabilityTypeGPU           = "GPU"
	MachineCapabilityTypeStorage       = "Storage"
	MachineCapabilityTypeNetwork       = "Network"
	MachineCapabilityTypeInfiniBand    = "InfiniBand"
	MachineCapabilityTypeDPU           = "DPU"
	MachineCapabilityDeviceTypeDPU     = "DPU"
	MachineCapabilityDeviceTypeNVLink  = "NVLink"
	MachineCapabilityDeviceTypeUnknown = "Unknown"

	// MachineCapabilityRelationName is the relation name for the MachineCapability model
	MachineCapabilityRelationName = "MachineCapability"

	// MachineCapabilityOrderByDefault default field to be used for ordering when none specified
	MachineCapabilityOrderByDefault = "created"
)

var (

	// MachineCapabilityTypeChoiceMap is a map of valid MachineCapability types
	MachineCapabilityTypeChoiceMap = map[string]bool{
		MachineCapabilityTypeCPU:        true,
		MachineCapabilityTypeMemory:     true,
		MachineCapabilityTypeGPU:        true,
		MachineCapabilityTypeStorage:    true,
		MachineCapabilityTypeNetwork:    true,
		MachineCapabilityTypeInfiniBand: true,
		MachineCapabilityTypeDPU:        true,
	}

	// MachineCapabilityOrderByFields is a list of valid order by fields for the MachineCapability model
	MachineCapabilityOrderByFields = []string{"type", "created", "updated"}

	// MachineCapabilityDeviceTypeChoiceMap is a map of valid MachineCapability device types
	MachineCapabilityDeviceTypeChoiceMap = map[string]bool{
		MachineCapabilityDeviceTypeDPU:    true,
		MachineCapabilityDeviceTypeNVLink: true,
	}
)

// MachineCapabilityCreateInput input parameters for Create method
type MachineCapabilityCreateInput struct {
	MachineID        *string
	InstanceTypeID   *uuid.UUID
	Type             string
	Name             string
	Frequency        *string
	Capacity         *string
	HardwareRevision *string
	Cores            *int
	Threads          *int
	Vendor           *string
	Count            *int
	DeviceType       *string
	InactiveDevices  []int
	Index            int
	Info             map[string]interface{}
}

// MachineCapabilityUpdateInput input parameters for Update method
type MachineCapabilityUpdateInput struct {
	ID               uuid.UUID
	MachineID        *string
	InstanceTypeID   *uuid.UUID
	Type             *string
	Name             *string
	Frequency        *string
	Capacity         *string
	HardwareRevision *string
	Cores            *int
	Threads          *int
	Vendor           *string
	Count            *int
	DeviceType       *string
	InactiveDevices  []int
	Index            *int
	Info             map[string]interface{}
}

// MachineCapability represents entries in the machine_capability table
// It describes capabilities of a Machine
type MachineCapability struct {
	bun.BaseModel `bun:"table:machine_capability,alias:mc"`

	ID               uuid.UUID              `bun:"type:uuid,pk"`
	MachineID        *string                `bun:"machine_id"`
	InstanceTypeID   *uuid.UUID             `bun:"instance_type_id,type:uuid"`
	InstanceType     *InstanceType          `bun:"rel:belongs-to,join:instance_type_id=id"`
	Type             string                 `bun:"type,notnull"`
	Name             string                 `bun:"name,notnull"`
	Frequency        *string                `bun:"frequency"`
	Capacity         *string                `bun:"capacity"`
	HardwareRevision *string                `bun:"hardware_revision"`
	Cores            *int                   `bun:"cores"`
	Threads          *int                   `bun:"threads"`
	Vendor           *string                `bun:"vendor"`
	Count            *int                   `bun:"count"`
	DeviceType       *string                `bun:"device_type"`
	InactiveDevices  []int                  `bun:"inactive_devices"`
	Index            int                    `bun:"index"`
	Info             map[string]interface{} `bun:"info,json_use_number"` // Any other attribute of the capability
	Created          time.Time              `bun:"created,nullzero,notnull,default:current_timestamp"`
	Updated          time.Time              `bun:"updated,nullzero,notnull,default:current_timestamp"`
	Deleted          *time.Time             `bun:"deleted,soft_delete"`

	// Deprecated fields: To be deleted
	ValueStr    *string `bun:"value_str"`
	ValueInt    *int    `bun:"value_int"`
	Description *string `bun:"description"`
}

// GetStrInfo returns the string value of the given key in the Info map
func (mc *MachineCapability) GetStrInfo(name string) *string {
	if mc.Info == nil {
		return nil
	}
	info, ok := mc.Info[name]
	if !ok {
		return nil
	}
	strInfo, ok := info.(string)
	if !ok {
		return nil
	}

	return &strInfo
}

// GetIntInfo returns the integer value of the given key in the Info map
func (mc *MachineCapability) GetIntInfo(name string) *int {
	if mc.Info == nil {
		return nil
	}
	info, ok := mc.Info[name]
	if !ok {
		return nil
	}

	var intInfo int

	// When info is read from DB, the value is of type json.Number
	jnInfo, ok := info.(json.Number)
	if ok {
		int64Info, err := jnInfo.Int64()
		if err != nil {
			return nil
		}

		intInfo = int(int64Info)
	} else {
		// When info is read from a freshly map, the value should be of type int
		intInfo, ok = info.(int)
		if !ok {
			return nil
		}
	}

	return &intInfo
}

// TODO: Add follow up migration to remove description, value_str and value_int

// GetIndentedJSON returns formatted json of MachineCapability
func (mc *MachineCapability) GetIndentedJSON() ([]byte, error) {
	return json.MarshalIndent(mc, "", "  ")
}

var _ bun.BeforeAppendModelHook = (*MachineCapability)(nil)

// BeforeAppendModel is a hook that is called before the model is appended to the query
func (mc *MachineCapability) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		mc.Created = db.GetCurTime()
		mc.Updated = db.GetCurTime()
	case *bun.UpdateQuery:
		mc.Updated = db.GetCurTime()
	}
	return nil
}

var _ bun.BeforeCreateTableHook = (*MachineCapability)(nil)

// BeforeCreateTable is a hook that is called before the table is created
// This is only used in tests
func (mc *MachineCapability) BeforeCreateTable(ctx context.Context,
	query *bun.CreateTableQuery) error {
	query.ForeignKey(`("machine_id") REFERENCES "machine" ("id")`).
		ForeignKey(`("instance_type_id") REFERENCES "instance_type" ("id")`)
	return nil
}

// MachineCapabilityDAO is an interface for interacting with the MachineCapability model
type MachineCapabilityDAO interface {
	//
	Create(ctx context.Context, tx *db.Tx, input MachineCapabilityCreateInput) (*MachineCapability, error)
	//
	GetByID(ctx context.Context, tx *db.Tx, id uuid.UUID, includeRelations []string) (*MachineCapability, error)
	//
	GetAll(ctx context.Context, tx *db.Tx, machineIDs []string, instanceTypeIDs []uuid.UUID, capabilityType *string,
		name *string, frequency *string, capacity *string, vendor *string,
		count *int, deviceType *string, inactiveDevices []int, includeRelations []string, offset *int, limit *int, orderBy *paginator.OrderBy) ([]MachineCapability, int, error)
	//
	GetAllDistinct(ctx context.Context, tx *db.Tx, machineIDs []string, instanceTypeID *uuid.UUID, capabilityType *string,
		name *string, frequency *string, capacity *string, vendor *string,
		count *int, deviceType *string, inactiveDevices []int, offset *int, limit *int, orderBy *paginator.OrderBy) ([]MachineCapability, int, error)
	//
	Update(ctx context.Context, tx *db.Tx, input MachineCapabilityUpdateInput) (*MachineCapability, error)
	//
	ClearFromParams(ctx context.Context, tx *db.Tx, id uuid.UUID,
		machineID, instanceTypeID, frequency, capacity, vendor, info bool) (*MachineCapability, error)
	//
	DeleteByID(ctx context.Context, tx *db.Tx, id uuid.UUID, purge bool) error
}

// MachineCapabilitySQLDAO is an implementation of the MachineCapabilityDAO interface
type MachineCapabilitySQLDAO struct {
	dbSession *db.Session
	MachineCapabilityDAO
	tracerSpan *stracer.TracerSpan
}

// CreateFromParams creates a new MachineCapability from the given parameters
// The returned MachineCapability will not have any related structs filled in
// since there are 2 operations (INSERT, SELECT), in this, it is required that
// this library call happens within a transaction
func (mcd MachineCapabilitySQLDAO) Create(
	ctx context.Context, tx *db.Tx,
	input MachineCapabilityCreateInput) (*MachineCapability, error) {
	// Create a child span and set the attributes for current request
	ctx, MachineCapabilityDAOSpan := mcd.tracerSpan.CreateChildInCurrentContext(ctx, "MachineCapabilityDAO.CreateFromParams")
	if MachineCapabilityDAOSpan != nil {
		defer MachineCapabilityDAOSpan.End()

		mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "name", input.Name)
	}

	if len(strings.TrimSpace(input.Type)) == 0 {
		return nil, errors.New("capabilityType is empty")
	}

	if input.MachineID == nil && input.InstanceTypeID == nil {
		return nil, errors.New("instanceTypeID or machineID needs to be specified")
	}

	m := &MachineCapability{
		ID:               uuid.New(),
		MachineID:        input.MachineID,
		InstanceTypeID:   input.InstanceTypeID,
		Type:             input.Type,
		Name:             input.Name,
		Frequency:        input.Frequency,
		Capacity:         input.Capacity,
		Vendor:           input.Vendor,
		Count:            input.Count,
		DeviceType:       input.DeviceType,
		Threads:          input.Threads,
		Cores:            input.Cores,
		HardwareRevision: input.HardwareRevision,
		Info:             input.Info,
		InactiveDevices:  input.InactiveDevices,
		Index:            input.Index,
	}

	_, err := db.GetIDB(tx, mcd.dbSession).NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return nil, err
	}

	nv, err := mcd.GetByID(ctx, tx, m.ID, []string{"InstanceType"})
	if err != nil {
		return nil, err
	}

	return nv, nil
}

// GetByID returns a MachineCapability by ID
// returns db.ErrDoesNotExist error if the record is not found
func (mcd MachineCapabilitySQLDAO) GetByID(ctx context.Context, tx *db.Tx, id uuid.UUID, includeRelations []string) (*MachineCapability, error) {
	// Create a child span and set the attributes for current request
	ctx, MachineCapabilityDAOSpan := mcd.tracerSpan.CreateChildInCurrentContext(ctx, "MachineCapabilityDAO.GetByID")
	if MachineCapabilityDAOSpan != nil {
		defer MachineCapabilityDAOSpan.End()

		mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "id", id.String())
	}

	m := &MachineCapability{}

	query := db.GetIDB(tx, mcd.dbSession).NewSelect().Model(m).Where("mc.id = ?", id)

	for _, relation := range includeRelations {
		query = query.Relation(relation)
	}

	err := query.Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, db.ErrDoesNotExist
		}
		return nil, err
	}

	return m, nil
}

// GetAll returns all MachineCapabilities filtered by the given parameters
// if orderBy is nil, then records are ordered by column specified in MachineCapabilityOrderByDefault in ascending order
func (mcd MachineCapabilitySQLDAO) GetAll(
	ctx context.Context, tx *db.Tx,
	machineIDs []string,
	instanceTypeIDs []uuid.UUID,
	capabilityType *string,
	name *string,
	frequency *string,
	capacity *string,
	vendor *string,
	count *int,
	deviceType *string,
	inactiveDevices []int,
	includeRelations []string,
	offset *int, limit *int, orderBy *paginator.OrderBy) ([]MachineCapability, int, error) {
	// Create a child span and set the attributes for current request
	ctx, MachineCapabilityDAOSpan := mcd.tracerSpan.CreateChildInCurrentContext(ctx, "MachineCapabilityDAO.GetAll")
	if MachineCapabilityDAOSpan != nil {
		defer MachineCapabilityDAOSpan.End()
	}

	mcs := []MachineCapability{}

	query := db.GetIDB(tx, mcd.dbSession).NewSelect().Model(&mcs)
	if machineIDs != nil {
		if len(machineIDs) == 1 {
			query = query.Where("mc.machine_id = ?", machineIDs[0])
		} else {
			query = query.Where("mc.machine_id IN (?)", bun.In(machineIDs))
		}

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "machine_ids", machineIDs)
		}
	}
	if instanceTypeIDs != nil {
		if len(machineIDs) == 1 {
			query = query.Where("mc.instance_type_id = ?", instanceTypeIDs[0])
		} else {
			query = query.Where("mc.instance_type_id IN (?)", bun.In(instanceTypeIDs))
		}

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "instance_type_ids", instanceTypeIDs)
		}
	}
	if capabilityType != nil {
		query = query.Where("mc.type = ?", *capabilityType)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "type", *capabilityType)
		}
	}
	if name != nil {
		query = query.Where("mc.name = ?", *name)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "name", *name)
		}
	}
	if frequency != nil {
		query = query.Where("mc.frequency = ?", *frequency)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "frequency", *frequency)
		}
	}
	if capacity != nil {
		query = query.Where("mc.capacity = ?", *capacity)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "capacity", *capacity)
		}
	}
	if vendor != nil {
		query = query.Where("mc.vendor = ?", *vendor)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "vendor", *vendor)
		}
	}
	if count != nil {
		query = query.Where("mc.count = ?", *count)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "count", *count)
		}
	}
	if deviceType != nil {
		query = query.Where("mc.device_type = ?", *deviceType)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "device_type", *deviceType)
		}
	}
	if inactiveDevices != nil {
		query = query.Where("mc.inactive_devices = ?", inactiveDevices)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "inactive_devices", inactiveDevices)
		}
	}

	for _, relation := range includeRelations {
		query = query.Relation(relation)
	}

	// if no order is passed, set default to make sure objects return always in the same order and pagination works properly
	if orderBy == nil {
		orderBy = paginator.NewDefaultOrderBy(MachineCapabilityOrderByDefault)
	}

	paginator, err := paginator.NewPaginator(ctx, query, offset, limit, orderBy, MachineCapabilityOrderByFields)
	if err != nil {
		return nil, 0, err
	}

	err = paginator.Query.Limit(paginator.Limit).Offset(paginator.Offset).Scan(ctx)
	if err != nil {
		return nil, 0, err
	}

	return mcs, paginator.Total, nil
}

// GetAllDistinct returns all MachineCapabilities that have distinct type, name, frequency, capacity, vendor, count, and device_type filtered by the given parameters
func (mcd MachineCapabilitySQLDAO) GetAllDistinct(
	ctx context.Context, tx *db.Tx,
	machineIDs []string,
	instanceTypeID *uuid.UUID,
	capabilityType *string,
	name *string,
	frequency *string,
	capacity *string,
	vendor *string,
	count *int,
	deviceType *string,
	inactiveDevices []int,
	offset *int, limit *int, orderBy *paginator.OrderBy) ([]MachineCapability, int, error) {
	// Create a child span and set the attributes for current request
	ctx, MachineCapabilityDAOSpan := mcd.tracerSpan.CreateChildInCurrentContext(ctx, "MachineCapabilityDAO.GetAllDistinct")
	if MachineCapabilityDAOSpan != nil {
		defer MachineCapabilityDAOSpan.End()
	}

	mcs := []MachineCapability{}

	query := db.GetIDB(tx, mcd.dbSession).NewSelect().Model(&mcs).ColumnExpr("DISTINCT ON (mc.type, mc.name, mc.frequency, mc.capacity, mc.vendor, mc.count, mc.device_type, mc.inactive_devices) mc.*")
	if machineIDs != nil {
		if len(machineIDs) == 1 {
			query = query.Where("mc.machine_id = ?", machineIDs[0])
		} else {
			query = query.Where("mc.machine_id IN (?)", bun.In(machineIDs))
		}

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "machine_ids", machineIDs)
		}
	}
	if instanceTypeID != nil {
		query = query.Where("mc.instance_type_id = ?", *instanceTypeID)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "instance_type_id", instanceTypeID.String())
		}
	}
	if capabilityType != nil {
		query = query.Where("mc.type = ?", *capabilityType)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "type", *capabilityType)
		}
	}
	if name != nil {
		query = query.Where("mc.name = ?", *name)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "name", *name)
		}
	}
	if frequency != nil {
		query = query.Where("mc.frequency = ?", *frequency)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "frequency", *frequency)
		}
	}
	if capacity != nil {
		query = query.Where("mc.capacity = ?", *capacity)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "capacity", *capacity)
		}
	}
	if vendor != nil {
		query = query.Where("mc.vendor = ?", *vendor)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "vendor", *vendor)
		}
	}
	if count != nil {
		query = query.Where("mc.count = ?", *count)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "count", *count)
		}
	}
	if deviceType != nil {
		query = query.Where("mc.device_type = ?", *deviceType)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "device_type", *deviceType)
		}
	}
	if inactiveDevices != nil {
		query = query.Where("mc.inactive_devices = ?", inactiveDevices)

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "inactive_devices", inactiveDevices)
		}
	}

	paginator, err := paginator.NewPaginator(ctx, query, offset, limit, orderBy, MachineCapabilityOrderByFields)
	if err != nil {
		return nil, 0, err
	}

	err = paginator.Query.Limit(paginator.Limit).Offset(paginator.Offset).Scan(ctx)
	if err != nil {
		return nil, 0, err
	}

	return mcs, paginator.Total, nil
}

// Update updates specified fields of an existing MachineCapability
// The updated fields are assumed to be set to non-null values
// since there are 2 operations (UPDATE, SELECT), in this, it is required that
// this library call happens within a transaction
func (mcd MachineCapabilitySQLDAO) Update(
	ctx context.Context, tx *db.Tx,
	input MachineCapabilityUpdateInput) (*MachineCapability, error) {
	// Create a child span and set the attributes for current request
	ctx, MachineCapabilityDAOSpan := mcd.tracerSpan.CreateChildInCurrentContext(ctx, "MachineCapabilityDAO.UpdateFromParams")
	if MachineCapabilityDAOSpan != nil {
		defer MachineCapabilityDAOSpan.End()
	}

	m := &MachineCapability{
		ID: input.ID,
	}

	updatedFields := []string{}

	if input.MachineID != nil {
		m.MachineID = input.MachineID
		updatedFields = append(updatedFields, "machine_id")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "machine_id", input.MachineID)
		}
	}
	if input.InstanceTypeID != nil {
		m.InstanceTypeID = input.InstanceTypeID
		updatedFields = append(updatedFields, "instance_type_id")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "instance_type_id", input.InstanceTypeID.String())
		}
	}
	if input.Type != nil {
		if len(strings.TrimSpace(*input.Type)) == 0 {
			return nil, errors.New("capabilityType is empty")
		}
		m.Type = *input.Type
		updatedFields = append(updatedFields, "type")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "type", *input.Type)
		}
	}
	if input.Name != nil {
		m.Name = *input.Name
		updatedFields = append(updatedFields, "name")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "name", *input.Name)
		}
	}
	if input.Frequency != nil {
		m.Frequency = input.Frequency
		updatedFields = append(updatedFields, "frequency")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "frequency", *input.Frequency)
		}
	}
	if input.Capacity != nil {
		m.Capacity = input.Capacity
		updatedFields = append(updatedFields, "capacity")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "capacity", *input.Capacity)
		}
	}
	if input.Vendor != nil {
		m.Vendor = input.Vendor
		updatedFields = append(updatedFields, "vendor")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "vendor", *input.Vendor)
		}
	}
	if input.Count != nil {
		m.Count = input.Count
		updatedFields = append(updatedFields, "count")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "count", *input.Count)
		}
	}
	if input.DeviceType != nil {
		m.DeviceType = input.DeviceType
		updatedFields = append(updatedFields, "device_type")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "device_type", *input.DeviceType)
		}
	}
	if input.Threads != nil {
		m.Threads = input.Threads
		updatedFields = append(updatedFields, "threads")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "threads", *input.Threads)
		}
	}

	if input.Cores != nil {
		m.Cores = input.Cores
		updatedFields = append(updatedFields, "cores")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "cores", *input.Cores)
		}
	}

	if input.HardwareRevision != nil {
		m.HardwareRevision = input.HardwareRevision
		updatedFields = append(updatedFields, "hardware_revision")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "hardware_revision", *input.HardwareRevision)
		}
	}

	if input.Index != nil {
		m.Index = *input.Index
		updatedFields = append(updatedFields, "index")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "index", *input.Index)
		}
	}

	if input.InactiveDevices != nil {
		m.InactiveDevices = input.InactiveDevices
		updatedFields = append(updatedFields, "inactive_devices")

		if MachineCapabilityDAOSpan != nil {
			mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "inactive_devices", fmt.Sprintf("%v", input.InactiveDevices))
		}
	}

	if input.Info != nil {
		m.Info = input.Info
		updatedFields = append(updatedFields, "info")
	}

	if len(updatedFields) > 0 {
		updatedFields = append(updatedFields, "updated")

		_, err := db.GetIDB(tx, mcd.dbSession).NewUpdate().Model(m).Column(updatedFields...).Where("id = ?", input.ID).Exec(ctx)
		if err != nil {
			return nil, err
		}
	}

	nv, err := mcd.GetByID(ctx, tx, m.ID, nil)
	if err != nil {
		return nil, err
	}

	return nv, nil
}

// ClearFromParams sets parameters of an existing Machine Capability to null values in db
// since there are 2 operations (UPDATE, SELECT), it is required that
// this must be within a transaction
func (mcd MachineCapabilitySQLDAO) ClearFromParams(
	ctx context.Context, tx *db.Tx,
	id uuid.UUID,
	machineID, instanceTypeID, frequency, capacity, vendor, info bool) (*MachineCapability, error) {
	// Create a child span and set the attributes for current request
	ctx, MachineCapabilityDAOSpan := mcd.tracerSpan.CreateChildInCurrentContext(ctx, "MachineCapabilityDAO.ClearFromParams")
	if MachineCapabilityDAOSpan != nil {
		defer MachineCapabilityDAOSpan.End()

		mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "id", id.String())
	}

	m := &MachineCapability{
		ID: id,
	}

	if machineID && instanceTypeID {
		return nil, fmt.Errorf("machineID and instanceTypeID cannot be cleared together: %w", db.ErrInvalidParams)
	}

	updatedFields := []string{}
	if machineID {
		m.MachineID = nil
		updatedFields = append(updatedFields, "machine_id")
	}
	if instanceTypeID {
		m.InstanceTypeID = nil
		updatedFields = append(updatedFields, "instance_type_id")
	}
	if frequency {
		m.Frequency = nil
		updatedFields = append(updatedFields, "frequency")
	}
	if capacity {
		m.Capacity = nil
		updatedFields = append(updatedFields, "capacity")
	}
	if vendor {
		m.Vendor = nil
		updatedFields = append(updatedFields, "vendor")
	}
	if info {
		m.Info = nil
		updatedFields = append(updatedFields, "info")
	}

	if len(updatedFields) > 0 {
		updatedFields = append(updatedFields, "updated")

		_, err := db.GetIDB(tx, mcd.dbSession).NewUpdate().Model(m).Column(updatedFields...).Where("id = ?", id).Exec(ctx)
		if err != nil {
			return nil, err
		}
	}

	nv, err := mcd.GetByID(ctx, tx, id, nil)
	if err != nil {
		return nil, err
	}
	return nv, nil
}

// DeleteByID deletes an MachineCapability by ID
// error is returned only if there is a db error
// if the object being deleted doesnt exist, error is not returned (idempotent delete)
func (mcd MachineCapabilitySQLDAO) DeleteByID(ctx context.Context, tx *db.Tx, id uuid.UUID, purge bool) error {
	// Create a child span and set the attributes for current request
	ctx, MachineCapabilityDAOSpan := mcd.tracerSpan.CreateChildInCurrentContext(ctx, "MachineCapabilityDAO.DeleteByID")
	if MachineCapabilityDAOSpan != nil {
		defer MachineCapabilityDAOSpan.End()

		mcd.tracerSpan.SetAttribute(MachineCapabilityDAOSpan, "id", id.String())
	}

	mc := &MachineCapability{
		ID: id,
	}

	var err error

	if purge {
		_, err = db.GetIDB(tx, mcd.dbSession).NewDelete().Model(mc).Where("id = ?", id).ForceDelete().Exec(ctx)
	} else {
		_, err = db.GetIDB(tx, mcd.dbSession).NewDelete().Model(mc).Where("id = ?", id).Exec(ctx)
	}
	if err != nil {
		return err
	}

	return nil
}

// NewMachineCapabilityDAO returns a new MachineCapabilityDAO
func NewMachineCapabilityDAO(dbSession *db.Session) MachineCapabilityDAO {
	return &MachineCapabilitySQLDAO{
		dbSession:  dbSession,
		tracerSpan: stracer.NewTracerSpan(),
	}
}
