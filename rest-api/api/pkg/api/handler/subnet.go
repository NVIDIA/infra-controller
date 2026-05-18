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

package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"

	"go.opentelemetry.io/otel/attribute"
	temporalClient "go.temporal.io/sdk/client"
	tp "go.temporal.io/sdk/temporal"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"

	"github.com/NVIDIA/infra-controller-rest/api/internal/config"
	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/handler/util/common"
	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/model"
	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/pagination"
	sc "github.com/NVIDIA/infra-controller-rest/api/pkg/client/site"
	auth "github.com/NVIDIA/infra-controller-rest/auth/pkg/authorization"
	cutil "github.com/NVIDIA/infra-controller-rest/common/pkg/util"
	cdb "github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	"github.com/NVIDIA/infra-controller-rest/db/pkg/db/ipam"
	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
	"github.com/NVIDIA/infra-controller-rest/db/pkg/db/paginator"
	swe "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/error"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
	"github.com/NVIDIA/infra-controller-rest/workflow/pkg/queue"
)

const DefaultReservedIPCount = 2

// ~~~~~ Create Handler ~~~~~ //

// CreateSubnetHandler is the API Handler for creating new Subnet
type CreateSubnetHandler struct {
	dbSession  *cdb.Session
	tc         temporalClient.Client
	scp        *sc.ClientPool
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewCreateSubnetHandler initializes and returns a new handler for creating Subnet
func NewCreateSubnetHandler(dbSession *cdb.Session, tc temporalClient.Client, scp *sc.ClientPool, cfg *config.Config) CreateSubnetHandler {
	return CreateSubnetHandler{
		dbSession:  dbSession,
		tc:         tc,
		scp:        scp,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Create a Subnet
// @Description Create a Subnet
// @Tags Subnet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param message body model.APISubnetCreateRequest true "Subnet creation request"
// @Success 201 {object} model.APISubnet
// @Router /v2/org/{org}/nico/subnet [post]
func (csh CreateSubnetHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Subnet", "Create", c, csh.tracerSpan)
	if handlerSpan != nil {
		defer handlerSpan.End()
	}
	if dbUser == nil {
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve current user", nil)
	}

	// Validate org
	ok, err := auth.ValidateOrgMembership(dbUser, org)
	if !ok {
		if err != nil {
			logger.Error().Err(err).Msg("error validating org membership for User in request")
		} else {
			logger.Warn().Msg("could not validate org membership for user, access denied")
		}
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, fmt.Sprintf("Failed to validate membership for org: %s", org), nil)
	}

	// Validate role, only Tenant Admins are allowed to interact with Subnet endpoints
	ok = auth.ValidateUserRoles(dbUser, org, nil, auth.TenantAdminRole)
	if !ok {
		logger.Warn().Msg("user does not have Tenant Admin role with org, access denied")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User does not have Tenant Admin role with org", nil)
	}

	// Validate request
	// Bind request data to API model
	apiRequest := model.APISubnetCreateRequest{}
	err = c.Bind(&apiRequest)
	if err != nil {
		logger.Warn().Err(err).Msg("error binding request data into API model")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Failed to parse request data, potentially invalid structure", nil)
	}
	// Validate request attributes
	verr := apiRequest.Validate()
	if verr != nil {
		logger.Warn().Err(verr).Msg("error validating Subnet creation request data")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Error validating Subnet creation request data", verr)
	}

	// Validate the tenant for which this Subnet is being created
	tenant, err := common.GetTenantForOrg(ctx, nil, csh.dbSession, org)
	if err != nil {
		logger.Warn().Err(err).Msg("error getting tenant from org")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Error retrieving Tenant from org", nil)
	}
	// Verify vpc in request
	vpc, err := common.GetVpcFromIDString(ctx, nil, apiRequest.VpcID, nil, csh.dbSession)
	if err != nil {
		logger.Warn().Err(err).Msg("error getting vpc in request")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Could not find VPC specified in request", nil)
	}
	if vpc.TenantID != tenant.ID {
		logger.Warn().Msg("tenant in vpc does not belong to tenant in org")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Tenant for VPC in request does not match tenant in org", nil)
	}

	// Verify if vpc is ethernet virtualized
	if vpc.NetworkVirtualizationType != nil && *vpc.NetworkVirtualizationType != cdbm.VpcEthernetVirtualizer {
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("VPC: %v specified in request must have Ethernet network virtualization type in order to create Subnets", vpc.ID), nil)
	}

	// Verify if vpc is ready
	if vpc.ControllerVpcID == nil || vpc.Status != cdbm.VpcStatusReady {
		logger.Warn().Msg(fmt.Sprintf("VPC: %v specified in request data must be in Ready state in order to create Subnet", apiRequest.VpcID))
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "VPC specified in request data must be in Ready state in order to create Subnet", nil)
	}

	// Verify if site is ready
	stDAO := cdbm.NewSiteDAO(csh.dbSession)
	site, err := stDAO.GetByID(ctx, nil, vpc.SiteID, nil, false)
	if err != nil {
		if err == cdb.ErrDoesNotExist {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Could not find Site associated with Subnet", nil)
		}
		logger.Error().Err(err).Msg("error retrieving Site from DB by ID")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Site with ID from VPC", nil)
	}

	if site.Status != cdbm.SiteStatusRegistered {
		logger.Warn().Msg(fmt.Sprintf("The Site: %v where the Subnet is being created must be in Registered state in order to proceed", vpc.SiteID.String()))
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "The Site where the Subnet is being created must be in Registered state in order to proceed", nil)
	}

	// Validate IPBlocks in request
	// NOTE: model validation ensures non-nil IPv4BlockID
	ipv4Block, err := common.GetIPBlockFromIDString(ctx, nil, *apiRequest.IPv4BlockID, csh.dbSession)
	if err != nil {
		logger.Warn().Err(err).Msg("error getting IPv4 IPBlock in request")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Error retrieving ipv4 IPBlock from request", nil)
	}
	// ipv4block is derived, check if it belongs to tenant via an allocation
	if ipv4Block.TenantID == nil || *ipv4Block.TenantID != tenant.ID {
		logger.Warn().Msg("IPv4 IPBlock in request does not belong to tenant")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "ipv4 IPBlock in request does not belong to tenant", nil)
	}
	if vpc.SiteID != ipv4Block.SiteID {
		logger.Warn().Msg("IPv4 Block specified in request and VPC do not belong to the same Site")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "IPv4 Block specified in request and VPC do not belong to the same Site", nil)
	}
	// NOTE: validation ensures that IPv6BlockID will be nil, ie, it is not supported yet
	// when IPv6 is supported, further validations must ensure that the RoutingType of v4 and v6 must match
	routingType := ipv4Block.RoutingType

	// Check for name uniqueness for the tenant, ie, Tenant cannot have another Subnet with same name at the Site
	// TODO consider doing this with an advisory lock for correctness
	sDAO := cdbm.NewSubnetDAO(csh.dbSession)
	sbs, tot, err := sDAO.GetAll(ctx, nil, cdbm.SubnetFilterInput{Names: []string{apiRequest.Name}, SiteIDs: []uuid.UUID{vpc.SiteID}, TenantIDs: []uuid.UUID{tenant.ID}}, paginator.PageInput{}, []string{})
	if err != nil {
		logger.Error().Err(err).Msg("db error checking for name uniqueness of tenant subnet")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to create Subnet due to DB error", nil)
	}
	if tot > 0 {
		logger.Warn().Str("tenantId", tenant.ID.String()).Str("name", apiRequest.Name).Msg("subnet with same name already exists for tenant")
		return cutil.NewAPIErrorResponse(c, http.StatusConflict, "A Subnet with specified name already exists for Tenant at this Site", validation.Errors{
			"id": errors.New(sbs[0].ID.String()),
		})
	}

	// Start a db tx
	tx, err := cdb.BeginTx(ctx, csh.dbSession, &sql.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("unable to start transaction")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Error creating subnet", nil)
	}
	// this variable is used in cleanup actions to indicate if this transaction committed
	txCommitted := false
	defer common.RollbackTx(ctx, tx, &txCommitted)

	// acquire an advisory lock on the parent IP block ID on which there could be contention
	// this lock is released when the transaction commits or rollsback
	err = tx.TryAcquireAdvisoryLock(ctx, cdb.GetAdvisoryLockIDFromString(fmt.Sprintf("%s-%s", tenant.ID.String(), ipv4Block.ID.String())), nil)
	if err != nil {
		// TODO add a retry here
		logger.Error().Err(err).Msg("Failed to acquire advisory lock on ipblock")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Error creating Subnet, detected multiple parallel request on IP Block by Tenant", nil)
	}

	// create an IPAM allocation for the subnet
	// allocate a child prefix in ipam
	ipamStorage := ipam.NewIpamStorage(csh.dbSession.DB, tx.GetBunTx())
	childPrefix, err := ipam.CreateChildIpamEntryForIPBlock(ctx, tx, csh.dbSession, ipamStorage, ipv4Block, apiRequest.PrefixLength)

	if err != nil {
		// printing parent prefix usage to debug the child prefix failure
		parentPrefix, serr := ipamStorage.ReadPrefix(ctx, ipv4Block.Prefix, ipam.GetIpamNamespaceForIPBlock(ctx, ipv4Block.RoutingType, ipv4Block.InfrastructureProviderID.String(), ipv4Block.SiteID.String()))
		if serr == nil {
			logger.Info().Str("IP Block ID", ipv4Block.ID.String()).Str("IP Block Prefix", ipv4Block.Prefix).Msgf("%+v\n", parentPrefix.Usage())
		}

		logger.Warn().Err(err).Msg("failed to create IPAM entry for subnet")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Could not create IPAM entry for Subnet. Details: %s", err.Error()), nil)
	}
	logger.Info().Str("childCidr", childPrefix.Cidr).Msg("created child cidr for subnet")

	// get the prefix and gateway IP addresses
	ipv4Prefix, _, err := ipam.ParseCidrIntoPrefixAndBlockSize(childPrefix.Cidr)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to parse cidr")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Could not create IPAM entry for Subnet. Details: %s", err.Error()), nil)
	}

	ipv4Gateway, err := ipam.GetFirstIPFromCidr(childPrefix.Cidr)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to get first ip in cidr")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Could not create IPAM entry for Subnet. Details: %s", err.Error()), nil)
	}

	// Create Subnet in DB
	subnet, err := sDAO.Create(
		ctx, tx, cdbm.SubnetCreateInput{
			Name:         apiRequest.Name,
			Description:  apiRequest.Description,
			Org:          org,
			SiteID:       site.ID,
			VpcID:        vpc.ID,
			TenantID:     tenant.ID,
			RoutingType:  &routingType,
			IPv4Prefix:   &ipv4Prefix,
			IPv4Gateway:  &ipv4Gateway,
			IPv4BlockID:  &ipv4Block.ID,
			PrefixLength: apiRequest.PrefixLength,
			Status:       cdbm.SubnetStatusPending,
			CreatedBy:    dbUser.ID,
		})
	if err != nil {
		logger.Error().Err(err).Msg("unable to create Subnet record in DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed creating subnet record", nil)
	}

	// Update the controller ID for the subnet.
	// We need this to match the subnet ID.  This was previously handled
	// by the async cloud workflow after successful creation on site.
	subnet, err = sDAO.Update(ctx, tx, cdbm.SubnetUpdateInput{SubnetId: subnet.ID, ControllerNetworkSegmentID: cdb.GetUUIDPtr(subnet.ID)})

	if err != nil {
		logger.Error().Err(err).Msg("unable to update Subnet record controllerNetworkSegmentId")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed updating new subnet record", nil)
	}
	// create the status detail record
	sdDAO := cdbm.NewStatusDetailDAO(csh.dbSession)
	ssd, serr := sdDAO.CreateFromParams(ctx, tx, subnet.ID.String(), *cdb.GetStrPtr(cdbm.SubnetStatusPending),
		cdb.GetStrPtr("received subnet creation request, pending"))
	if serr != nil {
		logger.Error().Err(serr).Msg("error creating Status Detail DB entry")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to create Status Detail for Subnet", nil)
	}
	if ssd == nil {
		logger.Error().Msg("Status Detail DB entry not returned from CreateFromParams")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to get new Status Detail for Subnet", nil)
	}

	// Get the temporal client for the site we are working with.
	stc, err := csh.scp.GetClientByID(subnet.SiteID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve Temporal client for Site")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve client for Site", nil)
	}

	var subnetMTU *int32 = nil
	if subnet.MTU != nil {
		mtu := int32(*subnet.MTU)
		subnetMTU = &mtu
	}

	var subnetDomainID *cwssaws.DomainId
	if subnet.DomainID != nil {
		subnetDomainID = &cwssaws.DomainId{Value: subnet.DomainID.String()}
	}
	prefixes := []*cwssaws.NetworkPrefix{
		{
			Gateway:      subnet.IPv4Gateway,
			ReserveFirst: DefaultReservedIPCount,
			Prefix:       fmt.Sprintf("%s/%d", *subnet.IPv4Prefix, subnet.PrefixLength),
		},
	}

	createSubnetRequest := &cwssaws.NetworkSegmentCreationRequest{
		Id:          &cwssaws.NetworkSegmentId{Value: common.GetSiteNetworkSegmentID(subnet).String()},
		Name:        subnet.Name,
		SubdomainId: subnetDomainID,
		VpcId:       &cwssaws.VpcId{Value: vpc.GetSiteID().String()},
		Mtu:         subnetMTU,
		Prefixes:    prefixes,
	}

	workflowOptions := temporalClient.StartWorkflowOptions{
		ID:                       "subnet-create-" + subnet.ID.String(),
		WorkflowExecutionTimeout: cutil.WorkflowExecutionTimeout,
		TaskQueue:                queue.SiteTaskQueue,
	}

	logger.Info().Msg("triggering Subnet create workflow")

	// Add context deadlines
	ctx, cancel := context.WithTimeout(ctx, cutil.WorkflowContextTimeout)
	defer cancel()

	// Trigger Site workflow
	we, err := stc.ExecuteWorkflow(ctx, workflowOptions, "CreateSubnetV2", createSubnetRequest)

	if err != nil {
		logger.Error().Err(err).Msg("failed to synchronously start Temporal workflow to create Subnet")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed start sync workflow to create Subnet on Site: %s", err), nil)
	}

	wid := we.GetID()
	logger.Info().Str("Workflow ID", wid).Msg("executed synchronous create Subnet workflow")

	// Block until the workflow has completed and returned success/error.
	err = we.Get(ctx, nil)
	if err != nil {
		var timeoutErr *tp.TimeoutError
		if errors.As(err, &timeoutErr) || err == context.DeadlineExceeded || ctx.Err() != nil {

			logger.Error().Err(err).Msg("failed to create Subnet, timeout occurred executing workflow on Site.")

			// Create a new context deadlines
			newctx, newcancel := context.WithTimeout(context.Background(), cutil.WorkflowContextNewAfterTimeout)
			defer newcancel()

			// Initiate termination workflow
			serr := stc.TerminateWorkflow(newctx, wid, "", "timeout occurred executing create Subnet workflow")
			if serr != nil {
				logger.Error().Err(serr).Msg("failed to execute terminate Temporal workflow for creating Subnet")
				return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to terminate synchronous Subnet creation workflow after timeout, Cloud and Site data may be de-synced: %s", serr), nil)
			}

			logger.Info().Str("Workflow ID", wid).Msg("initiated terminate synchronous create Subnet workflow successfully")

			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create Subnet, timeout occurred executing workflow on Site: %s", err), nil)
		}

		code, err := common.UnwrapWorkflowError(err)
		logger.Error().Err(err).Msg("failed to synchronously execute Temporal workflow to create Subnet")
		return cutil.NewAPIErrorResponse(c, code, fmt.Sprintf("Failed to execute sync workflow to create Subnet on Site: %s", err), nil)
	}

	logger.Info().Str("Workflow ID", wid).Msg("completed synchronous create Subnet workflow")

	// commit transaction
	err = tx.Commit()
	if err != nil {
		logger.Error().Err(err).Msg("error committing transaction")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to create Subnet, DB transaction error", nil)
	}
	// set committed so, deferred cleanup functions will do nothing
	txCommitted = true

	// create response
	apiInstance := model.NewAPISubnet(subnet, []cdbm.StatusDetail{*ssd})
	logger.Info().Msg("finishing API handler")
	return c.JSON(http.StatusCreated, apiInstance)
}

// ~~~~~ GetAll Handler ~~~~~ //

// GetAllSubnetHandler is the API Handler for getting all Subnets
type GetAllSubnetHandler struct {
	dbSession  *cdb.Session
	tc         temporalClient.Client
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewGetAllSubnetHandler initializes and returns a new handler for getting all Subnets
func NewGetAllSubnetHandler(dbSession *cdb.Session, tc temporalClient.Client, cfg *config.Config) GetAllSubnetHandler {
	return GetAllSubnetHandler{
		dbSession:  dbSession,
		tc:         tc,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Get all Subnets
// @Description Get all Subnets
// @Tags Subnet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param siteId query string false "Site ID"
// @Param vpcId query string true "ID of Vpc"
// @Param status query string false "Filter by status" e.g. 'Pending', 'Error'"
// @Param query query string false "Query input for full text search"
// @Param includeRelation query string false "Related entities to include in response e.g. 'Site', 'Vpc', 'Tenant', 'IPv4Block', 'IPv6Block'"
// @Param pageNumber query integer false "Page number of results returned"
// @Param pageSize query integer false "Number of results per page"
// @Param orderBy query string false "Order by field"
// @Success 200 {object} []model.APISubnet
// @Router /v2/org/{org}/nico/subnet [get]
func (gash GetAllSubnetHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Subnet", "GetAll", c, gash.tracerSpan)
	if handlerSpan != nil {
		defer handlerSpan.End()
	}
	if dbUser == nil {
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve current user", nil)
	}

	// Validate org
	ok, err := auth.ValidateOrgMembership(dbUser, org)
	if !ok {
		if err != nil {
			logger.Error().Err(err).Msg("error validating org membership for User in request")
		} else {
			logger.Warn().Msg("could not validate org membership for user, access denied")
		}
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, fmt.Sprintf("Failed to validate membership for org: %s", org), nil)
	}

	// Validate role, only Tenant Admins are allowed to interact with Subnet endpoints
	ok = auth.ValidateUserRoles(dbUser, org, nil, auth.TenantAdminRole)
	if !ok {
		logger.Warn().Msg("user does not have Tenant Admin role with org, access denied")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User does not have Tenant Admin role with org", nil)
	}

	// Validate paginantion request
	pageRequest := pagination.PageRequest{}
	err = c.Bind(&pageRequest)
	if err != nil {
		logger.Warn().Err(err).Msg("error binding pagination request data into API model")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Failed to parse request pagination data", nil)
	}

	// Validate request attributes
	err = pageRequest.Validate(cdbm.SubnetOrderByFields)
	if err != nil {
		logger.Warn().Err(err).Msg("error validating pagination request data")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Failed to validate pagination request data", err)
	}
	subnetFilter := cdbm.SubnetFilterInput{}

	// Validate the tenant for which this Subnet is being created
	tenant, err := common.GetTenantForOrg(ctx, nil, gash.dbSession, org)
	if err != nil {
		logger.Warn().Err(err).Msg("error getting tenant from org")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Error retrieving Tenant from org", nil)
	}
	subnetFilter.TenantIDs = []uuid.UUID{tenant.ID}

	// Get and validate includeRelation params
	qParams := c.QueryParams()
	qIncludeRelations, errMsg := common.GetAndValidateQueryRelations(qParams, cdbm.SubnetRelatedEntities)
	if errMsg != "" {
		logger.Warn().Msg(errMsg)
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, errMsg, nil)
	}

	// Get site ID from query param
	tsDAO := cdbm.NewTenantSiteDAO(gash.dbSession)
	var siteID *uuid.UUID
	siteIDStr := c.QueryParam("siteId")
	if siteIDStr != "" {
		site, err := common.GetSiteFromIDString(ctx, nil, siteIDStr, gash.dbSession)
		if err != nil {
			logger.Warn().Err(err).Msg("error getting site in request")
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Failed to retrieve Site specified in query param, invalid ID or DB error", nil)
		}
		siteID = &site.ID

		// Check Site association with Tenant
		_, err = tsDAO.GetByTenantIDAndSiteID(ctx, nil, tenant.ID, site.ID, nil)
		if err != nil {
			if err == cdb.ErrDoesNotExist {
				return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "Tenant does not have access to this Site", nil)
			}
			logger.Error().Err(err).Msg("error retrieving TenantSite from DB")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to determine Tenant access to Site, DB error", nil)
		}
		subnetFilter.SiteIDs = []uuid.UUID{*siteID}
	}

	// verify vpc if specified in query string
	qVpcID := c.QueryParam("vpcId")
	var vpcID *uuid.UUID
	if qVpcID != "" {
		vpc, err := common.GetVpcFromIDString(ctx, nil, qVpcID, nil, gash.dbSession)
		if err != nil {
			logger.Warn().Err(err).Msg("error getting vpc in request")
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, "Could not find VPC specified in request", nil)
		}
		if vpc.TenantID != tenant.ID {
			logger.Warn().Msg("tenant in vpc does not belong to tenant in org")
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Tenant for VPC in request does not match tenant in org", nil)
		}
		vpcID = &vpc.ID
		subnetFilter.VpcIDs = []uuid.UUID{*vpcID}
	}

	// Get query text for full text search from query param
	searchQuery := common.GetSearchQuery(c)
	if searchQuery != nil {
		gash.tracerSpan.SetAttribute(handlerSpan, attribute.String("query", *searchQuery), logger)
		subnetFilter.SearchQuery = searchQuery
	}

	// Get status from query param
	var status *string

	statusQuery := c.QueryParam("status")
	if statusQuery != "" {
		gash.tracerSpan.SetAttribute(handlerSpan, attribute.String("status", statusQuery), logger)
		_, ok := cdbm.SubnetStatusMap[statusQuery]
		if !ok {
			logger.Warn().Msg(fmt.Sprintf("invalid value in status query: %v", statusQuery))
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid Status value in query", nil)
		}
		status = &statusQuery
		subnetFilter.Statuses = []string{*status}
	}

	// Create response
	sDAO := cdbm.NewSubnetDAO(gash.dbSession)
	subnets, total, err := sDAO.GetAll(ctx, nil, subnetFilter, paginator.PageInput{
		Limit:   pageRequest.Limit,
		Offset:  pageRequest.Offset,
		OrderBy: pageRequest.OrderBy,
	}, qIncludeRelations)
	if err != nil {
		logger.Error().Err(err).Msg("error getting subnets from db")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Subnets", nil)
	}

	// Get status details
	sdDAO := cdbm.NewStatusDetailDAO(gash.dbSession)

	sdEntityIDs := []string{}
	for _, sn := range subnets {
		sdEntityIDs = append(sdEntityIDs, sn.ID.String())
	}
	ssds, serr := sdDAO.GetRecentByEntityIDs(ctx, nil, sdEntityIDs, common.RECENT_STATUS_DETAIL_COUNT)
	if serr != nil {
		logger.Warn().Err(serr).Msg("error retrieving Status Details for Subnets from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to populate status history for Subnets", nil)
	}
	ssdMap := map[string][]cdbm.StatusDetail{}
	for _, ssd := range ssds {
		cssd := ssd
		ssdMap[ssd.EntityID] = append(ssdMap[ssd.EntityID], cssd)
	}

	// Create response
	apiSubnets := []*model.APISubnet{}

	// get status details
	for _, sn := range subnets {
		cursn := sn
		apiSubnet := model.NewAPISubnet(&cursn, ssdMap[sn.ID.String()])
		apiSubnets = append(apiSubnets, apiSubnet)
	}

	// Create pagination response header
	pageReponse := pagination.NewPageResponse(*pageRequest.PageNumber, *pageRequest.PageSize, total, pageRequest.OrderByStr)
	pageHeader, err := json.Marshal(pageReponse)
	if err != nil {
		logger.Error().Err(err).Msg("error marshaling pagination response")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to generate pagination response header", nil)
	}

	c.Response().Header().Set(pagination.ResponseHeaderName, string(pageHeader))

	logger.Info().Msg("finishing API handler")

	return c.JSON(http.StatusOK, apiSubnets)
}

// ~~~~~ Get Handler ~~~~~ //

// GetSubnetHandler is the API Handler for retrieving Subnet
type GetSubnetHandler struct {
	dbSession  *cdb.Session
	tc         temporalClient.Client
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewGetSubnetHandler initializes and returns a new handler to retrieve Subnet
func NewGetSubnetHandler(dbSession *cdb.Session, tc temporalClient.Client, cfg *config.Config) GetSubnetHandler {
	return GetSubnetHandler{
		dbSession:  dbSession,
		tc:         tc,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Retrieve the Subnet
// @Description Retrieve the Subnet
// @Tags Subnet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param id path string true "ID of Subnet"
// @Param includeRelation query string false "Related entities to include in response e.g. 'Site', 'Vpc', 'Tenant', 'IPv4Block', 'IPv6Block'"
// @Success 200 {object} model.APISubnet
// @Router /v2/org/{org}/nico/subnet/{id} [get]
func (gsh GetSubnetHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Subnet", "Get", c, gsh.tracerSpan)
	if handlerSpan != nil {
		defer handlerSpan.End()
	}
	if dbUser == nil {
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve current user", nil)
	}

	// Validate org
	ok, err := auth.ValidateOrgMembership(dbUser, org)
	if !ok {
		if err != nil {
			logger.Error().Err(err).Msg("error validating org membership for User in request")
		} else {
			logger.Warn().Msg("could not validate org membership for user, access denied")
		}
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, fmt.Sprintf("Failed to validate membership for org: %s", org), nil)
	}

	// Validate role, only Tenant Admins are allowed to interact with Subnet endpoints
	ok = auth.ValidateUserRoles(dbUser, org, nil, auth.TenantAdminRole)
	if !ok {
		logger.Warn().Msg("user does not have Tenant Admin role with org, access denied")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User does not have Tenant Admin role with org", nil)
	}

	// Get and validate includeRelation params
	qParams := c.QueryParams()
	qIncludeRelations, errMsg := common.GetAndValidateQueryRelations(qParams, cdbm.SubnetRelatedEntities)
	if errMsg != "" {
		logger.Warn().Msg(errMsg)
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, errMsg, nil)
	}

	// Get subnet ID from URL param
	sStrID := c.Param("id")

	gsh.tracerSpan.SetAttribute(handlerSpan, attribute.String("subnet_id", sStrID), logger)

	sID, err := uuid.Parse(sStrID)
	if err != nil {
		logger.Warn().Err(err).Msg("error parsing id in url into uuid")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid Subnet ID in URL", nil)
	}

	sDAO := cdbm.NewSubnetDAO(gsh.dbSession)

	// Validate the tenant for which this Subnet is being updated
	tenant, err := common.GetTenantForOrg(ctx, nil, gsh.dbSession, org)
	if err != nil {
		logger.Warn().Err(err).Msg("error getting tenant from org")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Error retrieving Tenant from org", nil)
	}

	// Check that subnet exists
	subnet, err := sDAO.GetByID(ctx, nil, sID, qIncludeRelations)
	if err != nil {
		logger.Error().Err(err).Msg("error retrieving Subnet DB entity")
		if err == cdb.ErrDoesNotExist {
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, "Could not retrieve Subnet to update", nil)
		}
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Could not retrieve Subnet to update", nil)
	}

	// verify tenant matches
	if tenant.ID != subnet.TenantID {
		logger.Warn().Msg("tenant in subnet does not belong to tenant in org")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Tenant for subnet in request does not match tenant in org", nil)
	}

	// get status details for the response
	sdDAO := cdbm.NewStatusDetailDAO(gsh.dbSession)
	ssds, err := sdDAO.GetRecentByEntityIDs(ctx, nil, []string{subnet.ID.String()}, common.RECENT_STATUS_DETAIL_COUNT)
	if err != nil {
		logger.Error().Err(err).Msg("error retrieving Status Details for subnet from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Status Details for subnet", nil)
	}

	// Send response
	apiInstance := model.NewAPISubnet(subnet, ssds)
	logger.Info().Msg("finishing API handler")
	return c.JSON(http.StatusOK, apiInstance)
}

// ~~~~~ Update Handler ~~~~~ //

// UpdateSubnetHandler is the API Handler for updating a Subnet
type UpdateSubnetHandler struct {
	dbSession  *cdb.Session
	tc         temporalClient.Client
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewUpdateSubnetHandler initializes and returns a new handler for updating Subnet
func NewUpdateSubnetHandler(dbSession *cdb.Session, tc temporalClient.Client, cfg *config.Config) UpdateSubnetHandler {
	return UpdateSubnetHandler{
		dbSession:  dbSession,
		tc:         tc,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Update an existing Subnet
// @Description Update an existing Subnet
// @Tags Subnet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param id path string true "ID of Subnet"
// @Param message body model.APISubnetUpdateRequest true "Subnet update request"
// @Success 200 {object} model.APISubnet
// @Router /v2/org/{org}/nico/subnet/{id} [patch]
func (ush UpdateSubnetHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Subnet", "Update", c, ush.tracerSpan)
	if handlerSpan != nil {
		defer handlerSpan.End()
	}
	if dbUser == nil {
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve current user", nil)
	}

	// Validate org
	ok, err := auth.ValidateOrgMembership(dbUser, org)
	if !ok {
		if err != nil {
			logger.Error().Err(err).Msg("error validating org membership for User in request")
		} else {
			logger.Warn().Msg("could not validate org membership for user, access denied")
		}
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, fmt.Sprintf("Failed to validate membership for org: %s", org), nil)
	}

	// Validate role, only Tenant Admins are allowed to interact with Subnet endpoints
	ok = auth.ValidateUserRoles(dbUser, org, nil, auth.TenantAdminRole)
	if !ok {
		logger.Warn().Msg("user does not have Tenant Admin role with org, access denied")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User does not have Tenant Admin role with org", nil)
	}

	// Get subnet ID from URL param
	sStrID := c.Param("id")

	ush.tracerSpan.SetAttribute(handlerSpan, attribute.String("subnet_id", sStrID), logger)

	sID, err := uuid.Parse(sStrID)
	if err != nil {
		logger.Warn().Err(err).Msg("error parsing id in url into uuid")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid Subnet ID in URL", nil)
	}

	sDAO := cdbm.NewSubnetDAO(ush.dbSession)

	// Validate request
	// Bind request data to API model
	apiRequest := model.APISubnetUpdateRequest{}
	err = c.Bind(&apiRequest)
	if err != nil {
		logger.Warn().Err(err).Msg("error binding request data into API model")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Failed to parse request data, potentially invalid structure", nil)
	}
	// Validate request attributes
	verr := apiRequest.Validate()
	if verr != nil {
		logger.Warn().Err(verr).Msg("error validating Subnet update request data")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Error validating Subnet update request data", verr)
	}

	// Validate the tenant for which this Subnet is being updated
	tenant, err := common.GetTenantForOrg(ctx, nil, ush.dbSession, org)
	if err != nil {
		logger.Warn().Err(err).Msg("error getting tenant from org")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Error retrieving Tenant from org", nil)
	}

	// Check that subnet exists
	subnet, err := sDAO.GetByID(ctx, nil, sID, nil)
	if err != nil {
		logger.Error().Err(err).Msg("error retrieving Subnet DB entity")
		if err == cdb.ErrDoesNotExist {
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, "Could not retrieve Subnet to update", nil)
		}
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Could not retrieve Subnet to update", nil)
	}

	// verify tenant matches
	if tenant.ID != subnet.TenantID {
		logger.Warn().Msg("tenant in subnet does not belong to tenant in org")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Tenant for subnet in request does not match tenant in org", nil)
	}

	if apiRequest.Name != nil && *apiRequest.Name != subnet.Name {
		sbs, tot, serr := sDAO.GetAll(ctx, nil, cdbm.SubnetFilterInput{Names: []string{*apiRequest.Name}, SiteIDs: []uuid.UUID{subnet.SiteID}, TenantIDs: []uuid.UUID{tenant.ID}}, paginator.PageInput{}, []string{})
		if serr != nil {
			logger.Error().Err(serr).Msg("db error checking for name uniqueness of tenant subnet")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to create Subnet due to DB error", nil)
		}
		if tot > 0 {
			return cutil.NewAPIErrorResponse(c, http.StatusConflict, "Another Subnet with specified name already exists for Tenant", validation.Errors{
				"id": errors.New(sbs[0].ID.String()),
			})
		}
	}

	// start a database transaction
	tx, err := cdb.BeginTx(ctx, ush.dbSession, &sql.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("error updating subnet in DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to update Subnet", nil)
	}
	txCommitted := false
	defer common.RollbackTx(ctx, tx, &txCommitted)

	subnet, err = sDAO.Update(ctx, tx, cdbm.SubnetUpdateInput{SubnetId: sID, Name: apiRequest.Name, Description: apiRequest.Description})
	if err != nil {
		logger.Error().Err(err).Msg("error updating Subnet in DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to update Subnet", nil)
	}
	// get status details for the response
	sdDAO := cdbm.NewStatusDetailDAO(ush.dbSession)
	ssds, _, err := sdDAO.GetAllByEntityID(ctx, tx, subnet.ID.String(), nil, cdb.GetIntPtr(pagination.MaxPageSize), nil)
	if err != nil {
		logger.Error().Err(err).Msg("error retrieving Status Details for subnet from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Status Details for subnet", nil)
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		logger.Error().Err(err).Msg("error updating Subnet in DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to update Subnet", nil)
	}
	txCommitted = true

	// Send response
	apiInstance := model.NewAPISubnet(subnet, ssds)
	logger.Info().Msg("finishing API handler")
	return c.JSON(http.StatusOK, apiInstance)
}

// ~~~~~ Delete Handler ~~~~~ //

// DeleteSubnetHandler is the API Handler for deleting a Subnet
type DeleteSubnetHandler struct {
	dbSession  *cdb.Session
	tc         temporalClient.Client
	scp        *sc.ClientPool
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewDeleteSubnetHandler initializes and returns a new handler for deleting Subnet
func NewDeleteSubnetHandler(dbSession *cdb.Session, tc temporalClient.Client, scp *sc.ClientPool, cfg *config.Config) DeleteSubnetHandler {
	return DeleteSubnetHandler{
		dbSession:  dbSession,
		tc:         tc,
		scp:        scp,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Delete an existing Subnet
// @Description Delete an existing Subnet
// @Tags Subnet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param id path string true "ID of Subnet"
// @Success 202
// @Router /v2/org/{org}/nico/subnet/{id} [delete]
func (dsh DeleteSubnetHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Subnet", "Delete", c, dsh.tracerSpan)
	if handlerSpan != nil {
		defer handlerSpan.End()
	}
	if dbUser == nil {
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve current user", nil)
	}

	// Validate org
	ok, err := auth.ValidateOrgMembership(dbUser, org)
	if !ok {
		if err != nil {
			logger.Error().Err(err).Msg("error validating org membership for User in request")
		} else {
			logger.Warn().Msg("could not validate org membership for user, access denied")
		}
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, fmt.Sprintf("Failed to validate membership for org: %s", org), nil)
	}

	// Validate role, only Tenant Admins are allowed to interact with Subnet endpoints
	ok = auth.ValidateUserRoles(dbUser, org, nil, auth.TenantAdminRole)
	if !ok {
		logger.Warn().Msg("user does not have Tenant Admin role with org, access denied")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User does not have Tenant Admin role with org", nil)
	}

	// Get subnet ID from URL param
	sStrID := c.Param("id")

	dsh.tracerSpan.SetAttribute(handlerSpan, attribute.String("subnet_id", sStrID), logger)

	sID, err := uuid.Parse(sStrID)
	if err != nil {
		logger.Warn().Err(err).Msg("error parsing id in url into uuid")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid Subnet ID in URL", nil)
	}

	// Check that subnet exists
	sDAO := cdbm.NewSubnetDAO(dsh.dbSession)
	subnet, err := sDAO.GetByID(ctx, nil, sID, []string{"IPv4Block", cdbm.TenantRelationName, cdbm.SiteRelationName})
	if err != nil {
		logger.Error().Err(err).Msg("error retrieving Subnet DB entity")
		if err == cdb.ErrDoesNotExist {
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, "Could not find Subnet to delete", nil)
		}
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed not retrieve Subnet for deletion, DB error", nil)
	}

	if subnet.Tenant == nil {
		logger.Warn().Err(err).Msg("failed to retrieve Tenant details")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Tenant details", nil)
	}

	// Validate the tenant for which this Subnet is being updated
	if subnet.Tenant.Org != org {
		logger.Warn().Msg("org specified in request does not match org of Tenant associated with Subnet")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Org specified in request does not match org of Tenant associated with Subnet", nil)
	}

	// Verify that the Subnet is associated with a site and then that the site is
	// in a valid state.
	if subnet.Site == nil {
		logger.Error().Msg("failed to pull site data for Subnet")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Site details for Subnet", nil)
	}

	// Verify if site is ready
	if subnet.Site.Status != cdbm.SiteStatusRegistered {
		logger.Warn().Str("Site ID", subnet.SiteID.String()).Msg("Site associated with Subnet must be in Registered state in order to proceed")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Site associated with Subnet must be in Registered state in order to proceed", nil)
	}

	// Verify no instances are using the Subnet
	isDAO := cdbm.NewInterfaceDAO(dsh.dbSession)

	filterInput := cdbm.InterfaceFilterInput{
		SubnetID: &subnet.ID,
	}

	pageInput := paginator.PageInput{
		Offset:  nil,
		Limit:   cdb.GetIntPtr(0),
		OrderBy: nil,
	}

	_, ifcCount, err := isDAO.GetAll(ctx, nil, filterInput, pageInput, nil)
	if err != nil {
		logger.Error().Err(err).Msg("error retrieving Interfaces for Subnet from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Interfaces for Subnet, DB error", nil)
	}

	if ifcCount > 0 {
		logger.Warn().Msg("Interfaces exist for Subnet, cannot delete it")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Subnet is being used by one or more Instances and cannot be deleted", nil)
	}

	// Start a db tx
	tx, err := cdb.BeginTx(ctx, dsh.dbSession, &sql.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("unable to start transaction")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to delete Subnet, DB transaction error", nil)
	}
	// this variable is used in cleanup actions to indicate if this transaction committed
	txCommitted := false
	defer common.RollbackTx(ctx, tx, &txCommitted)

	// acquire an advisory lock on the parent IP block ID on which there could be contention
	// this lock is released when the transaction commits or rollsback
	err = tx.TryAcquireAdvisoryLock(ctx, cdb.GetAdvisoryLockIDFromString(subnet.IPv4BlockID.String()), nil)
	if err != nil {
		// TODO: Add a retry here
		logger.Error().Err(err).Msg("Failed to acquire advisory lock on ipblock")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to delete Subnet, DB lock error", nil)
	}

	// Set Subnet status to Deleting
	status := cdbm.SubnetStatusDeleting
	statusMsg := "Subnet deletion successfully initiated on Site"
	_, err = sDAO.Update(ctx, tx, cdbm.SubnetUpdateInput{SubnetId: subnet.ID, Status: &status})
	if err != nil {
		logger.Error().Err(err).Msg("error setting Subnet status to deleting")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to update Subnet status, DB error", nil)
	}

	sdDAO := cdbm.NewStatusDetailDAO(dsh.dbSession)
	_, err = sdDAO.CreateFromParams(ctx, tx, subnet.ID.String(), status, &statusMsg)
	if err != nil {
		logger.Error().Err(err).Msg("error creating Status Detail for Subnet")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to create Subnet status detail, DB error", nil)
	}

	// Get the temporal client for the site we are working with.
	stc, err := dsh.scp.GetClientByID(subnet.SiteID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve Temporal client for Site")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve client for Site", nil)
	}

	// Prepare the delete/release request workflow object
	deleteSubnetRequest := &cwssaws.NetworkSegmentDeletionRequest{
		Id: &cwssaws.NetworkSegmentId{Value: common.GetSiteNetworkSegmentID(subnet).String()},
	}

	workflowOptions := temporalClient.StartWorkflowOptions{
		ID:        "subnet-delete-" + subnet.ID.String(),
		TaskQueue: queue.SiteTaskQueue,
	}

	logger.Info().Msg("triggering Subnet delete workflow")

	// Trigger Site workflow to delete subnet Subnet
	// TODO: Once Site Agent offers DeleteSubnetV2 re-registered as SubnetSubnet then update workflow name here
	we, err := stc.ExecuteWorkflow(ctx, workflowOptions, "DeleteSubnetV2", deleteSubnetRequest)
	if err != nil {
		logger.Error().Err(err).Msg("failed to synchronously start Temporal workflow to delete Subnet")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to start sync workflow to delete Subnet on Site: %s", err), nil)
	}

	wid := we.GetID()
	logger.Info().Str("Workflow ID", wid).Msg("executed synchronous delete Subnet workflow")

	// Execute the workflow synchronously
	err = we.Get(ctx, nil)
	// Handle skippable errors
	if err != nil {
		// If this was a 404 back from NICo, we can treat the object as already having been deleted and allow things to proceed.
		var applicationErr *tp.ApplicationError
		if errors.As(err, &applicationErr) && slices.Contains(swe.ObjectNotFoundErrTypes(), applicationErr.Type()) {
			logger.Warn().Msg(swe.ErrTypeNICoObjectNotFound + " received from Site")
			// Reset error to nil
			err = nil
		}
	}

	// Check if err is still nil now that we've handled any skippable errors.
	if err != nil {
		var timeoutErr *tp.TimeoutError
		if errors.As(err, &timeoutErr) || ctx.Err() != nil {

			logger.Error().Err(err).Msg("failed to delete Subnet, timeout occurred executing workflow on Site.")

			// Create a new context deadlines
			newctx, newcancel := context.WithTimeout(context.Background(), cutil.WorkflowContextNewAfterTimeout)
			defer newcancel()

			// Initiate termination workflow
			serr := stc.TerminateWorkflow(newctx, wid, "", "timeout occurred executing delete Subnet workflow")
			if serr != nil {
				logger.Error().Err(serr).Msg("failed to terminate Temporal workflow for deleting Subnet")
				return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to terminate synchronous Subnet deletion workflow after timeout, Cloud and Site data may be de-synced: %s", serr), nil)
			}

			logger.Info().Str("Workflow ID", wid).Msg("initiated terminate synchronous delete Subnet workflow successfully")

			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to delete Subnet, timeout occurred executing workflow on Site: %s", err), nil)
		}

		code, err := common.UnwrapWorkflowError(err)
		logger.Error().Err(err).Msg("failed to synchronously execute Temporal workflow to delete Subnet")
		return cutil.NewAPIErrorResponse(c, code, fmt.Sprintf("Failed to execute sync workflow to delete Subnet on Site: %s", err), nil)
	}

	logger.Info().Str("Workflow ID", wid).Msg("completed synchronous delete Subnet workflow")

	// Commit the DB transaction after the synchronous workflow has completed without error
	err = tx.Commit()
	if err != nil {
		logger.Error().Err(err).Msg("error committing Subnet transaction to DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to delete Subnet, DB transaction error", nil)
	}

	// Set committed so, deferred cleanup functions will do nothing
	txCommitted = true

	// Create response
	logger.Info().Msg("finishing API handler")
	return c.String(http.StatusAccepted, "Deletion request was accepted")
}
