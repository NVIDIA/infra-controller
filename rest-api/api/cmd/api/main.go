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

package main

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	tClient "go.temporal.io/sdk/client"

	cdb "github.com/NVIDIA/infra-controller-rest/db/pkg/db"

	"github.com/NVIDIA/infra-controller-rest/api/internal/config"
	capis "github.com/NVIDIA/infra-controller-rest/api/internal/server"

	sc "github.com/NVIDIA/infra-controller-rest/api/pkg/client/site"

	// Imports for API doc generation
	_ "github.com/NVIDIA/infra-controller-rest/api/pkg/api/model"
)

const (
	// ZerologMessageFieldName specifies the field name for log message
	ZerologMessageFieldName = "msg"
	// ZerologLevelFieldName specifies the field name for log level
	ZerologLevelFieldName = "type"
)

// @title NVIDIA NICo REST API
// @version 1.0
// @description NICo REST API allows you to manage datacenter resources from Cloud
// @termsOfService https://ngc.nvidia.com/legal/terms

// @license.name Proprietary

// @BasePath /
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.LevelFieldName = ZerologLevelFieldName
	zerolog.MessageFieldName = ZerologMessageFieldName

	cfg := config.NewConfig()
	defer cfg.Close()

	dbConfig := cfg.GetDBConfig()

	// Initialize DB connection
	dbSession, err := cdb.NewSession(context.Background(), dbConfig.Host, dbConfig.Port, dbConfig.Name, dbConfig.User, dbConfig.Password, "")
	if err != nil {
		log.Panic().Err(err).Msg("failed to initialize DB session")
	} else {
		defer dbSession.Close()
	}

	// Initialize Temporal client and namespace client
	// Client objects are expensive so they are only initialized once
	tcfg, err := cfg.GetTemporalConfig()

	if err != nil {
		log.Panic().Err(err).Msg("failed to get Temporal config")
	}

	tc, tnc, err := capis.InitTemporalClients(tcfg, cfg.GetTracingEnabled())

	if err != nil {
		log.Panic().Err(err).Msg("failed to create Temporal clients")
	} else {
		defer tc.Close()
		defer tnc.Close()
	}

	_, err = tc.CheckHealth(context.Background(), &tClient.CheckHealthRequest{})
	if err != nil {
		log.Panic().Err(err).Msg("failed to check Temporal health")
	}

	scp := sc.NewClientPool(tcfg)

	// Initialize API Echo instance
	e := capis.InitAPIServer(cfg, dbSession, tc, tnc, scp)

	mconfig := cfg.GetMetricsConfig()
	if mconfig.Enabled {
		// Initialize Prometheus Echo instance
		ep := capis.InitMetricsServer(e, cfg)

		// Start Prometheus server
		log.Info().Msg("starting Metrics server")
		go func() {
			ep.Logger.Fatal(ep.Start(mconfig.GetListenAddr()))
		}()
	}

	// Start main server
	log.Info().Msg("starting API server")
	e.Logger.Fatal(e.Start(":8388"))
}
