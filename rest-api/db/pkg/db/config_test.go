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

package db

import (
	"testing"

	"github.com/NVIDIA/infra-controller-rest/common/pkg/credential"
	"github.com/stretchr/testify/assert"
)

func validConfig() Config {
	return Config{
		Host:       "db.example.internal",
		Port:       5432,
		DBName:     "forge",
		Credential: credential.New("forge", "s3cr3t"),
	}
}

// TestBuildDSN_NoCACert_UsesSSlModePrefer is a regression test for the bug
// introduced in v1.3.1 (PR #193) where BuildDSN started appending
// sslmode=disable when no CA certificate path is configured.
//
// In v1.0.4, NewSession built the DSN with no sslmode parameter at all, so
// pgx defaulted to "prefer" (try TLS, fall back to plaintext). This allowed
// connections to PostgreSQL clusters whose pg_hba.conf only contains hostssl
// rules (e.g. CloudNativePG defaults). The explicit sslmode=disable introduced
// by the refactor causes pg_hba.conf to reject those connections with:
//
//	FATAL: pg_hba.conf rejects connection for host "...", no encryption (SQLSTATE 28000)
func TestBuildDSN_NoCACert_UsesSSlModePrefer(t *testing.T) {
	cfg := validConfig()
	// No CACertificatePath set — mirrors the production API deployment, which
	// mounts no DB CA cert and therefore leaves CACertificatePath empty.

	dsn := cfg.BuildDSN()

	assert.Contains(t, dsn, "sslmode=prefer",
		"DSN must use sslmode=prefer when no CA cert is configured so that "+
			"TLS negotiation succeeds against servers with hostssl pg_hba rules")
	assert.NotContains(t, dsn, "sslmode=disable",
		"sslmode=disable breaks connections to CloudNativePG clusters whose "+
			"pg_hba.conf requires encrypted connections (regression since v1.3.1)")
}

func TestBuildDSN_WithCACert_UsesSSlModePreferAndRootCert(t *testing.T) {
	cfg := validConfig()
	cfg.CACertificatePath = "/var/secrets/db/ca.crt"

	dsn := cfg.BuildDSN()

	assert.Contains(t, dsn, "sslmode=prefer")
	assert.Contains(t, dsn, "sslrootcert=/var/secrets/db/ca.crt")
}
