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
	"github.com/spf13/cobra"

	"github.com/NVIDIA/infra-controller-rest/flow/pkg/client"
)

var ruleAssociateCmd = &cobra.Command{
	Use:   "associate",
	Short: "Associate a rule with a rack",
	Long: `Associate an operation rule with a specific rack.

The operation type and operation are automatically determined from the rule.
This will override any existing rule association for that operation on the rack.

Example:
  flow rule associate --rack-id R1 --rule-id abc123-def4-5678-90ab-cdef12345678`,
	RunE: runRuleAssociate,
}

var (
	assocRackID string
	assocRuleID string
)

func init() {
	ruleCmd.AddCommand(ruleAssociateCmd)

	ruleAssociateCmd.Flags().StringVar(&assocRackID, "rack-id", "", "Rack ID (required)")
	ruleAssociateCmd.Flags().StringVar(&assocRuleID, "rule-id", "", "Rule ID (required)")

	ruleAssociateCmd.MarkFlagRequired("rack-id")
	ruleAssociateCmd.MarkFlagRequired("rule-id")
}

// runRuleAssociate is the RunE handler for ruleAssociateCmd. It parses the
// rack and rule IDs from flags and calls AssociateRuleWithRack via the client.
func runRuleAssociate(cmd *cobra.Command, args []string) error {
	flowClient, err := client.New(newGlobalClientConfig())
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer flowClient.Close()

	rackID, err := uuid.Parse(assocRackID)
	if err != nil {
		return fmt.Errorf("invalid rack ID: %w", err)
	}

	ruleID, err := uuid.Parse(assocRuleID)
	if err != nil {
		return fmt.Errorf("invalid rule ID: %w", err)
	}

	err = flowClient.AssociateRuleWithRack(context.Background(), rackID, ruleID)
	if err != nil {
		return fmt.Errorf("failed to associate rule with rack: %w", err)
	}

	fmt.Printf("Successfully associated rule %s with rack %s\n",
		assocRuleID, assocRackID)

	return nil
}
