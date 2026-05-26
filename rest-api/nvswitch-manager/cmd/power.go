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
	"fmt"
	"time"

	pb "github.com/NVIDIA/infra-controller-rest/nvswitch-manager/internal/proto/v1"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	powerServerAddr string
	powerSwitchUUID string
	powerAction     string
)

var powerCmd = &cobra.Command{
	Use:   "power",
	Short: "Power control operations for NV-Switch trays",
	Long: `Power control operations for NV-Switch trays via the NSM gRPC service.

Supported actions:
  ForceOff, PowerCycle, GracefulShutdown, On, ForceOn, GracefulRestart, ForceRestart

Examples:
  nvswitch-manager power --switch-uuid <uuid> --action PowerCycle
  nvswitch-manager power --switch-uuid <uuid> --action GracefulShutdown
  nvswitch-manager power --switch-uuid <uuid> --action ForceOff`,
	Run: func(cmd *cobra.Command, args []string) {
		if powerSwitchUUID == "" {
			log.Fatal("--switch-uuid is required")
		}

		action := resolvePowerAction(powerAction)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		conn, err := grpc.DialContext(ctx, powerServerAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			log.Fatalf("Failed to connect to server %s: %v", powerServerAddr, err)
		}
		defer conn.Close()

		client := pb.NewNVSwitchManagerClient(conn)

		fmt.Printf("Performing %s on switch %s...\n", powerAction, powerSwitchUUID)

		resp, err := client.PowerControl(ctx, &pb.PowerControlRequest{
			Uuids:  []string{powerSwitchUUID},
			Action: action,
		})
		if err != nil {
			log.Fatalf("PowerControl RPC failed: %v", err)
		}

		for _, r := range resp.Responses {
			if r.Status == pb.StatusCode_SUCCESS {
				fmt.Printf("%s initiated successfully for switch %s\n", powerAction, r.Uuid)
			} else {
				fmt.Printf("%s failed for switch %s: %s\n", powerAction, r.Uuid, r.Error)
			}
		}
	},
}

func resolvePowerAction(s string) pb.PowerAction {
	switch s {
	case "ForceOff":
		return pb.PowerAction_POWER_ACTION_FORCE_OFF
	case "PowerCycle":
		return pb.PowerAction_POWER_ACTION_POWER_CYCLE
	case "GracefulShutdown":
		return pb.PowerAction_POWER_ACTION_GRACEFUL_SHUTDOWN
	case "On":
		return pb.PowerAction_POWER_ACTION_ON
	case "ForceOn":
		return pb.PowerAction_POWER_ACTION_FORCE_ON
	case "GracefulRestart":
		return pb.PowerAction_POWER_ACTION_GRACEFUL_RESTART
	case "ForceRestart":
		return pb.PowerAction_POWER_ACTION_FORCE_RESTART
	default:
		log.Fatalf("Unknown power action %q; valid actions: ForceOff, PowerCycle, GracefulShutdown, On, ForceOn, GracefulRestart, ForceRestart", s)
		return pb.PowerAction_POWER_ACTION_UNKNOWN
	}
}

func init() {
	rootCmd.AddCommand(powerCmd)

	powerCmd.Flags().StringVar(&powerServerAddr, "server", "localhost:50051", "NSM gRPC server address")
	powerCmd.Flags().StringVar(&powerSwitchUUID, "switch-uuid", "", "Switch UUID (required)")
	powerCmd.Flags().StringVar(&powerAction, "action", "PowerCycle", "Power action: ForceOff, PowerCycle, GracefulShutdown, On, ForceOn, GracefulRestart, ForceRestart")
}
