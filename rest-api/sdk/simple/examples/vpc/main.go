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

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/NVIDIA/infra-controller-rest/sdk/simple"
)

func main() {
	// NICO_BASE_URL, NICO_ORG, and NICO_TOKEN are required.
	// See sdk/simple/README.md for local dev (kind) setup.
	client, err := simple.NewClientFromEnv()
	if err != nil {
		fmt.Println("Error creating client:", err)
		os.Exit(1)
	}
	ctx := context.Background()
	if siteID := os.Getenv("NICO_SITE_ID"); siteID != "" {
		client.SetSiteID(siteID)
	}
	if err := client.Authenticate(ctx); err != nil {
		fmt.Printf("Error authenticating: %v\n", err)
		os.Exit(1)
	}

	// Example 1: List all VPCs
	fmt.Println("\nExample 1: Listing VPCs...")
	paginationFilter := &simple.PaginationFilter{
		PageSize: simple.IntPtr(20),
	}
	vpcs, pagination, apiErr := client.GetVpcs(ctx, nil, paginationFilter)
	if apiErr != nil {
		fmt.Printf("Error listing VPCs: %s\n", apiErr.Message)
		os.Exit(1)
	}
	fmt.Printf("Found %d VPCs on this page (total: %d)\n", len(vpcs), pagination.Total)
	for i, v := range vpcs {
		desc := ""
		if v.Description != nil {
			desc = *v.Description
		}
		fmt.Printf("  %d. ID=%s Name=%s Type=%s Desc=%s\n", i+1, v.ID, v.Name, v.NetworkVirtualizationType, desc)
	}

	// Example 2: Get a specific VPC by ID (if any exist)
	if len(vpcs) > 0 {
		vpcID := vpcs[0].ID
		fmt.Printf("\nExample 2: Getting VPC %s...\n", vpcID)
		vpc, apiErr := client.GetVpc(ctx, vpcID)
		if apiErr != nil {
			fmt.Printf("Error getting VPC: %s\n", apiErr.Message)
			os.Exit(1)
		}
		desc := ""
		if vpc.Description != nil {
			desc = *vpc.Description
		}
		fmt.Printf("Retrieved VPC: ID=%s Name=%s Type=%s Desc=%s\n",
			vpc.ID, vpc.Name, vpc.NetworkVirtualizationType, desc)
	}

	fmt.Println("\nVPC example completed successfully.")
}
