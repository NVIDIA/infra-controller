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

package nicoapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicMock(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	version, err := client.Version(ctx)
	assert.NoError(t, err)
	assert.Equal(t, version, "1.2.3")

	mID := "fm100ht09g4atrqgjb0b83b2to1qa1hfugks9mhutb0umcng1rkr54vliqg"
	serial := "12345"
	client.AddMachine(MachineDetail{MachineID: mID, ChassisSerial: &serial, FirmwareVersion: "1.2.3"})
	details, err2 := client.GetMachines(ctx)
	assert.NoError(t, err2)
	assert.Len(t, details, 1)
	assert.Equal(t, details[0].MachineID, mID)
	assert.Equal(t, details[0].FirmwareVersion, "1.2.3")
}
