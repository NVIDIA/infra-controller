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

package site

import (
	"os"
	"reflect"
	"testing"

	"github.com/NVIDIA/infra-controller-rest/api/internal/config"
	cconfig "github.com/NVIDIA/infra-controller-rest/common/pkg/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	temporalClient "go.temporal.io/sdk/client"
)

func TestNewClientPool(t *testing.T) {
	type args struct {
		tcfg *cconfig.TemporalConfig
	}

	keyPath, certPath := config.SetupTestCerts(t)
	defer os.Remove(keyPath)
	defer os.Remove(certPath)

	cfg := config.NewConfig()
	cfg.SetTemporalCertPath(certPath)
	cfg.SetTemporalKeyPath(keyPath)
	cfg.SetTemporalCaPath(certPath)

	tcfg, err := cfg.GetTemporalConfig()
	assert.NoError(t, err)

	tests := []struct {
		name string
		args args
		want *ClientPool
	}{
		{
			name: "test Site client pool initializer",
			args: args{
				tcfg: tcfg,
			},
			want: &ClientPool{
				tcfg:        tcfg,
				IDClientMap: map[string]temporalClient.Client{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClientPool(tt.args.tcfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSitePool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientPool_GetClientByID(t *testing.T) {
	type fields struct {
		tcfg *cconfig.TemporalConfig
	}
	type args struct {
		siteID uuid.UUID
	}

	keyPath, certPath := config.SetupTestCerts(t)
	defer os.Remove(keyPath)
	defer os.Remove(certPath)

	cfg := config.NewConfig()
	cfg.SetTemporalCertPath(certPath)
	cfg.SetTemporalKeyPath(keyPath)
	cfg.SetTemporalCaPath(certPath)

	tcfg, err := cfg.GetTemporalConfig()
	assert.NoError(t, err)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    temporalClient.Client
		wantErr bool
	}{
		{
			name: "test retrieving client for given site ID",
			fields: fields{
				tcfg: tcfg,
			},
			args: args{
				siteID: uuid.New(),
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewClientPool(tt.fields.tcfg)
			_, err := cp.GetClientByID(tt.args.siteID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientPool.GetClientByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
