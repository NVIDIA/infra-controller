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

package model

import (
	"fmt"
	cdb "github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAPIMachineValidationTestCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		desc      string
		obj       APIMachineValidationTestCreateRequest
		expectErr bool
	}{
		{
			desc:      "no error",
			obj:       APIMachineValidationTestCreateRequest{Name: "test-1", Command: "/bin/sh/test1", Args: "-p 12"},
			expectErr: false,
		},
		{
			desc:      "error no Name",
			obj:       APIMachineValidationTestCreateRequest{Command: "/bin/sh/test1", Args: "-p 12"},
			expectErr: true,
		},
		{
			desc:      "error no Command",
			obj:       APIMachineValidationTestCreateRequest{Name: "test-1", Args: "-p 12"},
			expectErr: true,
		},
		{
			desc:      "error no args",
			obj:       APIMachineValidationTestCreateRequest{Name: "test-1", Command: "/bin/sh/test1"},
			expectErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.obj.Validate()
			assert.Equal(t, tc.expectErr, err != nil)
			if err != nil {
				fmt.Println(err.Error())
			}
		})
	}
}

func TestAPIMachineValidationExternalConfigCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		desc      string
		obj       APIMachineValidationExternalConfigCreateRequest
		expectErr bool
	}{
		{
			desc:      "no error",
			obj:       APIMachineValidationExternalConfigCreateRequest{Name: "test-1", Description: cdb.GetStrPtr("test description"), Config: []byte{0, 1, 12}},
			expectErr: false,
		},
		{
			desc:      "no error with no description",
			obj:       APIMachineValidationExternalConfigCreateRequest{Name: "test-1", Config: []byte{0, 1, 12}},
			expectErr: false,
		},
		{
			desc:      "error no Name",
			obj:       APIMachineValidationExternalConfigCreateRequest{Config: []byte{0, 1, 12}},
			expectErr: true,
		},
		{
			desc:      "error no Config",
			obj:       APIMachineValidationExternalConfigCreateRequest{Name: "test-1"},
			expectErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.obj.Validate()
			assert.Equal(t, tc.expectErr, err != nil)
			if err != nil {
				fmt.Println(err.Error())
			}
		})
	}
}
