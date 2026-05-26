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

package elektra

import (
	zlog "github.com/rs/zerolog/log"

	"github.com/NVIDIA/infra-controller-rest/site-agent/pkg/components/config"
	"github.com/NVIDIA/infra-controller-rest/site-agent/pkg/components/managers"
	"github.com/NVIDIA/infra-controller-rest/site-agent/pkg/datatypes/elektratypes"
)

// Interface - Managers' interface
type Interface interface {
	Managers() managers.Manager
}

// Elektra - Managers struct
type Elektra struct {
	manager *managers.Manager
}

// Init - initializes the cluster
func (Cluster *Elektra) Init() (err error) {
	zlog.Info().Msg("Elektra: Initializing Elektra cluster")
	Cluster.Managers().Init()
	return nil
}

// Start () Start the Cluster
func (Cluster *Elektra) Start() (err error) {
	zlog.Info().Msg("Elektra: Starting Elektra cluster")
	Cluster.Managers().Start()
	return nil
}

// Managers () Instantiate the Managers
func (Cluster *Elektra) Managers() *managers.Manager {
	return Cluster.manager
}

// NewElektraAPI - Instantiate new struct
func NewElektraAPI(superElektra *elektratypes.Elektra, utMode bool) (*Elektra, error) {
	zlog.Info().Msg("Elektra: Initializing Config Manager")
	var eb Elektra
	var err error
	// Initialize Global Config
	// Load configuration
	if superElektra != nil {
		// Configuration
		zlog.Info().Msg("Elektra: Loading configuration")
		superElektra.Conf = config.NewElektraConfig(utMode)
		eb.manager, err = managers.NewInstance(superElektra)
		zlog.Info().Interface("config", superElektra.Conf).Msg("Elektra: Config Manager initialized")
	}

	return &eb, err
}
