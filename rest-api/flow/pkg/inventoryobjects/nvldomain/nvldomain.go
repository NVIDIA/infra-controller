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

package nvldomain

import (
	"errors"

	identifier "github.com/NVIDIA/infra-controller-rest/flow/pkg/common/Identifier"
	"github.com/google/uuid"
)

type NVLDomain struct {
	Identifier      identifier.Identifier   `json:"identifier"`
	RackIdentifiers []identifier.Identifier `json:"rack_identifiers"`
}

func (d *NVLDomain) Validate() error {
	if d == nil {
		return errors.New("nvl domain is not specfied")
	}

	return d.Identifier.Validate()
}

func (d *NVLDomain) ID() uuid.UUID {
	return d.Identifier.ID
}

func (d *NVLDomain) Name() string {
	return d.Identifier.Name
}
