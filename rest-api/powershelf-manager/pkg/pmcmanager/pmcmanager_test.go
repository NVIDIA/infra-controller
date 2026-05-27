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
package pmcmanager

import (
	"context"
	"testing"

	"github.com/NVIDIA/infra-controller-rest/powershelf-manager/pkg/redfish"

	"github.com/stretchr/testify/assert"
)

func TestRedfishTx_NilPMC(t *testing.T) {
	pm := &PmcManager{}

	err := pm.RedfishTx(context.Background(), nil, func(_ *redfish.RedfishClient) error {
		t.Fatal("tx should not be called with nil PMC")
		return nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "null PMC")
}
