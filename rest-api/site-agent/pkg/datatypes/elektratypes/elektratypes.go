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

package elektratypes

import (
	"os"

	"github.com/NVIDIA/infra-controller-rest/site-agent/pkg/conftypes"
	"github.com/NVIDIA/infra-controller-rest/site-agent/pkg/datatypes/managertypes"
	"github.com/rs/zerolog"
	"go.uber.org/atomic"
)

// Elektra is the main struct for the Elektra plugin
type Elektra struct {
	// Main structure of Elektra
	// All information is contained in this structure
	Managers *managertypes.Managers
	Conf     *conftypes.Config
	// HealthStatus current health state
	HealthStatus atomic.Uint64
	Version      string
	Log          zerolog.Logger
}

// NewElektraTypes - create new Elektra Type
func NewElektraTypes() *Elektra {
	return &Elektra{
		Version:  "0.0.1",
		Managers: managertypes.NewManagerType(),
		Conf:     conftypes.NewConfType(),
		Log:      zerolog.New(os.Stderr).With().Timestamp().Logger(),
	}
}
