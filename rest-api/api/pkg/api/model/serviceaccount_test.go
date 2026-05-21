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

	cdb "github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewAPIServiceAccount(t *testing.T) {
	type args struct {
		serviceAccountEnabled bool
		dbProvider            *cdbm.InfrastructureProvider
		dbTenant              *cdbm.Tenant
	}

	dbProvider := &cdbm.InfrastructureProvider{
		ID: uuid.New(),
	}
	dbTenant := &cdbm.Tenant{
		ID: uuid.New(),
	}

	tests := []struct {
		name string
		args args
		want *APIServiceAccount
	}{
		{
			name: "test NewAPIServiceAccount with service account enabled",
			args: args{
				serviceAccountEnabled: true,
				dbProvider:            dbProvider,
				dbTenant:              dbTenant,
			},
			want: &APIServiceAccount{
				Enabled:                  true,
				InfrastructureProviderID: cdb.GetStrPtr(dbProvider.ID.String()),
				TenantID:                 cdb.GetStrPtr(dbTenant.ID.String()),
			},
		},
		{
			name: "test NewAPIServiceAccount with service account disabled",
			args: args{
				serviceAccountEnabled: false,
			},
			want: &APIServiceAccount{
				Enabled: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAPIServiceAccount(tt.args.serviceAccountEnabled, tt.args.dbProvider, tt.args.dbTenant)
			assert.Equal(t, tt.want, got)
		})
	}
}
