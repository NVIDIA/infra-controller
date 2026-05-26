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
	"reflect"
	"testing"
	"time"

	cdb "github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	cdbm "github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
	"github.com/google/uuid"
)

func TestNewAPIInfrastructureProvider(t *testing.T) {
	type args struct {
		dbip *cdbm.InfrastructureProvider
	}

	dbip := &cdbm.InfrastructureProvider{
		ID:             uuid.New(),
		Name:           "test-infrastructure-provider",
		DisplayName:    nil,
		Org:            "test-org",
		OrgDisplayName: cdb.GetStrPtr("Org Display name"),
		Created:        time.Now(),
		Updated:        time.Now(),
	}

	ipAPIInfrastructureProvider := APIInfrastructureProvider{
		ID:             dbip.ID.String(),
		Org:            dbip.Org,
		OrgDisplayName: dbip.OrgDisplayName,
		Created:        dbip.Created,
		Updated:        dbip.Updated,
	}

	tests := []struct {
		name string
		args args
		want *APIInfrastructureProvider
	}{
		{
			name: "test initializing API model for Infrastructure Provider",
			args: args{
				dbip: dbip,
			},
			want: &ipAPIInfrastructureProvider,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAPIInfrastructureProvider(tt.args.dbip); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAPIInfrastructureProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAPIInfrastructureProviderSummary(t *testing.T) {
	dbip := &cdbm.InfrastructureProvider{
		ID:             uuid.New(),
		Name:           "test-infrastructure-provider",
		DisplayName:    nil,
		Org:            "test-org",
		OrgDisplayName: cdb.GetStrPtr("Org Display name"),
		Created:        time.Now(),
		Updated:        time.Now(),
	}

	type args struct {
		dbip *cdbm.InfrastructureProvider
	}
	tests := []struct {
		name string
		args args
		want *APIInfrastructureProviderSummary
	}{
		{
			name: "test init API summary model for Infrastructure Provider",
			args: args{
				dbip: dbip,
			},
			want: &APIInfrastructureProviderSummary{
				Org:            dbip.Org,
				OrgDisplayName: dbip.OrgDisplayName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAPIInfrastructureProviderSummary(tt.args.dbip); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAPIInfrastructureProviderSummary() = %v, want %v", got, tt.want)
			}
		})
	}
}
