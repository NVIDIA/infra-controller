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
	"time"

	"github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	"github.com/NVIDIA/infra-controller-rest/db/pkg/db/paginator"
	"github.com/google/uuid"

	"github.com/uptrace/bun"

	stracer "github.com/NVIDIA/infra-controller-rest/db/pkg/tracer"
)

const (
	// VpcPrefixStatusReady status is ready
	VpcPrefixStatusReady = "Ready"
	// VpcPrefixStatusError status is error
	VpcPrefixStatusError = "Error"
	// VpcPrefixStatusDeleting indicates that the VpcPrefix is being deleted
	VpcPrefixStatusDeleting = "Deleting"
	// VpcPrefixStatusDeleted indicates that the VpcPrefix has been deleted
	VpcPrefixStatusDeleted = "Deleted"
	// VpcPrefixRelationName is the relation name for the VpcPrefix model
	VpcPrefixRelationName = "VpcPrefix"

	// VpcPrefixOrderByDefault default field to be used for ordering when none specified
	VpcPrefixOrderByDefault = "created"
)

var (
	// VpcPrefixOrderByFields is a list of valid order by fields for the VpcPrefix model
	VpcPrefixOrderByFields = []string{"name", "status", "created", "updated"}
	// VpcPrefixRelatedEntities is a list of valid relation by fields for the VpcPrefix model
	VpcPrefixRelatedEntities = map[string]bool{
		SiteRelationName:    true,
		VpcRelationName:     true,
		TenantRelationName:  true,
		IPBlockRelationName: true,
	}
	// VpcPrefixStatusMap is a list of valid status for the VpcPrefix model
	VpcPrefixStatusMap = map[string]bool{
		VpcPrefixStatusReady:    true,
		VpcPrefixStatusError:    true,
		VpcPrefixStatusDeleting: true,
		VpcPrefixStatusDeleted:  true,
	}
)

// VpcPrefix is a network construct for bare-metal machines
type VpcPrefix struct {
	bun.BaseModel `bun:"table:vpc_prefix,alias:vp"`

	ID              uuid.UUID  `bun:"type:uuid,pk"`
	Name            string     `bun:"name,notnull"`
	Org             string     `bun:"org,notnull"`
	SiteID          uuid.UUID  `bun:"site_id,type:uuid,notnull"`
	Site            *Site      `bun:"rel:belongs-to,join:site_id=id"`
	VpcID           uuid.UUID  `bun:"vpc_id,type:uuid,notnull"`
	Vpc             *Vpc       `bun:"rel:belongs-to,join:vpc_id=id"`
	TenantID        uuid.UUID  `bun:"tenant_id,type:uuid"`
	Tenant          *Tenant    `bun:"rel:belongs-to,join:tenant_id=id"`
	IPBlockID       *uuid.UUID `bun:"ip_block_id,type:uuid"`
	IPBlock         *IPBlock   `bun:"rel:belongs-to,join:ip_block_id=id"`
	Prefix          string     `bun:"prefix,notnull"`
	PrefixLength    int        `bun:"prefix_length,notnull"`
	Status          string     `bun:"status,notnull"`
	IsMissingOnSite bool       `bun:"is_missing_on_site,notnull"`
	Created         time.Time  `bun:"created,nullzero,notnull,default:current_timestamp"`
	Updated         time.Time  `bun:"updated,nullzero,notnull,default:current_timestamp"`
	Deleted         *time.Time `bun:"deleted,soft_delete"`
	CreatedBy       uuid.UUID  `bun:"type:uuid,notnull"`
}

// VpcPrefixCreateInput input parameters for Create method
type VpcPrefixCreateInput struct {
	VpcPrefixID  *uuid.UUID
	Name         string
	TenantOrg    string
	SiteID       uuid.UUID
	VpcID        uuid.UUID
	TenantID     uuid.UUID
	IpBlockID    *uuid.UUID
	Prefix       string
	PrefixLength int
	Status       string
	CreatedBy    uuid.UUID
}

// VpcPrefixUpdateInput input parameters for Update method
type VpcPrefixUpdateInput struct {
	VpcPrefixID     uuid.UUID
	Name            *string
	TenantOrg       *string
	VpcID           *uuid.UUID
	TenantID        *uuid.UUID
	IpBlockID       *uuid.UUID
	Prefix          *string
	PrefixLength    *int
	Status          *string
	IsMissingOnSite *bool
}

// VpcPrefixFilterInput input parameters for Filter method
type VpcPrefixFilterInput struct {
	VpcPrefixIDs  []uuid.UUID
	Names         []string
	VpcIDs        []uuid.UUID
	TenantOrgs    []string
	TenantIDs     []uuid.UUID
	IpBlockIDs    []uuid.UUID
	SiteIDs       []uuid.UUID
	Statuses      []string
	SearchQuery   *string
	Prefixes      []string
	PrefixLengths []int
}

var _ bun.BeforeAppendModelHook = (*VpcPrefix)(nil)

// BeforeAppendModel is a hook that is called before the model is appended to the query
func (vp *VpcPrefix) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		vp.Created = db.GetCurTime()
		vp.Updated = db.GetCurTime()
	case *bun.UpdateQuery:
		vp.Updated = db.GetCurTime()
	}
	return nil
}

var _ bun.BeforeCreateTableHook = (*VpcPrefix)(nil)

// BeforeCreateTable is a hook that is called before the table is created
func (vp *VpcPrefix) BeforeCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	query.ForeignKey(`("site_id") REFERENCES "site" ("id")`).
		ForeignKey(`("vpc_id") REFERENCES "vpc" ("id")`).
		ForeignKey(`("tenant_id") REFERENCES "tenant" ("id")`).
		ForeignKey(`("ip_block_id") REFERENCES "ip_block" ("id")`)
	return nil
}

// VpcPrefixDAO is an interface for interacting with the VpcPrefix model
type VpcPrefixDAO interface {
	//
	Create(ctx context.Context, tx *db.Tx, input VpcPrefixCreateInput) (*VpcPrefix, error)
	//
	GetByID(ctx context.Context, tx *db.Tx, id uuid.UUID, includeRelations []string) (*VpcPrefix, error)
	//
	GetAll(ctx context.Context, tx *db.Tx, filter VpcPrefixFilterInput, page paginator.PageInput, includeRelations []string) ([]VpcPrefix, int, error)
	//
	Update(ctx context.Context, tx *db.Tx, input VpcPrefixUpdateInput) (*VpcPrefix, error)
	//
	Delete(ctx context.Context, tx *db.Tx, id uuid.UUID) error
}

// VpcPrefixSQLDAO is an implementation of the VpcPrefixDAO interface
type VpcPrefixSQLDAO struct {
	dbSession *db.Session
	VpcPrefixDAO
	tracerSpan *stracer.TracerSpan
}

// Create creates a new VpcPrefix from the given parameters
func (vpsd VpcPrefixSQLDAO) Create(ctx context.Context, tx *db.Tx, input VpcPrefixCreateInput) (*VpcPrefix, error) {
	// Create a child span and set the attributes for current request
	ctx, vpDAOSpan := vpsd.tracerSpan.CreateChildInCurrentContext(ctx, "VpcPrefixDAO.Create")
	if vpDAOSpan != nil {
		defer vpDAOSpan.End()

		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "name", input.Name)
	}

	id := input.VpcPrefixID
	if id == nil {
		id = db.GetUUIDPtr(uuid.New())
	}

	vpp := &VpcPrefix{
		ID:              *id,
		Name:            input.Name,
		Org:             input.TenantOrg,
		SiteID:          input.SiteID,
		VpcID:           input.VpcID,
		TenantID:        input.TenantID,
		IPBlockID:       input.IpBlockID,
		Prefix:          input.Prefix,
		PrefixLength:    input.PrefixLength,
		IsMissingOnSite: false,
		Status:          input.Status,
		CreatedBy:       input.CreatedBy,
	}

	_, err := db.GetIDB(tx, vpsd.dbSession).NewInsert().Model(vpp).Exec(ctx)
	if err != nil {
		return nil, err
	}

	nvp, err := vpsd.GetByID(ctx, tx, vpp.ID, nil)
	if err != nil {
		return nil, err
	}

	return nvp, nil
}

// GetByID returns a VpcPrefix by ID
// includeRelation can be a subset of Vpc
// returns db.ErrDoesNotExist error if the record is not found
func (vpsd VpcPrefixSQLDAO) GetByID(ctx context.Context, tx *db.Tx, id uuid.UUID, includeRelations []string) (*VpcPrefix, error) {
	// Create a child span and set the attributes for current request
	ctx, vpDAOSpan := vpsd.tracerSpan.CreateChildInCurrentContext(ctx, "VpcPrefixDAO.GetByID")
	if vpDAOSpan != nil {
		defer vpDAOSpan.End()

		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "id", id.String())
	}

	vpp := &VpcPrefix{}

	query := db.GetIDB(tx, vpsd.dbSession).NewSelect().Model(vpp).Where("vp.id = ?", id)

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

	return vpp, nil
}

// GetAll returns all VpcPrefixs filtering by Vpc, Domain, Tenant
// errors are returned only when there is a db related error
// if records not found, then error is nil, but length of returned slice is 0
// if orderBy is nil, then records are ordered by column specified in VpcPrefixOrderByDefault in ascending order
func (vpsd VpcPrefixSQLDAO) GetAll(ctx context.Context, tx *db.Tx, filter VpcPrefixFilterInput, page paginator.PageInput, includeRelations []string) ([]VpcPrefix, int, error) {
	// Create a child span and set the attributes for current request
	ctx, vpDAOSpan := vpsd.tracerSpan.CreateChildInCurrentContext(ctx, "VpcPrefixDAO.GetAll")
	if vpDAOSpan != nil {
		defer vpDAOSpan.End()
	}

	vps := []VpcPrefix{}

	query := db.GetIDB(tx, vpsd.dbSession).NewSelect().Model(&vps)
	if filter.VpcPrefixIDs != nil {
		query = query.Where("vp.id IN (?)", bun.In(filter.VpcPrefixIDs))
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "vpc_prefix_ids", filter.VpcPrefixIDs)
	}
	if filter.Names != nil {
		query = query.Where("vp.name IN (?)", bun.In(filter.Names))
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "name", filter.Names)
	}
	if filter.SiteIDs != nil {
		query = query.Where("vp.site_id IN (?)", bun.In(filter.SiteIDs))
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "site_id", filter.SiteIDs)
	}
	if filter.VpcIDs != nil {
		query = query.Where("vp.vpc_id IN (?)", bun.In(filter.VpcIDs))
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "vpc_id", filter.VpcIDs)
	}
	if filter.TenantIDs != nil {
		query = query.Where("vp.tenant_id IN (?)", bun.In(filter.TenantIDs))
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "tenant_id", filter.TenantIDs)
	}
	if filter.IpBlockIDs != nil {
		query = query.Where("vp.ip_block_id IN (?)", bun.In(filter.IpBlockIDs))
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "ip_block_id", filter.IpBlockIDs)
	}
	if filter.Prefixes != nil {
		query = query.Where("vp.prefix IN (?)", bun.In(filter.Prefixes))
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "prefix", filter.Prefixes)
	}
	if filter.PrefixLengths != nil {
		query = query.Where("vp.prefix_length IN (?)", bun.In(filter.PrefixLengths))
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "prefix_length", filter.PrefixLengths)
	}
	if filter.Statuses != nil {
		query = query.Where("vp.status IN (?)", bun.In(filter.Statuses))
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "status", filter.Statuses)
	}
	searchQuery, normalizedTokens, ok := db.NormalizeSearchQuery(filter.SearchQuery)
	if ok {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				Where("to_tsvector('english', (coalesce(vp.name, ' ') || ' ' || coalesce(vp.status, ' '))) @@ to_tsquery('english', ?)", *normalizedTokens).
				WhereOr("vp.name ILIKE ?", "%"+searchQuery+"%").
				WhereOr("vp.status ILIKE ?", "%"+searchQuery+"%")
		})
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "search_query", searchQuery)
	}

	for _, relation := range includeRelations {
		query = query.Relation(relation)
	}

	// if no order is passed, set default to make sure objects return always in the same order and pagination works properly
	if page.OrderBy == nil {
		page.OrderBy = paginator.NewDefaultOrderBy(VpcPrefixOrderByDefault)
	}

	paginator, err := paginator.NewPaginator(ctx, query, page.Offset, page.Limit, page.OrderBy, VpcPrefixOrderByFields)
	if err != nil {
		return nil, 0, err
	}

	err = paginator.Query.Limit(paginator.Limit).Offset(paginator.Offset).Scan(ctx)
	if err != nil {
		return nil, 0, err
	}

	return vps, paginator.Total, nil
}

// Update updates specified fields of an existing VpcPrefix
// The updated fields are assumed to be set to non-null values
// For setting to null values, use: Clear
// since there are 2 operations (UPDATE, SELECT), in this, it is required that
// this library call happens within a transaction
func (vpsd VpcPrefixSQLDAO) Update(ctx context.Context, tx *db.Tx, input VpcPrefixUpdateInput) (*VpcPrefix, error) {
	// Create a child span and set the attributes for current request
	ctx, vpDAOSpan := vpsd.tracerSpan.CreateChildInCurrentContext(ctx, "VpcPrefixDAO.Update")
	if vpDAOSpan != nil {
		defer vpDAOSpan.End()

		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "id", input.VpcPrefixID)
	}

	vp := &VpcPrefix{
		ID: input.VpcPrefixID,
	}
	updatedFields := []string{}

	if input.Name != nil {
		vp.Name = *input.Name
		updatedFields = append(updatedFields, "name")
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "name", *input.Name)
	}
	if input.TenantOrg != nil {
		vp.Org = *input.TenantOrg
		updatedFields = append(updatedFields, "org")
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "org", *input.TenantOrg)
	}
	if input.VpcID != nil {
		vp.VpcID = *input.VpcID
		updatedFields = append(updatedFields, "vpc_id")
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "vpc_id", input.VpcID.String())
	}
	if input.TenantID != nil {
		vp.TenantID = *input.TenantID
		updatedFields = append(updatedFields, "tenant_id")
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "tenant_id", input.TenantID.String())
	}
	if input.IpBlockID != nil {
		vp.IPBlockID = input.IpBlockID
		updatedFields = append(updatedFields, "ip_block_id")
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "ip_block_id", input.IpBlockID.String())
	}
	if input.Prefix != nil {
		vp.Prefix = *input.Prefix
		updatedFields = append(updatedFields, "prefix")
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "prefix", *input.Prefix)
	}
	if input.PrefixLength != nil {
		vp.PrefixLength = *input.PrefixLength
		updatedFields = append(updatedFields, "prefix_length")
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "prefix_length", *input.PrefixLength)
	}
	if input.Status != nil {
		vp.Status = *input.Status
		updatedFields = append(updatedFields, "status")
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "status", *input.Status)
	}
	if input.IsMissingOnSite != nil {
		vp.IsMissingOnSite = *input.IsMissingOnSite
		updatedFields = append(updatedFields, "is_missing_on_site")
		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "is_missing_on_site", *input.IsMissingOnSite)
	}

	if len(updatedFields) > 0 {
		updatedFields = append(updatedFields, "updated")

		_, err := db.GetIDB(tx, vpsd.dbSession).NewUpdate().Model(vp).Column(updatedFields...).Where("id = ?", input.VpcPrefixID).Exec(ctx)
		if err != nil {
			return nil, err
		}
	}

	nvp, err := vpsd.GetByID(ctx, tx, vp.ID, nil)

	if err != nil {
		return nil, err
	}
	return nvp, nil
}

// Delete deletes an VpcPrefix by ID
// error is returned only if there is a db error
// if the object being deleted doesnt exist, error is not returned (idempotent delete)
func (vpsd VpcPrefixSQLDAO) Delete(ctx context.Context, tx *db.Tx, id uuid.UUID) error {
	// Create a child span and set the attributes for current request
	ctx, vpDAOSpan := vpsd.tracerSpan.CreateChildInCurrentContext(ctx, "VpcPrefixDAO.Delete")
	if vpDAOSpan != nil {
		defer vpDAOSpan.End()

		vpsd.tracerSpan.SetAttribute(vpDAOSpan, "id", id.String())
	}

	vp := &VpcPrefix{
		ID: id,
	}

	_, err := db.GetIDB(tx, vpsd.dbSession).NewDelete().Model(vp).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// NewVpcPrefixDAO returns a new VpcPrefixDAO
func NewVpcPrefixDAO(dbSession *db.Session) VpcPrefixDAO {
	return &VpcPrefixSQLDAO{
		dbSession:  dbSession,
		tracerSpan: stracer.NewTracerSpan(),
	}
}
