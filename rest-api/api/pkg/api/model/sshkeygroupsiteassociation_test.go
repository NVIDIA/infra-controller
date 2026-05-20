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

package model

import (
	"testing"
	"time"

	"github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewAPISSHKeyGroupSiteAssociation(t *testing.T) {
	skgsa := cdbm.SSHKeyGroupSiteAssociation{
		ID:            uuid.New(),
		SSHKeyGroupID: uuid.New(),
		SiteID:        uuid.New(),
		Version:       db.GetStrPtr("1234"),
		Status:        cdbm.SSHKeyGroupSiteAssociationStatusSyncing,
		Created:       time.Now(),
		Updated:       time.Now(),
	}
	apiskgsa := NewAPISSHKeyGroupSiteAssociation(&skgsa, nil)
	assert.Equal(t, apiskgsa.ControllerKeySetVersion, skgsa.Version)
	assert.Equal(t, apiskgsa.Status, skgsa.Status)
}
