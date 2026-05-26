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

package common

import (
	"errors"
	"fmt"
	"strings"

	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

// Target represents a batch of components of the same type for activity execution.
// Workflow passes only component IDs to activity (not full objects).
type Target struct {
	Type         devicetypes.ComponentType
	ComponentIDs []string
}

// Validate returns an error if the Target has an unknown component type or no component IDs.
func (t *Target) Validate() error {
	if t.Type == devicetypes.ComponentTypeUnknown {
		return errors.New("component type is unknown")
	}

	if len(t.ComponentIDs) == 0 {
		return errors.New("component IDs are required")
	}

	return nil
}

// String returns a human-readable representation for logging.
func (t *Target) String() string {
	return fmt.Sprintf(
		"[type: %s, component_ids: %s]",
		devicetypes.ComponentTypeToString(t.Type),
		strings.Join(t.ComponentIDs, ","),
	)
}
