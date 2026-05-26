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

package nvswitchregistry

import (
	"context"

	"github.com/NVIDIA/infra-controller-rest/nvswitch-manager/pkg/objects/nvswitch"

	"github.com/google/uuid"
)

// Registry defines the interface for NV-Switch tray storage.
type Registry interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Register creates a new NV-Switch entry or updates an existing one.
	// Returns the UUID and whether it was newly created.
	Register(ctx context.Context, tray *nvswitch.NVSwitchTray) (uuid.UUID, bool, error)

	// Get retrieves an NV-Switch by UUID.
	Get(ctx context.Context, id uuid.UUID) (*nvswitch.NVSwitchTray, error)

	// GetByBMCMAC retrieves an NV-Switch by BMC MAC address.
	GetByBMCMAC(ctx context.Context, bmcMAC string) (*nvswitch.NVSwitchTray, error)

	// List returns all registered NV-Switches.
	List(ctx context.Context) ([]*nvswitch.NVSwitchTray, error)

	// Delete removes an NV-Switch by UUID.
	Delete(ctx context.Context, id uuid.UUID) error
}
