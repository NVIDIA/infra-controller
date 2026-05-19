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

var ruleDeleteCmd = &cobra.Command{
	Use:   "delete <rule-id>",
	Short: "Delete an operation rule",
	Long:  `Delete an operation rule by ID. This will also remove all rack associations for this rule.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRuleDelete,
}

func init() {
	ruleCmd.AddCommand(ruleDeleteCmd)
}

// runRuleDelete is the RunE handler for ruleDeleteCmd. It parses the rule ID
// from the positional argument and calls DeleteOperationRule via the client.
func runRuleDelete(cmd *cobra.Command, args []string) error {
	ruleIDStr := args[0]

	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		return fmt.Errorf("invalid rule ID: %w", err)
	}

	flowClient, err := client.New(newGlobalClientConfig())
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer flowClient.Close()

	err = flowClient.DeleteOperationRule(context.Background(), ruleID)
	if err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	fmt.Printf("Successfully deleted rule %s\n", ruleIDStr)
	return nil
}
