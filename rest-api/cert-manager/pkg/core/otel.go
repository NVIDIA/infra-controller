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

package core

import (
	"context"

	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// tracer is the OTel tracer to use for the core package
var tracer oteltrace.Tracer

func init() {
	tracer = otel.Tracer("nvmetal/cloud-cert-manager/pkg/core")
}

// StartOTELDaemon starts a go routine that waits on the provided context to quit and then shuts down the daemon
func StartOTELDaemon(ctx context.Context) {
	log := GetLogger(ctx)

	// Ignore this is most likely disabled
	log.Infof("Skipping OTEL startup - not supported")
}
