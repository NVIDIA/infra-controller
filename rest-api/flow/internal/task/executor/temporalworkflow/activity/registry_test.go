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

package activity

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager"
)

// TestActivities_All_ContainsAllActivities verifies that All() returns every
// expected activity name with a non-nil function value.
func TestActivities_All_ContainsAllActivities(t *testing.T) {
	acts := New(nil, nil)
	all := acts.All()

	expectedNames := []string{
		NameInjectExpectation,
		NamePowerControl,
		NameGetPowerStatus,
		NameUpdateTaskStatus,
		NameFirmwareControl,
		NameGetFirmwareStatus,
		NameBringUpControl,
		NameGetBringUpStatus,
		NameVerifyFirmwareConsistency,
	}
	require.Len(t, all, len(expectedNames), "unexpected number of activities")

	for _, name := range expectedNames {
		assert.Contains(t, all, name, "expected activity %q to be present", name)
		assert.NotNil(t, all[name], "expected function for activity %q to be non-nil", name)
	}
}

// TestActivities_All_ReturnsCopy verifies that mutating the returned map does
// not affect subsequent calls — each call produces an independent map.
func TestActivities_All_ReturnsCopy(t *testing.T) {
	acts := New(nil, nil)
	first := acts.All()
	firstLen := len(first)

	first["should-not-persist"] = func() {}

	second := acts.All()
	assert.Equal(t, firstLen, len(second), "registry size should be unchanged after mutating the returned map")
	assert.NotContains(t, second, "should-not-persist")
}

// TestActivities_Isolation verifies that two Activities instances do not share
// state: mutations to one instance's map must not affect the other.
func TestActivities_Isolation(t *testing.T) {
	a1 := New(nil, nil)
	a2 := New(nil, nil)

	m1 := a1.All()
	m1["isolation-sentinel"] = func() {}

	m2 := a2.All()
	assert.NotContains(t, m2, "isolation-sentinel", "mutations to one instance's map must not bleed into another instance's map")
}

// TestActivities_ValidAndGetComponentManager_NilRegistry verifies that manager
// lookup returns a clear configuration error when no registry is configured.
func TestActivities_ValidAndGetComponentManager_NilRegistry(t *testing.T) {
	acts := New(nil, nil)
	_, err := acts.validAndGetComponentManager(newActivityTestTarget())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, componentmanager.ErrRegistryNotConfigured))
}
