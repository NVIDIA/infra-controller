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

package api

import (
	"net/http"
	"testing"

	"github.com/NVIDIA/infra-controller-rest/api/internal/config"
	"github.com/NVIDIA/infra-controller-rest/api/pkg/api/handler/util/common"
	sc "github.com/NVIDIA/infra-controller-rest/api/pkg/client/site"
	cdb "github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	"github.com/stretchr/testify/assert"

	temporalClient "go.temporal.io/sdk/client"
	tmocks "go.temporal.io/sdk/mocks"
)

func TestNewAPIRoutes(t *testing.T) {
	type args struct {
		dbSession *cdb.Session
		tc        temporalClient.Client
		tnc       temporalClient.NamespaceClient
		scp       *sc.ClientPool
		cfg       *config.Config
	}

	tc := &tmocks.Client{}
	tnc := &tmocks.NamespaceClient{}

	cfg := common.GetTestConfig()
	tcfg, _ := cfg.GetTemporalConfig()
	scp := sc.NewClientPool(tcfg)

	routeCount := map[string]int{
		"metadata":                 1,
		"service-account":          1,
		"infrastructure-provider":  4,
		"tenant":                   4,
		"tenant-account":           5,
		"site":                     6,
		"vpc":                      6,
		"vpcpeering":               4,
		"vpcprefix":                5,
		"ip-block":                 6,
		"instance":                 8,
		"interface":                1,
		"infiniband-interface":     2,
		"infiniband-partition":     5,
		"nvlink-interface":         2,
		"nvlink-logical-partition": 4,
		"expected-machine":         7,
		"expected-power-shelf":     5,
		"expected-rack":            7,
		"expected-switch":          5,
		"instance-type":            5,
		"machine":                  5,
		"allocation":               6,
		"subnet":                   5,
		"machine-instance-type":    3,
		"user":                     1,
		"operating-system":         5,
		"sshkey":                   5,
		"sshkeygroup":              5,
		"machine-capability":       1,
		"audit":                    2,
		"network-security-group":   5,
		"machine-validation":       11,
		"dpu-extension-service":    7,
		"sku":                      2,
		"rack":                     12,
		"tray":                     8,
		"stats":                    4,
	}

	totalRouteCount := 0
	for _, v := range routeCount {
		totalRouteCount += v
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "test initializing API routes",
			args: args{
				dbSession: &cdb.Session{},
				tc:        tc,
				tnc:       tnc,
				scp:       scp,
				cfg:       cfg,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAPIRoutes(tt.args.dbSession, tt.args.tc, tt.args.tnc, tt.args.scp, tt.args.cfg)

			assert.Equal(t, totalRouteCount, len(got))

			for _, route := range got {
				assert.Contains(t, route.Path, "/org/:orgName/"+cfg.GetAPIName())
			}

			expectedMachineBatchPath := "/org/:orgName/" + cfg.GetAPIName() + "/expected-machine/batch"
			assertRouteExists(t, got, http.MethodPost, expectedMachineBatchPath)
			assertRouteExists(t, got, http.MethodPatch, expectedMachineBatchPath)
			assertRouteBefore(t, got, http.MethodPatch, expectedMachineBatchPath, http.MethodPatch, "/org/:orgName/"+cfg.GetAPIName()+"/expected-machine/:id")
		})
	}
}

func assertRouteExists(t *testing.T, routes []Route, method, path string) {
	t.Helper()

	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return
		}
	}

	assert.Failf(t, "route not found", "missing %s %s", method, path)
}

func assertRouteBefore(t *testing.T, routes []Route, firstMethod, firstPath, secondMethod, secondPath string) {
	t.Helper()

	firstIndex := -1
	secondIndex := -1
	for i, route := range routes {
		if route.Method == firstMethod && route.Path == firstPath {
			firstIndex = i
		}
		if route.Method == secondMethod && route.Path == secondPath {
			secondIndex = i
		}
	}

	assert.NotEqual(t, -1, firstIndex, "missing %s %s", firstMethod, firstPath)
	assert.NotEqual(t, -1, secondIndex, "missing %s %s", secondMethod, secondPath)
	assert.Less(t, firstIndex, secondIndex, "%s %s must be registered before %s %s", firstMethod, firstPath, secondMethod, secondPath)
}
