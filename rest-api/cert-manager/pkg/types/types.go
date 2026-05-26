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

// Package types defines shared types for cert-manager
package types

import (
	"context"
	"fmt"
	"strings"
)

// CertificateIssuer defines a certificate issuer interface
type CertificateIssuer interface {
	NewCertificate(ctx context.Context, req *CertificateRequest) (string, string, error)
	GetCACertificate(ctx context.Context) (string, error)
	GetCRL(ctx context.Context) (string, error)
	RawCertificate(ctx context.Context, sans string, ttl int) (string, string, error)
}

// CertificateRequest defines a request
type CertificateRequest struct {
	// Name should be unique for a particular CertificateType
	Name string `json:"name,omitempty"`

	// App identifies the specific app
	App string `json:"app,omitempty"`

	TTL int `json:"ttl,omitempty"`
}

// CertificateResponse defines a response
type CertificateResponse struct {
	Key         string `json:"key,omitempty"`
	Certificate string `json:"certificate,omitempty"`
}

// UniqueName returns a sans per node/app combination
func (r *CertificateRequest) UniqueName(baseDNS string) string {
	var sans string

	if r.App == "" {
		sans = r.Name
	} else {
		sans = fmt.Sprintf("%s.%s", r.App, r.Name)
	}

	if strings.HasSuffix(sans, ".com") || strings.HasSuffix(sans, ".local") {
		return sans
	}

	// add baseDNS
	return fmt.Sprintf("%s.%s", sans, baseDNS)
}
