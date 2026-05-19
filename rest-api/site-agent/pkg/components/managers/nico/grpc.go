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

package nico

import (
	"sync"

	computils "github.com/NVIDIA/infra-controller-rest/site-agent/pkg/components/utils"
	"github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"
	"github.com/gogo/status"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
)

// checkCertsOnce is a local variable to ensure the go routine for checking if the certificate has changed only gets
// kicked off once even if creategRPC gets called multiple times
var checkCertsOnce sync.Once

func creategRPC() (conn *client.NICoCoreClient, err error) {
	// Initialize contextual logger
	logger := log.With().Str("Method", "CarbideClient.creategRPC").Logger()
	logger.Info().Msg("GRPC: Starting GRPC client")

	// Initialize the GRPC client configuration
	ManagerAccess.Data.EB.Managers.NICo.Client.Config = &client.NICoCoreClientConfig{
		Address:        ManagerAccess.Conf.EB.NICo.Address,
		Secure:         ManagerAccess.Conf.EB.NICo.Secure,
		ServerCAPath:   ManagerAccess.Conf.EB.NICo.ServerCAPath,
		SkipServerAuth: ManagerAccess.Conf.EB.NICo.SkipServerAuth,
		ClientCertPath: ManagerAccess.Conf.EB.NICo.ClientCertPath,
		ClientKeyPath:  ManagerAccess.Conf.EB.NICo.ClientKeyPath,
		ClientMetrics:  makeGrpcClientMetrics(),
	}

	logger.Info().Interface("GRPCConfig", ManagerAccess.Data.EB.Managers.NICo.Client.Config).Msg("Initializing GRPC client")

	// Get initial certificate MD5 hashes
	initialClientMD5, initialServerMD5, err := ManagerAccess.Data.EB.Managers.NICo.Client.GetInitialCertMD5()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get initial certificate MD5 hashes")
		return nil, err
	}
	newClient, err := client.NewNICoCoreClient(ManagerAccess.Data.EB.Managers.NICo.Client.Config)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize GRPC client")
		return nil, err
	}

	// Since this is initial creation, there's no old client to manage. SwapClient still used for consistency.
	_ = ManagerAccess.Data.EB.Managers.NICo.Client.SwapClient(newClient)
	logger.Info().Msg("Successfully initialized GRPC client")

	// Start the certificate check and reload routine in a background goroutine
	checkCertsOnce.Do(func() {
		go ManagerAccess.Data.EB.Managers.NICo.Client.CheckAndReloadCerts(initialClientMD5, initialServerMD5)
		logger.Info().Msg("Started certificate reload routine")
	})

	return ManagerAccess.Data.EB.Managers.NICo.GetClient(), nil
}

// CreateGRPCClient - creates the grpc connection handle
func (Carbide *API) CreateGRPCClient() error {
	// Initialize the GRPC client
	// We can handle advanced features later
	_, err := creategRPC()
	if err != nil {
		ManagerAccess.Data.EB.Managers.NICo.State.HealthStatus.Store(uint64(computils.CompUnhealthy))
	} else {
		ManagerAccess.Data.EB.Managers.NICo.State.HealthStatus.Store(uint64(computils.CompNotKnown))
	}

	return err
}

// GetGRPCClient - gets the grpc connection handle
func (Carbide *API) GetGRPCClient() *client.NICoCoreClient {
	return ManagerAccess.Data.EB.Managers.NICo.GetClient()
}

// isGRPCUp Is grpc connection functional
func isGRPCUp(c codes.Code) bool {
	switch c {
	case codes.Unavailable, codes.Unauthenticated:
		return false
	}
	return true
}

// UpdateGRPCClientState - updates carbide state
func (Carbide *API) UpdateGRPCClientState(err error) {
	defer computils.UpdateState(ManagerAccess.Data.EB)
	if err == nil {
		ManagerAccess.Data.EB.Managers.NICo.State.GrpcSucc.Inc()
		ManagerAccess.Data.EB.Managers.NICo.State.HealthStatus.Store(uint64(computils.CompHealthy))
		return
	}
	ManagerAccess.Data.EB.Managers.NICo.State.GrpcFail.Inc()
	ManagerAccess.Data.EB.Managers.NICo.State.Err = err.Error()
	log.Error().Err(err).Msg("GRPC: Failed to send request to GRPC server")
	st, ok := status.FromError(err)
	if ok {
		if !isGRPCUp(st.Code()) {
			ManagerAccess.Data.EB.Managers.NICo.State.HealthStatus.Store(uint64(computils.CompUnhealthy))
			log.Error().Err(err).Msg("GRPC: connection down")
		} else {
			log.Info().Msgf("GRPC application error %v", st.Code())
		}
	}
}
