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

package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/NVIDIA/infra-controller-rest/flow/pkg/types"
)

// componentCmd is the parent command for component operation subcommands.
var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Component operations",
	Long:  `Commands for querying and comparing components (expected vs actual).`,
}

func init() {
	rootCmd.AddCommand(componentCmd)
}

// parseComponentTypeToTypes converts string to types.ComponentType
func parseComponentTypeToTypes(s string) types.ComponentType {
	switch strings.ToLower(s) {
	case "compute":
		return types.ComponentTypeCompute
	case "nvlswitch", "nvl-switch":
		return types.ComponentTypeNVSwitch
	case "powershelf", "power-shelf":
		return types.ComponentTypePowerShelf
	case "torswitch", "tor-switch":
		return types.ComponentTypeTORSwitch
	case "ums":
		return types.ComponentTypeUMS
	case "cdu":
		return types.ComponentTypeCDU
	default:
		return types.ComponentTypeUnknown
	}
}
