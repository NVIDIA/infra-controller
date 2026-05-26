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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NVIDIA/infra-controller-rest/db/pkg/db"
	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

func TestLabelsFromProtoMetadata(t *testing.T) {
	tests := []struct {
		name string
		md   *cwssaws.Metadata
		want map[string]string
	}{
		{
			name: "nil metadata",
			md:   nil,
			want: nil,
		},
		{
			name: "nil labels slice",
			md:   &cwssaws.Metadata{Labels: nil},
			want: nil,
		},
		{
			name: "empty labels slice",
			md:   &cwssaws.Metadata{Labels: []*cwssaws.Label{}},
			want: map[string]string{},
		},
		{
			name: "single label with value",
			md: &cwssaws.Metadata{
				Labels: []*cwssaws.Label{
					{Key: "environment", Value: db.GetStrPtr("production")},
				},
			},
			want: map[string]string{"environment": "production"},
		},
		{
			name: "multiple labels",
			md: &cwssaws.Metadata{
				Labels: []*cwssaws.Label{
					{Key: "environment", Value: db.GetStrPtr("production")},
					{Key: "rack", Value: db.GetStrPtr("rack-1")},
					{Key: "datacenter", Value: db.GetStrPtr("dc1")},
				},
			},
			want: map[string]string{
				"environment": "production",
				"rack":        "rack-1",
				"datacenter":  "dc1",
			},
		},
		{
			name: "label with nil value yields empty string",
			md: &cwssaws.Metadata{
				Labels: []*cwssaws.Label{
					{Key: "flag", Value: nil},
				},
			},
			want: map[string]string{"flag": ""},
		},
		{
			name: "label with empty key is skipped",
			md: &cwssaws.Metadata{
				Labels: []*cwssaws.Label{
					{Key: "", Value: db.GetStrPtr("value")},
					{Key: "valid", Value: db.GetStrPtr("data")},
				},
			},
			want: map[string]string{"valid": "data"},
		},
		{
			name: "nil label entry is skipped",
			md: &cwssaws.Metadata{
				Labels: []*cwssaws.Label{
					nil,
					{Key: "valid", Value: db.GetStrPtr("data")},
				},
			},
			want: map[string]string{"valid": "data"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := LabelsFromProtoMetadata(tc.md)
			if tc.want == nil {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}
