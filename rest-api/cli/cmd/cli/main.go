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
	"fmt"
	"os"

	appcli "github.com/NVIDIA/infra-controller-rest/cli/pkg"
	"github.com/NVIDIA/infra-controller-rest/cli/tui"
	"github.com/NVIDIA/infra-controller-rest/openapi"
	"github.com/urfave/cli/v2"
)

func main() {
	app, err := appcli.NewApp(openapi.Spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
	app.Commands = append(app.Commands, &cli.Command{
		Name:    "tui",
		Aliases: []string{"i"},
		Usage:   "Start interactive TUI mode with config selector",
		Action: func(c *cli.Context) error {
			return tui.RunTUI(c.String("config"))
		},
	})
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
