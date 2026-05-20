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
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"go.opentelemetry.io/otel/attribute"
	temporalClient "go.temporal.io/sdk/client"
	tp "go.temporal.io/sdk/temporal"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/labstack/echo/v4"

	cdb "github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
	cdbp "github.com/NVIDIA/infra-controller-rest/db/pkg/db/paginator"
	swe "github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/error"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"

	"github.com/NVIDIA/infra-controller-rest/workflow/pkg/queue"

	"github.com/NVIDIA/infra-controller-rest/api/internal/config"
	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/handler/util/common"
	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/model"
	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/model/util"
	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/pagination"
	auth "github.com/NVIDIA/infra-controller-rest/auth/pkg/authorization"
	cutil "github.com/NVIDIA/infra-controller-rest/common/pkg/util"

	sc "github.com/NVIDIA/infra-controller-rest/api/pkg/client/site"
)

const MachineMissingDelayThreshold = 24 * time.Hour

// ~~~~~ Utility for Gets ~~~~~ //
func getAPIMachines(ctx context.Context, ms []cdbm.Machine, logger zerolog.Logger, tx *cdb.Tx, dbSession *cdb.Session, includeMetadata bool, isProviderOrPrivilegedTenant bool) ([]*model.APIMachine, *cutil.APIError) {
	// Get status details
	sdDAO := cdbm.NewStatusDetailDAO(dbSession)

	sdEntityIDs := []string{}
	for _, m := range ms {
		sdEntityIDs = append(sdEntityIDs, m.ID)
	}
	ssds, serr := sdDAO.GetRecentByEntityIDs(ctx, tx, sdEntityIDs, common.RECENT_STATUS_DETAIL_COUNT)
	if serr != nil {
		logger.Error().Err(serr).Msg("error retrieving Status Details for Machines from DB")
		return nil, cutil.NewAPIError(http.StatusInternalServerError, "Failed to retrieve Status Details for Machines, DB error", nil)
	}
	ssdMap := map[string][]cdbm.StatusDetail{}
	for _, ssd := range ssds {
		cssd := ssd
		ssdMap[ssd.EntityID] = append(ssdMap[ssd.EntityID], cssd)
	}

	// Get Instance details
	instanceDAO := cdbm.NewInstanceDAO(dbSession)

	mids := []string{}
	for _, m := range ms {
		mids = append(mids, m.ID)
	}
	instances, _, serr := instanceDAO.GetAll(ctx, tx, cdbm.InstanceFilterInput{MachineIDs: mids}, cdbp.PageInput{Limit: cdb.GetIntPtr(cdbp.TotalLimit)}, []string{cdbm.TenantRelationName})
	if serr != nil {
		logger.Error().Err(serr).Msg("error retrieving Instances for Machines")
		return nil, cutil.NewAPIError(http.StatusInternalServerError, "Failed to retrieve Instances for Machines, DB error", nil)
	}
	insMap := map[string]*cdbm.Instance{}
	for _, ins := range instances {
		curIns := ins
		insMap[*ins.MachineID] = &curIns
	}

	// Create response
	apiMs := make([]*model.APIMachine, 0, len(ms))

	mcDAO := cdbm.NewMachineCapabilityDAO(dbSession)
	miDAO := cdbm.NewMachineInterfaceDAO(dbSession)

	mcs, _, err := mcDAO.GetAll(ctx, tx, mids, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, cdb.GetIntPtr(cdbp.TotalLimit), nil)
	if err != nil {
		// Continue in spite of the error
		logger.Error().Err(err).Msg("error retrieving Machine Capabilities for Machine from DB")
		return nil, cutil.NewAPIError(http.StatusInternalServerError, "Failed to retrieve Capabilities for Machines, DB error", nil)
	}
	midToMCMap := map[string][]cdbm.MachineCapability{}

	for _, mc := range mcs {
		tmpmc := mc
		midToMCMap[*mc.MachineID] = append(midToMCMap[*mc.MachineID], tmpmc)
	}

	mis, _, err := miDAO.GetAll(
		ctx,
		tx,
		cdbm.MachineInterfaceFilterInput{
			MachineIDs: mids,
		},
		cdbp.PageInput{Limit: cdb.GetIntPtr(cdbp.TotalLimit)},
		nil,
	)
	if err != nil {
		// Continue in spite of the error
		logger.Error().Err(err).Msg("error retrieving Machine Interfaces for Machine from DB")
		return nil, cutil.NewAPIError(http.StatusInternalServerError, "Failed to retrieve Interfaces for Machines, DB error", nil)
	}

	midToMIMap := map[string][]cdbm.MachineInterface{}
	for _, mi := range mis {
		tmpmi := mi
		midToMIMap[mi.MachineID] = append(midToMIMap[mi.MachineID], tmpmi)
	}

	// Get machine capability, machine interface, and status details
	for _, m := range ms {
		tmpm := m
		apim := model.NewAPIMachine(&tmpm, midToMCMap[m.ID], midToMIMap[m.ID], ssdMap[m.ID], insMap[m.ID], includeMetadata, isProviderOrPrivilegedTenant)
		apiMs = append(apiMs, apim)
	}
	return apiMs, nil
}

// isProviderOrTenant returns the Infrastructure Provider and Tenant for the org if the user is a Provider Admin or Tenant Admin
func isProviderOrTenant(ctx context.Context, logger zerolog.Logger, dbSession *cdb.Session, org string, userOrgDetails *cdbm.Org) (*cdbm.InfrastructureProvider, *cdbm.Tenant, *cutil.APIError) {
	isProvider := auth.ValidateUserRolesInOrg(*userOrgDetails, nil, auth.ProviderAdminRole, auth.ProviderViewerRole)

	var infrastructureProvider *cdbm.InfrastructureProvider
	var tenant *cdbm.Tenant
	var err error

	if isProvider {
		infrastructureProvider, err = common.GetInfrastructureProviderForOrg(ctx, nil, dbSession, org)
		if err != nil {
			if errors.Is(err, common.ErrOrgInstrastructureProviderNotFound) {
				return nil, nil, cutil.NewAPIError(http.StatusNotFound, "Infrastructure Provider not found in org", nil)
			}
			logger.Error().Err(err).Msg("error getting infrastructure provider for org")
			return nil, nil, cutil.NewAPIError(http.StatusInternalServerError, "Failed to retrieve infrastructure provider for org, DB error", nil)
		}
	} else {
		isTenant := auth.ValidateUserRolesInOrg(*userOrgDetails, nil, auth.TenantAdminRole)
		if !isTenant {
			logger.Warn().Msg("user does not have required role, access denied")
			return nil, nil, cutil.NewAPIError(http.StatusForbidden, "User doesn't have Provider Admin or Tenant Admin role", nil)
		}

		// Get Tenant for org
		tenant, err = common.GetTenantForOrg(ctx, nil, dbSession, org)
		if err != nil {
			if errors.Is(err, common.ErrOrgTenantNotFound) {
				logger.Warn().Msg("organization doesn't have a Tenant associated, access denied")
				return nil, nil, cutil.NewAPIError(http.StatusForbidden, "Organization doesn't have a Provider or Tenant associated", nil)
			}
			logger.Error().Err(err).Msg("error retrieving Tenant for Organization")
			return nil, nil, cutil.NewAPIError(http.StatusInternalServerError, "Failed to retrieve Tenant for Organization, DB error", nil)
		}

		// Check Tenant config
		if tenant.Config == nil {
			logger.Warn().Msg("unexpected empty Tenant configuration in DB, defaulting to empty config")
			tenant.Config = &cdbm.TenantConfig{}
		}
	}

	return infrastructureProvider, tenant, nil
}

// ~~~~~ GetAll Handler ~~~~~ //

// GetAllMachineHandler is the API Handler for getting all Machines
type GetAllMachineHandler struct {
	dbSession  *cdb.Session
	tc         temporalClient.Client
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewGetAllMachineHandler initializes and returns a new handler for getting all Machines
func NewGetAllMachineHandler(dbSession *cdb.Session, tc temporalClient.Client, cfg *config.Config) GetAllMachineHandler {
	return GetAllMachineHandler{
		dbSession:  dbSession,
		tc:         tc,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Get all Machines
// @Description Get all Machines
// @Tags Machine
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param siteId query string true "ID of Site"
// @Param id query string true "ID of Machine"
// @Param hasInstance query boolean false "Filter Machines by whether machine is assigned to an Instance."
// @Param hasInstanceType query boolean false "Filter by assigned an InstanceType to include in response"
// @Param instanceTypeId query string true "Filter by InstanceType ID"
// @Param tenantId query string false "Filter by Tenant ID"
// @Param capabilityType query string true "Filter by CapabilityType" e.g "'InfiniBand', 'CPU'"
// @Param capabilityName query string true "Filter by CapabilityName" e.g. "'MT2910 Family [ConnectX-7]', 'Dell Ent NVMe CM6 RI 1.92TB'"
// @Param status query string false "Filter by status" e.g. 'Pending', 'Error'"
// @Param hwSkuDeviceType query string false "Filter by hardware SKU device type" e.g. 'gpu', 'cpu', 'storage', 'cache'"
// @Param query query string false "Query input for full text search"
// @Param includeMetadata query boolean false "Include metadata info in response"
// @Param includeRelation query string false "Related entities to include in response e.g. 'InfrastructureProvider', 'Site', 'InstanceType'"
// @Param pageNumber query integer false "Page number of results returned"
// @Param pageSize query integer false "Number of results per page"
// @Param orderBy query string false "Order by field"
// @Success 200 {object} []model.APIMachine
// @Router /v2/org/{org}/nico/machine [get]
func (gamh GetAllMachineHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Machine", "GetAll", c, gamh.tracerSpan)
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

	// Validate role, only Provider Admins or Tenant Admins with TargetedInstanceCreation capability are allowed to retrieve Machines
	userOrgDetails, _ := dbUser.OrgData.GetOrgByName(org)
	infrastructureProvider, tenant, apiError := isProviderOrTenant(ctx, logger, gamh.dbSession, org, userOrgDetails)
	if apiError != nil {
		return cutil.NewAPIErrorResponse(c, apiError.Code, apiError.Message, apiError.Data)
	}

	// Validate pagination request
	pageRequest := pagination.PageRequest{}
	err = c.Bind(&pageRequest)
	if err != nil {
		logger.Warn().Err(err).Msg("error binding pagination request data into API model")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Failed to parse request pagination data", nil)
	}

	// Validate request attributes
	err = pageRequest.Validate(cdbm.MachineOrderByFields)
	if err != nil {
		logger.Warn().Err(err).Msg("error validating pagination request data")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Failed to validate pagination request data", err)
	}

	// Get and validate includeRelation params
	qParams := c.QueryParams()
	qIncludeRelations, errMsg := common.GetAndValidateQueryRelations(qParams, cdbm.MachineRelatedEntities)
	if errMsg != "" {
		logger.Warn().Msg(errMsg)
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, errMsg, nil)
	}

	filterInput := cdbm.MachineFilterInput{}

	// Validate other query params
	if infrastructureProvider != nil {
		filterInput.InfrastructureProviderID = &infrastructureProvider.ID
	}

	mDAO := cdbm.NewMachineDAO(gamh.dbSession)

	// Validate site id if provided
	qSiteID := c.QueryParam("siteId")
	if qSiteID != "" {
		site, serr := common.GetSiteFromIDString(ctx, nil, qSiteID, gamh.dbSession)
		if serr != nil {
			if serr == cdb.ErrDoesNotExist {
				return cutil.NewAPIErrorResponse(c, http.StatusNotFound, "Could not find Site specified in query", nil)
			}
			logger.Error().Err(serr).Str("Site ID", qSiteID).Msg("error retrieving Site specified in query")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Error retrieving Site specified in query", nil)
		}

		if infrastructureProvider != nil {
			// Check if Site belongs to org's Infrastructure Provider
			if site.InfrastructureProviderID != infrastructureProvider.ID {
				logger.Error().Msg("Site's Infrastructure Provider doesn't match org's Infrastructure Provider")
				return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "Site specified in query doesn't belong to org's Infrastructure provider", nil)
			}
		} else if tenant != nil {
			// Check if Tenant is privileged
			if !tenant.Config.TargetedInstanceCreation {
				logger.Warn().Msg("Tenant doesn't have targeted Instance creation capability, access denied")
				return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "Tenant must have targeted Instance creation capability in order to retrieve Machines", nil)
			}

			// Check if privileged Tenant has an account with Infrastructure Provider
			taDAO := cdbm.NewTenantAccountDAO(gamh.dbSession)
			_, taCount, serr := taDAO.GetAll(ctx, nil, cdbm.TenantAccountFilterInput{
				InfrastructureProviderID: &site.InfrastructureProviderID,
				TenantIDs:                []uuid.UUID{tenant.ID},
			}, cdbp.PageInput{}, []string{})
			if serr != nil {
				logger.Error().Err(serr).Msg("error retrieving Tenant Account for Site")
				return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Error retrieving Tenant Account for Site", nil)
			}

			if taCount == 0 {
				logger.Error().Msg("privileged Tenant doesn't have an account with Infrastructure Provider")
				return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "Privileged Tenant must have an account with Provider of Site specified in query", nil)
			}
		}

		filterInput.SiteID = &site.ID
	} else if tenant != nil {
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Site ID must be specified in query when retrieving Machines as a privileged Tenant", nil)
	}

	// Validate InstanceType ID if provided
	qInstanceTypeID := qParams["instanceTypeId"]
	if len(qInstanceTypeID) > 0 {
		gamh.tracerSpan.SetAttribute(handlerSpan, attribute.StringSlice("instanceTypeId", qInstanceTypeID), logger)
		for _, instanceTypeID := range qInstanceTypeID {
			instancetype, serr := common.GetInstanceTypeFromIDString(ctx, nil, instanceTypeID, gamh.dbSession)

			if serr != nil {
				instanceTypeIdError := validation.Errors{
					"instanceTypeId": errors.New(instanceTypeID),
				}
				if serr == cdb.ErrDoesNotExist {
					return cutil.NewAPIErrorResponse(c, http.StatusNotFound, "Could not find Instance Type specified in query", instanceTypeIdError)
				}
				logger.Error().Err(serr).Str("Instance Type ID", instanceTypeID).Msg("error retreiving Instance Type specified in query")
				return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Error retrieving Instance Type specified in query", instanceTypeIdError)
			}

			filterInput.InstanceTypeIDs = append(filterInput.InstanceTypeIDs, instancetype.ID)
		}
	}

	// Check if `hasInstanceType` query params
	qHasInstanceType := c.QueryParam("hasInstanceType")
	if qHasInstanceType != "" {
		gamh.tracerSpan.SetAttribute(handlerSpan, attribute.String("hasInstanceType", qHasInstanceType), logger)
		hiType, serr := strconv.ParseBool(qHasInstanceType)
		if serr != nil {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid value specified for hasInstanceType in query", nil)
		}

		if !hiType && len(filterInput.InstanceTypeIDs) > 0 {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "hasInstanceType cannot be false when and instanceTypeId is specified in query", nil)
		}

		filterInput.HasInstanceType = cdb.GetBoolPtr(hiType)
	}

	// Check `includeMetadata` in query
	includeMetadata := false
	qim := c.QueryParam("includeMetadata")
	if qim != "" {
		includeMetadata, err = strconv.ParseBool(qim)
		if err != nil {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid value specified for `includeMetadata` query param", nil)
		}
	}

	// Get Machine ID from query param
	idQuery := qParams["id"]
	if len(idQuery) > 0 {
		gamh.tracerSpan.SetAttribute(handlerSpan, attribute.StringSlice("id", idQuery), logger)
		filterInput.MachineIDs = append(filterInput.MachineIDs, idQuery...)
	}

	qTenantIDStrs := qParams["tenantId"]
	if len(qTenantIDStrs) > 0 {
		gamh.tracerSpan.SetAttribute(handlerSpan, attribute.StringSlice("tenantId", qTenantIDStrs), logger)
		tenantAccountDAO := cdbm.NewTenantAccountDAO(gamh.dbSession)
		tenantIDs := make([]uuid.UUID, 0, len(qTenantIDStrs))
		for _, tenantIDStr := range qTenantIDStrs {
			tenantID, err := uuid.Parse(tenantIDStr)
			if err != nil {
				return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid tenant ID specified in query", nil)
			}
			tenantIDs = append(tenantIDs, tenantID)
		}

		tenantAccounts, _, err := tenantAccountDAO.GetAll(ctx, nil, cdbm.TenantAccountFilterInput{
			TenantIDs:                tenantIDs,
			InfrastructureProviderID: &infrastructureProvider.ID,
		}, cdbp.PageInput{}, nil)
		if err != nil {
			logger.Error().Err(err).Msg("error retrieving Tenant Accounts for tenant IDs specified in query")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Error retrieving Tenant Accounts for Tenants specified in query", nil)
		}
		tenantIDsMap := make(map[uuid.UUID]bool)
		for _, tenantAccount := range tenantAccounts {
			tenantIDsMap[*tenantAccount.TenantID] = true
		}
		for _, tenantID := range tenantIDs {
			if !tenantIDsMap[tenantID] {
				return cutil.NewAPIErrorResponse(c, http.StatusForbidden, fmt.Sprintf("Tenant ID %s specified in query param does not have an account with current org's Provider", tenantID.String()), nil)
			}
		}

		// Get all instances matching the specified tenant ID(s)
		instanceDAO := cdbm.NewInstanceDAO(gamh.dbSession)
		matchingInstances, _, err := instanceDAO.GetAll(ctx, nil, cdbm.InstanceFilterInput{TenantIDs: tenantIDs}, cdbp.PageInput{Limit: cdb.GetIntPtr(cdbp.TotalLimit)}, nil)
		if err != nil {
			logger.Error().Err(err).Msg("error retrieving instances for machine ID filtering")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve instances for machine ID filtering", nil)
		}
		for _, ins := range matchingInstances {
			if ins.MachineID != nil {
				filterInput.MachineIDs = append(filterInput.MachineIDs, *ins.MachineID)
			}
		}
		if len(matchingInstances) == 0 {
			filterInput.MachineIDs = []string{}
		}
	}

	//	Check if `hasInstance` query params
	qHasInstance := c.QueryParam("hasInstance")
	if qHasInstance != "" {
		gamh.tracerSpan.SetAttribute(handlerSpan, attribute.String("hasInstance", qHasInstance), logger)
		hi, serr := strconv.ParseBool(qHasInstance)
		if serr != nil {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid value specified for `hasInstance` in query", nil)
		}

		if filterInput.SiteID == nil {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "`hasInstance` cannot be specified when `siteId` is not specified in query", nil)
		}

		if !hi && len(qTenantIDStrs) > 0 {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "`hasInstance` cannot be false when `tenantId` is specified in query", nil)
		}

		filterInput.IsAssigned = cdb.GetBoolPtr(hi)
	}

	// Validate capability type from query param if it is provided
	qCPtype := c.QueryParam("capabilityType")
	if qCPtype != "" {
		_, ok := cdbm.MachineCapabilityTypeChoiceMap[qCPtype]
		if !ok {
			logger.Warn().Msg(fmt.Sprintf("invalid capabilityType value in query: %v", qCPtype))
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid capabilityType value in query: %v", qCPtype), nil)
		}
		filterInput.CapabilityType = &qCPtype
	}

	// Validate capability name from query param if it is provided
	capNameQuery := qParams["capabilityName"]
	if len(capNameQuery) > 0 {
		gamh.tracerSpan.SetAttribute(handlerSpan, attribute.StringSlice("capabilityName", capNameQuery), logger)
		filterInput.CapabilityNames = append(filterInput.CapabilityNames, capNameQuery...)
	}

	// Get query text for full text search from query param
	searchQuery := common.GetSearchQuery(c)
	if searchQuery != nil {
		filterInput.SearchQuery = searchQuery
		gamh.tracerSpan.SetAttribute(handlerSpan, attribute.String("query", *searchQuery), logger)
	}

	// Get status from query param
	statusQuery := qParams["status"]
	if len(statusQuery) > 0 {
		gamh.tracerSpan.SetAttribute(handlerSpan, attribute.StringSlice("status", statusQuery), logger)
		for _, status := range statusQuery {
			_, ok := cdbm.MachineStatusMap[status]
			if !ok {
				logger.Warn().Msg(fmt.Sprintf("invalid value in status query: %v", statusQuery))
				return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid Status value in query", nil)
			}
			filterInput.Statuses = append(filterInput.Statuses, status)
		}
	}

	// Get isMissingOnSite from query param
	qIsMissingOnSite := c.QueryParam("isMissingOnSite")
	if qIsMissingOnSite != "" {
		gamh.tracerSpan.SetAttribute(handlerSpan, attribute.String("isMissingOnSite", qIsMissingOnSite), logger)
		isMissingOnSite, err := strconv.ParseBool(qIsMissingOnSite)
		if err != nil {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid value specified for `isMissingOnSite` query param", nil)
		}

		filterInput.IsMissingOnSite = cdb.GetBoolPtr(isMissingOnSite)
	}

	// Get hwSkuDeviceType from query param
	hwSkuDeviceTypeQuery := qParams["hwSkuDeviceType"]
	if len(hwSkuDeviceTypeQuery) > 0 {
		gamh.tracerSpan.SetAttribute(handlerSpan, attribute.StringSlice("hwSkuDeviceType", hwSkuDeviceTypeQuery), logger)
		for _, hwSkuDeviceType := range hwSkuDeviceTypeQuery {
			// HwSkuDeviceType is a free-form string field, no validation needed
			filterInput.HwSkuDeviceTypes = append(filterInput.HwSkuDeviceTypes, hwSkuDeviceType)
		}
	}

	// Create response
	pageInput := cdbp.PageInput{
		Offset:  pageRequest.Offset,
		Limit:   pageRequest.Limit,
		OrderBy: pageRequest.OrderBy,
	}
	ms, total, err := mDAO.GetAll(ctx, nil, filterInput, pageInput, qIncludeRelations)
	if err != nil {
		logger.Error().Err(err).Msg("error getting Machines from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Machines", nil)
	}
	apiMs, apiErr := getAPIMachines(ctx, ms, logger, nil, gamh.dbSession, includeMetadata, true)
	if apiErr != nil {
		return cutil.NewAPIErrorResponse(c, apiErr.Code, apiErr.Message, apiErr.Data)
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

	return c.JSON(http.StatusOK, apiMs)
}

// ~~~~~ Get Handler ~~~~~ //

// GetMachineHandler is the API Handler for retrieving Machine
type GetMachineHandler struct {
	dbSession  *cdb.Session
	tc         temporalClient.Client
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewGetMachineHandler initializes and returns a new handler to retrieve Machine
func NewGetMachineHandler(dbSession *cdb.Session, tc temporalClient.Client, cfg *config.Config) GetMachineHandler {
	return GetMachineHandler{
		dbSession:  dbSession,
		tc:         tc,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Retrieve the Machine
// @Description Retrieve the Machine
// @Tags Machine
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param id path string true "ID of Machine"
// @Param includeMetadata query boolean false "Include metadata info in response"
// @Success 200 {object} model.APIMachine
// @Router /v2/org/{org}/nico/machine/{id} [get]
func (gmh GetMachineHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Machine", "Get", c, gmh.tracerSpan)
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

	// Validate role, only Provider Admins or Tenant Admins with TargetedInstanceCreation capability are allowed to retrieve Machines
	userOrgDetails, _ := dbUser.OrgData.GetOrgByName(org)
	infrastructureProvider, tenant, apiError := isProviderOrTenant(ctx, logger, gmh.dbSession, org, userOrgDetails)
	if apiError != nil {
		return cutil.NewAPIErrorResponse(c, apiError.Code, apiError.Message, apiError.Data)
	}

	// Get and validate includeRelation params
	qParams := c.QueryParams()
	qIncludeRelations, errMsg := common.GetAndValidateQueryRelations(qParams, cdbm.MachineRelatedEntities)
	if errMsg != "" {
		logger.Warn().Msg(errMsg)
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, errMsg, nil)
	}

	// Check `includeMetadata` in query
	includeMetadata := false
	qim := c.QueryParam("includeMetadata")
	if qim != "" {
		includeMetadata, err = strconv.ParseBool(qim)
		if err != nil {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid value specified for `includeMetadata` query param", nil)
		}
	}

	// Get machine ID from URL param
	mID := c.Param("id")

	gmh.tracerSpan.SetAttribute(handlerSpan, attribute.String("machine_id", mID), logger)

	mDAO := cdbm.NewMachineDAO(gmh.dbSession)
	// Check that Machine exists
	machine, err := mDAO.GetByID(ctx, nil, mID, qIncludeRelations, false)
	if err != nil {
		if err == cdb.ErrDoesNotExist {
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, "Could not find Machine with specified ID", nil)
		}
		logger.Error().Err(err).Msg("error retrieving Machine DB entity")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Could not retrieve Machine", nil)
	}

	isAssociated := false
	isProviderOrPrivilegedTenant := false

	if infrastructureProvider != nil {
		if machine.InfrastructureProviderID != infrastructureProvider.ID {
			logger.Error().Msg("machine's infrastructure provider doesn't match org")
			return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "Machine doesn't belong to org's Infrastructure provider", nil)
		}

		isAssociated = true
		isProviderOrPrivilegedTenant = true
	} else if tenant != nil {
		// Check if Tenant is privileged
		if tenant.Config.TargetedInstanceCreation {
			// Check if privileged Tenant has an account with Infrastructure Provider
			taDAO := cdbm.NewTenantAccountDAO(gmh.dbSession)
			_, taCount, serr := taDAO.GetAll(ctx, nil, cdbm.TenantAccountFilterInput{
				InfrastructureProviderID: &machine.InfrastructureProviderID,
				TenantIDs:                []uuid.UUID{tenant.ID},
			}, cdbp.PageInput{}, []string{})
			if serr != nil {
				logger.Error().Err(serr).Msg("error retrieving Tenant Account for Site")
				return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Error retrieving Tenant Account for Site", nil)
			}

			if taCount > 0 {
				isAssociated = true
				isProviderOrPrivilegedTenant = true
			}
		} else {
			// if not privileged, check if the machine is associated with an Instance belonging to the org's Tenant
			instanceDAO := cdbm.NewInstanceDAO(gmh.dbSession)
			_, iCount, serr := instanceDAO.GetAll(ctx, nil, cdbm.InstanceFilterInput{
				TenantIDs:  []uuid.UUID{tenant.ID},
				MachineIDs: []string{machine.ID},
			}, cdbp.PageInput{}, nil)
			if serr != nil {
				logger.Error().Err(serr).Msg("error retrieving Instances for tenant")
				return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to determine Tenant's association with Machine", nil)
			}

			if iCount > 0 {
				isAssociated = true
			}
		}
	}

	if !isAssociated {
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "Machine details can only be retrieved by it's Provider, privileged Tenants or Tenants who have an associated Instance", nil)
	}

	// Create response
	apiMs, apiErr := getAPIMachines(ctx, []cdbm.Machine{*machine}, logger, nil, gmh.dbSession, includeMetadata, isProviderOrPrivilegedTenant)
	if apiErr != nil {
		return cutil.NewAPIErrorResponse(c, apiErr.Code, apiErr.Message, apiErr.Data)
	}

	logger.Info().Msg("finishing API handler")

	return c.JSON(http.StatusOK, apiMs[0])
}

// ~~~~~ Update Handler ~~~~~ //

// UpdateMachineHandler is the API Handler for updating a Machine
type UpdateMachineHandler struct {
	dbSession  *cdb.Session
	tc         temporalClient.Client
	scp        *sc.ClientPool
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewUpdateMachineHandler initializes and returns a new handler to update Machine
func NewUpdateMachineHandler(dbSession *cdb.Session, tc temporalClient.Client, scp *sc.ClientPool, cfg *config.Config) UpdateMachineHandler {
	return UpdateMachineHandler{
		dbSession:  dbSession,
		tc:         tc,
		scp:        scp,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Update an existing Machine
// @Description Update an existing Machine for the org
// @Tags machine
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param id path string true "ID of Machine"
// @Param message body model.APIMachineUpdateRequest true "Machine update request"
// @Success 200 {object} model.APIMachine
// @Router /v2/org/{org}/nico/machine/{id} [patch]
func (umh UpdateMachineHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Machine", "Update", c, umh.tracerSpan)
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

	// Validate role, only Provider Admins or Tenant Admins with TargetedInstanceCreation capability are allowed to update Machine
	userOrgDetails, _ := dbUser.OrgData.GetOrgByName(org)
	infrastructureProvider, tenant, apiError := isProviderOrTenant(ctx, logger, umh.dbSession, org, userOrgDetails)
	if apiError != nil {
		return cutil.NewAPIErrorResponse(c, apiError.Code, apiError.Message, apiError.Data)
	}
	// Get machine ID from URL param
	mID := c.Param("id")

	umh.tracerSpan.SetAttribute(handlerSpan, attribute.String("machine_id", mID), logger)

	mDAO := cdbm.NewMachineDAO(umh.dbSession)
	// Check that Machine exists
	machine, err := mDAO.GetByID(ctx, nil, mID, []string{cdbm.SiteRelationName}, false)
	if err != nil {
		if err == cdb.ErrDoesNotExist {
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, "Could not find Machine specified in URL", nil)
		}
		logger.Error().Err(err).Msg("error retrieving Machine DB entity")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Machine specified in URL", nil)
	}

	if machine.Site == nil {
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Site detail for Machine", nil)
	}

	if infrastructureProvider != nil {
		if machine.InfrastructureProviderID != infrastructureProvider.ID {
			logger.Error().Msg("machine's infrastructure provider doesn't match org")
			return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "Machine doesn't belong to org's Infrastructure provider", nil)
		}
	} else if tenant != nil {
		// Check if Tenant is privileged
		if tenant.Config.TargetedInstanceCreation {
			// Check if privileged Tenant has an account with Infrastructure Provider
			taDAO := cdbm.NewTenantAccountDAO(umh.dbSession)
			_, taCount, serr := taDAO.GetAll(ctx, nil, cdbm.TenantAccountFilterInput{
				InfrastructureProviderID: &machine.InfrastructureProviderID,
				TenantIDs:                []uuid.UUID{tenant.ID},
			}, cdbp.PageInput{}, []string{})
			if serr != nil {
				logger.Error().Err(serr).Msg("error retrieving Tenant Account for Site")
				return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Error retrieving Tenant Account for Site", nil)
			}

			if taCount == 0 {
				logger.Error().Msg("privileged Tenant doesn't have an account with Infrastructure Provider")
				return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "PrivilegedTenant must have an account with Machine's Provider in order to modify it", nil)
			}
		}
	}

	// Validate request
	// Bind request data to API model
	apiRequest := model.APIMachineUpdateRequest{}
	err = c.Bind(&apiRequest)
	if err != nil {
		logger.Warn().Err(err).Msg("error binding request data into API model")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Failed to parse request data, potentially invalid structure", nil)
	}

	// Validate request attributes
	verr := apiRequest.Validate()
	if verr != nil {
		logger.Warn().Err(verr).Msg("error validating Machine update request data")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Error validating Machine update request data", verr)
	}

	// Prevent assigning or clearing Instance Type on assigned machines
	if apiRequest.InstanceTypeID != nil || apiRequest.ClearInstanceType != nil {
		if infrastructureProvider == nil {
			return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "Only Provider Admins can update or clear Machine's Instance Type", nil)
		}

		if machine.IsAssigned {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Machine is currently in use by an Instance, so its Instance Type cannot be modified", nil)
		}
	}

	// Prevent assigning Instance Type to Machine in non-ready state, but allow clearing Instance Type
	if apiRequest.InstanceTypeID != nil && machine.Status != cdbm.MachineStatusReady {
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Machine is in %v state. Instance Type can only be assigned to a Machine in `Ready` state", machine.Status), nil)
	}

	// Retrieve Instance Type
	var newit *cdbm.InstanceType

	if apiRequest.InstanceTypeID != nil {
		itDAO := cdbm.NewInstanceTypeDAO(umh.dbSession)

		parseID, serr := uuid.Parse(*apiRequest.InstanceTypeID)
		if serr != nil {
			logger.Warn().Err(serr).Msg("error parsing Instance Type ID in request")
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Instance Type ID specified in request is not valid", nil)
		}

		newit, serr = itDAO.GetByID(ctx, nil, parseID, []string{cdbm.SiteRelationName})
		if serr != nil {
			if serr == cdb.ErrDoesNotExist {
				return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Could not find Instance Type specified in request", nil)
			}
			logger.Error().Err(err).Msg("error retrieving InstanceType from DB by ID")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Instance Type specified in request", nil)
		}

		// Unlikely but check that Site relation was retrieved for Instance Type
		if newit.Site == nil {
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Instance Type specified in request doesn't have a Site associated", nil)
		}

		// Check if Machine is already associated with the Instance Type
		if machine.InstanceTypeID != nil && *machine.InstanceTypeID == newit.ID {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Machine is already associated with Instance Type specified in request", nil)
		}

		// Check if new Instance Type belong to org's Provider
		if newit.InfrastructureProviderID != infrastructureProvider.ID {
			return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "Instance Type specified in request doesn't belong to org's Infrastructure Provider", nil)
		}

		// Check that Machine and new Instance Type both belong to the same Site
		if *newit.SiteID != machine.SiteID {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Instance Type specified in request doesn't belong to the same Site as Machine", nil)
		}
	}

	if apiRequest.ClearInstanceType != nil && *apiRequest.ClearInstanceType && machine.InstanceTypeID == nil {
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Machine does not have an Instance Type assigned", nil)
	}

	// Verify if Capabilties of Machine matches with Instance Type's Capabilities
	if apiRequest.InstanceTypeID != nil {
		isMatch, _, apiErr := common.MatchInstanceTypeCapabilitiesForMachines(ctx, logger, umh.dbSession, newit.ID, []string{machine.ID})
		if apiErr != nil {
			return cutil.NewAPIErrorResponse(c, apiErr.Code, apiErr.Message, apiErr.Data)
		}

		if !isMatch {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Capabilities for Machine: %v do not match Instance Type's Capabilities", machine.ID), nil)
		}
	}

	var um *cdbm.Machine

	// Check if Site has connectivity
	if machine.Site.Status != cdbm.SiteStatusRegistered {
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Site is not in Registered state, unable to update Machine", nil)
	}

	// Get Temporal site client
	stc, err := umh.scp.GetClientByID(machine.SiteID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve Temporal client for Site")
		return err
	}

	// Save/clear Instance Type in DB and execute workflow on Site if required
	if apiRequest.InstanceTypeID != nil || (apiRequest.ClearInstanceType != nil && *apiRequest.ClearInstanceType) {
		// NOTE: Don't check if Machine is missing from Site when clearing.
		// That would prevent people from cleaning up machines in
		// cases where it had been assigned an instancetype in the past.
		// Also, if the machine doesn't exist on site, the site has no knowledge
		// of the instancetype, anyway.
		if apiRequest.InstanceTypeID != nil && machine.IsMissingOnSite {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Machine is currently missing on Site, cannot change Instance Type", nil)
		}

		var itTimeoutResp func() error
		err = cdb.WithTx(ctx, umh.dbSession, func(itTx *cdb.Tx) error {
			// Lock the machine itself so the read/delete/create sequence on its
			// instance-type association is serialized per machine. Without this,
			// two concurrent PATCHes against a machine with no current
			// association can both observe totalEmits == 0 and race into
			// CreateFromParams/Update.
			derr := itTx.TryAcquireAdvisoryLock(ctx, cdb.GetAdvisoryLockIDFromString(machine.ID), nil)
			if derr != nil {
				logger.Error().Err(derr).Str("Machine ID", machine.ID).Msg("failed to acquire advisory lock on Machine")
				return cutil.NewAPIError(http.StatusInternalServerError, "Failed to update Machine/InstanceType association", nil)
			}

			// Check if Machine/InstanceType association already exists filter by machine
			mitDAO := cdbm.NewMachineInstanceTypeDAO(umh.dbSession)
			emits, totalEmits, derr := mitDAO.GetAll(ctx, itTx, &machine.ID, nil, nil, nil, nil, nil)
			if derr != nil {
				logger.Error().Err(derr).Msg("error retrieving Machine/InstanceType association from DB")
				return cutil.NewAPIError(http.StatusInternalServerError, "Failed to check for existing InstanceType association for Machine", nil)
			}

			// Request validation guarantees that either we have a new Instance Type or existing Instance Type needs to be cleared
			// In either case, we remove existing Instance Type association
			if totalEmits > 0 {
				if totalEmits != 1 {
					logger.Error().Msg("more than 1 Machine/InstanceType association found for Machine")
					return cutil.NewAPIError(http.StatusInternalServerError, "Machine is associated with more than 1 Instance Type, data consistency detected", nil)
				}

				emit := emits[0]

				// Get the lock for old instancetype
				lockID := emit.InstanceTypeID.String()
				aerr := itTx.TryAcquireAdvisoryLock(ctx, cdb.GetAdvisoryLockIDFromString(lockID), nil)
				if aerr != nil {
					logger.Error().Err(aerr).Str("Lock ID", lockID).Msg("failed to acquire Advisory Lock")
					return cutil.NewAPIError(http.StatusInternalServerError, "Failed to update Machine/InstanceType association", nil)
				}

				// Remove Machine/InstanceType association
				serr := mitDAO.DeleteByID(ctx, itTx, emit.ID, false)
				if serr != nil {
					logger.Error().Err(serr).Msg("error deleting Machine/InstanceType association in DB")
					return cutil.NewAPIError(http.StatusInternalServerError, "Failed to remove existing Machine/InstanceType association", nil)
				}

				// Check if the above deletion of Machine/InstanceType association will violate Allocation Constraints
				ok, serr := common.CheckMachinesForInstanceTypeAllocation(ctx, itTx, umh.dbSession, logger, emit.InstanceTypeID, 0)
				if serr != nil {
					logger.Error().Err(serr).Str("Instance Type ID", emit.InstanceTypeID.String()).Msg("error checking Machine allocations for current Instance Type")
					return cutil.NewAPIError(http.StatusInternalServerError, "Failed to check Machine allocations for existing Instance Type", nil)
				}

				if !ok {
					logger.Warn().Str("resourceId", emit.InstanceTypeID.String()).Msg("Machine cannot be dissociated from existing Instance Type as it will violate Allocation Constraints")
					return cutil.NewAPIError(http.StatusBadRequest, "Machine cannot be dissociated from existing Instance Type as it will violate Allocation Constraints", nil)
				}

				// Clear Instance Type for Machine
				clearInput := cdbm.MachineClearInput{
					MachineID:      machine.ID,
					InstanceTypeID: true,
				}
				um, serr = mDAO.Clear(ctx, itTx, clearInput)
				if serr != nil {
					logger.Error().Err(serr).Msg("error clearing Instance Type for Machine in DB")
					return cutil.NewAPIError(http.StatusInternalServerError, "Failed to update InstanceType for Machine", nil)
				}
			}

			if newit != nil {
				// Create new Machine/InstanceType association
				_, serr := mitDAO.CreateFromParams(ctx, itTx, machine.ID, newit.ID)
				if serr != nil {
					logger.Error().Err(serr).Msg("error creating Machine/InstanceType association")
					return cutil.NewAPIError(http.StatusInternalServerError, "Failed to create Machine/InstanceType association", nil)
				}

				// Update Machine and set new Instance Type
				updateInput := cdbm.MachineUpdateInput{
					MachineID:      machine.ID,
					InstanceTypeID: cdb.GetUUIDPtr(newit.ID),
				}
				um, serr = mDAO.Update(ctx, itTx, updateInput)
				if serr != nil {
					logger.Error().Err(serr).Msg("error updating Machine's Instance Type in DB")
					return cutil.NewAPIError(http.StatusInternalServerError, "Failed to update Machine with new Instance Type", nil)
				}
			}

			// raise error if data inconsistency exists
			if um == nil {
				logger.Error().Msg("error updating Machine's Instance Type in DB")
				return cutil.NewAPIError(http.StatusInternalServerError, "Data inconsistencies detected in Instance Type association for this Machine", nil)
			}

			// Make the synchronous call to the site
			// Clear instance type on site if the request cleared it in cloud.
			// Earlier checks block a request to clear if there is no instance type assigned.
			if apiRequest.ClearInstanceType != nil && *apiRequest.ClearInstanceType {
				// Prepare the create request workflow object
				removeInstanceTypeRequest := &cwssaws.RemoveMachineInstanceTypeAssociationRequest{
					MachineId: machine.ID,
				}

				workflowOptions := temporalClient.StartWorkflowOptions{
					ID:                       "remove-machine-instance-type-association" + machine.InstanceTypeID.String(),
					TaskQueue:                queue.SiteTaskQueue,
					WorkflowExecutionTimeout: cutil.WorkflowExecutionTimeout,
				}

				logger.Info().Msg("triggering RemoveMachineInstanceTypeAssociation workflow")

				// Add context deadlines
				wfCtx, cancel := context.WithTimeout(ctx, cutil.WorkflowContextTimeout)
				defer cancel()

				// Trigger Site workflow
				we, wferr := stc.ExecuteWorkflow(wfCtx, workflowOptions, "RemoveMachineInstanceTypeAssociation", removeInstanceTypeRequest)
				if wferr != nil {
					logger.Error().Err(wferr).Msg("failed to synchronously start Temporal workflow to remove Machine association with InstanceType")
					return cutil.NewAPIError(http.StatusInternalServerError, fmt.Sprintf("Failed start sync workflow to remove Machine association with Instance Type on Site: %s", wferr), nil)
				}

				wid := we.GetID()
				logger.Info().Str("Workflow ID", wid).Msg("executed synchronous RemoveMachineInstanceTypeAssociation workflow")

				// Block until the workflow has completed and returned success/error.
				wferr = we.Get(wfCtx, nil)

				if wferr != nil {
					// If this was a 404 back from NICo, the machine was not found, and we can
					// treat the object as already having been deleted and allow things to proceed.
					var applicationErr *tp.ApplicationError
					if errors.As(wferr, &applicationErr) {
						if slices.Contains(swe.ObjectNotFoundErrTypes(), applicationErr.Type()) {
							logger.Warn().Msg(swe.ErrTypeNICoObjectNotFound + " received from Site")
							// Reset error to nil
							wferr = nil
						}
					}
				}

				if wferr != nil {
					var timeoutErr *tp.TimeoutError
					if errors.As(wferr, &timeoutErr) || wferr == context.DeadlineExceeded || wfCtx.Err() != nil {
						capturedErr := wferr
						itTimeoutResp = func() error {
							return common.TerminateWorkflowOnTimeOut(c, logger, stc, wid, capturedErr, "MachineInstanceType", "RemoveMachineInstanceTypeAssociation")
						}
						return cutil.NewAPIError(http.StatusInternalServerError, "workflow timeout", nil)
					}

					code, unwrapped := common.UnwrapWorkflowError(wferr)
					logger.Error().Err(unwrapped).Msg("failed to synchronously execute Temporal workflow to remove Machine association with InstanceType")
					return cutil.NewAPIError(code, fmt.Sprintf("Failed to execute sync workflow to remove Machine association with Instance Type on Site: %s", unwrapped), nil)
				}

				logger.Info().Str("Workflow ID", wid).Msg("completed synchronous RemoveMachineInstanceTypeAssociation workflow")
			} else if newit != nil {
				// NOTE: If the machine was missing on site from the POV of cloud, the request would have
				// been rejected before getting here.

				// If the request updated to a different instancetype, send that to the site.
				associateMachinesRequest := &cwssaws.AssociateMachinesWithInstanceTypeRequest{
					InstanceTypeId: newit.ID.String(),
					MachineIds:     []string{machine.ID},
				}

				workflowOptions := temporalClient.StartWorkflowOptions{
					ID:                       "associate-machines-with-instance-type-" + newit.ID.String(),
					TaskQueue:                queue.SiteTaskQueue,
					WorkflowExecutionTimeout: cutil.WorkflowExecutionTimeout,
				}

				logger.Info().Msg("triggering AssociateMachinesWithInstanceType workflow")

				// Add context deadlines
				wfCtx, cancel := context.WithTimeout(ctx, cutil.WorkflowContextTimeout)
				defer cancel()

				// Trigger Site workflow
				we, wferr := stc.ExecuteWorkflow(wfCtx, workflowOptions, "AssociateMachinesWithInstanceType", associateMachinesRequest)
				if wferr != nil {
					logger.Error().Err(wferr).Msg("failed to synchronously start Temporal workflow to associate Machines with InstanceType")
					return cutil.NewAPIError(http.StatusInternalServerError, fmt.Sprintf("Failed start sync workflow to associate Machines with Instance Type on Site: %s", wferr), nil)
				}

				wid := we.GetID()
				logger.Info().Str("Workflow ID", wid).Msg("executed synchronous AssociateMachinesWithInstanceType workflow")

				// Block until the workflow has completed and returned success/error.
				wferr = we.Get(wfCtx, nil)

				if wferr != nil {
					var timeoutErr *tp.TimeoutError
					if errors.As(wferr, &timeoutErr) || wferr == context.DeadlineExceeded || wfCtx.Err() != nil {
						capturedErr := wferr
						itTimeoutResp = func() error {
							return common.TerminateWorkflowOnTimeOut(c, logger, stc, wid, capturedErr, "MachineInstanceType", "AssociateMachinesWithInstanceType")
						}
						return cutil.NewAPIError(http.StatusInternalServerError, "workflow timeout", nil)
					}

					code, unwrapped := common.UnwrapWorkflowError(wferr)
					logger.Error().Err(unwrapped).Msg("failed to synchronously execute Temporal workflow to associate Machines with InstanceType")
					return cutil.NewAPIError(code, fmt.Sprintf("Failed to execute sync workflow to  associate Machines with Instance Type on Site: %s", unwrapped), nil)
				}

				logger.Info().Str("Workflow ID", wid).Msg("completed synchronous AssociateMachinesWithInstanceType workflow")

			}

			return nil
		})
		if itTimeoutResp != nil {
			return itTimeoutResp()
		}
		if err != nil {
			return common.HandleTxError(c, logger, err, "Failed to update Machine, DB transaction error")
		}
	}

	// Save/clear maintenance mode in DB and execute workflow on Site if required
	if apiRequest.SetMaintenanceMode != nil {
		// Check if Machine is missing from Site
		if machine.IsMissingOnSite {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Machine is currently missing on Site, cannot update maintenance mode", nil)
		}

		// timeoutResp lets the closure signal a post-rollback handler — the
		// TerminateWorkflow call has to run after the closure returns so that
		// the DB tx unwinds before we make the second remote call. nil means
		// no timeout occurred and the normal flow continues.
		var timeoutResp func() error

		err = cdb.WithTx(ctx, umh.dbSession, func(mnTx *cdb.Tx) error {
			// Update records in DB
			status := cdbm.MachineStatusMaintenance
			statusMessage := "Machine is in maintenance mode"
			if apiRequest.MaintenanceMessage != nil {
				statusMessage = fmt.Sprintf("%s: %s", statusMessage, *apiRequest.MaintenanceMessage)
			}

			if !*apiRequest.SetMaintenanceMode {
				// TODO: Inspect Machine metadata object to determine the appropriate status
				status = cdbm.MachineStatusInitializing
				statusMessage = "Machine is in initialization state"
			}

			updateInput := cdbm.MachineUpdateInput{
				MachineID:          machine.ID,
				IsInMaintenance:    apiRequest.SetMaintenanceMode,
				MaintenanceMessage: apiRequest.MaintenanceMessage,
				Status:             &status,
			}
			var derr error
			um, derr = mDAO.Update(ctx, mnTx, updateInput)
			if derr != nil {
				logger.Error().Err(derr).Msg("error updating Machine's maintenance mode in DB")
				return cutil.NewAPIError(http.StatusInternalServerError, "Failed to update Machine maintenance mode, DB error", nil)
			}

			// Clear maintenance message if maintenance mode is being disabled
			if !*apiRequest.SetMaintenanceMode {
				// Clear Maintenance Message
				clearInput := cdbm.MachineClearInput{
					MachineID:          machine.ID,
					MaintenanceMessage: true,
				}
				um, derr = mDAO.Clear(ctx, mnTx, clearInput)
				if derr != nil {
					logger.Error().Err(derr).Msg("error clearing maintenance message for Machine in DB")
					return cutil.NewAPIError(http.StatusInternalServerError, "Failed to clear Machine maintenance message, DB error", nil)
				}
			}

			// Add status detail
			sdDAO := cdbm.NewStatusDetailDAO(umh.dbSession)
			_, derr = sdDAO.CreateFromParams(ctx, mnTx, machine.ID, status, &statusMessage)
			if derr != nil {
				logger.Error().Err(derr).Msg("error creating Status Detail for Machine in DB")
				return cutil.NewAPIError(http.StatusInternalServerError, "Failed to create status detail for Machine, DB error", nil)
			}

			// Trigger Site workflow to set/remove maintenance mode
			wfOpts := temporalClient.StartWorkflowOptions{
				ID:                       "site-set-maintenance-" + machine.ID,
				WorkflowExecutionTimeout: cutil.WorkflowExecutionTimeout,
				TaskQueue:                queue.SiteTaskQueue,
			}

			// If maintenance mode is being removed and Machine is currently not in maintenance mode then raise error
			if !*apiRequest.SetMaintenanceMode && machine.Status != cdbm.MachineStatusMaintenance {
				return cutil.NewAPIError(http.StatusBadRequest, "Machine is currently not in maintenance mode, cannot remove maintenance mode", nil)
			}

			wfReq := &cwssaws.MaintenanceRequest{HostId: &cwssaws.MachineId{Id: machine.ID}}
			if *apiRequest.SetMaintenanceMode {
				wfReq.Operation = cwssaws.MaintenanceOperation_Enable
				wfReq.Reference = apiRequest.MaintenanceMessage
			} else {
				wfReq.Operation = cwssaws.MaintenanceOperation_Disable
			}

			// Add context deadlines
			wfCtx, cancel := context.WithTimeout(ctx, cutil.WorkflowContextTimeout)
			defer cancel()

			we, wferr := stc.ExecuteWorkflow(wfCtx, wfOpts, "SetMachineMaintenance", wfReq)
			if wferr != nil {
				logger.Error().Err(wferr).Msg("failed to synchronously start Temporal workflow to set/remove maintenance mode")
				return cutil.NewAPIError(http.StatusInternalServerError, fmt.Sprintf("Failed start sync workflow to set/remove maintenance mode on Site: %s", wferr), nil)
			}

			wid := we.GetID()
			logger.Info().Str("Workflow ID", wid).Msg("executed synchronous set/remove maintenance mode workflow")

			// Execute the workflow synchronously
			wferr = we.Get(wfCtx, nil)
			if wferr != nil {
				var timeoutErr *tp.TimeoutError
				if errors.As(wferr, &timeoutErr) || wferr == context.DeadlineExceeded || wfCtx.Err() != nil {
					logger.Error().Err(wferr).Msg("failed to set/remove Machine maintenance mode, timeout occurred executing workflow on Site.")
					timeoutResp = func() error {
						return common.TerminateWorkflowOnTimeOut(c, logger, stc, wid, wferr, "Machine", "SetMachineMaintenance")
					}
					return cutil.NewAPIError(http.StatusInternalServerError, "Machine maintenance mode workflow timed out", nil)
				}

				code, unwrapped := common.UnwrapWorkflowError(wferr)
				logger.Error().Err(unwrapped).Msg("failed to synchronously execute Temporal workflow to set/remove maintenance mode")
				return cutil.NewAPIError(code, fmt.Sprintf("Failed to execute sync workflow to set/remove maintenance mode on Site: %s", unwrapped), nil)
			}

			logger.Info().Str("Workflow ID", wid).Msg("completed synchronous set/remove maintenance mode workflow")

			return nil
		})
		if timeoutResp != nil {
			return timeoutResp()
		}
		if err != nil {
			return common.HandleTxError(c, logger, err, "Failed to update Machine, DB transaction error")
		}
	}

	// Save labels in DB and execute metadata update workflow on Site if required
	if apiRequest.Labels != nil && !maps.Equal(apiRequest.Labels, machine.Labels) {
		// Check if Machine is missing from Site
		if machine.IsMissingOnSite {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Machine is currently missing on Site, cannot update labels", nil)
		}

		// timeoutResp lets the closure signal a post-rollback handler — the
		// TerminateWorkflow call has to run after the closure returns so that
		// the DB tx unwinds before we make the second remote call. nil means
		// no timeout occurred and the normal flow continues.
		var timeoutResp func() error

		err = cdb.WithTx(ctx, umh.dbSession, func(lTx *cdb.Tx) error {
			// Update labels
			updateInput := cdbm.MachineUpdateInput{
				MachineID: machine.ID,
				Labels:    apiRequest.Labels,
			}

			var derr error
			um, derr = mDAO.Update(ctx, lTx, updateInput)
			if derr != nil {
				logger.Error().Err(derr).Msg("error updating Machine labels in DB")
				return cutil.NewAPIError(http.StatusInternalServerError, "Failed to update Machine labels, DB error", nil)
			}

			// Trigger Site workflow to update labels with Machine metadata
			wfOpts := temporalClient.StartWorkflowOptions{
				ID:                       "site-update-machine-metadata-" + machine.ID,
				WorkflowExecutionTimeout: cutil.WorkflowExecutionTimeout,
				TaskQueue:                queue.SiteTaskQueue,
			}

			labels := util.ProtobufLabelsFromAPILabels(apiRequest.Labels)

			machineName := machine.ID
			if machine.Metadata != nil && machine.Metadata.Metadata != nil {
				machineName = machine.Metadata.Metadata.Name
			}

			wfReq := &cwssaws.MachineMetadataUpdateRequest{
				MachineId: &cwssaws.MachineId{
					Id: machine.ID,
				},
				Metadata: &cwssaws.Metadata{
					Name:   machineName, // Site Controller sets Machine ID as name and it must be specified to update labels
					Labels: labels,
				},
			}

			// Add context deadlines
			wfCtx, cancel := context.WithTimeout(ctx, cutil.WorkflowContextTimeout)
			defer cancel()

			we, wferr := stc.ExecuteWorkflow(wfCtx, wfOpts, "UpdateMachineMetadata", wfReq)
			if wferr != nil {
				logger.Error().Err(wferr).Msg("failed to synchronously start Temporal workflow to update Machine metadata")
				return cutil.NewAPIError(http.StatusInternalServerError, fmt.Sprintf("Failed start sync workflow to update Machine labels on Site: %s", wferr), nil)
			}

			wid := we.GetID()

			// Execute the workflow synchronously
			logger.Info().Str("Workflow ID", wid).Msg("executing synchronous Machine metadata update workflow")

			wferr = we.Get(wfCtx, nil)
			if wferr != nil {
				var timeoutErr *tp.TimeoutError
				if errors.As(wferr, &timeoutErr) || wferr == context.DeadlineExceeded || wfCtx.Err() != nil {
					logger.Error().Err(wferr).Msg("failed to update Machine metadata, timeout occurred executing workflow on Site.")
					timeoutResp = func() error {
						return common.TerminateWorkflowOnTimeOut(c, logger, stc, wid, wferr, "Machine", "UpdateMachineMetadata")
					}
					return cutil.NewAPIError(http.StatusInternalServerError, "Machine metadata update workflow timed out", nil)
				}

				code, unwrapped := common.UnwrapWorkflowError(wferr)

				logger.Error().Err(unwrapped).Msg("failed to synchronously execute Temporal workflow to update Machine metadata")

				return cutil.NewAPIError(code, fmt.Sprintf("Failed to execute sync workflow to update Machine labels on Site: %s", unwrapped), nil)
			}

			logger.Info().Str("Workflow ID", wid).Msg("completed synchronous Machine metadata update workflow")

			return nil
		})
		if timeoutResp != nil {
			return timeoutResp()
		}
		if err != nil {
			return common.HandleTxError(c, logger, err, "Failed to update Machine, DB transaction error")
		}
	}

	// Create response
	if um == nil {
		um = machine
	}

	apiMs, apiErr := getAPIMachines(ctx, []cdbm.Machine{*um}, logger, nil, umh.dbSession, true, true)
	if apiErr != nil {
		return cutil.NewAPIErrorResponse(c, apiErr.Code, apiErr.Message, apiErr.Data)
	}

	logger.Info().Msg("finishing API handler")

	return c.JSON(http.StatusOK, apiMs[0])
}

// GetMachineStatusDetailsHandler is the API Handler for getting Machine StatusDetail records
type GetMachineStatusDetailsHandler struct {
	dbSession  *cdb.Session
	tracerSpan *cutil.TracerSpan
}

// NewGetMachineStatusDetailsHandler initializes and returns a new handler to retrieve Machine StatusDetail records
func NewGetMachineStatusDetailsHandler(dbSession *cdb.Session) GetMachineStatusDetailsHandler {
	return GetMachineStatusDetailsHandler{
		dbSession:  dbSession,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Get Machine StatusDetails
// @Description Get all StatusDetails for Machine
// @Tags Machine
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param id path string true "ID of Machine"
// @Success 200 {object} []model.APIStatusDetail
// @Router /v2/org/{org}/nico/machine/{id}/status-history [get]
func (gmsdh GetMachineStatusDetailsHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Machine", "Get", c, gmsdh.tracerSpan)
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

	// Get machine ID from URL param
	machineID := c.Param("id")

	gmsdh.tracerSpan.SetAttribute(handlerSpan, attribute.String("machine_id", machineID), logger)

	mDAO := cdbm.NewMachineDAO(gmsdh.dbSession)
	// Check that Machine exists
	m, err := mDAO.GetByID(ctx, nil, machineID, nil, false)
	if err != nil {
		if err == cdb.ErrDoesNotExist {
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, "Could not find Machine with specified ID", nil)
		}
		logger.Error().Err(err).Msg("error retrieving Machine DB entity")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Could not retrieve Machine", nil)
	}

	isAssociated := false

	// check org has infra provider
	orgInfrastructureProvider, err := common.GetInfrastructureProviderForOrg(ctx, nil, gmsdh.dbSession, org)
	if err != nil {
		if err != common.ErrOrgInstrastructureProviderNotFound {
			logger.Error().Err(err).Msg("error getting infrastructure provider for org")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve infrastructure provider for org, DB error", nil)
		}
	} else if m.InfrastructureProviderID != orgInfrastructureProvider.ID {
		logger.Error().Msg("machine's infrastructure provider doesn't match org")
	} else {
		// Validate role, only Provider Admins are allowed to proceed from here
		ok = auth.ValidateUserRoles(dbUser, org, nil, auth.ProviderAdminRole, auth.ProviderViewerRole)
		if !ok {
			logger.Warn().Msg("user does not have Provider Admin role, access denied")
			return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User doesn't have Provider Admin role with org", nil)
		}
		isAssociated = true
	}

	// check if this retrieve is for a machine corresponding to the org's tenant's instance
	if !isAssociated {
		tn, err := common.GetTenantForOrg(ctx, nil, gmsdh.dbSession, org)
		if err == nil {
			instanceDAO := cdbm.NewInstanceDAO(gmsdh.dbSession)
			instances, _, serr := instanceDAO.GetAll(ctx, nil, cdbm.InstanceFilterInput{TenantIDs: []uuid.UUID{tn.ID}, MachineIDs: []string{m.ID}}, cdbp.PageInput{}, []string{cdbm.TenantRelationName})
			if serr != nil {
				logger.Error().Err(serr).Msg("error retrieving Instances for tenant")
				return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to determine Tenant's association with Machine", nil)
			}

			if len(instances) > 0 {
				// Validate role, only Tenant Admins are allowed to proceed from here
				ok = auth.ValidateUserRoles(dbUser, org, nil, auth.TenantAdminRole)
				if !ok {
					logger.Warn().Msg("user does not have Tenant Admin role, access denied")
					return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User doesn't have Tenant Admin role with org", nil)
				}
				isAssociated = true
			}
		} else {
			// TODO: We must distinguish between DB connection error and no Tenant found
			logger.Info().Err(err).Msg("error retrieving Tenant for org")
		}
	}

	if !isAssociated {
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Machine is neither owned by org's Provider, nor it is associated with an Instance belonging to the org's Tenant", nil)
	}

	// handle retrieving and building status details response
	apiSds, err := handleEntityStatusDetails(ctx, c, gmsdh.dbSession, machineID, logger)
	if err != nil {
		return err
	}

	logger.Info().Msg("finishing API handler")

	return c.JSON(http.StatusOK, apiSds)
}

// ~~~~~ Delete Handler ~~~~~ //

// DeleteMachineHandler is the API Handler for updating a Machine
type DeleteMachineHandler struct {
	dbSession  *cdb.Session
	tc         temporalClient.Client
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewDeleteMachineHandler initializes and returns a new handler to update Machine
func NewDeleteMachineHandler(dbSession *cdb.Session, tc temporalClient.Client, cfg *config.Config) DeleteMachineHandler {
	return DeleteMachineHandler{
		dbSession:  dbSession,
		tc:         tc,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Delete an existing Machine from Cloud
// @Description Delete an existing Machine for the org
// @Tags machine
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param id path string true "ID of Machine"
// @Success 202 {object}
// @Router /v2/org/{org}/nico/machine/{id} [delete]
func (umh DeleteMachineHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Machine", "Delete", c, umh.tracerSpan)
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

	// Validate role, only Provider Admins are allowed to proceed from here
	ok = auth.ValidateUserRoles(dbUser, org, nil, auth.ProviderAdminRole)
	if !ok {
		logger.Warn().Msg("user does not have Provider Admin role, access denied")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User doesn't have Provider Admin role with org", nil)
	}

	// Get machine ID from URL param
	mID := c.Param("id")

	logger = log.With().Str("Machine", mID).Logger()

	umh.tracerSpan.SetAttribute(handlerSpan, attribute.String("machine_id", mID), logger)

	err = cdb.WithTx(ctx, umh.dbSession, func(tx *cdb.Tx) error {
		mDAO := cdbm.NewMachineDAO(umh.dbSession)
		// Check that Machine exists
		// We do this twice:
		// The first time is to grab a row-level lock with FOR UPDATE and without relations because they'd prevent FOR UPDATE
		// We then query again to get the rest of the details we need.
		// We use the ForUpdate option to grab a row-level lock on the machine.
		_, derr := mDAO.GetByID(ctx, tx, mID, nil, true)
		if derr != nil {
			if derr == cdb.ErrDoesNotExist {
				return cutil.NewAPIError(http.StatusNotFound, "Could not find Machine specified in URL", nil)
			}
			logger.Error().Err(derr).Msg("error retrieving Machine DB entity")
			return cutil.NewAPIError(http.StatusInternalServerError, "Failed to retrieve Machine specified in URL", nil)
		}

		machine, derr := mDAO.GetByID(ctx, tx, mID, []string{cdbm.SiteRelationName, cdbm.InstanceTypeRelationName}, false)
		if derr != nil {
			if derr == cdb.ErrDoesNotExist {
				return cutil.NewAPIError(http.StatusNotFound, "Could not find Machine specified in URL", nil)
			}
			logger.Error().Err(derr).Msg("error retrieving Machine DB entity")
			return cutil.NewAPIError(http.StatusInternalServerError, "Failed to retrieve Machine specified in URL", nil)
		}

		// Check org has infra provider
		orgInfrastructureProvider, derr := common.GetInfrastructureProviderForOrg(ctx, tx, umh.dbSession, org)
		if derr != nil {
			if derr == common.ErrOrgInstrastructureProviderNotFound {
				return cutil.NewAPIError(http.StatusBadRequest, "Org doesn't have an Infrastructure Provider associated", nil)
			}
			logger.Error().Err(derr).Msg("error getting Infrastructure Provider for org")
			return cutil.NewAPIError(http.StatusInternalServerError, "Failed to retrieve infrastructure provider for org, DB error", nil)
		}

		// Check if Machine belongs to org's Provider
		if machine.InfrastructureProviderID != orgInfrastructureProvider.ID {
			return cutil.NewAPIError(http.StatusForbidden, "Machine specified in URL is not owned by org's Infrastructure Provider", nil)
		}

		if machine.Site == nil {
			logger.Error().Msg("no Site relation found for Machine")
			return cutil.NewAPIError(http.StatusInternalServerError, "Failed to retrieve Site detail for Machine", nil)
		}

		// Prevent deleting if seen on site
		if !machine.IsMissingOnSite {
			logger.Error().Msg("Machine exists on Site and cannot be deleted")
			return cutil.NewAPIError(http.StatusBadRequest, "Machine exists on Site and cannot be deleted", nil)
		}

		// Even if IsMissingOnSite is true, we want to make sure it's been missing for a little while
		statusDAO := cdbm.NewStatusDetailDAO(umh.dbSession)
		statuses, _, derr := statusDAO.GetAllByEntityID(ctx, tx, machine.ID, nil, cdb.GetIntPtr(1), nil)

		if derr != nil {
			logger.Error().Err(derr).Msg("error while retrieving StatusDetail for Machine")
			return cutil.NewAPIError(http.StatusInternalServerError, "Failed to delete Machine", nil)
		}

		if len(statuses) == 0 {
			logger.Error().Msg("IsMissingOnSite is true but no status seen from Site")
			return cutil.NewAPIError(http.StatusBadRequest, "Machine does not have a status detail indicating when it went missing, unable to proceed with deletion", nil)
		}

		lastStatus := statuses[0]

		if lastStatus.Message == nil || *lastStatus.Message != "Machine is missing on Site" {
			lastStatusMessage := ""
			if lastStatus.Message != nil {
				lastStatusMessage = *lastStatus.Message
			}
			logger.Error().Msgf("IsMissingOnSite is true but most recent status `%s` does not match", lastStatusMessage)
			return cutil.NewAPIError(http.StatusBadRequest, "Latest status detail for Machine is not regarding it's missing state, unable to proceed with deletion", nil)
		}

		// If the most recent status shows missing on site but it has not been in that state for very long,
		// reject because maybe this was just a force-delete on site and the machine will come back.
		curTime := time.Now()
		deletionThreshold := lastStatus.Created.Add(MachineMissingDelayThreshold)

		if curTime.Before(deletionThreshold) {
			timeSince := curTime.Sub(lastStatus.Created)
			timeRemaining := deletionThreshold.Sub(curTime)

			logger.Warn().Msgf("Machine cannot be deleted as it has been missing on Site for %d hour(s) and %d minute(s) only", int(timeSince.Hours()), int(timeSince.Minutes())%60)

			return cutil.NewAPIError(
				http.StatusBadRequest,
				fmt.Sprintf(
					"Machine missing on site less than %d hour(s), reattempt after %d hour(s) and %v minute(s)",
					int(MachineMissingDelayThreshold.Hours()),
					int(timeRemaining.Hours()),
					int(timeRemaining.Minutes())%60,
				),
				nil,
			)
		}

		// Prevent deleting if an instance exists.
		// This technically shouldn't be necessary since we check for association
		// with an instance type before allowing deletion, and you can't get an
		// instance without an instance type, but this check here lets us be helpful
		// to the user by giving them some details about what they need to clean up.
		iDAO := cdbm.NewInstanceDAO(umh.dbSession)
		instances, _, derr := iDAO.GetAll(
			ctx, tx,
			cdbm.InstanceFilterInput{MachineIDs: []string{machine.ID}},
			cdbp.PageInput{
				Limit: cdb.GetIntPtr(1),
			},
			[]string{cdbm.TenantRelationName},
		)

		if derr != nil {
			logger.Error().Err(derr).Msg("error pulling instance details for Machine in DB")
			return cutil.NewAPIError(http.StatusInternalServerError, "Failed to query Instance assocations for Machine", nil)
		}

		if len(instances) > 0 {
			instance := instances[0]

			return cutil.NewAPIError(
				http.StatusBadRequest,
				fmt.Sprintf("Machine is attached to Instance: `%s` owned by Tenant: `%s`. Please ask the Tenant to delete the Instance first", instance.Name, instance.Tenant.Name),
				nil,
			)
		}

		if machine.InstanceType != nil {
			return cutil.NewAPIError(
				http.StatusBadRequest,
				fmt.Sprintf("Machine has Instance Type: Name: `%s` ID: `%s` assigned to it. Please unassign before deleting Machine", machine.InstanceType.Name, machine.InstanceType.ID),
				nil,
			)
		}

		// Clean up capabilities
		mcDAO := cdbm.NewMachineCapabilityDAO(umh.dbSession)
		caps, _, derr := mcDAO.GetAll(ctx, tx, []string{machine.ID}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, cdb.GetIntPtr(cdbp.TotalLimit), nil)
		if derr != nil {
			logger.Error().Err(derr).Msg("error pulling machine capabilities for Machine in DB")
			return cutil.NewAPIError(http.StatusInternalServerError, "Failed to retrieve Capabilities for Machine, DB error", nil)
		}

		for _, cap := range caps {
			derr := mcDAO.DeleteByID(ctx, tx, cap.ID, false)
			if derr != nil {
				logger.Error().Err(derr).Msg("error deleting machine capabilities for Machine in DB")
				return cutil.NewAPIError(http.StatusInternalServerError, "Failed to delete Capability for Machine, DB error", nil)
			}
		}

		// Clean up interfaces
		mifcDAO := cdbm.NewMachineInterfaceDAO(umh.dbSession)
		ifcs, _, derr := mifcDAO.GetAll(
			ctx,
			tx,
			cdbm.MachineInterfaceFilterInput{
				MachineIDs: []string{machine.ID},
			},
			cdbp.PageInput{Limit: cdb.GetIntPtr(cdbp.TotalLimit)},
			nil,
		)
		if derr != nil {
			logger.Error().Err(derr).Msg("error pulling machine interfaces for Machine in DB")
			return cutil.NewAPIError(http.StatusInternalServerError, "Failed to retrieve Interfaces for Machine, DB error", nil)
		}

		for _, ifc := range ifcs {
			derr := mifcDAO.Delete(ctx, tx, ifc.ID, false)
			if derr != nil {
				logger.Error().Err(derr).Msg("error deleting machine interfaces for Machine in DB")
				return cutil.NewAPIError(http.StatusInternalServerError, "Failed to delete Interface for Machine, DB error", nil)
			}
		}

		// Delete the machine
		derr = mDAO.Delete(ctx, tx, machine.ID, false)
		if derr != nil {
			logger.Error().Err(derr).Msg("error deleting Machine in DB")
			return cutil.NewAPIError(http.StatusInternalServerError, "Failed to delete Machine, DB error", nil)
		}

		return nil
	})
	if err != nil {
		return common.HandleTxError(c, logger, err, "Failed to delete Machine, DB transaction error")
	}

	logger.Info().Msg("finishing API handler")

	return c.JSON(http.StatusAccepted, nil)
}
