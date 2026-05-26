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

package claim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNgcClaims_ValidateOrg(t *testing.T) {
	type fields struct {
		Access []NgcAccessClaim
	}
	type args struct {
		orgName string
	}

	orgName := "test-org"

	ngcOrgClaim := NgcAccessClaim{
		Type:    "group/ngc-stg",
		Name:    orgName,
		Actions: []string{},
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "validate and accept org in claim",
			fields: fields{
				Access: []NgcAccessClaim{
					ngcOrgClaim,
				},
			},
			args: args{
				orgName: orgName,
			},
			want: true,
		},
		{
			name: "validate and reject org in claim",
			fields: fields{
				Access: []NgcAccessClaim{
					ngcOrgClaim,
				},
			},
			args: args{
				orgName: "invalid-org",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nc := &NgcKasClaims{
				Access: tt.fields.Access,
			}

			assert.Equal(t, tt.want, nc.ValidateOrg(tt.args.orgName))
		})
	}
}
