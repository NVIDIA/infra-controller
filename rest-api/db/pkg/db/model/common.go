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
	cwssaws "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

var protoJsonUnmarshalOptions = protojson.UnmarshalOptions{
	AllowPartial:   true,
	DiscardUnknown: true,
}

// LabelsFromProtoMetadata converts a workflow Metadata's Labels list to a
// string map. Returns nil when md is nil or md.Labels is nil; returns an
// empty map (not nil) when Labels is non-nil but empty, so callers can
// distinguish "no labels reported" from "labels explicitly cleared".
func LabelsFromProtoMetadata(md *cwssaws.Metadata) map[string]string {
	if md == nil || md.Labels == nil {
		return nil
	}
	result := map[string]string{}
	for _, label := range md.Labels {
		if label == nil || label.Key == "" {
			continue
		}
		value := ""
		if label.Value != nil {
			value = *label.Value
		}
		result[label.Key] = value
	}
	return result
}
