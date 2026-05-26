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

package processors

import (
	"github.com/NVIDIA/infra-controller-rest/auth/pkg/config"
	commonConfig "github.com/NVIDIA/infra-controller-rest/common/pkg/config"
	cdb "github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	temporalClient "go.temporal.io/sdk/client"
)

// NewKeycloakProcessor creates a new Keycloak token processor
func NewKeycloakProcessor(dbSession *cdb.Session, kcfg *config.KeycloakConfig) config.TokenProcessor {
	return &KeycloakProcessor{
		dbSession:      dbSession,
		keycloakConfig: kcfg,
	}
}

// NewSSAProcessor creates a new SSA token processor
func NewSSAProcessor(dbSession *cdb.Session) config.TokenProcessor {
	return &SSAProcessor{
		dbSession: dbSession,
	}
}

// NewKASProcessor creates a new KAS token processor
func NewKASProcessor(dbSession *cdb.Session, tc temporalClient.Client, encCfg *commonConfig.PayloadEncryptionConfig) config.TokenProcessor {
	return &KASProcessor{
		dbSession: dbSession,
		tc:        tc,
		encCfg:    encCfg,
	}
}

// NewCustomProcessor creates a new custom token processor
func NewCustomProcessor(dbSession *cdb.Session) config.TokenProcessor {
	return &CustomProcessor{
		dbSession: dbSession,
	}
}

// InitializeProcessors sets up all token processors in the JWTOriginConfig
func InitializeProcessors(joCfg *config.JWTOriginConfig, dbSession *cdb.Session, tc temporalClient.Client, encCfg *commonConfig.PayloadEncryptionConfig, kcfg *config.KeycloakConfig) {
	for _, origin := range []string{config.TokenOriginKeycloak, config.TokenOriginKasSsa, config.TokenOriginKasLegacy, config.TokenOriginCustom} {
		switch origin {
		case config.TokenOriginKeycloak:
			processor := NewKeycloakProcessor(dbSession, kcfg)
			joCfg.SetProcessorForOrigin(origin, processor)
		case config.TokenOriginKasSsa:
			processor := NewSSAProcessor(dbSession)
			joCfg.SetProcessorForOrigin(origin, processor)
		case config.TokenOriginKasLegacy:
			processor := NewKASProcessor(dbSession, tc, encCfg)
			joCfg.SetProcessorForOrigin(origin, processor)
		case config.TokenOriginCustom:
			processor := NewCustomProcessor(dbSession)
			joCfg.SetProcessorForOrigin(origin, processor)
		}
	}
}
