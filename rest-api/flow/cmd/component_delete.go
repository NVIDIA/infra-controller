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

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/NVIDIA/infra-controller-rest/flow/pkg/client"
)

var (
	deleteComponentID string
)

// newDeleteCmd returns a configured cobra.Command for soft-deleting a component
// from the inventory by UUID.
func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a component",
		Long: `Soft-delete a component from the inventory table.

Required:
  --id : Component UUID to delete

Examples:
  flow component delete --id "component-uuid"
`,
		Run: func(cmd *cobra.Command, args []string) {
			doDeleteComponent()
		},
	}

	cmd.Flags().StringVar(&deleteComponentID, "id", "", "Component UUID (required)")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func init() {
	componentCmd.AddCommand(newDeleteCmd())
}

// doDeleteComponent parses the component UUID from the flag and calls
// DeleteComponent via the gRPC client.
func doDeleteComponent() {
	compID, err := uuid.Parse(deleteComponentID)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid component UUID")
	}

	c, err := client.New(newGlobalClientConfig())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create client")
	}
	defer c.Close()

	ctx := context.Background()
	if err := c.DeleteComponent(ctx, compID); err != nil {
		log.Fatal().Err(err).Msg("Failed to delete component")
	}

	fmt.Printf("Component %s deleted successfully.\n", compID)
}
