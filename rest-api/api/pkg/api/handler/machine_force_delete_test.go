// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/NVIDIA/infra-controller/rest-api/api/pkg/api/handler/util/common"
	"github.com/NVIDIA/infra-controller/rest-api/api/pkg/api/model"
	cdbm "github.com/NVIDIA/infra-controller/rest-api/db/pkg/db/model"
	corev1 "github.com/NVIDIA/infra-controller/rest-api/proto/core/gen/v1"
)

func TestDeleteMachineHandlerForce(t *testing.T) {
	t.Run("true proxies the complete force-delete request and returns resource IDs", func(t *testing.T) {
		coreResp := &corev1.AdminForceDeleteMachineResponse{
			AllDone:                       true,
			ManagedHostMachineId:          "machine-1",
			ManagedHostMachineInterfaceId: "interface-1",
			InstanceId:                    "instance-1",
			ManagedHostBmcIp:              "192.0.2.10",
			DpuBmcIp:                      "192.0.2.11",
			UfmUnregistrations:            2,
			UfmUnregistrationPending:      true,
			InitialLockdownState:          "Enabled",
			MachineUnlocked:               true,
			DpuMachineIds:                 []string{"dpu-1", "dpu-2"},
			DpuMachineInterfaceIds:        []string{"dpu-interface-1", "dpu-interface-2"},
			HostInterfacesDeleted:         true,
			DpuInterfacesDeleted:          true,
			HostBmcInterfaceAssociated:    true,
			DpuBmcInterfaceAssociated:     true,
			HostBmcInterfaceDeleted:       true,
			DpuBmcInterfaceDeleted:        true,
		}
		fixture := common.NewTestSetupProviderMachineHandlerFixture(t, coreResp)
		handler := NewDeleteMachineHandler(fixture.DBSession, fixture.SiteClientPool, fixture.Config)

		isAssigned := true
		_, err := cdbm.NewMachineDAO(fixture.DBSession).Update(context.Background(), nil, cdbm.MachineUpdateInput{
			MachineID:  fixture.MachineID,
			IsAssigned: &isAssigned,
		})
		require.NoError(t, err)

		rec := fixture.Request(t, handler.Handle, http.MethodDelete, "/?force=true", nil, "")
		require.Equal(t, http.StatusAccepted, rec.Code, rec.Body.String())
		require.Equal(t, "application/vnd.nvidia.nico.machine-force-delete+json", rec.Header().Get("Content-Type"))
		require.Equal(t, corev1.Forge_AdminForceDeleteMachine_FullMethodName, fixture.ProxiedReq.FullMethod)

		var coreReq corev1.AdminForceDeleteMachineRequest
		require.NoError(t, protojson.Unmarshal(fixture.ProxiedReq.RequestJSON, &coreReq))
		require.Equal(t, fixture.MachineID, coreReq.GetHostQuery())
		require.True(t, coreReq.GetDeleteInterfaces())
		require.True(t, coreReq.GetDeleteBmcInterfaces())
		require.False(t, coreReq.GetDeleteBmcCredentials())
		require.False(t, coreReq.GetAllowDeleteWithOrphanedDpfCrds())
		require.False(t, coreReq.GetDeleteBmcSuppressions())
		require.False(t, coreReq.GetDeleteRetainedBootInterfaces())

		var apiResp model.APIMachineForceDeleteResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &apiResp))
		require.Equal(t, model.APIMachineForceDeleteResponse{
			AllDone:                       true,
			ManagedHostMachineID:          "machine-1",
			ManagedHostMachineInterfaceID: "interface-1",
			InstanceID:                    "instance-1",
			ManagedHostBMCIP:              "192.0.2.10",
			DPUBMCIP:                      "192.0.2.11",
			UFMUnregistrations:            2,
			UFMUnregistrationPending:      true,
			InitialLockdownState:          "Enabled",
			MachineUnlocked:               true,
			HostInterfacesDeleted:         true,
			DPUInterfacesDeleted:          true,
			HostBMCInterfaceAssociated:    true,
			DPUBMCInterfaceAssociated:     true,
			HostBMCInterfaceDeleted:       true,
			DPUBMCInterfaceDeleted:        true,
			DPUMachineIDs:                 []string{"dpu-1", "dpu-2"},
			DPUMachineInterfaceIDs:        []string{"dpu-interface-1", "dpu-interface-2"},
		}, apiResp)
	})

	t.Run("false preserves normal deletion safeguards", func(t *testing.T) {
		fixture := common.NewTestSetupProviderMachineHandlerFixture(t, nil)
		handler := NewDeleteMachineHandler(fixture.DBSession, fixture.SiteClientPool, fixture.Config)

		rec := fixture.Request(t, handler.Handle, http.MethodDelete, "/?force=false", nil, "")
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		require.Empty(t, fixture.ProxiedReq.FullMethod)
		require.Contains(t, rec.Body.String(), "Machine exists on Site and cannot be deleted")
	})

	t.Run("invalid value is rejected before deletion", func(t *testing.T) {
		fixture := common.NewTestSetupProviderMachineHandlerFixture(t, nil)
		handler := NewDeleteMachineHandler(fixture.DBSession, fixture.SiteClientPool, fixture.Config)

		rec := fixture.Request(t, handler.Handle, http.MethodDelete, "/?force=definitely", nil, "")
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		require.Empty(t, fixture.ProxiedReq.FullMethod)
		require.Contains(t, rec.Body.String(), "force")
	})
}
