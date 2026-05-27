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

package managerapi

import (
	"context"

	"github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"
)

// CarbideExpansion - Carbide Expansion
type CarbideExpansion interface{}

// NICoInterface - interface to NICo
type NICoInterface interface {
	// List all the apis of Carbide here
	Init()
	Start()
	CreateGRPCClient() error
	GetGRPCClient() *client.NICoCoreClient
	UpdateGRPCClientState(err error)
	CreateGRPCClientActivity(ctx context.Context, ResourceID string) (client *client.NICoCoreClient, err error)
	RegisterGRPC()
	GetState() []string
	GetGRPCClientVersion() int64
	CarbideExpansion
}
