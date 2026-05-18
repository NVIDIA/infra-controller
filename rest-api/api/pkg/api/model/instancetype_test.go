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
	"github.com/stretchr/testify/assert"
)

func TestNewAPIInstanceType(t *testing.T) {
	type args struct {
		dbit  *cdbm.InstanceType
		dbsds []cdbm.StatusDetail
		mcs   []cdbm.MachineCapability
		mit   []cdbm.MachineInstanceType
	}

	dbit := &cdbm.InstanceType{
		ID:                       uuid.New(),
		Name:                     "test-name",
		Description:              cdb.GetStrPtr("test-description"),
		ControllerMachineType:    cdb.GetStrPtr("test-controller-machine-type"),
		InfrastructureProviderID: uuid.New(),
		SiteID:                   cdb.GetUUIDPtr(uuid.New()),
		Status:                   "test-status",
		Created:                  time.Now(),
		Updated:                  time.Now(),
	}

	dbsd := cdbm.StatusDetail{
		ID:       uuid.New(),
		EntityID: dbit.ID.String(),
		Status:   "test-status",
		Message:  cdb.GetStrPtr("test-message"),
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	dbmc := cdbm.MachineCapability{
		ID:             uuid.New(),
		InstanceTypeID: &dbit.ID,
		Type:           "test-type",
		Name:           "test-name",
		Capacity:       cdb.GetStrPtr("test-capacity"),
		Count:          cdb.GetIntPtr(2),
		Created:        time.Now(),
		Updated:        time.Now(),
	}

	mit := cdbm.MachineInstanceType{
		ID:             uuid.New(),
		MachineID:      uuid.New().String(),
		InstanceTypeID: dbit.ID,
	}

	tests := []struct {
		name string
		args args
		want *APIInstanceType
	}{
		{
			name: "test new API Instance Type initializer",
			args: args{
				dbit:  dbit,
				dbsds: []cdbm.StatusDetail{dbsd},
				mcs:   []cdbm.MachineCapability{dbmc},
				mit:   []cdbm.MachineInstanceType{},
			},
			want: &APIInstanceType{
				ID:                       dbit.ID.String(),
				Name:                     dbit.Name,
				Description:              dbit.Description,
				ControllerMachineType:    dbit.ControllerMachineType,
				InfrastructureProviderID: dbit.InfrastructureProviderID.String(),
				SiteID:                   dbit.SiteID.String(),
				Status:                   dbit.Status,
				Created:                  dbit.Created,
				Updated:                  dbit.Updated,
				StatusHistory: []APIStatusDetail{
					{
						Status:  dbsd.Status,
						Message: dbsd.Message,
						Created: dbsd.Created,
						Updated: dbsd.Updated,
					},
				},
				MachineCapabilities: []APIMachineCapability{
					{
						Type:     dbmc.Type,
						Name:     dbmc.Name,
						Capacity: dbmc.Capacity,
						Count:    dbmc.Count,
					},
				},
				MachineInstanceTypes: []APIMachineInstanceType{},
			},
		},
		{
			name: "test new API Instance Type initializer with deprecation",
			args: args{
				dbit:  dbit,
				dbsds: []cdbm.StatusDetail{dbsd},
				mcs:   []cdbm.MachineCapability{dbmc},
				mit:   []cdbm.MachineInstanceType{mit},
			},
			want: func() *APIInstanceType {
				expected := &APIInstanceType{
					ID:                       dbit.ID.String(),
					Name:                     dbit.Name,
					Description:              dbit.Description,
					ControllerMachineType:    dbit.ControllerMachineType,
					InfrastructureProviderID: dbit.InfrastructureProviderID.String(),
					SiteID:                   dbit.SiteID.String(),
					Status:                   dbit.Status,
					Created:                  dbit.Created,
					Updated:                  dbit.Updated,
					StatusHistory: []APIStatusDetail{
						{
							Status:  dbsd.Status,
							Message: dbsd.Message,
							Created: dbsd.Created,
							Updated: dbsd.Updated,
						},
					},
					MachineCapabilities: []APIMachineCapability{
						{
							Type:     dbmc.Type,
							Name:     dbmc.Name,
							Capacity: dbmc.Capacity,
							Count:    dbmc.Count,
						},
					},
					MachineInstanceTypes: []APIMachineInstanceType{
						*NewAPIMachineInstanceType(&mit),
					},
				}

				return expected
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAPIInstanceType(tt.args.dbit, tt.args.dbsds, tt.args.mcs, tt.args.mit, nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAPIInstanceType() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAPIInstanceTypeCreateRequest_Validate(t *testing.T) {
	type fields struct {
		Name                  string
		Description           *string
		SiteID                string
		Labels                map[string]string
		ControllerMachineType *string
		MachineCapabilities   []APIMachineCapability
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test valid Instance Type create request",
			fields: fields{
				Name:        "test-name",
				Description: cdb.GetStrPtr("test-description"),
				SiteID:      uuid.New().String(),
				Labels: map[string]string{
					"name":        "a-nv100-instance",
					"description": "",
				},
				ControllerMachineType: cdb.GetStrPtr("test-controller-machine-type"),
				MachineCapabilities: []APIMachineCapability{
					{
						Type:     cdbm.MachineCapabilityTypeCPU,
						Name:     "AMD Opteron Series x10",
						Capacity: cdb.GetStrPtr("3.0GHz"),
						Count:    cdb.GetIntPtr(2),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test invalid Instance Type create request - invalid Site ID",
			fields: fields{
				Name:                  "test-name",
				Description:           cdb.GetStrPtr("test-description"),
				SiteID:                "",
				ControllerMachineType: cdb.GetStrPtr("test-controller-machine-type"),
				MachineCapabilities: []APIMachineCapability{
					{
						Type:     cdbm.MachineCapabilityTypeCPU,
						Name:     "AMD Opteron Series x10",
						Capacity: cdb.GetStrPtr("3.0GHz"),
						Count:    cdb.GetIntPtr(2),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test invalid Instance Type create request - invalid Labels",
			fields: fields{
				Name:        "test-name",
				Description: cdb.GetStrPtr("test-description"),
				SiteID:      uuid.New().String(),
				Labels: map[string]string{
					"name": "a-nv100-instance",
					"":     "test",
				},
				ControllerMachineType: cdb.GetStrPtr("test-controller-machine-type"),
				MachineCapabilities: []APIMachineCapability{
					{
						Type:     cdbm.MachineCapabilityTypeCPU,
						Name:     "AMD Opteron Series x10",
						Capacity: cdb.GetStrPtr("3.0GHz"),
						Count:    cdb.GetIntPtr(2),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test invalid Instance Type create request - invalid Machine Capability type",
			fields: fields{
				Name:                  "test-name",
				Description:           cdb.GetStrPtr("test-description"),
				SiteID:                uuid.New().String(),
				ControllerMachineType: cdb.GetStrPtr("test-controller-machine-type"),
				MachineCapabilities: []APIMachineCapability{
					{
						Type:     "test-type",
						Name:     "test-name",
						Capacity: cdb.GetStrPtr("test-capacity"),
						Count:    cdb.GetIntPtr(1),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test invalid Instance Type create request - multiple Machine Capability specified with same name",
			fields: fields{
				Name:                  "test-name",
				Description:           cdb.GetStrPtr("test-description"),
				SiteID:                uuid.New().String(),
				ControllerMachineType: cdb.GetStrPtr("test-controller-machine-type"),
				MachineCapabilities: []APIMachineCapability{
					{
						Type:  cdbm.MachineCapabilityTypeCPU,
						Name:  "test-name",
						Count: cdb.GetIntPtr(1),
					},
					{
						Type:     cdbm.MachineCapabilityTypeCPU,
						Name:     "test-name",
						Capacity: cdb.GetStrPtr("test-capacity"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test invalid Instance Type create request - invalid Machine Capability device type",
			fields: fields{
				Name:                  "test-name",
				Description:           cdb.GetStrPtr("test-description"),
				SiteID:                uuid.New().String(),
				ControllerMachineType: cdb.GetStrPtr("test-controller-machine-type"),
				MachineCapabilities: []APIMachineCapability{
					{
						Type:       cdbm.MachineCapabilityTypeNetwork,
						Name:       "test-name",
						DeviceType: cdb.GetStrPtr("test-device-type"),
						Count:      cdb.GetIntPtr(1),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itcr := APIInstanceTypeCreateRequest{
				Name:                  tt.fields.Name,
				Description:           tt.fields.Description,
				SiteID:                tt.fields.SiteID,
				Labels:                tt.fields.Labels,
				ControllerMachineType: tt.fields.ControllerMachineType,
				MachineCapabilities:   tt.fields.MachineCapabilities,
			}
			err := itcr.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAPIInstanceTypeUpdateRequest_Validate(t *testing.T) {
	type fields struct {
		Name        *string
		Description *string
		Labels      map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test valid Instance Type update request",
			fields: fields{
				Name:        cdb.GetStrPtr("test-name"),
				Description: cdb.GetStrPtr("test-description"),
				Labels: map[string]string{
					"name":        "a-nv100-instance",
					"description": "",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itur := APIInstanceTypeUpdateRequest{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Labels:      tt.fields.Labels,
			}
			err := itur.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
