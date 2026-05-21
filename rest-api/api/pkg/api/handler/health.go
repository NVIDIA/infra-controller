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
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/model"
)

// HealthCheckHandler is an API handler to return health status of the API server
type HealthCheckHandler struct{}

// NewHealthCheckHandler creates and returns a new handler
func NewHealthCheckHandler() HealthCheckHandler {
	return HealthCheckHandler{}
}

// Handle godoc
// @Summary Returns the health status of API server
// @Description Returns the health status of the API server
// @Tags health
// @Accept */*
// @Produce json
// @Success 200 {object} model.APIHealthCheck
// @Router /healthz [get]
func (hch HealthCheckHandler) Handle(c echo.Context) error {
	ahc := model.NewAPIHealthCheck(true, nil)
	return c.JSON(http.StatusOK, ahc)
}
