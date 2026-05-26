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

package claim

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// NgcOrgClaimTypePrefix is the prefix for access claim that contains NGC organization name
	// e.g. Staging: "group/ngc-stg", Production: "group/ngc"
	NgcOrgClaimTypePrefix = "group/ngc"
	// NgcAudience describes the audience value present in NGC tokens
	NgcAudience = "ngc"

	// SsaScopeKas is the scope required to access KAS
	SsaScopeKas = "kas"
)

// NgcAccessClaim represent the custom NGC KAS access claims
type NgcAccessClaim struct {
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}

// NgcKasLegacyClaims represent the custom JWT claims used by NGC KAS
type NgcKasClaims struct {
	Access []NgcAccessClaim `json:"access"`
	jwt.RegisteredClaims
}

// ValidateOrg checks whether a specified org name is included in claims
func (nc *NgcKasClaims) ValidateOrg(orgName string) bool {
	isValid := false
	for _, claim := range nc.Access {
		if strings.HasPrefix(claim.Type, NgcOrgClaimTypePrefix) && claim.Name == orgName {
			isValid = true
			break
		}
	}

	return isValid
}

// SsaClaims represent the custom JWT claims used by SSA
type SsaClaims struct {
	Scopes []string `json:"scopes"`
	jwt.RegisteredClaims
}

// ValidateScope checks whether a specified scope is included in claims
func (sc *SsaClaims) ValidateScope(scope string) bool {
	for _, s := range sc.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}
