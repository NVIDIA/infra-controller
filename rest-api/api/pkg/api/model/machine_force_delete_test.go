// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"encoding/json"
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIMachineForceDeleteResponseZeroValueJSON(t *testing.T) {
	payload, err := json.Marshal(APIMachineForceDeleteResponse{})
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(payload, &decoded))
	require.ElementsMatch(t, []string{
		"allDone",
		"managedHostMachineId",
		"managedHostMachineInterfaceId",
		"instanceId",
		"managedHostBmcIp",
		"dpuBmcIp",
		"ufmUnregistrations",
		"ufmUnregistrationPending",
		"initialLockdownState",
		"machineUnlocked",
		"hostInterfacesDeleted",
		"dpuInterfacesDeleted",
		"hostBmcInterfaceAssociated",
		"dpuBmcInterfaceAssociated",
		"hostBmcInterfaceDeleted",
		"dpuBmcInterfaceDeleted",
		"dpuMachineIds",
		"dpuMachineInterfaceIds",
	}, slices.Collect(maps.Keys(decoded)))
	require.Nil(t, decoded["dpuMachineIds"])
	require.Nil(t, decoded["dpuMachineInterfaceIds"])
}
