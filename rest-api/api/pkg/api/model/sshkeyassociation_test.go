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

	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewAPISSHKeyAssociation(t *testing.T) {
	ska := cdbm.SSHKeyAssociation{
		ID:            uuid.New(),
		SSHKeyID:      uuid.New(),
		SSHKeyGroupID: uuid.New(),
		Created:       time.Now(),
		Updated:       time.Now(),
	}
	apiska := NewAPISSHKeyAssociation(&ska)
	assert.Equal(t, apiska.ID, ska.ID.String())
	assert.Equal(t, apiska.SSHKeyID, ska.SSHKeyID.String())
	assert.Equal(t, apiska.SSHKeyGroupID, ska.SSHKeyGroupID.String())
}
