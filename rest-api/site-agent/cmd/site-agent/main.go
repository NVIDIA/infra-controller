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
	"os"
	"os/signal"
	"syscall"

	"github.com/NVIDIA/infra-controller-rest/site-agent/pkg/metadata"

	components "github.com/NVIDIA/infra-controller-rest/site-agent/pkg/components"
	"github.com/NVIDIA/infra-controller-rest/site-agent/pkg/datatypes/elektratypes"
	"github.com/rs/zerolog/log"
)

// InitElektra initializes the Elektra site agent framework
func InitElektra() {
	// Initialize Elektra microservice
	log.Info().Msg("Elektra: Initializing Elektra service")

	// TODO: this is for verification that we can get version, will move it to a metric after
	log.Info().Msgf("Elektra: version=%s, time=%s", metadata.Version, metadata.BuildTime)

	// Initialize Elektra Data Structures
	elektraTypes := elektratypes.NewElektraTypes()

	// Initialize Elektra API
	api, initErr := components.NewElektraAPI(elektraTypes, false)
	if initErr != nil {
		log.Error().Err(initErr).Msg("Elektra: Failed to initialize Elektra API")
	} else {
		log.Info().Msg("Elektra: Successfully initialized Elektra API")
	}

	// Initialize Elektra Managers
	api.Init()

	// Start Elektra Managers
	api.Start()
}

func main() {
	InitElektra()
	// sleep
	// Wait forever
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-termChan:
		return
	}
}
