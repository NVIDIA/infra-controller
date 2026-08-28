// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/NVIDIA/infra-controller/rest-api/api/pkg/api/handler/util/common"
	cdbm "github.com/NVIDIA/infra-controller/rest-api/db/pkg/db/model"
	corev1 "github.com/NVIDIA/infra-controller/rest-api/proto/core/gen/v1"
)

func TestDeleteMachineHandlerForce(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantStatus  int
		wantMessage string
		wantProxy   bool
	}{
		{name: "true proxies the complete force-delete request", query: "/?force=true", wantStatus: http.StatusAccepted, wantProxy: true},
		{name: "false preserves normal deletion safeguards", query: "/?force=false", wantStatus: http.StatusBadRequest, wantMessage: "Machine exists on Site and cannot be deleted"},
		{name: "invalid value is rejected before deletion", query: "/?force=definitely", wantStatus: http.StatusBadRequest, wantMessage: "Invalid force query parameter, expected a boolean value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := common.NewTestSetupProviderMachineHandlerFixture(t, &corev1.AdminForceDeleteMachineResponse{})
			handler := NewDeleteMachineHandler(fixture.DBSession, fixture.SiteClientPool, fixture.Config)

			if tt.wantProxy {
				isAssigned := true
				_, err := cdbm.NewMachineDAO(fixture.DBSession).Update(context.Background(), nil, cdbm.MachineUpdateInput{
					MachineID:  fixture.MachineID,
					IsAssigned: &isAssigned,
				})
				require.NoError(t, err)
			}

			rec := fixture.Request(t, handler.Handle, http.MethodDelete, tt.query, nil, "")
			require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())

			if !tt.wantProxy {
				require.Empty(t, fixture.ProxiedReq.FullMethod)
				require.Contains(t, rec.Body.String(), tt.wantMessage)
				return
			}

			require.Equal(t, corev1.Forge_AdminForceDeleteMachine_FullMethodName, fixture.ProxiedReq.FullMethod)
			var coreReq corev1.AdminForceDeleteMachineRequest
			require.NoError(t, protojson.Unmarshal(fixture.ProxiedReq.RequestJSON, &coreReq))
			require.True(t, proto.Equal(&coreReq, &corev1.AdminForceDeleteMachineRequest{
				HostQuery:                   fixture.MachineID,
				DeleteInterfaces:            true,
				DeleteBmcInterfaces:         true,
				AllowDeleteWithInstanceType: true,
			}))
			require.JSONEq(t, `{"message":"Deletion request was accepted"}`, rec.Body.String())
		})
	}
}
