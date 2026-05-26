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
	"testing"
)

func TestHealthzConfig(t *testing.T) {
	type args struct {
		enabled bool
		port    int
	}

	hccfg := HealthzConfig{
		Enabled: true,
		Port:    6930,
	}

	tests := []struct {
		name string
		args args
		want *HealthzConfig
	}{
		{
			name: "initialize Healthz config",
			args: args{
				enabled: true,
				port:    hccfg.Port,
			},
			want: &hccfg,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHealthzConfig(tt.args.enabled, tt.args.port)

			if p := got.Port; p != tt.want.Port {
				t.Errorf("got.Port = %v, want %v", p, tt.want.Port)
			}

			if got := got.GetListenAddr(); got != tt.want.GetListenAddr() {
				t.Errorf("GetListenAddr() = %v, want %v", got, tt.want.GetListenAddr())
			}
		})
	}
}
