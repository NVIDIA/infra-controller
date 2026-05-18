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

// Package certs implements certificate management
package certs

import (
	"github.com/NVIDIA/infra-controller-rest/cert-manager/pkg/types"
)

// CertificateIssuer is an alias for types.CertificateIssuer for backward compatibility
type CertificateIssuer = types.CertificateIssuer

// CertificateRequest is an alias for types.CertificateRequest for backward compatibility
type CertificateRequest = types.CertificateRequest

// CertificateResponse is an alias for types.CertificateResponse for backward compatibility
type CertificateResponse = types.CertificateResponse
