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
	"fmt"
	"time"

	"github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	"github.com/NVIDIA/infra-controller-rest/db/pkg/db/paginator"
	stracer "github.com/NVIDIA/infra-controller-rest/db/pkg/tracer"
	"github.com/google/uuid"

	"github.com/uptrace/bun"
)

const (
	// InstanceStatusPending indicates that the Instance provisioning hasn't started yet
	InstanceStatusPending = "Pending"
	// InstanceStatusProvisioning indicates that the Instance provisioning is in progress
	InstanceStatusProvisioning = "Provisioning"
	// InstanceStatusConfiguring indicates that the Instance is being configured
	InstanceStatusConfiguring = "Configuring"
	// InstanceStatusReady indicates that the Instance provisioning is complete
	InstanceStatusReady = "Ready"
	// InstanceStatusUpdating indicates that the Instance is receiving system updates
	InstanceStatusUpdating = "Updating"
	// InstanceStatusError indicates that the Instance provisioning has failed
	InstanceStatusError = "Error"
	// InstanceStatusTerminating indicates that the Instance is being terminated
	InstanceStatusTerminating = "Terminating"
	// InstanceStatusTerminated indicates that the Instance has been terminated
	InstanceStatusTerminated = "Terminated"
	// InstanceStatusUnknown indicates that the Instance status is unknown
	InstanceStatusUnknown = "Unknown"

	// InstancePowerStatusBootCompleted status is bootcompleted
	InstancePowerStatusBootCompleted = "BootCompleted"
	// InstancePowerStatusRebooting status is rebooting
	InstancePowerStatusRebooting = "Rebooting"
	// InstancePowerStatusError status is error
	InstancePowerStatusError = "Error"

	// InstanceRelationName is the relation name for the Instance model
	InstanceRelationName = "Instance"

	// names of order by fields
	instanceOrderByName                        = "name"
	instanceOrderByStatus                      = "status"
	instanceOrderByCreated                     = "created"
	instanceOrderByUpdated                     = "updated"
	instanceOrderByMachineID                   = "machine_id"
	instanceOrderByTenantOrgDisplayNameExt     = "tenant_org_display_name"
	instanceOrderByTenantOrgDisplayNameInt     = "tenant.org_display_name"
	instanceOrderByInstanceTypeNameExt         = "instance_type_name"
	instanceOrderByInstanceTypeNameInt         = "instance_type.name"
	instanceOrderByNetworkSecurityGroupNameExt = "network_security_group.name"
	instanceOrderByNetworkSecurityGroupNameInt = "network_security_group.name"
	instanceOrderByHasInfiniBandExt            = "has_infiniband"
	instanceOrderByHasInfiniBandInt            = "mc_type"
	// InstanceOrderByDefault default field to be used for ordering when none specified
	InstanceOrderByDefault = instanceOrderByCreated
)

var (
	// InstanceOrderByFields is a list of valid order by fields for the Instance model
	InstanceOrderByFields = []string{
		instanceOrderByName,
		instanceOrderByStatus,
		instanceOrderByCreated,
		instanceOrderByUpdated,
		instanceOrderByMachineID,
		instanceOrderByTenantOrgDisplayNameExt,
		instanceOrderByInstanceTypeNameExt,
		instanceOrderByHasInfiniBandExt,
		instanceOrderByNetworkSecurityGroupNameExt,
		instanceOrderByNetworkSecurityGroupNameInt,
	}
	// internal list of fields that can be used for ordering
	instanceOrderByFieldsInt = []string{
		instanceOrderByName,
		instanceOrderByStatus,
		instanceOrderByCreated,
		instanceOrderByUpdated,
		instanceOrderByMachineID,
		instanceOrderByTenantOrgDisplayNameInt,
		instanceOrderByInstanceTypeNameInt,
		instanceOrderByHasInfiniBandInt,
		instanceOrderByNetworkSecurityGroupNameInt,
	}
	// mapping of sort fields and required relation (for those that need it)
	instanceOrderByFieldToRelation = map[string]string{
		instanceOrderByTenantOrgDisplayNameExt:     TenantRelationName,
		instanceOrderByInstanceTypeNameExt:         InstanceTypeRelationName,
		instanceOrderByNetworkSecurityGroupNameExt: NetworkSecurityGroupRelationName,
	}
	// mapping of external sort by field to internal
	instanceOrderByFieldExtToInt = map[string]string{
		instanceOrderByTenantOrgDisplayNameExt: instanceOrderByTenantOrgDisplayNameInt,
		instanceOrderByInstanceTypeNameExt:     instanceOrderByInstanceTypeNameInt,
		instanceOrderByHasInfiniBandExt:        instanceOrderByHasInfiniBandInt,
	}
	// InstanceRelatedEntities is a list of valid relation by fields for the Instance model
	InstanceRelatedEntities = map[string]bool{
		InfrastructureProviderRelationName: true,
		SiteRelationName:                   true,
		InstanceTypeRelationName:           true,
		NetworkSecurityGroupRelationName:   true,
		TenantRelationName:                 true,
		VpcRelationName:                    true,
		MachineRelationName:                true,
		OperatingSystemRelationName:        true,
	}
	// InstanceStatusMap is a list of valid status for the Instance model
	InstanceStatusMap = map[string]bool{
		InstanceStatusPending:            true,
		InstanceStatusReady:              true,
		InstanceStatusUpdating:           true,
		InstanceStatusError:              true,
		InstanceStatusConfiguring:        true,
		InstanceStatusProvisioning:       true,
		InstanceStatusTerminating:        true,
		InstanceStatusTerminated:         true,
		InstancePowerStatusBootCompleted: true,
		InstancePowerStatusRebooting:     true,
	}
)

// Instance is a bare-metal machine that has been provisioned for a tenant
type Instance struct {
	bun.BaseModel `bun:"table:instance,alias:i"`

	ID                                     uuid.UUID                               `bun:"type:uuid,pk"`
	Name                                   string                                  `bun:"name,notnull"`
	Description                            *string                                 `bun:"description"`
	TenantID                               uuid.UUID                               `bun:"tenant_id,type:uuid,notnull"`
	Tenant                                 *Tenant                                 `bun:"rel:belongs-to,join:tenant_id=id"`
	InfrastructureProviderID               uuid.UUID                               `bun:"infrastructure_provider_id,type:uuid,notnull"`
	InfrastructureProvider                 *InfrastructureProvider                 `bun:"rel:belongs-to,join:infrastructure_provider_id=id"`
	SiteID                                 uuid.UUID                               `bun:"site_id,type:uuid,notnull"`
	Site                                   *Site                                   `bun:"rel:belongs-to,join:site_id=id"`
	NetworkSecurityGroupID                 *string                                 `bun:"network_security_group_id"`
	NetworkSecurityGroup                   *NetworkSecurityGroup                   `bun:"rel:belongs-to,join:network_security_group_id=id"`
	NetworkSecurityGroupPropagationDetails *NetworkSecurityGroupPropagationDetails `bun:"network_security_group_propagation_details,type:jsonb"`
	InstanceTypeID                         *uuid.UUID                              `bun:"instance_type_id,type:uuid"`
	InstanceType                           *InstanceType                           `bun:"rel:belongs-to,join:instance_type_id=id"`
	VpcID                                  uuid.UUID                               `bun:"vpc_id,type:uuid,notnull"`
	Vpc                                    *Vpc                                    `bun:"rel:belongs-to,join:vpc_id=id"`
	MachineID                              *string                                 `bun:"machine_id"`
	Machine                                *Machine                                `bun:"rel:belongs-to,join:machine_id=id"`
	ControllerInstanceID                   *uuid.UUID                              `bun:"controller_instance_id,type:uuid"`
	Hostname                               *string                                 `bun:"hostname"`
	OperatingSystemID                      *uuid.UUID                              `bun:"operating_system_id,type:uuid"`
	OperatingSystem                        *OperatingSystem                        `bun:"rel:belongs-to,join:operating_system_id=id"`
	IpxeScript                             *string                                 `bun:"ipxe_script"`
	AlwaysBootWithCustomIpxe               bool                                    `bun:"always_boot_with_custom_ipxe,notnull"`
	PhoneHomeEnabled                       bool                                    `bun:"phone_home_enabled,notnull"`
	UserData                               *string                                 `bun:"user_data"`
	Labels                                 map[string]string                       `bun:"labels,type:jsonb"`
	IsUpdatePending                        bool                                    `bun:"is_update_pending,notnull"`
	InfinityRCRStatus                      *string                                 `bun:"infinity_rcr_status"`
	TpmEkCertificate                       *string                                 `bun:"tpm_ek_certificate"`
	Status                                 string                                  `bun:"status,notnull"`
	PowerStatus                            *string                                 `bun:"power_status"`
	IsMissingOnSite                        bool                                    `bun:"is_missing_on_site,notnull"`
	Created                                time.Time                               `bun:"created,nullzero,notnull,default:current_timestamp"`
	Updated                                time.Time                               `bun:"updated,nullzero,notnull,default:current_timestamp"`
	Deleted                                *time.Time                              `bun:"deleted,soft_delete"`
	CreatedBy                              uuid.UUID                               `bun:"created_by,type:uuid,notnull"`
	// Not for display, used by the query that sorts on machine capability type, specifically InfiniBand type
	MCType string `bun:"mc_type,scanonly"`
}

// InstanceCreateInput input parameters for Create method
type InstanceCreateInput struct {
	Name                                   string
	Description                            *string
	TenantID                               uuid.UUID
	InfrastructureProviderID               uuid.UUID
	SiteID                                 uuid.UUID
	InstanceTypeID                         *uuid.UUID
	NetworkSecurityGroupID                 *string
	NetworkSecurityGroupPropagationDetails *NetworkSecurityGroupPropagationDetails
	VpcID                                  uuid.UUID
	MachineID                              *string
	ControllerInstanceID                   *uuid.UUID
	Hostname                               *string
	OperatingSystemID                      *uuid.UUID
	IpxeScript                             *string
	AlwaysBootWithCustomIpxe               bool
	PhoneHomeEnabled                       bool
	UserData                               *string
	Labels                                 map[string]string
	IsUpdatePending                        bool
	InfinityRCRStatus                      *string
	TpmEkCertificate                       *string
	Status                                 string
	PowerStatus                            *string
	CreatedBy                              uuid.UUID
}

// InstanceUpdateInput input parameters for Update method
type InstanceUpdateInput struct {
	InstanceID                             uuid.UUID
	Name                                   *string
	Description                            *string
	TenantID                               *uuid.UUID
	InfrastructureProviderID               *uuid.UUID
	SiteID                                 *uuid.UUID
	InstanceTypeID                         *uuid.UUID
	NetworkSecurityGroupID                 *string
	NetworkSecurityGroupPropagationDetails *NetworkSecurityGroupPropagationDetails
	VpcID                                  *uuid.UUID
	MachineID                              *string
	ControllerInstanceID                   *uuid.UUID
	Hostname                               *string
	OperatingSystemID                      *uuid.UUID
	IpxeScript                             *string
	AlwaysBootWithCustomIpxe               *bool
	PhoneHomeEnabled                       *bool
	UserData                               *string
	Labels                                 map[string]string
	IsUpdatePending                        *bool
	InfinityRCRStatus                      *string
	TpmEkCertificate                       *string
	Status                                 *string
	PowerStatus                            *string
	IsMissingOnSite                        *bool
}

// InstanceClearInput input parameters for Clear method
type InstanceClearInput struct {
	InstanceID                             uuid.UUID
	Description                            bool
	MachineID                              bool
	ControllerInstanceID                   bool
	NetworkSecurityGroupID                 bool
	NetworkSecurityGroupPropagationDetails bool
	Hostname                               bool
	OperatingSystemID                      bool
	IpxeScript                             bool
	UserData                               bool
	Labels                                 bool
	TpmEkCertificate                       bool
}

// InstanceFilterInput input parameters for GetAll method
type InstanceFilterInput struct {
	InstanceIDs               []uuid.UUID
	Names                     []string
	TenantIDs                 []uuid.UUID
	InfrastructureProviderIDs []uuid.UUID
	SiteIDs                   []uuid.UUID
	InstanceTypeIDs           []uuid.UUID
	NetworkSecurityGroupIDs   []string
	VpcIDs                    []uuid.UUID
	MachineIDs                []string
	ControllerInstanceIDs     []uuid.UUID
	OperatingSystemIDs        []uuid.UUID
	Statuses                  []string
	SearchQuery               *string
}

var _ bun.BeforeAppendModelHook = (*Instance)(nil)

// BeforeAppendModel is a hook that is called before the model is appended to the query
func (i *Instance) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		i.Created = db.GetCurTime()
		i.Updated = db.GetCurTime()
	case *bun.UpdateQuery:
		i.Updated = db.GetCurTime()
	}
	return nil
}

var _ bun.BeforeCreateTableHook = (*Instance)(nil)

// BeforeCreateTable is a hook that is called before the table is created
func (i *Instance) BeforeCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	query.ForeignKey(`("tenant_id") REFERENCES "tenant" ("id")`).
		ForeignKey(`("infrastructure_provider_id") REFERENCES "infrastructure_provider" ("id")`).
		ForeignKey(`("site_id") REFERENCES "site" ("id")`).
		ForeignKey(`("instance_type_id") REFERENCES "instance_type" ("id")`).
		ForeignKey(`("vpc_id") REFERENCES "vpc" ("id")`).
		ForeignKey(`("machine_id") REFERENCES "machine" ("id")`).
		ForeignKey(`("operating_system_id") REFERENCES "operating_system" ("id")`).
		ForeignKey(`("network_security_group_id") REFERENCES "network_security_group" ("id")`)
	return nil
}

// InstanceDAO is an interface for interacting with the Instance model
type InstanceDAO interface {
	//
	Create(ctx context.Context, tx *db.Tx, input InstanceCreateInput) (*Instance, error)
	//
	CreateMultiple(ctx context.Context, tx *db.Tx, inputs []InstanceCreateInput) ([]Instance, error)
	//
	GetByID(ctx context.Context, tx *db.Tx, id uuid.UUID, includeRelations []string) (*Instance, error)
	//
	GetCountByStatus(ctx context.Context, tx *db.Tx, tenantID *uuid.UUID, siteID *uuid.UUID) (map[string]int, error)
	//
	GetAll(ctx context.Context, tx *db.Tx, filter InstanceFilterInput, page paginator.PageInput, includeRelations []string) ([]Instance, int, error)
	//
	Update(ctx context.Context, tx *db.Tx, input InstanceUpdateInput) (*Instance, error)
	// UpdateMultiple used to update multiple rows
	UpdateMultiple(ctx context.Context, tx *db.Tx, inputs []InstanceUpdateInput) ([]Instance, error)
	//
	Clear(ctx context.Context, tx *db.Tx, input InstanceClearInput) (*Instance, error)
	//
	Delete(ctx context.Context, tx *db.Tx, id uuid.UUID) error
	// GetCount returns total count of rows for specified filter
	GetCount(ctx context.Context, tx *db.Tx, filter InstanceFilterInput) (count int, err error)
}

// InstanceSQLDAO is an implementation of the InstanceDAO interface
type InstanceSQLDAO struct {
	dbSession *db.Session
	InstanceDAO
	tracerSpan *stracer.TracerSpan
}

// Create creates a new Instance from the given parameters
// The returned Instance will not have any related structs (InfrastructureProvider/Site etc) filled in
// since there are 2 operations (INSERT, SELECT), in this, it is required that
// this library call happens within a transaction
func (isd InstanceSQLDAO) Create(ctx context.Context, tx *db.Tx, input InstanceCreateInput) (*Instance, error) {
	// Create a child span and set the attributes for current request
	ctx, instanceDAOSpan := isd.tracerSpan.CreateChildInCurrentContext(ctx, "InstanceDAO.Create")
	if instanceDAOSpan != nil {
		defer instanceDAOSpan.End()
		isd.tracerSpan.SetAttribute(instanceDAOSpan, "name", input.Name)
	}

	results, err := isd.CreateMultiple(ctx, tx, []InstanceCreateInput{input})
	if err != nil {
		return nil, err
	}
	return &results[0], nil
}

// GetByID returns a Instance by ID
// includeRelation can be a subset of "Tenant", "InfrastructureProvider"
// "Site", "InstanceType", "Vpc", "Machine", "OperatingSystem", "NetworkSecurityGroup"
// Allocation relations are intentionally omitted because direct instance-allocation linkage was removed.
// returns db.ErrDoesNotExist error if the record is not found
func (isd InstanceSQLDAO) GetByID(ctx context.Context, tx *db.Tx, id uuid.UUID, includeRelations []string) (*Instance, error) {
	i := &Instance{}
	// Create a child span and set the attributes for current request
	ctx, instanceDAOSpan := isd.tracerSpan.CreateChildInCurrentContext(ctx, "InstanceDAO.GetByID")
	if instanceDAOSpan != nil {
		defer instanceDAOSpan.End()
		isd.tracerSpan.SetAttribute(instanceDAOSpan, "id", id.String())
	}

	query := db.GetIDB(tx, isd.dbSession).NewSelect().Model(i).Where("i.id = ?", id)

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

	return i, nil
}

// GetCountByStatus returns count of Instances for given status
// Errors are returned only when there is a db related error
// if records not found, then error is nil, but length of returned map is 0
func (isd InstanceSQLDAO) GetCountByStatus(ctx context.Context, tx *db.Tx, tenantID *uuid.UUID, siteID *uuid.UUID) (map[string]int, error) {
	i := &Instance{}
	// Create a child span and set the attributes for current request
	ctx, instanceDAOSpan := isd.tracerSpan.CreateChildInCurrentContext(ctx, "InstanceDAO.GetCountByStatus")
	if instanceDAOSpan != nil {
		defer instanceDAOSpan.End()
	}

	var statusQueryResults []map[string]interface{}
	query := db.GetIDB(tx, isd.dbSession).NewSelect().Model(i)
	if tenantID != nil {
		query = query.Where("i.tenant_id = ?", *tenantID)

		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "tenant_id", tenantID.String())
		}
	}
	if siteID != nil {
		query = query.Where("i.site_id = ?", *siteID)

		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "site_id", siteID.String())
		}
	}

	err := query.Column("i.status").ColumnExpr("COUNT(*) AS total_count").GroupExpr("i.status").Scan(ctx, &statusQueryResults)
	if err != nil {
		return nil, err
	}

	// creare results map by holding key as status value with total count
	results := map[string]int{
		"total":                    0,
		InstanceStatusPending:      0,
		InstanceStatusProvisioning: 0,
		InstanceStatusConfiguring:  0,
		InstanceStatusReady:        0,
		InstanceStatusUpdating:     0,
		InstanceStatusTerminating:  0,
		InstanceStatusError:        0,
	}
	if len(statusQueryResults) > 0 {
		for _, statusMap := range statusQueryResults {
			results[statusMap["status"].(string)] = int(statusMap["total_count"].(int64))

			results["total"] += int(statusMap["total_count"].(int64))
		}
	}
	return results, nil
}

func (isd InstanceSQLDAO) setQueryWithFilter(filter InstanceFilterInput, query *bun.SelectQuery, instanceDAOSpan *stracer.CurrentContextSpan) (*bun.SelectQuery, error) {
	// Single-item IN queries are optimized by the query planner to =
	if filter.InstanceIDs != nil {
		query = query.Where("i.id IN (?)", bun.In(filter.InstanceIDs))
		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "instance_ids", filter.InstanceIDs)
		}
	}

	if filter.Names != nil {
		query = query.Where("i.name IN (?)", bun.In(filter.Names))
		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "names", filter.Names)
		}
	}

	if filter.TenantIDs != nil {
		query = query.Where("i.tenant_id IN (?)", bun.In(filter.TenantIDs))
		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "tenant_ids", filter.TenantIDs)
		}
	}

	if filter.InfrastructureProviderIDs != nil {
		query = query.Where("i.infrastructure_provider_id IN (?)", bun.In(filter.InfrastructureProviderIDs))
		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "infrastructure_provider_ids", filter.InfrastructureProviderIDs)
		}
	}

	if filter.SiteIDs != nil {
		query = query.Where("i.site_id IN (?)", bun.In(filter.SiteIDs))
		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "site_ids", filter.SiteIDs)
		}
	}

	if filter.InstanceTypeIDs != nil {
		query = query.Where("i.instance_type_id IN (?)", bun.In(filter.InstanceTypeIDs))
		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "instance_type_ids", filter.InstanceTypeIDs)
		}
	}

	if filter.NetworkSecurityGroupIDs != nil {
		query = query.Where("i.network_security_group_id IN (?)", bun.In(filter.NetworkSecurityGroupIDs))
		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "network_security_group_ids", filter.NetworkSecurityGroupIDs)
		}
	}

	if filter.VpcIDs != nil {

		// Attach interface data with an outer join.
		// We seem to have a few scenarios (nvlink, ib, etc)
		// where an instance could have a VPC but not have
		// ethernet interfaces associated.
		query = query.Join("LEFT OUTER JOIN interface ifc").
			JoinOn("ifc.instance_id = i.id").
			JoinOn("ifc.deleted IS NULL")

		// Attach vpc_prefix data with an outer join
		query = query.Join("LEFT OUTER JOIN vpc_prefix vp").
			JoinOn("vp.id = ifc.vpc_prefix_id").
			JoinOn("vp.deleted IS NULL")

		isd.tracerSpan.SetAttribute(instanceDAOSpan, "vpc_ids", filter.VpcIDs)

		// Filter on VPC IDs
		// Match instances by either their primary VPC (`i.vpc_id`) or any
		// interface-attached VPC prefix (`vp.vpc_id`).
		// We need to check for both so that we cover legacy VPCs with network segments and
		// VPCs with VPC prefixes.
		query = query.Where("(vp.vpc_id IN (?) OR i.vpc_id IN (?))", bun.In(filter.VpcIDs), bun.In(filter.VpcIDs))

		// Now boil everything down to only the unique instance records.
		query = query.Distinct()

	}

	if filter.MachineIDs != nil {
		query = query.Where("i.machine_id IN (?)", bun.In(filter.MachineIDs))
		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "machine_ids", filter.MachineIDs)
		}
	}

	if filter.ControllerInstanceIDs != nil {
		query = query.Where("i.controller_instance_id IN (?)", bun.In(filter.ControllerInstanceIDs))
		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "controller_instance_ids", filter.ControllerInstanceIDs)
		}
	}

	if filter.OperatingSystemIDs != nil {
		query = query.Where("i.operating_system_id IN (?)", bun.In(filter.OperatingSystemIDs))
		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "operating_system_ids", filter.OperatingSystemIDs)
		}
	}

	if filter.Statuses != nil {
		query = query.Where("i.status IN (?)", bun.In(filter.Statuses))
		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "statuses", filter.Statuses)
		}
	}

	searchQuery, searchTokens, ok := db.NormalizeSearchQuery(filter.SearchQuery)
	if ok {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				Where("to_tsvector('english', (coalesce(i.name, ' ') || ' ' || coalesce(i.status, ' ') || ' ' || coalesce(i.labels::text, ' '))) @@ to_tsquery('english', ?)", *searchTokens).
				WhereOr("i.name ILIKE ?", "%"+searchQuery+"%").
				WhereOr("i.status ILIKE ?", "%"+searchQuery+"%").
				WhereOr("i.description ILIKE ?", "%"+searchQuery+"%").
				WhereOr("i.labels::text ILIKE ?", "%"+searchQuery+"%")
		})

		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "search_query", searchQuery)
		}
	}
	return query, nil
}

// GetAll returns all Instances filtered by the fields in InstanceFilterInput:
// InstanceIDs, Names, TenantIDs, InfrastructureProviderIDs, SiteIDs, InstanceTypeIDs,
// VpcIDs, MachineIDs, ControllerInstanceIDs, OperatingSystemIDs, IDsNotIn, SearchQuery,
// Statuses, TenantOrgName, Labels, NetworkSecurityGroupIDs, and Hostnames.
// Allocation-based filters are intentionally omitted because direct instance-allocation linkage was removed.
// errors are returned only when there is a db related error
// if records not found, then error is nil, but length of returned slice is 0
// if page.OrderBy is nil, then records are ordered by column specified in InstanceOrderByDefault in ascending order
func (isd InstanceSQLDAO) GetAll(ctx context.Context, tx *db.Tx, filter InstanceFilterInput, page paginator.PageInput, includeRelations []string) ([]Instance, int, error) {
	// Create a child span and set the attributes for current request
	ctx, instanceDAOSpan := isd.tracerSpan.CreateChildInCurrentContext(ctx, "InstanceDAO.GetAll")
	if instanceDAOSpan != nil {
		defer instanceDAOSpan.End()
	}

	var instances []Instance

	query := db.GetIDB(tx, isd.dbSession).NewSelect().Model(&instances).ColumnExpr("i.*")

	query, err := isd.setQueryWithFilter(filter, query, instanceDAOSpan)
	if err != nil {
		return instances, 0, err
	}

	var multiOrderBy []*paginator.OrderBy
	if page.OrderBy != nil {
		multiOrderBy = append(multiOrderBy, page.OrderBy)
		// handle sorting by presence of infiniband
		if page.OrderBy.Field == instanceOrderByHasInfiniBandExt {
			query = query.ColumnExpr("mc.type AS mc_type")
			query = query.Join("LEFT JOIN machine_capability AS mc ON i.machine_id = mc.machine_id AND mc.type = 'InfiniBand'").Distinct()
		}
	}
	if page.OrderBy == nil || page.OrderBy.Field != InstanceOrderByDefault {
		// add default sort to make sure objects returned in same order
		multiOrderBy = append(multiOrderBy, paginator.NewDefaultOrderBy(InstanceOrderByDefault))
	}

	for _, orderBy := range multiOrderBy {
		// validate order by
		if relationName := instanceOrderByFieldToRelation[orderBy.Field]; relationName != "" {
			if !db.IsStrInSlice(relationName, includeRelations) {
				// add relation, so that we can sort on joined data
				includeRelations = append(includeRelations, relationName)
			}
		}
		// convert to internal
		if internalName := instanceOrderByFieldExtToInt[orderBy.Field]; internalName != "" {
			orderBy.Field = internalName
		}
	}

	for _, relation := range includeRelations {
		query = query.Relation(relation)
	}

	paginator, err := paginator.NewPaginatorMultiOrderBy(ctx, query, page.Offset, page.Limit, multiOrderBy, instanceOrderByFieldsInt)
	if err != nil {
		return nil, 0, err
	}

	err = paginator.Query.Limit(paginator.Limit).Offset(paginator.Offset).Scan(ctx)
	if err != nil {
		return nil, 0, err
	}

	return instances, paginator.Total, nil
}

// GetCount returns total count of rows for specified filter
func (isd InstanceSQLDAO) GetCount(ctx context.Context, tx *db.Tx, filter InstanceFilterInput) (count int, err error) {
	// Create a child span and set the attributes for current request
	ctx, instanceDAOSpan := isd.tracerSpan.CreateChildInCurrentContext(ctx, "InstanceDAO.GetCount")
	if instanceDAOSpan != nil {
		defer instanceDAOSpan.End()
	}

	query := db.GetIDB(tx, isd.dbSession).NewSelect().Model((*Instance)(nil))
	query, err = isd.setQueryWithFilter(filter, query, instanceDAOSpan)
	if err != nil {
		return 0, err
	}

	return query.Count(ctx)
}

// Update updates specified fields of an existing Instance
// The updated fields are assumed to be set to non-null values
// For setting to null values, use: Clear
// since there are 2 operations (UPDATE, SELECT), in this, it is required that
// this library call happens within a transaction
func (isd InstanceSQLDAO) Update(ctx context.Context, tx *db.Tx, input InstanceUpdateInput) (*Instance, error) {
	// Create a child span and set the attributes for current request
	ctx, instanceDAOSpan := isd.tracerSpan.CreateChildInCurrentContext(ctx, "InstanceDAO.Update")
	if instanceDAOSpan != nil {
		defer instanceDAOSpan.End()
		// Detailed per-field tracing is recorded in the UpdateMultiple child span.
	}

	results, err := isd.UpdateMultiple(ctx, tx, []InstanceUpdateInput{input})
	if err != nil {
		return nil, err
	}
	return &results[0], nil
}

// Clear sets parameters of an existing Instance to null values in db
// parameters when true, the are set to null in db
// since there are 2 operations (UPDATE, SELECT), it is required that
// this must be within a transaction
func (isd InstanceSQLDAO) Clear(ctx context.Context, tx *db.Tx, input InstanceClearInput) (*Instance, error) {
	// Create a child span and set the attributes for current request
	ctx, instanceDAOSpan := isd.tracerSpan.CreateChildInCurrentContext(ctx, "InstanceDAO.Clear")
	if instanceDAOSpan != nil {
		defer instanceDAOSpan.End()
	}

	i := &Instance{
		ID: input.InstanceID,
	}

	updatedFields := []string{}

	if input.Description {
		i.Description = nil
		updatedFields = append(updatedFields, "description")
	}
	if input.MachineID {
		i.MachineID = nil
		updatedFields = append(updatedFields, "machine_id")
	}
	if input.ControllerInstanceID {
		i.ControllerInstanceID = nil
		updatedFields = append(updatedFields, "controller_instance_id")
	}
	if input.Hostname {
		i.Hostname = nil
		updatedFields = append(updatedFields, "hostname")
	}
	if input.OperatingSystemID {
		i.OperatingSystemID = nil
		updatedFields = append(updatedFields, "operating_system_id")
	}
	if input.IpxeScript {
		i.IpxeScript = nil
		updatedFields = append(updatedFields, "ipxe_script")
	}
	if input.UserData {
		i.UserData = nil
		updatedFields = append(updatedFields, "user_data")
	}
	if input.Labels {
		i.Labels = nil
		updatedFields = append(updatedFields, "labels")
	}
	if input.NetworkSecurityGroupID {
		i.NetworkSecurityGroupID = nil
		updatedFields = append(updatedFields, "network_security_group_id")
	}
	if input.NetworkSecurityGroupPropagationDetails {
		i.NetworkSecurityGroupPropagationDetails = nil
		updatedFields = append(updatedFields, "network_security_group_propagation_details")
	}
	if input.TpmEkCertificate {
		i.TpmEkCertificate = nil
		updatedFields = append(updatedFields, "tpm_ek_certificate")
	}

	if len(updatedFields) > 0 {
		updatedFields = append(updatedFields, "updated")

		_, err := db.GetIDB(tx, isd.dbSession).NewUpdate().Model(i).Column(updatedFields...).Where("id = ?", input.InstanceID).Exec(ctx)
		if err != nil {
			return nil, err
		}
	}

	nv, err := isd.GetByID(ctx, tx, i.ID, nil)
	if err != nil {
		return nil, err
	}
	return nv, nil
}

// Delete deletes an Instance by ID
// error is returned only if there is a db error
// if the object being deleted doesnt exist, error is not returned (idempotent delete)
func (isd InstanceSQLDAO) Delete(ctx context.Context, tx *db.Tx, id uuid.UUID) error {
	// Create a child span and set the attributes for current request
	ctx, instanceDAOSpan := isd.tracerSpan.CreateChildInCurrentContext(ctx, "InstanceDAO.Delete")
	if instanceDAOSpan != nil {
		defer instanceDAOSpan.End()

		isd.tracerSpan.SetAttribute(instanceDAOSpan, "id", id.String())
	}

	i := &Instance{
		ID: id,
	}

	_, err := db.GetIDB(tx, isd.dbSession).NewDelete().Model(i).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// CreateMultiple creates multiple Instances from the given parameters
// The returned Instances will not have any related structs filled in
// since there are 2 operations (INSERT, SELECT), in this, it is required that
// this library call happens within a transaction
func (isd InstanceSQLDAO) CreateMultiple(ctx context.Context, tx *db.Tx, inputs []InstanceCreateInput) ([]Instance, error) {
	if len(inputs) > db.MaxBatchItems {
		return nil, fmt.Errorf("batch size %d exceeds maximum allowed %d", len(inputs), db.MaxBatchItems)
	}

	// Create a child span and set the attributes for current request
	ctx, instanceDAOSpan := isd.tracerSpan.CreateChildInCurrentContext(ctx, "InstanceDAO.CreateMultiple")
	if instanceDAOSpan != nil {
		defer instanceDAOSpan.End()
		isd.tracerSpan.SetAttribute(instanceDAOSpan, "batch_size", len(inputs))
	}

	if len(inputs) == 0 {
		return []Instance{}, nil
	}

	instances := make([]Instance, 0, len(inputs))
	ids := make([]uuid.UUID, 0, len(inputs))

	for _, input := range inputs {
		i := Instance{
			ID:                                     uuid.New(),
			Name:                                   input.Name,
			Description:                            input.Description,
			TenantID:                               input.TenantID,
			InfrastructureProviderID:               input.InfrastructureProviderID,
			SiteID:                                 input.SiteID,
			InstanceTypeID:                         input.InstanceTypeID,
			NetworkSecurityGroupID:                 input.NetworkSecurityGroupID,
			NetworkSecurityGroupPropagationDetails: input.NetworkSecurityGroupPropagationDetails,
			VpcID:                                  input.VpcID,
			MachineID:                              input.MachineID,
			ControllerInstanceID:                   input.ControllerInstanceID,
			Hostname:                               input.Hostname,
			OperatingSystemID:                      input.OperatingSystemID,
			IpxeScript:                             input.IpxeScript,
			AlwaysBootWithCustomIpxe:               input.AlwaysBootWithCustomIpxe,
			PhoneHomeEnabled:                       input.PhoneHomeEnabled,
			UserData:                               input.UserData,
			IsUpdatePending:                        input.IsUpdatePending,
			InfinityRCRStatus:                      input.InfinityRCRStatus,
			TpmEkCertificate:                       input.TpmEkCertificate,
			Status:                                 input.Status,
			PowerStatus:                            input.PowerStatus,
			CreatedBy:                              input.CreatedBy,
			Labels:                                 input.Labels,
		}
		instances = append(instances, i)
		ids = append(ids, i.ID)
	}

	_, err := db.GetIDB(tx, isd.dbSession).NewInsert().Model(&instances).Exec(ctx)
	if err != nil {
		return nil, err
	}

	// Fetch the created instances
	var result []Instance
	err = db.GetIDB(tx, isd.dbSession).NewSelect().Model(&result).Where("i.id IN (?)", bun.In(ids)).Scan(ctx)
	if err != nil {
		return nil, err
	}

	// Sort result to match input order (O(n) direct index placement)
	// This check should never fail since we just inserted these records with the exact ids
	if len(result) != len(ids) {
		return nil, fmt.Errorf("unexpected result count: got %d, expected %d", len(result), len(ids))
	}
	idToIndex := make(map[uuid.UUID]int, len(ids))
	for i, id := range ids {
		idToIndex[id] = i
	}
	sorted := make([]Instance, len(result))
	for _, item := range result {
		sorted[idToIndex[item.ID]] = item
	}

	return sorted, nil
}

// UpdateMultiple updates multiple Instances with the given parameters using a single bulk UPDATE query
// All inputs should update the same set of fields for optimal performance
// The updated fields are assumed to be set to non-null values
// since there are 2 operations (UPDATE, SELECT), it is required that
// this library call happens within a transaction
func (isd InstanceSQLDAO) UpdateMultiple(ctx context.Context, tx *db.Tx, inputs []InstanceUpdateInput) ([]Instance, error) {
	if len(inputs) > db.MaxBatchItems {
		return nil, fmt.Errorf("batch size %d exceeds maximum allowed %d", len(inputs), db.MaxBatchItems)
	}

	// Create a child span and set the attributes for current request
	ctx, instanceDAOSpan := isd.tracerSpan.CreateChildInCurrentContext(ctx, "InstanceDAO.UpdateMultiple")
	if instanceDAOSpan != nil {
		defer instanceDAOSpan.End()
		isd.tracerSpan.SetAttribute(instanceDAOSpan, "batch_size", len(inputs))
	}

	if len(inputs) == 0 {
		return []Instance{}, nil
	}

	// Build instances and collect columns to update
	instances := make([]*Instance, 0, len(inputs))
	ids := make([]uuid.UUID, 0, len(inputs))
	columnsSet := make(map[string]bool)

	// Limit per-item tracing to avoid overly-large spans; see db.MaxBatchItemsToTrace for details
	traceItems := len(inputs)
	if traceItems > db.MaxBatchItemsToTrace {
		traceItems = db.MaxBatchItemsToTrace
		if instanceDAOSpan != nil {
			isd.tracerSpan.SetAttribute(instanceDAOSpan, "items_truncated", "true")
		}
	}

	for idx, input := range inputs {
		i := &Instance{
			ID: input.InstanceID,
		}
		columns := []string{}
		addTrace := instanceDAOSpan != nil && idx < traceItems
		prefix := fmt.Sprintf("items.%d.", idx)

		// Field-level tracing: only trace fields that are actually being updated for this item
		// This keeps spans focused and avoids recording null/unchanged values
		if input.Name != nil {
			i.Name = *input.Name
			columns = append(columns, "name")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"name", *input.Name)
			}
		}
		if input.Description != nil {
			i.Description = input.Description
			columns = append(columns, "description")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"description", *input.Description)
			}
		}
		if input.TenantID != nil {
			i.TenantID = *input.TenantID
			columns = append(columns, "tenant_id")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"tenant_id", input.TenantID.String())
			}
		}
		if input.InfrastructureProviderID != nil {
			i.InfrastructureProviderID = *input.InfrastructureProviderID
			columns = append(columns, "infrastructure_provider_id")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"infrastructure_provider_id", input.InfrastructureProviderID.String())
			}
		}
		if input.SiteID != nil {
			i.SiteID = *input.SiteID
			columns = append(columns, "site_id")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"site_id", input.SiteID.String())
			}
		}
		if input.InstanceTypeID != nil {
			i.InstanceTypeID = input.InstanceTypeID
			columns = append(columns, "instance_type_id")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"instance_type_id", input.InstanceTypeID.String())
			}
		}
		if input.NetworkSecurityGroupID != nil {
			i.NetworkSecurityGroupID = input.NetworkSecurityGroupID
			columns = append(columns, "network_security_group_id")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"network_security_group_id", *input.NetworkSecurityGroupID)
			}
		}
		if input.VpcID != nil {
			i.VpcID = *input.VpcID
			columns = append(columns, "vpc_id")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"vpc_id", input.VpcID.String())
			}
		}
		if input.MachineID != nil {
			i.MachineID = input.MachineID
			columns = append(columns, "machine_id")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"machine_id", *input.MachineID)
			}
		}
		if input.ControllerInstanceID != nil {
			i.ControllerInstanceID = input.ControllerInstanceID
			columns = append(columns, "controller_instance_id")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"controller_instance_id", input.ControllerInstanceID.String())
			}
		}
		if input.Hostname != nil {
			i.Hostname = input.Hostname
			columns = append(columns, "hostname")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"hostname", *input.Hostname)
			}
		}
		if input.OperatingSystemID != nil {
			i.OperatingSystemID = input.OperatingSystemID
			columns = append(columns, "operating_system_id")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"operating_system_id", input.OperatingSystemID.String())
			}
		}
		if input.IpxeScript != nil {
			i.IpxeScript = input.IpxeScript
			columns = append(columns, "ipxe_script")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"ipxe_script", *input.IpxeScript)
			}
		}
		if input.AlwaysBootWithCustomIpxe != nil {
			i.AlwaysBootWithCustomIpxe = *input.AlwaysBootWithCustomIpxe
			columns = append(columns, "always_boot_with_custom_ipxe")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"always_boot_with_custom_ipxe", fmt.Sprintf("%t", *input.AlwaysBootWithCustomIpxe))
			}
		}
		if input.PhoneHomeEnabled != nil {
			i.PhoneHomeEnabled = *input.PhoneHomeEnabled
			columns = append(columns, "phone_home_enabled")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"phone_home_enabled", fmt.Sprintf("%t", *input.PhoneHomeEnabled))
			}
		}
		if input.UserData != nil {
			i.UserData = input.UserData
			columns = append(columns, "user_data")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"user_data", *input.UserData)
			}
		}
		if input.Labels != nil {
			i.Labels = input.Labels
			columns = append(columns, "labels")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"labels", input.Labels)
			}
		}
		if input.IsUpdatePending != nil {
			i.IsUpdatePending = *input.IsUpdatePending
			columns = append(columns, "is_update_pending")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"is_update_pending", fmt.Sprintf("%t", *input.IsUpdatePending))
			}
		}
		if input.InfinityRCRStatus != nil {
			i.InfinityRCRStatus = input.InfinityRCRStatus
			columns = append(columns, "infinity_rcr_status")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"infinity_rcr_status", *input.InfinityRCRStatus)
			}
		}
		if input.Status != nil {
			i.Status = *input.Status
			columns = append(columns, "status")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"status", *input.Status)
			}
		}
		if input.PowerStatus != nil {
			i.PowerStatus = input.PowerStatus
			columns = append(columns, "power_status")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"power_status", *input.PowerStatus)
			}
		}
		if input.IsMissingOnSite != nil {
			i.IsMissingOnSite = *input.IsMissingOnSite
			columns = append(columns, "is_missing_on_site")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"is_missing_on_site", fmt.Sprintf("%t", *input.IsMissingOnSite))
			}
		}
		if input.NetworkSecurityGroupPropagationDetails != nil {
			i.NetworkSecurityGroupPropagationDetails = input.NetworkSecurityGroupPropagationDetails
			columns = append(columns, "network_security_group_propagation_details")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"network_security_group_propagation_details", input.NetworkSecurityGroupPropagationDetails)
			}
		}
		if input.TpmEkCertificate != nil {
			i.TpmEkCertificate = input.TpmEkCertificate
			columns = append(columns, "tpm_ek_certificate")
			if addTrace {
				isd.tracerSpan.SetAttribute(instanceDAOSpan, prefix+"tpm_ek_certificate", *input.TpmEkCertificate)
			}
		}

		instances = append(instances, i)
		ids = append(ids, input.InstanceID)
		for _, col := range columns {
			columnsSet[col] = true
		}

	}

	// Build column list
	columns := make([]string, 0, len(columnsSet)+1)
	for col := range columnsSet {
		columns = append(columns, col)
	}
	columns = append(columns, "updated")

	// Execute bulk update
	_, err := db.GetIDB(tx, isd.dbSession).NewUpdate().
		Model(&instances).
		Column(columns...).
		Bulk().
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	// Fetch the updated instances
	var result []Instance
	err = db.GetIDB(tx, isd.dbSession).NewSelect().Model(&result).Where("i.id IN (?)", bun.In(ids)).Scan(ctx)
	if err != nil {
		return nil, err
	}

	// Sort result to match input order (O(n) direct index placement)
	// This check should never fail since we just updated these records with the exact ids
	if len(result) != len(ids) {
		return nil, fmt.Errorf("unexpected result count: got %d, expected %d", len(result), len(ids))
	}
	idToIndex := make(map[uuid.UUID]int, len(ids))
	for i, id := range ids {
		idToIndex[id] = i
	}
	sorted := make([]Instance, len(result))
	for _, item := range result {
		sorted[idToIndex[item.ID]] = item
	}

	return sorted, nil
}

// NewInstanceDAO returns a new InstanceDAO
func NewInstanceDAO(dbSession *db.Session) InstanceDAO {
	return &InstanceSQLDAO{
		dbSession:  dbSession,
		tracerSpan: stracer.NewTracerSpan(),
	}
}
