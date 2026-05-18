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

// MetadataHandler is an API handler to return system information about the API
type MetadataHandler struct{}

// NewMetadataHandler creates and returns a new handler
func NewMetadataHandler() MetadataHandler {
	return MetadataHandler{}
}

// Handle godoc
// @Summary Returns system information about the API
// @Description Returns system information about the API
// @Tags metadata
// @Accept */*
// @Produce json
// @Success 200 {object} model.APIMetadata
// @Router /v2/org/{org}/nico/metadata [get]
func (mdh MetadataHandler) Handle(c echo.Context) error {
	amd := model.NewAPIMetadata()
	return c.JSON(http.StatusOK, amd)
}
