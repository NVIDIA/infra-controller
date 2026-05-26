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
package firmwaremanager

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/NVIDIA/infra-controller-rest/powershelf-manager/pkg/objects/powershelf"
)

// ErrNotFound is returned by Get when no firmware update record exists for
// the requested (mac, component) key.
var ErrNotFound = errors.New("firmware update not found")

// FirmwareUpdateStore abstracts firmware update persistence so the Manager can
// operate against either a Postgres-backed or in-memory backend.
type FirmwareUpdateStore interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// CreateOrReplace upserts a firmware update keyed by (mac, component).
	CreateOrReplace(ctx context.Context, mac net.HardwareAddr, component powershelf.Component, versionFrom, versionTo string) (*FirmwareUpdateRecord, error)

	// Get retrieves the firmware update for (mac, component).
	// Returns ErrNotFound if no record exists.
	Get(ctx context.Context, mac net.HardwareAddr, component powershelf.Component) (*FirmwareUpdateRecord, error)

	// GetAllPending returns all non-terminal firmware updates.
	GetAllPending(ctx context.Context) ([]*FirmwareUpdateRecord, error)

	// SetState transitions a record to a new state with an optional error message.
	SetState(ctx context.Context, mac net.HardwareAddr, component powershelf.Component, newState powershelf.FirmwareState, errMsg string) error
}

// FirmwareUpdateRecord is a storage-agnostic representation of a firmware update
// used by the Manager. Both the Postgres and in-memory backends produce this type.
type FirmwareUpdateRecord struct {
	PmcMacAddress      net.HardwareAddr
	Component          powershelf.Component
	VersionFrom        string
	VersionTo          string
	State              powershelf.FirmwareState
	JobID              string
	ErrorMessage       string
	LastTransitionTime time.Time
	UpdatedAt          time.Time
}

// IsTerminal returns true if the firmware update is in a terminal state.
func (r *FirmwareUpdateRecord) IsTerminal() bool {
	return r.State == powershelf.FirmwareStateCompleted || r.State == powershelf.FirmwareStateFailed
}
