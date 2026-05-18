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

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	tclient "go.temporal.io/sdk/client"
	tp "go.temporal.io/sdk/temporal"

	"github.com/NVIDIA/infra-controller-rest/api/internal/config"
	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/handler/util/common"
	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/model"
	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/pagination"
	sc "github.com/NVIDIA/infra-controller-rest/api/pkg/client/site"
	auth "github.com/NVIDIA/infra-controller-rest/auth/pkg/authorization"
	cutil "github.com/NVIDIA/infra-controller-rest/common/pkg/util"
	cdb "github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
	"github.com/NVIDIA/infra-controller-rest/db/pkg/db/paginator"
	"github.com/NVIDIA/infra-controller-rest/workflow/pkg/queue"

	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

// ~~~~~ Create Handler ~~~~~ //

// CreateDpuExtensionServiceHandler is the API Handler for creating new DPU Extension Service
type CreateDpuExtensionServiceHandler struct {
	dbSession  *cdb.Session
	tc         tclient.Client
	scp        *sc.ClientPool
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewCreateDpuExtensionServiceHandler initializes and returns a new handler for creating DPU Extension Service
func NewCreateDpuExtensionServiceHandler(dbSession *cdb.Session, tc tclient.Client, scp *sc.ClientPool, cfg *config.Config) CreateDpuExtensionServiceHandler {
	return CreateDpuExtensionServiceHandler{
		dbSession:  dbSession,
		tc:         tc,
		scp:        scp,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Create a DPU Extension Service
// @Description Create a DPU Extension Service for the current Tenant
// @Tags DPU Extension Service
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param message body model.APIDpuExtensionServiceCreateRequest true "DPU Extension Service creation request"
// @Success 201 {object} model.APIDpuExtensionService
// @Router /v2/org/{org}/nico/dpu-extension-service [post]
func (cdesh CreateDpuExtensionServiceHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Create", "DpuExtensionService", c, cdesh.tracerSpan)
	if handlerSpan != nil {
		defer handlerSpan.End()
	}

	// Is DB user missing?
	if dbUser == nil {
		logger.Error().Msg("invalid User object found in request context")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve current user", nil)
	}

	// Validate the tenant for which this DPU Extension Service is being created
	tenant, err := common.GetTenantForOrg(ctx, nil, cdesh.dbSession, org)
	if err != nil {
		if err == common.ErrOrgTenantNotFound {
			logger.Warn().Err(err).Msg("Org does not have a Tenant associated")
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Org does not have a Tenant associated", nil)
		}
		logger.Error().Err(err).Msg("unable to retrieve Tenant for org")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Tenant for org", nil)
	}

	// Validate role, only Tenant Admins are allowed to create DPU Extension Service
	ok := auth.ValidateUserRoles(dbUser, org, nil, auth.TenantAdminRole)
	if !ok {
		logger.Warn().Msg("user does not have Tenant Admin role with org, access denied")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User does not have Tenant Admin role with org", nil)
	}

	// Validate request
	// Bind request data to API model
	apiRequest := model.APIDpuExtensionServiceCreateRequest{}
	err = c.Bind(&apiRequest)
	if err != nil {
		logger.Warn().Err(err).Msg("error binding request data into API model")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Failed to parse request data, potentially invalid structure", nil)
	}

	cdesh.tracerSpan.SetAttribute(handlerSpan, attribute.String("name", apiRequest.Name), logger)

	// Validate request attributes
	verr := apiRequest.Validate()
	if verr != nil {
		logger.Warn().Err(verr).Msg("error validating DPU Extension Service creation request data")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Error validating DPU Extension Service creation request data", verr)
	}

	// Retrieve the Site from the DB
	site, err := common.GetSiteFromIDString(ctx, nil, apiRequest.SiteID, cdesh.dbSession)
	if err != nil {
		if errors.Is(err, cdb.ErrDoesNotExist) {
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Site specified in request data does not exist", nil)
		}
		logger.Error().Err(err).Msg("error retrieving Site from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Site specified in request data, DB error", nil)
	}

	// Verify Tenant has access to Site
	tsDAO := cdbm.NewTenantSiteDAO(cdesh.dbSession)
	tenantSites, _, err := tsDAO.GetAll(
		ctx,
		nil,
		cdbm.TenantSiteFilterInput{
			TenantIDs: []uuid.UUID{tenant.ID},
			SiteIDs:   []uuid.UUID{site.ID},
		},
		paginator.PageInput{Limit: cdb.GetIntPtr(1)},
		nil,
	)
	if err != nil {
		logger.Error().Err(err).Msg("error retrieving TenantSite from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to validate Site access for Tenant, DB error", nil)
	}
	if len(tenantSites) == 0 {
		logger.Warn().Msg("Tenant does not have access to Site specified in request")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Tenant does not have access to Site specified in request", nil)
	}

	// Validate that site is in Registered state
	if site.Status != cdbm.SiteStatusRegistered {
		logger.Warn().Msg("Site specified in request data is not in Registered state")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Site specified in request data is not in Registered state, cannot create DPU Extension Service", nil)
	}

	// Check for duplicate DPU Extension Service name for this Tenant
	desDAO := cdbm.NewDpuExtensionServiceDAO(cdesh.dbSession)
	existingServices, _, err := desDAO.GetAll(
		ctx,
		nil,
		cdbm.DpuExtensionServiceFilterInput{
			TenantIDs: []uuid.UUID{tenant.ID},
			Names:     []string{apiRequest.Name},
		},
		paginator.PageInput{Limit: cdb.GetIntPtr(1)},
		nil,
	)
	if err != nil {
		logger.Error().Err(err).Msg("error checking for duplicate DPU Extension Service name")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to validate DPU Extension Service name uniqueness, DB error", nil)
	}
	if len(existingServices) > 0 {
		logger.Warn().Msg("DPU Extension Service with this name already exists for Tenant")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "DPU Extension Service with this name already exists", nil)
	}

	// Start a db transaction
	tx, err := cdb.BeginTx(ctx, cdesh.dbSession, &sql.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("unable to start transaction")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to create DPU Extension Service, DB transaction error", nil)
	}
	txCommitted := false
	defer common.RollbackTx(ctx, tx, &txCommitted)

	// Create the DPU Extension Service in DB
	dpuExtensionService, err := desDAO.Create(
		ctx,
		tx,
		cdbm.DpuExtensionServiceCreateInput{
			Name:        apiRequest.Name,
			Description: apiRequest.Description,
			ServiceType: apiRequest.ServiceType,
			SiteID:      site.ID,
			TenantID:    tenant.ID,
			Status:      cdbm.DpuExtensionServiceStatusPending,
			CreatedBy:   dbUser.ID,
		},
	)
	if err != nil {
		logger.Error().Err(err).Msg("error creating DPU Extension Service record in DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to create DPU Extension Service, DB error", nil)
	}

	// Create a status detail record for the DPU Extension Service
	sdDAO := cdbm.NewStatusDetailDAO(cdesh.dbSession)
	statusDetail, err := sdDAO.CreateFromParams(ctx, tx, dpuExtensionService.ID.String(), cdbm.DpuExtensionServiceStatusPending,
		cdb.GetStrPtr("Received DPU Extension Service creation request, pending processing"))
	if err != nil {
		logger.Error().Err(err).Msg("error creating Status Detail DB entry")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to create Status Detail for DPU Extension Service", nil)
	}

	statusDetails := []cdbm.StatusDetail{}
	if statusDetail != nil {
		statusDetails = append(statusDetails, *statusDetail)
	}

	createDpuExtensionServiceRequest := &cwssaws.CreateDpuExtensionServiceRequest{
		ServiceId:            cdb.GetStrPtr(dpuExtensionService.ID.String()),
		ServiceName:          apiRequest.Name,
		Description:          apiRequest.Description,
		TenantOrganizationId: org,
		Data:                 apiRequest.Data,
	}

	if apiRequest.ServiceType == model.DpuExtensionServiceTypeKubernetesPod {
		createDpuExtensionServiceRequest.ServiceType = cwssaws.DpuExtensionServiceType_KUBERNETES_POD
	}

	if apiRequest.Credentials != nil {
		createDpuExtensionServiceRequest.Credential = &cwssaws.DpuExtensionServiceCredential{
			RegistryUrl: apiRequest.Credentials.RegistryURL,
			Type: &cwssaws.DpuExtensionServiceCredential_UsernamePassword{
				UsernamePassword: &cwssaws.UsernamePassword{
					Username: *apiRequest.Credentials.Username,
					Password: *apiRequest.Credentials.Password,
				},
			},
		}
	}

	if apiRequest.Observability != nil {
		createDpuExtensionServiceRequest.Observability = apiRequest.Observability.ToProto()
	}

	logger.Info().Msg("triggering DPU Extension Service create workflow on Site")

	// Create workflow options
	workflowOptions := tclient.StartWorkflowOptions{
		ID:                       "dpu-extension-service-create-" + dpuExtensionService.ID.String(),
		WorkflowExecutionTimeout: cutil.WorkflowExecutionTimeout,
		TaskQueue:                queue.SiteTaskQueue,
	}

	// Get the temporal client for the site we are working with
	stc, err := cdesh.scp.GetClientByID(site.ID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve Temporal client for Site")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve workflow client for Site", nil)
	}

	// Add context deadlines
	ctxWithTimeout, cancel := context.WithTimeout(ctx, cutil.WorkflowContextTimeout)
	defer cancel()

	// Trigger Site workflow
	workflowRun, err := stc.ExecuteWorkflow(ctxWithTimeout, workflowOptions, "CreateDpuExtensionService", createDpuExtensionServiceRequest)
	if err != nil {
		logger.Error().Err(err).Msg("failed to schedule CreateDpuExtensionService workflow on Site")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to schedule DPU Extension Service creation workflow", nil)
	}

	workflowID := workflowRun.GetID()

	logger.Info().Str("Workflow ID", workflowID).Msg("executing CreateDpuExtensionService workflow on Site in sync mode")

	// Execute sync workflow on Site
	var controllerDpuExtensionService *cwssaws.DpuExtensionService

	err = workflowRun.Get(ctxWithTimeout, &controllerDpuExtensionService)
	if err != nil {
		var timeoutErr *tp.TimeoutError
		if errors.As(err, &timeoutErr) {
			logger.Error().Err(err).Msg("timed out executing DPU Extension Service creation workflow on Site")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Timed out executing DPU Extension Service creation workflow on Site: %s", err), nil)
		}

		code, uwerr := common.UnwrapWorkflowError(err)
		logger.Error().Err(uwerr).Msg("failed to execute DPU Extension Service creation workflow on Site")
		return cutil.NewAPIErrorResponse(c, code, fmt.Sprintf("Failed to execute DPU Extension Service creation workflow on Site: %s", uwerr), nil)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		logger.Error().Err(err).Msg("error committing DPU Extension Service transaction to DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to create DPU Extension Service, DB transaction error", nil)
	}
	txCommitted = true

	// Best effort to update the DPU Extension Service version and versionInfo from received response
	updatedDpuExtensionService := dpuExtensionService

	if controllerDpuExtensionService != nil && controllerDpuExtensionService.LatestVersionInfo != nil {
		version := controllerDpuExtensionService.LatestVersionInfo.Version
		activeVersions := controllerDpuExtensionService.ActiveVersions
		versionInfo := &cdbm.DpuExtensionServiceVersionInfo{}
		versionInfo.FromProto(controllerDpuExtensionService.LatestVersionInfo, dpuExtensionService.Created)
		status := cdbm.DpuExtensionServiceStatusReady

		updatedDpuExtensionService, err = desDAO.Update(ctx, nil, cdbm.DpuExtensionServiceUpdateInput{
			DpuExtensionServiceID: dpuExtensionService.ID,
			Version:               &version,
			VersionInfo:           versionInfo,
			ActiveVersions:        activeVersions,
			Status:                &status,
		})
		if err != nil {
			logger.Error().Err(err).Msg("error updating DPU Extension Service record in DB")
			// Don't fail the request, the service will get updated on next inventory sync
		} else {
			statusDetail, serr := sdDAO.CreateFromParams(ctx, nil, dpuExtensionService.ID.String(), cdbm.DpuExtensionServiceStatusReady,
				cdb.GetStrPtr("DPU Extension Service is ready for deployment"))
			if serr != nil {
				logger.Error().Err(serr).Msg("error creating Status Detail DB entry")
			} else {
				statusDetails = append(statusDetails, *statusDetail)
			}
		}
	}

	// Create response
	apiDpuExtensionService := model.NewAPIDpuExtensionService(updatedDpuExtensionService, statusDetails)

	logger.Info().Msg("finishing API handler")
	return c.JSON(http.StatusCreated, apiDpuExtensionService)
}

// ~~~~~ GetAll Handler ~~~~~ //

// GetAllDpuExtensionServiceHandler is the API Handler for getting all DPU Extension Services
type GetAllDpuExtensionServiceHandler struct {
	dbSession  *cdb.Session
	tc         tclient.Client
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewGetAllDpuExtensionServiceHandler initializes and returns a new handler for getting all DPU Extension Services
func NewGetAllDpuExtensionServiceHandler(dbSession *cdb.Session, tc tclient.Client, cfg *config.Config) GetAllDpuExtensionServiceHandler {
	return GetAllDpuExtensionServiceHandler{
		dbSession:  dbSession,
		tc:         tc,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Get all DPU Extension Services
// @Description Get all DPU Extension Services for the current Tenant
// @Tags DPU Extension Service
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param siteId query string false "Filter by Site ID"
// @Param status query string false "Filter by Status"
// @Param query query string false "Search query for name, description and status"
// @Param includeRelation query string false "Related entities to include in response e.g. 'Site', 'Tenant'"
// @Param pageNumber query integer false "Page number of results returned"
// @Param pageSize query integer false "Number of results per page"
// @Param orderBy query string false "Order by field"
// @Success 200 {array} model.APIDpuExtensionService
// @Router /v2/org/{org}/nico/dpu-extension-service [get]
func (gadesh GetAllDpuExtensionServiceHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("GetAll", "DpuExtensionService", c, gadesh.tracerSpan)
	if handlerSpan != nil {
		defer handlerSpan.End()
	}

	// Is DB user missing?
	if dbUser == nil {
		logger.Error().Msg("invalid User object found in request context")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve current user", nil)
	}

	// Validate the tenant
	tenant, err := common.GetTenantForOrg(ctx, nil, gadesh.dbSession, org)
	if err != nil {
		if err == common.ErrOrgTenantNotFound {
			logger.Warn().Err(err).Msg("Org does not have a Tenant associated")
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Org does not have a Tenant associated", nil)
		}
		logger.Error().Err(err).Msg("unable to retrieve Tenant for org")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Tenant for org", nil)
	}

	// Validate role, only Tenant Admins are allowed to get DPU Extension Services
	ok := auth.ValidateUserRoles(dbUser, org, nil, auth.TenantAdminRole)
	if !ok {
		logger.Warn().Msg("user does not have Tenant Admin role with org, access denied")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User does not have Tenant Admin role with org", nil)
	}

	filterInput := cdbm.DpuExtensionServiceFilterInput{
		TenantIDs: []uuid.UUID{tenant.ID},
	}

	// Get Site ID from query param if specified
	siteIDStr := c.QueryParam("siteId")
	if siteIDStr != "" {
		site, err := common.GetSiteFromIDString(ctx, nil, siteIDStr, gadesh.dbSession)
		if err != nil {
			if errors.Is(err, cdb.ErrDoesNotExist) {
				return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Site specified in query does not exist", nil)
			}
			logger.Error().Err(err).Msg("error retrieving Site from DB")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Site specified in query, DB error", nil)
		}

		// Verify Tenant has access to Site
		tsDAO := cdbm.NewTenantSiteDAO(gadesh.dbSession)
		tenantSites, _, err := tsDAO.GetAll(
			ctx,
			nil,
			cdbm.TenantSiteFilterInput{
				TenantIDs: []uuid.UUID{tenant.ID},
				SiteIDs:   []uuid.UUID{site.ID},
			},
			paginator.PageInput{Limit: cdb.GetIntPtr(1)},
			nil,
		)
		if err != nil {
			logger.Error().Err(err).Msg("error retrieving TenantSite from DB")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to validate Site access for Tenant, DB error", nil)
		}
		if len(tenantSites) == 0 {
			logger.Warn().Msg("Tenant does not have access to Site specified in query")
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Tenant does not have access to Site specified in query", nil)
		}

		filterInput.SiteIDs = []uuid.UUID{site.ID}
		gadesh.tracerSpan.SetAttribute(handlerSpan, attribute.String("site_id", siteIDStr), logger)
	}

	// Get status from query param
	statusQuery := c.QueryParam("status")
	if statusQuery != "" {
		_, ok := cdbm.DpuExtensionServiceStatusMap[statusQuery]
		if !ok {
			logger.Warn().Msg(fmt.Sprintf("invalid value in status query: %v", statusQuery))
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid Status value in query", nil)
		}
		filterInput.Statuses = []string{statusQuery}
		gadesh.tracerSpan.SetAttribute(handlerSpan, attribute.String("status", statusQuery), logger)
	}

	// Get query text for full text search from query param
	searchQuery := common.GetSearchQuery(c)
	if searchQuery != nil {
		filterInput.SearchQuery = searchQuery
		gadesh.tracerSpan.SetAttribute(handlerSpan, attribute.String("query", *searchQuery), logger)
	}

	// Get and validate includeRelation params
	qParams := c.QueryParams()
	qIncludeRelations, errStr := common.GetAndValidateQueryRelations(qParams, cdbm.DpuExtensionServiceRelatedEntities)
	if errStr != "" {
		logger.Warn().Msg(errStr)
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, errStr, nil)
	}

	// Validate pagination request
	pageRequest := pagination.PageRequest{}
	err = c.Bind(&pageRequest)
	if err != nil {
		logger.Warn().Err(err).Msg("error binding pagination request data into API model")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Failed to parse request pagination data", nil)
	}

	// Validate pagination attributes
	err = pageRequest.Validate(cdbm.DpuExtensionServiceOrderByFields)
	if err != nil {
		logger.Warn().Err(err).Msg("error validating pagination request data")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Failed to validate pagination request data", err)
	}

	// Get DPU Extension Services from DB
	desDAO := cdbm.NewDpuExtensionServiceDAO(gadesh.dbSession)
	dpuExtensionServices, total, err := desDAO.GetAll(
		ctx,
		nil,
		filterInput,
		paginator.PageInput{
			Offset:  pageRequest.Offset,
			Limit:   pageRequest.Limit,
			OrderBy: pageRequest.OrderBy,
		},
		qIncludeRelations,
	)
	if err != nil {
		logger.Error().Err(err).Msg("error retrieving DPU Extension Services from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve DPU Extension Services, DB error", nil)
	}

	// Get status details for all DPU Extension Services
	sdDAO := cdbm.NewStatusDetailDAO(gadesh.dbSession)
	dpuExtensionServiceIDs := []string{}
	for _, des := range dpuExtensionServices {
		dpuExtensionServiceIDs = append(dpuExtensionServiceIDs, des.ID.String())
	}

	statusDetails, err := sdDAO.GetRecentByEntityIDs(ctx, nil, dpuExtensionServiceIDs, common.RECENT_STATUS_DETAIL_COUNT)
	if err != nil {
		logger.Error().Err(err).Msg("error retrieving Status Details from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Status Details, DB error", nil)
	}

	statusDetailsMap := map[string][]cdbm.StatusDetail{}
	for _, sd := range statusDetails {
		csd := sd
		statusDetailsMap[sd.EntityID] = append(statusDetailsMap[sd.EntityID], csd)
	}

	// Build API response
	apiDpuExtensionServices := []model.APIDpuExtensionService{}
	for _, des := range dpuExtensionServices {
		cdes := des
		sds := statusDetailsMap[des.ID.String()]
		apiDpuExtensionServices = append(apiDpuExtensionServices, *model.NewAPIDpuExtensionService(&cdes, sds))
	}

	// Create pagination response header
	pageResponse := pagination.NewPageResponse(*pageRequest.PageNumber, *pageRequest.PageSize, total, pageRequest.OrderByStr)
	pageHeader, err := json.Marshal(pageResponse)
	if err != nil {
		logger.Error().Err(err).Msg("error marshaling pagination response")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to generate pagination response header", nil)
	}

	c.Response().Header().Set(pagination.ResponseHeaderName, string(pageHeader))

	logger.Info().Msg("finishing API handler")

	return c.JSON(http.StatusOK, apiDpuExtensionServices)
}

// ~~~~~ Get Handler ~~~~~ //

// GetDpuExtensionServiceHandler is the API Handler for retrieving a DPU Extension Service
type GetDpuExtensionServiceHandler struct {
	dbSession  *cdb.Session
	tc         tclient.Client
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewGetDpuExtensionServiceHandler initializes and returns a new handler to retrieve DPU Extension Service
func NewGetDpuExtensionServiceHandler(dbSession *cdb.Session, tc tclient.Client, cfg *config.Config) GetDpuExtensionServiceHandler {
	return GetDpuExtensionServiceHandler{
		dbSession:  dbSession,
		tc:         tc,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Retrieve a DPU Extension Service
// @Description Retrieve a DPU Extension Service by ID for the current Tenant
// @Tags DPU Extension Service
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param dpuExtensionServiceId path string true "ID of DPU Extension Service"
// @Param includeRelation query string false "Related entities to include in response e.g. 'Site', 'Tenant'"
// @Success 200 {object} model.APIDpuExtensionService
// @Router /v2/org/{org}/nico/dpu-extension-service/{dpuExtensionServiceId} [get]
func (gdesh GetDpuExtensionServiceHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Get", "DpuExtensionService", c, gdesh.tracerSpan)
	if handlerSpan != nil {
		defer handlerSpan.End()
	}

	// Is DB user missing?
	if dbUser == nil {
		logger.Error().Msg("invalid User object found in request context")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve current user", nil)
	}

	// Validate the tenant
	tenant, err := common.GetTenantForOrg(ctx, nil, gdesh.dbSession, org)
	if err != nil {
		if err == common.ErrOrgTenantNotFound {
			logger.Warn().Err(err).Msg("Org does not have a Tenant associated")
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Org does not have a Tenant associated", nil)
		}
		logger.Error().Err(err).Msg("unable to retrieve Tenant for org")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Tenant for org", nil)
	}

	// Validate role, only Tenant Admins are allowed to get DPU Extension Service
	ok := auth.ValidateUserRoles(dbUser, org, nil, auth.TenantAdminRole)
	if !ok {
		logger.Warn().Msg("user does not have Tenant Admin role with org, access denied")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User does not have Tenant Admin role with org", nil)
	}

	// Get DPU Extension Service ID from URL param
	dpuExtensionServiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid DPU Extension Service ID in URL", nil)
	}

	logger = logger.With().Str("DPU Extension Service ID", dpuExtensionServiceID.String()).Logger()

	gdesh.tracerSpan.SetAttribute(handlerSpan, attribute.String("dpu_extension_service_id", dpuExtensionServiceID.String()), logger)

	// Get and validate includeRelation params
	qParams := c.QueryParams()
	qIncludeRelations, errStr := common.GetAndValidateQueryRelations(qParams, cdbm.DpuExtensionServiceRelatedEntities)
	if errStr != "" {
		logger.Warn().Msg(errStr)
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, errStr, nil)
	}

	// Get DPU Extension Service from DB by ID
	desDAO := cdbm.NewDpuExtensionServiceDAO(gdesh.dbSession)
	dpuExtensionService, err := desDAO.GetByID(ctx, nil, dpuExtensionServiceID, qIncludeRelations)
	if err != nil {
		if errors.Is(err, cdb.ErrDoesNotExist) {
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, fmt.Sprintf("Could not find DPU Extension Service with ID: %s", dpuExtensionServiceID.String()), nil)
		}
		logger.Error().Err(err).Msg("error retrieving DPU Extension Service from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve DPU Extension Service, DB error", nil)
	}

	// Validate that DPU Extension Service belongs to the Tenant
	if dpuExtensionService.TenantID != tenant.ID {
		logger.Warn().Msg("DPU Extension Service does not belong to current Tenant")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "DPU Extension Service does not belong to current Tenant", nil)
	}

	// Get status details
	sdDAO := cdbm.NewStatusDetailDAO(gdesh.dbSession)
	statusDetails, err := sdDAO.GetRecentByEntityIDs(ctx, nil, []string{dpuExtensionService.ID.String()}, common.RECENT_STATUS_DETAIL_COUNT)
	if err != nil {
		logger.Error().Err(err).Msg("error retrieving Status Details from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Status Details, DB error", nil)
	}

	// Create response
	apiDpuExtensionService := model.NewAPIDpuExtensionService(dpuExtensionService, statusDetails)

	logger.Info().Msg("finishing API handler")
	return c.JSON(http.StatusOK, apiDpuExtensionService)
}

// ~~~~~ Update Handler ~~~~~ //

// UpdateDpuExtensionServiceHandler is the API Handler for updating a DPU Extension Service
type UpdateDpuExtensionServiceHandler struct {
	dbSession  *cdb.Session
	tc         tclient.Client
	cfg        *config.Config
	scp        *sc.ClientPool
	tracerSpan *cutil.TracerSpan
}

// NewUpdateDpuExtensionServiceHandler initializes and returns a new handler for updating DPU Extension Service
func NewUpdateDpuExtensionServiceHandler(dbSession *cdb.Session, tc tclient.Client, scp *sc.ClientPool, cfg *config.Config) UpdateDpuExtensionServiceHandler {
	return UpdateDpuExtensionServiceHandler{
		dbSession:  dbSession,
		tc:         tc,
		scp:        scp,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Update a DPU Extension Service
// @Description Update a DPU Extension Service by ID. A new version will be created if data or credentials are modified.
// @Tags DPU Extension Service
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param dpuExtensionServiceId path string true "ID of DPU Extension Service"
// @Param message body model.APIDpuExtensionServiceUpdateRequest true "DPU Extension Service update request"
// @Success 200 {object} model.APIDpuExtensionService
// @Router /v2/org/{org}/nico/dpu-extension-service/{dpuExtensionServiceId} [patch]
func (udesh UpdateDpuExtensionServiceHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Update", "DpuExtensionService", c, udesh.tracerSpan)
	if handlerSpan != nil {
		defer handlerSpan.End()
	}

	// Is DB user missing?
	if dbUser == nil {
		logger.Error().Msg("invalid User object found in request context")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve current user", nil)
	}

	// Validate the tenant
	tenant, err := common.GetTenantForOrg(ctx, nil, udesh.dbSession, org)
	if err != nil {
		if err == common.ErrOrgTenantNotFound {
			logger.Warn().Err(err).Msg("Org does not have a Tenant associated")
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Org does not have a Tenant associated", nil)
		}
		logger.Error().Err(err).Msg("unable to retrieve Tenant for org")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Tenant for org", nil)
	}

	// Validate role, only Tenant Admins are allowed to update DPU Extension Service
	ok := auth.ValidateUserRoles(dbUser, org, nil, auth.TenantAdminRole)
	if !ok {
		logger.Warn().Msg("user does not have Tenant Admin role with org, access denied")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User does not have Tenant Admin role with org", nil)
	}

	// Get DPU Extension Service ID from URL param
	dpuExtensionServiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid DPU Extension Service ID in URL", nil)
	}
	logger = logger.With().Str("DPU Extension Service ID", dpuExtensionServiceID.String()).Logger()

	udesh.tracerSpan.SetAttribute(handlerSpan, attribute.String("dpu_extension_service_id", dpuExtensionServiceID.String()), logger)

	// Validate request
	// Bind request data to API model
	apiRequest := model.APIDpuExtensionServiceUpdateRequest{}
	err = c.Bind(&apiRequest)
	if err != nil {
		logger.Warn().Err(err).Msg("error binding request data into API model")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Failed to parse request data, potentially invalid structure", nil)
	}

	// Validate request attributes
	verr := apiRequest.Validate()
	if verr != nil {
		logger.Warn().Err(verr).Msg("error validating DPU Extension Service update request data")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Error validating DPU Extension Service update request data", verr)
	}

	// Get DPU Extension Service from DB by ID
	desDAO := cdbm.NewDpuExtensionServiceDAO(udesh.dbSession)
	dpuExtensionService, err := desDAO.GetByID(ctx, nil, dpuExtensionServiceID, []string{cdbm.SiteRelationName})
	if err != nil {
		if errors.Is(err, cdb.ErrDoesNotExist) {
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, fmt.Sprintf("Could not find DPU Extension Service with ID: %s", dpuExtensionServiceID.String()), nil)
		}
		logger.Error().Err(err).Msg("error retrieving DPU Extension Service from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve DPU Extension Service, DB error", nil)
	}

	// Validate that DPU Extension Service belongs to the Tenant
	if dpuExtensionService.TenantID != tenant.ID {
		logger.Warn().Msg("DPU Extension Service does not belong to current Tenant")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "DPU Extension Service does not belong to current Tenant", nil)
	}

	// Check if name is being updated and if it's unique
	if apiRequest.Name != nil && *apiRequest.Name != dpuExtensionService.Name {
		existingServices, _, err := desDAO.GetAll(
			ctx,
			nil,
			cdbm.DpuExtensionServiceFilterInput{
				TenantIDs: []uuid.UUID{tenant.ID},
				Names:     []string{*apiRequest.Name},
			},
			paginator.PageInput{Limit: cdb.GetIntPtr(1)},
			nil,
		)
		if err != nil {
			logger.Error().Err(err).Msg("error checking for duplicate DPU Extension Service name")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to validate DPU Extension Service name uniqueness, DB error", nil)
		}
		if len(existingServices) > 0 {
			logger.Warn().Msg("DPU Extension Service with this name already exists for Tenant")
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "DPU Extension Service with this name already exists", nil)
		}
	}

	// Start a db tx
	tx, err := cdb.BeginTx(ctx, udesh.dbSession, &sql.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("unable to start transaction")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to update DPU Extension Service, DB transaction error", nil)
	}
	txCommitted := false
	defer common.RollbackTx(ctx, tx, &txCommitted)

	// Update DPU Extension Service in DB
	var updateInput cdbm.DpuExtensionServiceUpdateInput
	updateInput.DpuExtensionServiceID = dpuExtensionService.ID

	if apiRequest.Name != nil {
		updateInput.Name = apiRequest.Name
	}

	if apiRequest.Description != nil {
		updateInput.Description = apiRequest.Description
	}

	updatedDpuExtensionService, err := desDAO.Update(ctx, tx, updateInput)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update DPU Extension Service record in DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to update DPU Extension Service, DB error", nil)
	}

	// Get updated status details
	sdDAO := cdbm.NewStatusDetailDAO(udesh.dbSession)
	statusDetails, err := sdDAO.GetRecentByEntityIDs(ctx, nil, []string{updatedDpuExtensionService.ID.String()}, common.RECENT_STATUS_DETAIL_COUNT)
	if err != nil {
		logger.Error().Err(err).Msg("error retrieving Status Details from DB")
		// Don't fail the request, just use what we have
	}

	// Trigger workflow to update DPU Extension Service
	updateDpuExtensionServiceRequest := &cwssaws.UpdateDpuExtensionServiceRequest{
		ServiceId:   updatedDpuExtensionService.ID.String(),
		ServiceName: apiRequest.Name,
		Description: apiRequest.Description,
	}

	if apiRequest.Data != nil {
		updateDpuExtensionServiceRequest.Data = *apiRequest.Data
	}

	if apiRequest.Credentials != nil {
		updateDpuExtensionServiceRequest.Credential = &cwssaws.DpuExtensionServiceCredential{
			RegistryUrl: apiRequest.Credentials.RegistryURL,
			Type: &cwssaws.DpuExtensionServiceCredential_UsernamePassword{
				UsernamePassword: &cwssaws.UsernamePassword{
					Username: *apiRequest.Credentials.Username,
					Password: *apiRequest.Credentials.Password,
				},
			},
		}
	}

	if apiRequest.Observability != nil {
		updateDpuExtensionServiceRequest.Observability = apiRequest.Observability.ToProto()
	}

	logger.Info().Msg("triggering DPU Extension Service update workflow on Site")

	workflowOptions := tclient.StartWorkflowOptions{
		ID:                       "dpu-extension-service-update-" + updatedDpuExtensionService.ID.String(),
		WorkflowExecutionTimeout: cutil.WorkflowExecutionTimeout,
		TaskQueue:                queue.SiteTaskQueue,
	}

	// Get the temporal client for the site we are working with
	stc, err := udesh.scp.GetClientByID(updatedDpuExtensionService.SiteID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve Temporal client for Site")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve workflow client for Site", nil)
	}

	// Add context deadlines
	ctxWithTimeout, cancel := context.WithTimeout(ctx, cutil.WorkflowContextTimeout)
	defer cancel()

	// Trigger Site workflow
	workflowRun, err := stc.ExecuteWorkflow(ctxWithTimeout, workflowOptions, "UpdateDpuExtensionService", updateDpuExtensionServiceRequest)
	if err != nil {
		logger.Error().Err(err).Msg("failed to schedule UpdateDpuExtensionService workflow on Site")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to schedule DPU Extension Service update workflow", nil)
	}

	workflowID := workflowRun.GetID()

	logger.Info().Str("Workflow ID", workflowID).Msg("executing UpdateDpuExtensionService workflow on Site in sync mode")

	// Execute sync workflow on Site
	var controllerDpuExtensionService *cwssaws.DpuExtensionService

	err = workflowRun.Get(ctxWithTimeout, &controllerDpuExtensionService)
	if err != nil {
		var timeoutErr *tp.TimeoutError
		if errors.As(err, &timeoutErr) {
			logger.Error().Err(err).Msg("timed out executing DPU Extension Service update workflow on Site")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Timed out executing DPU Extension Service update workflow on Site: %v", err), nil)
		}

		code, uwerr := common.UnwrapWorkflowError(err)
		logger.Error().Err(uwerr).Msg("failed to execute DPU Extension Service update workflow on Site")
		return cutil.NewAPIErrorResponse(c, code, fmt.Sprintf("Failed to execute DPU Extension Service update workflow on Site: %s", uwerr), nil)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		logger.Error().Err(err).Msg("error committing DPU Extension Service transaction to DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to update DPU Extension Service, DB transaction error", nil)
	}
	txCommitted = true

	// Best effort to update the DPU Extension Service version and versionInfo from received response
	reUpdatedDpuExtensionService := updatedDpuExtensionService

	if controllerDpuExtensionService != nil && controllerDpuExtensionService.LatestVersionInfo != nil {
		version := controllerDpuExtensionService.LatestVersionInfo.Version
		activeVersions := controllerDpuExtensionService.ActiveVersions
		versionInfo := &cdbm.DpuExtensionServiceVersionInfo{}
		versionInfo.FromProto(controllerDpuExtensionService.LatestVersionInfo, updatedDpuExtensionService.Updated)
		status := cdbm.DpuExtensionServiceStatusReady

		reUpdatedDpuExtensionService, err = desDAO.Update(ctx, nil, cdbm.DpuExtensionServiceUpdateInput{
			DpuExtensionServiceID: dpuExtensionService.ID,
			Version:               &version,
			VersionInfo:           versionInfo,
			ActiveVersions:        activeVersions,
			Status:                &status,
		})

		if err != nil {
			logger.Error().Err(err).Msg("error updating DPU Extension Service record in DB")
			// Don't fail the request, the service will get updated on next inventory sync
		}
	}

	// Create response
	apiDpuExtensionService := model.NewAPIDpuExtensionService(reUpdatedDpuExtensionService, statusDetails)

	logger.Info().Msg("finishing API handler")
	return c.JSON(http.StatusOK, apiDpuExtensionService)
}

// ~~~~~ Delete Handler ~~~~~ //

// DeleteDpuExtensionServiceHandler is the API Handler for deleting a DPU Extension Service
type DeleteDpuExtensionServiceHandler struct {
	dbSession  *cdb.Session
	tc         tclient.Client
	scp        *sc.ClientPool
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewDeleteDpuExtensionServiceHandler initializes and returns a new handler for deleting DPU Extension Service
func NewDeleteDpuExtensionServiceHandler(dbSession *cdb.Session, tc tclient.Client, scp *sc.ClientPool, cfg *config.Config) DeleteDpuExtensionServiceHandler {
	return DeleteDpuExtensionServiceHandler{
		dbSession:  dbSession,
		tc:         tc,
		scp:        scp,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Delete a DPU Extension Service
// @Description Delete a DPU Extension Service by ID. All versions will be deleted.
// @Tags DPU Extension Service
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param dpuExtensionServiceId path string true "ID of DPU Extension Service"
// @Success 204 "No Content"
// @Router /v2/org/{org}/nico/dpu-extension-service/{dpuExtensionServiceId} [delete]
func (ddesh DeleteDpuExtensionServiceHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("Delete", "DpuExtensionService", c, ddesh.tracerSpan)
	if handlerSpan != nil {
		defer handlerSpan.End()
	}

	// Is DB user missing?
	if dbUser == nil {
		logger.Error().Msg("invalid User object found in request context")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve current user", nil)
	}

	// Validate the tenant
	tenant, err := common.GetTenantForOrg(ctx, nil, ddesh.dbSession, org)
	if err != nil {
		if err == common.ErrOrgTenantNotFound {
			logger.Warn().Err(err).Msg("Org does not have a Tenant associated")
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Org does not have a Tenant associated", nil)
		}
		logger.Error().Err(err).Msg("unable to retrieve Tenant for org")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Tenant for org", nil)
	}

	// Validate role, only Tenant Admins are allowed to delete DPU Extension Service
	ok := auth.ValidateUserRoles(dbUser, org, nil, auth.TenantAdminRole)
	if !ok {
		logger.Warn().Msg("user does not have Tenant Admin role with org, access denied")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User does not have Tenant Admin role with org", nil)
	}

	// Get DPU Extension Service ID from URL param
	dpuExtensionServiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid DPU Extension Service ID in URL", nil)
	}
	logger = logger.With().Str("DPU Extension Service ID", dpuExtensionServiceID.String()).Logger()

	ddesh.tracerSpan.SetAttribute(handlerSpan, attribute.String("dpu_extension_service_id", dpuExtensionServiceID.String()), logger)

	// Get DPU Extension Service from DB by ID
	desDAO := cdbm.NewDpuExtensionServiceDAO(ddesh.dbSession)
	dpuExtensionService, err := desDAO.GetByID(ctx, nil, dpuExtensionServiceID, nil)
	if err != nil {
		if errors.Is(err, cdb.ErrDoesNotExist) {
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, fmt.Sprintf("Could not find DPU Extension Service with ID: %s", dpuExtensionServiceID.String()), nil)
		}
		logger.Error().Err(err).Msg("error retrieving DPU Extension Service from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve DPU Extension Service, DB error", nil)
	}

	// Validate that DPU Extension Service belongs to the Tenant
	if dpuExtensionService.TenantID != tenant.ID {
		logger.Warn().Msg("DPU Extension Service does not belong to current Tenant")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "DPU Extension Service does not belong to current Tenant", nil)
	}

	// Check if any deployments are active
	desdDAO := cdbm.NewDpuExtensionServiceDeploymentDAO(ddesh.dbSession)
	activeDeployments, _, err := desdDAO.GetAll(
		ctx,
		nil,
		cdbm.DpuExtensionServiceDeploymentFilterInput{
			DpuExtensionServiceIDs: []uuid.UUID{dpuExtensionService.ID},
		},
		paginator.PageInput{Limit: cdb.GetIntPtr(1)},
		nil,
	)
	if err != nil {
		logger.Error().Err(err).Msg("error checking for active deployments")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to check for active deployments, DB error", nil)
	}
	if len(activeDeployments) > 0 {
		logger.Warn().Msg("Cannot delete DPU Extension Service with active deployments")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Cannot delete DPU Extension Service with active deployments", nil)
	}

	// Start a db tx
	tx, err := cdb.BeginTx(ctx, ddesh.dbSession, &sql.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("unable to start transaction")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to delete DPU Extension Service, DB transaction error", nil)
	}
	txCommitted := false
	defer common.RollbackTx(ctx, tx, &txCommitted)

	// Update status to Deleting
	_, err = desDAO.Update(
		ctx,
		tx,
		cdbm.DpuExtensionServiceUpdateInput{
			DpuExtensionServiceID: dpuExtensionService.ID,
			Status:                cdb.GetStrPtr(cdbm.DpuExtensionServiceStatusDeleting),
		},
	)
	if err != nil {
		logger.Error().Err(err).Msg("unable to update DPU Extension Service status to Deleting")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to update DPU Extension Service status to Deleting, DB error", nil)
	}

	// Delete the DPU Extension Service
	err = desDAO.Delete(ctx, tx, dpuExtensionService.ID)
	if err != nil {
		logger.Error().Err(err).Msg("unable to delete DPU Extension Service")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to delete DPU Extension Service, DB error", nil)
	}

	// Trigger workflow to delete DPU Extension Service
	deleteDpuExtensionServiceRequest := &cwssaws.DeleteDpuExtensionServiceRequest{
		ServiceId: dpuExtensionService.ID.String(),
	}

	// Get the temporal client for the site we are working with
	stc, err := ddesh.scp.GetClientByID(dpuExtensionService.SiteID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve Temporal client for Site")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve workflow client for Site", nil)
	}

	workflowOptions := tclient.StartWorkflowOptions{
		ID:                       "dpu-extension-service-delete-" + dpuExtensionService.ID.String(),
		WorkflowExecutionTimeout: cutil.WorkflowExecutionTimeout,
		TaskQueue:                queue.SiteTaskQueue,
	}

	// Trigger Site workflow
	apiErr := common.ExecuteSyncWorkflow(ctx, logger, stc, "DeleteDpuExtensionService", workflowOptions, deleteDpuExtensionServiceRequest)
	if apiErr != nil {
		return cutil.NewAPIErrorResponse(c, apiErr.Code, apiErr.Message, apiErr.Data)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		logger.Error().Err(err).Msg("error committing DPU Extension Service delete transaction to DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to delete DPU Extension Service, DB transaction error", nil)
	}
	txCommitted = true

	logger.Info().Msg("finishing API handler")

	return c.NoContent(http.StatusNoContent)
}

// ~~~~~ Get Version Handler ~~~~~ //

// GetDpuExtensionServiceVersionHandler is the API Handler for retrieving a DPU Extension Service version
type GetDpuExtensionServiceVersionHandler struct {
	dbSession  *cdb.Session
	tc         tclient.Client
	scp        *sc.ClientPool
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewGetDpuExtensionServiceVersionHandler initializes and returns a new handler for retrieving DPU Extension Service version
func NewGetDpuExtensionServiceVersionHandler(dbSession *cdb.Session, tc tclient.Client, scp *sc.ClientPool, cfg *config.Config) GetDpuExtensionServiceVersionHandler {
	return GetDpuExtensionServiceVersionHandler{
		dbSession:  dbSession,
		tc:         tc,
		scp:        scp,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Retrieve a DPU Extension Service version
// @Description Retrieve a specific version of a DPU Extension Service for the current Tenant
// @Tags DPU Extension Service
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param dpuExtensionServiceId path string true "ID of DPU Extension Service"
// @Param versionId path string true "Version ID"
// @Success 200 {object} model.APIDpuExtensionServiceVersionInfo
// @Router /v2/org/{org}/nico/dpu-extension-service/{dpuExtensionServiceId}/version/{versionId} [get]
func (gdesvh GetDpuExtensionServiceVersionHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("GetVersion", "DpuExtensionService", c, gdesvh.tracerSpan)
	if handlerSpan != nil {
		defer handlerSpan.End()
	}

	// Is DB user missing?
	if dbUser == nil {
		logger.Error().Msg("invalid User object found in request context")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve current user", nil)
	}

	// Validate the tenant
	tenant, err := common.GetTenantForOrg(ctx, nil, gdesvh.dbSession, org)
	if err != nil {
		if err == common.ErrOrgTenantNotFound {
			logger.Warn().Err(err).Msg("Org does not have a Tenant associated")
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Org does not have a Tenant associated", nil)
		}
		logger.Error().Err(err).Msg("unable to retrieve Tenant for org")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Tenant for org", nil)
	}

	// Validate role, only Tenant Admins are allowed to get DPU Extension Service version
	ok := auth.ValidateUserRoles(dbUser, org, nil, auth.TenantAdminRole)
	if !ok {
		logger.Warn().Msg("user does not have Tenant Admin role with org, access denied")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User does not have Tenant Admin role with org", nil)
	}

	// Get DPU Extension Service ID from URL param
	dpuExtensionServiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid DPU Extension Service ID in URL", nil)
	}

	// Get version ID from URL param
	versionID := c.Param("version")

	logger = logger.With().
		Str("DPU Extension Service ID", dpuExtensionServiceID.String()).
		Str("Version ID", versionID).
		Logger()

	gdesvh.tracerSpan.SetAttribute(handlerSpan, attribute.String("dpu_extension_service_id", dpuExtensionServiceID.String()), logger)
	gdesvh.tracerSpan.SetAttribute(handlerSpan, attribute.String("version_id", versionID), logger)

	// Get DPU Extension Service from DB by ID
	desDAO := cdbm.NewDpuExtensionServiceDAO(gdesvh.dbSession)
	dpuExtensionService, err := desDAO.GetByID(ctx, nil, dpuExtensionServiceID, nil)
	if err != nil {
		if errors.Is(err, cdb.ErrDoesNotExist) {
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, fmt.Sprintf("Could not find DPU Extension Service with ID: %s", dpuExtensionServiceID.String()), nil)
		}
		logger.Error().Err(err).Msg("error retrieving DPU Extension Service from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve DPU Extension Service, DB error", nil)
	}

	// Validate that DPU Extension Service belongs to the Tenant
	if dpuExtensionService.TenantID != tenant.ID {
		logger.Warn().Msg("DPU Extension Service does not belong to current Tenant")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "DPU Extension Service does not belong to current Tenant", nil)
	}

	// Get version info from Site DPU Extension Service
	getDpuVersionInfoRequest := &cwssaws.GetDpuExtensionServiceVersionsInfoRequest{
		ServiceId: dpuExtensionService.ID.String(),
		Versions:  []string{versionID},
	}

	// Get the temporal client for the site we are working with
	stc, err := gdesvh.scp.GetClientByID(dpuExtensionService.SiteID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve Temporal client for Site")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve workflow client for Site", nil)
	}

	workflowOptions := tclient.StartWorkflowOptions{
		ID:                       "dpu-extension-service-get-versions-info-" + dpuExtensionService.ID.String() + "-" + versionID,
		WorkflowExecutionTimeout: cutil.WorkflowExecutionTimeout,
		TaskQueue:                queue.SiteTaskQueue,
	}

	// Trigger Site workflow
	// Add context deadlines
	ctxWithTimeout, cancel := context.WithTimeout(ctx, cutil.WorkflowContextTimeout)
	defer cancel()

	// Trigger Site workflow
	workflowRun, err := stc.ExecuteWorkflow(ctxWithTimeout, workflowOptions, "GetDpuExtensionServiceVersionsInfo", getDpuVersionInfoRequest)
	if err != nil {
		logger.Error().Err(err).Msg("failed to schedule DPU Extension Service version info retrieval workflow on Site")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to schedule DPU Extension Service version info retrieval workflow", nil)
	}

	workflowID := workflowRun.GetID()

	logger = logger.With().Str("Workflow ID", workflowID).Logger()

	logger.Info().Msg("executing sync Temporal workflow on Site")

	// Execute sync workflow on Site
	var versionInfos *cwssaws.DpuExtensionServiceVersionInfoList
	err = workflowRun.Get(ctxWithTimeout, &versionInfos)
	if err != nil {
		var timeoutErr *tp.TimeoutError
		if errors.As(err, &timeoutErr) {
			logger.Error().Err(err).Msg("timed out executing DPU Extension Service version info retrieval workflow on Site")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Timed out executing DPU Extension Service version info retrieval workflow on Site: %s", err), nil)
		}

		code, uwerr := common.UnwrapWorkflowError(err)
		logger.Error().Err(uwerr).Msg("failed to execute DPU Extension Service version info retrieval workflow on Site")
		return cutil.NewAPIErrorResponse(c, code, fmt.Sprintf("Failed to execute DPU Extension Service version info retrieval workflow on Site: %s", uwerr), nil)
	}

	if len(versionInfos.VersionInfos) == 0 {
		return cutil.NewAPIErrorResponse(c, http.StatusNotFound, "Could not find version info for DPU Extension Service", nil)
	}

	versionInfo := versionInfos.VersionInfos[0]

	apiVersionInfo := &model.APIDpuExtensionServiceVersionInfo{}
	apiVersionInfo.FromProto(versionInfo, dpuExtensionService.Updated)

	logger.Info().Msg("finishing API handler")

	return c.JSON(http.StatusOK, apiVersionInfo)
}

// ~~~~~ Delete Version Handler ~~~~~ //

// DeleteDpuExtensionServiceVersionHandler is the API Handler for deleting a DPU Extension Service version
type DeleteDpuExtensionServiceVersionHandler struct {
	dbSession  *cdb.Session
	tc         tclient.Client
	scp        *sc.ClientPool
	cfg        *config.Config
	tracerSpan *cutil.TracerSpan
}

// NewDeleteDpuExtensionServiceVersionHandler initializes and returns a new handler for deleting DPU Extension Service version
func NewDeleteDpuExtensionServiceVersionHandler(dbSession *cdb.Session, tc tclient.Client, scp *sc.ClientPool, cfg *config.Config) DeleteDpuExtensionServiceVersionHandler {
	return DeleteDpuExtensionServiceVersionHandler{
		dbSession:  dbSession,
		tc:         tc,
		scp:        scp,
		cfg:        cfg,
		tracerSpan: cutil.NewTracerSpan(),
	}
}

// Handle godoc
// @Summary Delete a DPU Extension Service version
// @Description Delete a specific version of a DPU Extension Service. The version being deleted cannot have active deployments.
// @Tags DPU Extension Service
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param org path string true "Name of NGC organization"
// @Param dpuExtensionServiceId path string true "ID of DPU Extension Service"
// @Param versionId path string true "Version ID"
// @Success 202 "Accepted"
// @Router /v2/org/{org}/nico/dpu-extension-service/{dpuExtensionServiceId}/version/{versionId} [delete]
func (ddesvh DeleteDpuExtensionServiceVersionHandler) Handle(c echo.Context) error {
	org, dbUser, ctx, logger, handlerSpan := common.SetupHandler("DeleteVersion", "DpuExtensionService", c, ddesvh.tracerSpan)
	if handlerSpan != nil {
		defer handlerSpan.End()
	}

	// Is DB user missing?
	if dbUser == nil {
		logger.Error().Msg("invalid User object found in request context")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve current user", nil)
	}

	// Validate the tenant
	tenant, err := common.GetTenantForOrg(ctx, nil, ddesvh.dbSession, org)
	if err != nil {
		if err == common.ErrOrgTenantNotFound {
			logger.Warn().Err(err).Msg("Org does not have a Tenant associated")
			return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Org does not have a Tenant associated", nil)
		}
		logger.Error().Err(err).Msg("unable to retrieve Tenant for org")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve Tenant for org", nil)
	}

	// Validate role, only Tenant Admins are allowed to delete DPU Extension Service version
	ok := auth.ValidateUserRoles(dbUser, org, nil, auth.TenantAdminRole)
	if !ok {
		logger.Warn().Msg("user does not have Tenant Admin role with org, access denied")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "User does not have Tenant Admin role with org", nil)
	}

	// Get DPU Extension Service ID from URL param
	dpuExtensionServiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Invalid DPU Extension Service ID in URL", nil)
	}

	// Get version ID from URL param
	versionID := c.Param("version")

	logger = logger.With().
		Str("DPU Extension Service ID", dpuExtensionServiceID.String()).
		Str("Version ID", versionID).
		Logger()

	ddesvh.tracerSpan.SetAttribute(handlerSpan, attribute.String("dpu_extension_service_id", dpuExtensionServiceID.String()), logger)
	ddesvh.tracerSpan.SetAttribute(handlerSpan, attribute.String("version_id", versionID), logger)

	// Get DPU Extension Service from DB by ID
	desDAO := cdbm.NewDpuExtensionServiceDAO(ddesvh.dbSession)
	dpuExtensionService, err := desDAO.GetByID(ctx, nil, dpuExtensionServiceID, nil)
	if err != nil {
		if errors.Is(err, cdb.ErrDoesNotExist) {
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, fmt.Sprintf("Could not find DPU Extension Service with ID: %s", dpuExtensionServiceID.String()), nil)
		}
		logger.Error().Err(err).Msg("error retrieving DPU Extension Service from DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve DPU Extension Service, DB error", nil)
	}

	// Validate that DPU Extension Service belongs to the Tenant
	if dpuExtensionService.TenantID != tenant.ID {
		logger.Warn().Msg("DPU Extension Service does not belong to current Tenant")
		return cutil.NewAPIErrorResponse(c, http.StatusForbidden, "DPU Extension Service does not belong to current Tenant", nil)
	}

	// Verify version exists
	if dpuExtensionService.Version == nil || *dpuExtensionService.Version != versionID {
		// Verify if version is in active versions list
		versionFound := false
		for _, version := range dpuExtensionService.ActiveVersions {
			if version == versionID {
				versionFound = true
				break
			}
		}
		if !versionFound {
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, fmt.Sprintf("Version: %s not found for DPU Extension Service: %s", versionID, dpuExtensionServiceID.String()), nil)
		}
	}

	// Check if this version is currently deployed
	desdDAO := cdbm.NewDpuExtensionServiceDeploymentDAO(ddesvh.dbSession)
	activeDeployments, _, err := desdDAO.GetAll(
		ctx,
		nil,
		cdbm.DpuExtensionServiceDeploymentFilterInput{
			DpuExtensionServiceIDs: []uuid.UUID{dpuExtensionService.ID},
			Versions:               []string{versionID},
		},
		paginator.PageInput{Limit: cdb.GetIntPtr(1)},
		nil,
	)
	if err != nil {
		logger.Error().Err(err).Msg("error checking for active deployments of version")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to check for active deployments, DB error", nil)
	}
	if len(activeDeployments) > 0 {
		logger.Warn().Msg("Cannot delete version with active deployments")
		return cutil.NewAPIErrorResponse(c, http.StatusBadRequest, "Cannot delete version with active deployments", nil)
	}

	// Start a db tx
	tx, err := cdb.BeginTx(ctx, ddesvh.dbSession, &sql.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("unable to start transaction")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to delete DPU Extension Service version, DB transaction error", nil)
	}
	txCommitted := false
	defer common.RollbackTx(ctx, tx, &txCommitted)

	remainingVersions := []string{}
	for _, version := range dpuExtensionService.ActiveVersions {
		if version != versionID {
			remainingVersions = append(remainingVersions, version)
		}
	}

	fetchLatestRemainingVersion := false

	// Check if this was the last version
	if len(remainingVersions) == 0 {
		logger.Info().Msg("since deleted version was the last version, deleting DPU Extension Service from DB")
		// Delete the DPU Extension Service record from DB
		err := desDAO.Delete(ctx, tx, dpuExtensionService.ID)
		if err != nil {
			logger.Error().Err(err).Msg("error deleting DPU Extension Service from DB")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to delete DPU Extension Service version, error deleting parent DPU Extension Service", nil)
		}
	} else if dpuExtensionService.Version != nil {
		// Update active versions
		_, err = desDAO.Update(ctx, tx, cdbm.DpuExtensionServiceUpdateInput{
			DpuExtensionServiceID: dpuExtensionService.ID,
			ActiveVersions:        remainingVersions,
			Status:                cdb.GetStrPtr(cdbm.DpuExtensionServiceStatusReady),
		})
		if err != nil {
			logger.Error().Err(err).Msg("error updating DPU Extension Service record in DB after deleting individual version")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to update active versions after DPU Extension Service version deletion, DB error", nil)
		}

		// If latest version is not equal to the remaining latest version, then fetch the latest remaining version
		if *dpuExtensionService.Version != remainingVersions[0] {
			fetchLatestRemainingVersion = true
		}
	} else {
		// DPU Extension Service doesn't have version field populated, so we need to fetch the latest remaining version
		fetchLatestRemainingVersion = true
	}

	if fetchLatestRemainingVersion && dpuExtensionService.VersionInfo != nil {
		// Clear version info since latest version was deleted and latest version info is now incorrect
		_, err = desDAO.Clear(ctx, tx, cdbm.DpuExtensionServiceClearInput{
			DpuExtensionServiceID: dpuExtensionService.ID,
			VersionInfo:           true,
		})
		if err != nil {
			logger.Error().Err(err).Msg("error clearing version info after DPU Extension Service version deletion")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to clear version info for deleted version, DB error", nil)
		}
	}

	// Trigger workflow to delete DPU Extension Service version
	deleteDpuExtensionServiceVersionRequest := &cwssaws.DeleteDpuExtensionServiceRequest{
		ServiceId: dpuExtensionService.ID.String(),
		Versions:  []string{versionID},
	}

	// Get the temporal client for the site we are working with
	stc, err := ddesvh.scp.GetClientByID(dpuExtensionService.SiteID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve Temporal client for Site")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve workflow client for Site", nil)
	}

	workflowOptions := tclient.StartWorkflowOptions{
		ID:                       "dpu-extension-service-delete-version-" + dpuExtensionService.ID.String() + "-" + versionID,
		WorkflowExecutionTimeout: cutil.WorkflowExecutionTimeout,
		TaskQueue:                queue.SiteTaskQueue,
	}

	// Trigger Site workflow
	apiErr := common.ExecuteSyncWorkflow(ctx, logger, stc, "DeleteDpuExtensionService", workflowOptions, deleteDpuExtensionServiceVersionRequest)
	if apiErr != nil {
		return cutil.NewAPIErrorResponse(c, apiErr.Code, apiErr.Message, apiErr.Data)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		logger.Error().Err(err).Msg("error committing DPU Extension Service transaction to DB")
		return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to delete DPU Extension Service version, DB transaction error", nil)
	}
	txCommitted = true

	// Best effort to update the DPU Extension Service record in REST layer with the latest remaining version
	if fetchLatestRemainingVersion {
		workflowOptions := tclient.StartWorkflowOptions{
			ID:                       "dpu-extension-service-get-versions-info-" + dpuExtensionService.ID.String() + "-" + remainingVersions[0],
			WorkflowExecutionTimeout: cutil.WorkflowExecutionTimeout,
			TaskQueue:                queue.SiteTaskQueue,
		}

		// Trigger Site workflow
		// Add context deadlines
		ctxWithTimeout, cancel := context.WithTimeout(ctx, cutil.WorkflowContextTimeout)
		defer cancel()

		getDpuVersionInfoRequest := &cwssaws.GetDpuExtensionServiceVersionsInfoRequest{
			ServiceId: dpuExtensionService.ID.String(),
			Versions:  []string{remainingVersions[0]},
		}

		// Trigger Site workflow
		workflowRun, err := stc.ExecuteWorkflow(ctxWithTimeout, workflowOptions, "GetDpuExtensionServiceVersionsInfo", getDpuVersionInfoRequest)
		if err != nil {
			logger.Error().Err(err).Msg("failed to schedule DPU Extension Service version info retrieval workflow on Site")
			return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, "Failed to schedule DPU Extension Service version info retrieval workflow", nil)
		}

		workflowID := workflowRun.GetID()

		logger = logger.With().Str("Workflow ID", workflowID).Logger()

		logger.Info().Msg("executing sync Temporal workflow on Site")

		// Execute sync workflow on Site
		var controllerVersionInfos *cwssaws.DpuExtensionServiceVersionInfoList
		err = workflowRun.Get(ctxWithTimeout, &controllerVersionInfos)
		if err != nil {
			var timeoutErr *tp.TimeoutError
			if errors.As(err, &timeoutErr) {
				logger.Error().Err(err).Msg("timed out executing DPU Extension Service version info retrieval workflow on Site")
				return cutil.NewAPIErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Timed out executing DPU Extension Service version info retrieval workflow on Site: %s", err), nil)
			}

			code, uwerr := common.UnwrapWorkflowError(err)
			logger.Error().Err(uwerr).Msg("failed to execute DPU Extension Service version info retrieval workflow on Site")
			return cutil.NewAPIErrorResponse(c, code, fmt.Sprintf("Failed to execute DPU Extension Service version info retrieval workflow on Site: %s", uwerr), nil)
		}

		if len(controllerVersionInfos.VersionInfos) == 0 {
			return cutil.NewAPIErrorResponse(c, http.StatusNotFound, "Could not find latest remaining version details for DPU Extension Service", nil)
		}

		controllerVersionInfo := controllerVersionInfos.VersionInfos[0]

		if controllerVersionInfo != nil {
			versionInfo := &cdbm.DpuExtensionServiceVersionInfo{}
			versionInfo.FromProto(controllerVersionInfo, dpuExtensionService.Updated)
			versionInfo.Version = remainingVersions[0]

			_, err = desDAO.Update(ctx, nil, cdbm.DpuExtensionServiceUpdateInput{
				DpuExtensionServiceID: dpuExtensionService.ID,
				Version:               &remainingVersions[0],
				VersionInfo:           versionInfo,
				ActiveVersions:        remainingVersions,
				Status:                cdb.GetStrPtr(cdbm.DpuExtensionServiceStatusReady),
			})
			if err != nil {
				logger.Error().Err(err).Msg("error updating DPU Extension Service record in DB after deleting individual version")
			}
		}
	}

	logger.Info().Msg("finishing API handler")

	return c.NoContent(http.StatusNoContent)
}
