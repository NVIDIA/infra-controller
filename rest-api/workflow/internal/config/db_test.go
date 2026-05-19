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

package config

import (
	"reflect"
	"testing"

	cconfig "github.com/NVIDIA/infra-controller-rest/common/pkg/config"
)

func TestNewDBConfig(t *testing.T) {
	type args struct {
		host     string
		port     int
		name     string
		user     string
		password string
	}

	dbcfg := cconfig.DBConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "forge",
		User:     "forge",
		Password: "test123",
	}

	tests := []struct {
		name string
		args args
		want *cconfig.DBConfig
	}{
		{
			name: "initialize database config",
			args: args{
				host:     dbcfg.Host,
				port:     dbcfg.Port,
				name:     dbcfg.Name,
				user:     dbcfg.User,
				password: dbcfg.Password,
			},
			want: &dbcfg,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cconfig.NewDBConfig(tt.args.host, tt.args.port, tt.args.name, tt.args.user, tt.args.password)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDBConfig() = %v, want %v", got, tt.want)
			}

			if got := got.GetHostPort(); got != tt.want.GetHostPort() {
				t.Errorf("GetHostPort() = %v, want %v", got, tt.want.GetHostPort())
			}
		})
	}
}
