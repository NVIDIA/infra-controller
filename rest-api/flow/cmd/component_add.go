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
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/NVIDIA/infra-controller-rest/flow/pkg/client"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/types"
)

var (
	addRackID          string
	addName            string
	addType            string
	addManufacturer    string
	addSerialNumber    string
	addModel           string
	addFirmwareVersion string
	addSlotID          int
	addTrayIndex       int
	addHostID          int
	addDescription     string
	addBmcMAC          string
	addBmcIP           string
	addBmcType         string
)

// newAddCmd returns a configured cobra.Command for adding a new component
// to an existing rack in the inventory.
func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a component to an existing rack",
		Long: `Add a new component to an existing rack in the inventory table.

Required:
  --rack-id          : Rack UUID to add the component to
  --name             : Component name
  --type             : Component type (compute, nvlswitch, powershelf, torswitch, ums, cdu)
  --manufacturer     : Manufacturer
  --serial-number    : Serial number

Optional:
  --model            : Model name
  --firmware-version : Firmware version
  --slot-id          : Slot ID (position)
  --tray-index       : Tray index (position)
  --host-id          : Host ID (position)
  --description      : Description (JSON string)
  --bmc-mac          : BMC MAC address
  --bmc-ip           : BMC IP address
  --bmc-type         : BMC type (host, dpu). Default: host

Examples:
  # Add a compute node
  flow component add --rack-id "uuid" --name "node-01" --type compute \
    --manufacturer "NVIDIA" --serial-number "SN123" --slot-id 1 --tray-index 0 --host-id 1

  # Add a powershelf with BMC
  flow component add --rack-id "uuid" --name "ps-01" --type powershelf \
    --manufacturer "NVIDIA" --serial-number "PS123" --bmc-mac "aa:bb:cc:dd:ee:ff" --bmc-ip "10.0.0.1"
`,
		Run: func(cmd *cobra.Command, args []string) {
			doAddComponent()
		},
	}

	cmd.Flags().StringVar(&addRackID, "rack-id", "", "Rack UUID (required)")
	cmd.Flags().StringVar(&addName, "name", "", "Component name (required)")
	cmd.Flags().StringVarP(&addType, "type", "t", "", "Component type: compute, nvlswitch, powershelf, torswitch, ums, cdu (required)")
	cmd.Flags().StringVar(&addManufacturer, "manufacturer", "", "Manufacturer (required)")
	cmd.Flags().StringVar(&addSerialNumber, "serial-number", "", "Serial number (required)")
	cmd.Flags().StringVar(&addModel, "model", "", "Model name")
	cmd.Flags().StringVar(&addFirmwareVersion, "firmware-version", "", "Firmware version")
	cmd.Flags().IntVar(&addSlotID, "slot-id", 0, "Slot ID")
	cmd.Flags().IntVar(&addTrayIndex, "tray-index", 0, "Tray index")
	cmd.Flags().IntVar(&addHostID, "host-id", 0, "Host ID")
	cmd.Flags().StringVar(&addDescription, "description", "", "Description (JSON string)")
	cmd.Flags().StringVar(&addBmcMAC, "bmc-mac", "", "BMC MAC address")
	cmd.Flags().StringVar(&addBmcIP, "bmc-ip", "", "BMC IP address")
	cmd.Flags().StringVar(&addBmcType, "bmc-type", "host", "BMC type: host, dpu")

	_ = cmd.MarkFlagRequired("rack-id")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("manufacturer")
	_ = cmd.MarkFlagRequired("serial-number")

	return cmd
}

func init() {
	componentCmd.AddCommand(newAddCmd())
}

// doAddComponent builds a types.Component from the CLI flags and calls
// AddComponent via the gRPC client, printing the created component as JSON.
func doAddComponent() {
	rackID, err := uuid.Parse(addRackID)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid rack UUID")
	}

	compType := parseComponentTypeToTypes(addType)
	if compType == types.ComponentTypeUnknown {
		log.Fatal().Str("type", addType).Msg("Invalid component type")
	}

	comp := &types.Component{
		Type: compType,
		Info: types.DeviceInfo{
			ID:           uuid.New(),
			Name:         addName,
			Manufacturer: addManufacturer,
			SerialNumber: addSerialNumber,
			Model:        addModel,
			Description:  addDescription,
		},
		FirmwareVersion: addFirmwareVersion,
		Position: types.InRackPosition{
			SlotID:    addSlotID,
			TrayIndex: addTrayIndex,
			HostID:    addHostID,
		},
		RackID: rackID,
	}

	// Add BMC if provided
	if addBmcMAC != "" {
		bmcType := types.BMCTypeHost
		if addBmcType == "dpu" {
			bmcType = types.BMCTypeDPU
		}

		bmc := types.BMC{
			Type: bmcType,
		}
		bmc.MAC, err = net.ParseMAC(addBmcMAC)
		if err != nil {
			log.Fatal().Err(err).Msg("Invalid BMC MAC address")
		}
		if addBmcIP != "" {
			bmc.IP = net.ParseIP(addBmcIP)
		}
		comp.BMCs = []types.BMC{bmc}
	}

	c, err := client.New(newGlobalClientConfig())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create client")
	}
	defer c.Close()

	ctx := context.Background()
	created, err := c.AddComponent(ctx, comp)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to add component")
	}

	data, err := json.MarshalIndent(created, "", "  ")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal JSON")
	}
	fmt.Println(string(data))
}
