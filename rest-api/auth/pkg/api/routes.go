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
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/NVIDIA/infra-controller-rest/auth/pkg/authentication"
	"github.com/NVIDIA/infra-controller-rest/auth/pkg/config"

	cah "github.com/NVIDIA/infra-controller-rest/auth/pkg/api/handler"
)

// AuthRoute represents an authentication route
type AuthRoute struct {
	Method  string
	Path    string
	Handler AuthHandler
}

// AuthHandler represents an authentication handler
type AuthHandler interface {
	Handle(c echo.Context) error
}

// NewAuthRoutes creates new authentication routes using cloud-auth services
// This function provides a complete authentication API that can be used by any service
func NewAuthRoutes(keycloakConfig *config.KeycloakConfig) []AuthRoute {
	if keycloakConfig == nil {
		log.Error().Msg("keycloak config is not initialized, cannot create authentication routes")
		return nil
	}

	// Initialize Keycloak auth service
	keycloakAuth := authentication.NewKeycloakAuthService(keycloakConfig)

	// Create handlers
	loginHandler := cah.NewLoginHandler(keycloakAuth)
	callbackHandler := cah.NewCallbackHandler(keycloakAuth)
	logoutHandler := cah.NewLogoutHandler(keycloakAuth)
	refreshHandler := cah.NewRefreshTokenHandler(keycloakAuth)

	return []AuthRoute{
		{
			Method:  "POST",
			Path:    "/login",
			Handler: loginHandler,
		},
		{
			Method:  "POST",
			Path:    "/callback",
			Handler: callbackHandler,
		},
		{
			Method:  "POST",
			Path:    "/logout",
			Handler: logoutHandler,
		},
		{
			Method:  "POST",
			Path:    "/refresh",
			Handler: refreshHandler,
		},
	}
}
