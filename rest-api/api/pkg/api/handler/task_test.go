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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/handler/util/common"
	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/model"
	sc "github.com/NVIDIA/infra-controller-rest/api/pkg/client/site"
	authz "github.com/NVIDIA/infra-controller-rest/auth/pkg/authorization"
	"github.com/NVIDIA/infra-controller-rest/common/pkg/otelecho"
	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
	flowv1 "github.com/NVIDIA/infra-controller-rest/workflow-schema/flow/protobuf/v1"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	oteltrace "go.opentelemetry.io/otel/trace"
	tmocks "go.temporal.io/sdk/mocks"
)

func TestGetTaskHandler_Handle(t *testing.T) {
	e := echo.New()
	dbSession := testRackInitDB(t)
	defer dbSession.Close()

	cfg := common.GetTestConfig()
	tcfg, _ := cfg.GetTemporalConfig()
	scp := sc.NewClientPool(tcfg)

	org := "test-org"
	_, site, _ := testRackSetupTestData(t, dbSession, org)

	siteNoFlow := &cdbm.Site{
		ID:                       uuid.New(),
		Name:                     "test-site-no-flow",
		Org:                      org,
		InfrastructureProviderID: site.InfrastructureProviderID,
		Status:                   cdbm.SiteStatusRegistered,
		Config:                   &cdbm.SiteConfig{},
	}
	_, err := dbSession.DB.NewInsert().Model(siteNoFlow).Exec(context.Background())
	assert.Nil(t, err)

	providerUser := testRackBuildUser(t, dbSession, "provider-user-task-get", org, []string{authz.ProviderAdminRole})
	tenantUser := testRackBuildUser(t, dbSession, "tenant-user-task-get", org, []string{authz.TenantAdminRole})

	handler := NewGetTaskHandler(dbSession, nil, scp, cfg)

	taskUUID := uuid.New().String()

	mockTask := &flowv1.Task{
		Id:          &flowv1.UUID{Id: taskUUID},
		Operation:   "power_on",
		RackId:      &flowv1.UUID{Id: uuid.New().String()},
		Description: "Power on rack",
		Status:      flowv1.TaskStatus_TASK_STATUS_RUNNING,
		Message:     "Processing",
	}

	tracer := oteltrace.NewNoopTracerProvider().Tracer("test")
	ctx := context.Background()

	tests := []struct {
		name           string
		reqOrg         string
		user           *cdbm.User
		taskUUID       string
		queryParams    map[string]string
		mockTasks      []*flowv1.Task
		expectedStatus int
	}{
		{
			name:     "success - get task by ID",
			reqOrg:   org,
			user:     providerUser,
			taskUUID: taskUUID,
			queryParams: map[string]string{
				"siteId": site.ID.String(),
			},
			mockTasks:      []*flowv1.Task{mockTask},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "failure - task not found (empty result)",
			reqOrg:   org,
			user:     providerUser,
			taskUUID: taskUUID,
			queryParams: map[string]string{
				"siteId": site.ID.String(),
			},
			mockTasks:      []*flowv1.Task{},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "failure - Flow not enabled on site",
			reqOrg:   org,
			user:     providerUser,
			taskUUID: taskUUID,
			queryParams: map[string]string{
				"siteId": siteNoFlow.ID.String(),
			},
			expectedStatus: http.StatusPreconditionFailed,
		},
		{
			name:        "failure - missing siteId",
			reqOrg:      org,
			user:        providerUser,
			taskUUID:    taskUUID,
			queryParams: map[string]string{
				// no siteId
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "failure - invalid siteId",
			reqOrg:   org,
			user:     providerUser,
			taskUUID: taskUUID,
			queryParams: map[string]string{
				"siteId": uuid.New().String(),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "failure - tenant access denied",
			reqOrg:   org,
			user:     tenantUser,
			taskUUID: taskUUID,
			queryParams: map[string]string{
				"siteId": site.ID.String(),
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTemporalClient := &tmocks.Client{}
			mockWorkflowRun := &tmocks.WorkflowRun{}
			mockWorkflowRun.On("GetID").Return("test-workflow-id")
			if tt.mockTasks != nil {
				mockWorkflowRun.Mock.On("Get", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					resp := args.Get(1).(*flowv1.GetTasksByIDsResponse)
					resp.Tasks = tt.mockTasks
				}).Return(nil)
			}
			mockTemporalClient.Mock.On("ExecuteWorkflow", mock.Anything, mock.Anything, "GetRackTask", mock.Anything).Return(mockWorkflowRun, nil)
			scp.IDClientMap[site.ID.String()] = mockTemporalClient

			q := url.Values{}
			for k, v := range tt.queryParams {
				q.Set(k, v)
			}
			path := fmt.Sprintf("/v2/org/%s/nico/rack/task/%s?%s", tt.reqOrg, tt.taskUUID, q.Encode())

			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			ec := e.NewContext(req, rec)
			ec.SetParamNames("orgName", "id")
			ec.SetParamValues(tt.reqOrg, tt.taskUUID)
			ec.Set("user", tt.user)

			ctx = context.WithValue(ctx, otelecho.TracerKey, tracer)
			ec.SetRequest(ec.Request().WithContext(ctx))

			err := handler.Handle(ec)

			if tt.expectedStatus != rec.Code {
				t.Errorf("GetTaskHandler.Handle() status = %v, want %v, response: %v, err: %v", rec.Code, tt.expectedStatus, rec.Body.String(), err)
			}

			require.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedStatus != http.StatusOK {
				return
			}

			var apiTask model.APIRackTask
			err = json.Unmarshal(rec.Body.Bytes(), &apiTask)
			assert.NoError(t, err)
			assert.Equal(t, taskUUID, apiTask.ID)
			assert.Equal(t, "Running", apiTask.Status)
			assert.Equal(t, "Power on rack", apiTask.Description)
			assert.Equal(t, "Processing", apiTask.Message)
		})
	}
}

func TestCancelTaskHandler_Handle(t *testing.T) {
	e := echo.New()
	dbSession := testRackInitDB(t)
	defer dbSession.Close()

	cfg := common.GetTestConfig()
	tcfg, _ := cfg.GetTemporalConfig()
	scp := sc.NewClientPool(tcfg)

	org := "test-org"
	_, site, _ := testRackSetupTestData(t, dbSession, org)

	siteNoFlow := &cdbm.Site{
		ID:                       uuid.New(),
		Name:                     "test-site-no-flow-cancel",
		Org:                      org,
		InfrastructureProviderID: site.InfrastructureProviderID,
		Status:                   cdbm.SiteStatusRegistered,
		Config:                   &cdbm.SiteConfig{},
	}
	_, err := dbSession.DB.NewInsert().Model(siteNoFlow).Exec(context.Background())
	assert.Nil(t, err)

	providerUser := testRackBuildUser(t, dbSession, "provider-user-task-cancel", org, []string{authz.ProviderAdminRole})
	tenantUser := testRackBuildUser(t, dbSession, "tenant-user-task-cancel", org, []string{authz.TenantAdminRole})

	handler := NewCancelTaskHandler(dbSession, nil, scp, cfg)

	taskUUID := uuid.New().String()

	cancelledTask := &flowv1.Task{
		Id:          &flowv1.UUID{Id: taskUUID},
		Operation:   "power_on",
		RackId:      &flowv1.UUID{Id: uuid.New().String()},
		Description: "Power on rack",
		Status:      flowv1.TaskStatus_TASK_STATUS_TERMINATED,
		Message:     "Cancelled by user",
	}

	tracer := oteltrace.NewNoopTracerProvider().Tracer("test")
	ctx := context.Background()

	tests := []struct {
		name           string
		reqOrg         string
		user           *cdbm.User
		taskUUID       string
		body           any
		mockTask       *flowv1.Task
		mockExecErr    error
		expectedStatus int
	}{
		{
			name:           "success - cancel task returns 202 Accepted",
			reqOrg:         org,
			user:           providerUser,
			taskUUID:       taskUUID,
			body:           model.APICancelTaskRequest{SiteID: site.ID.String()},
			mockTask:       cancelledTask,
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "failure - Flow not enabled on site",
			reqOrg:         org,
			user:           providerUser,
			taskUUID:       taskUUID,
			body:           model.APICancelTaskRequest{SiteID: siteNoFlow.ID.String()},
			expectedStatus: http.StatusPreconditionFailed,
		},
		{
			name:           "failure - missing siteId",
			reqOrg:         org,
			user:           providerUser,
			taskUUID:       taskUUID,
			body:           model.APICancelTaskRequest{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "failure - invalid siteId",
			reqOrg:         org,
			user:           providerUser,
			taskUUID:       taskUUID,
			body:           model.APICancelTaskRequest{SiteID: uuid.New().String()},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "failure - invalid task UUID",
			reqOrg:         org,
			user:           providerUser,
			taskUUID:       "not-a-uuid",
			body:           model.APICancelTaskRequest{SiteID: site.ID.String()},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "failure - tenant access denied",
			reqOrg:         org,
			user:           tenantUser,
			taskUUID:       taskUUID,
			body:           model.APICancelTaskRequest{SiteID: site.ID.String()},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "failure - workflow scheduling error",
			reqOrg:         org,
			user:           providerUser,
			taskUUID:       taskUUID,
			body:           model.APICancelTaskRequest{SiteID: site.ID.String()},
			mockExecErr:    errors.New("temporal scheduling failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTemporalClient := &tmocks.Client{}
			mockWorkflowRun := &tmocks.WorkflowRun{}
			mockWorkflowRun.On("GetID").Return("test-workflow-id")
			if tt.mockTask != nil {
				mockWorkflowRun.Mock.On("Get", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					resp := args.Get(1).(*flowv1.CancelTaskResponse)
					resp.Task = tt.mockTask
				}).Return(nil)
			}
			mockTemporalClient.Mock.On("ExecuteWorkflow", mock.Anything, mock.Anything, "CancelRackTask", mock.Anything).Return(mockWorkflowRun, tt.mockExecErr)
			scp.IDClientMap[site.ID.String()] = mockTemporalClient

			path := fmt.Sprintf("/v2/org/%s/nico/rack/task/%s/cancel", tt.reqOrg, tt.taskUUID)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(bodyBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			ec := e.NewContext(req, rec)
			ec.SetParamNames("orgName", "id")
			ec.SetParamValues(tt.reqOrg, tt.taskUUID)
			ec.Set("user", tt.user)

			ctx = context.WithValue(ctx, otelecho.TracerKey, tracer)
			ec.SetRequest(ec.Request().WithContext(ctx))

			err = handler.Handle(ec)

			if tt.expectedStatus != rec.Code {
				t.Errorf("CancelTaskHandler.Handle() status = %v, want %v, response: %v, err: %v", rec.Code, tt.expectedStatus, rec.Body.String(), err)
			}

			require.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedStatus != http.StatusAccepted {
				return
			}

			var apiTask model.APIRackTask
			err = json.Unmarshal(rec.Body.Bytes(), &apiTask)
			assert.NoError(t, err)
			assert.Equal(t, taskUUID, apiTask.ID)
			assert.Equal(t, "Terminated", apiTask.Status)
			assert.Equal(t, "Cancelled by user", apiTask.Message)
		})
	}
}
