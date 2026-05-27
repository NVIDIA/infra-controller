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

package helpers

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRewriteAPINamePath(t *testing.T) {
	t.Run("default api name leaves path unchanged", func(t *testing.T) {
		path := "/v2/org/test-org/nico/metadata"
		assert.Equal(t, path, RewriteAPINamePath(path, ""))
		assert.Equal(t, path, RewriteAPINamePath(path, "nico"))
	})

	t.Run("custom api name rewrites org scoped paths", func(t *testing.T) {
		assert.Equal(
			t,
			"/v2/org/test-org/nico/metadata",
			RewriteAPINamePath("/v2/org/test-org/nico/metadata", "nico"),
		)
	})

	t.Run("non matching paths are left unchanged", func(t *testing.T) {
		path := "/healthz"
		assert.Equal(t, path, RewriteAPINamePath(path, "nico"))
	})
}

func TestNormalizeAPIName(t *testing.T) {
	t.Run("empty string returns default", func(t *testing.T) {
		assert.Equal(t, DefaultAPIName, NormalizeAPIName(""))
	})

	t.Run("whitespace only returns default", func(t *testing.T) {
		assert.Equal(t, DefaultAPIName, NormalizeAPIName("   "))
	})

	t.Run("embedded slash returns default", func(t *testing.T) {
		assert.Equal(t, DefaultAPIName, NormalizeAPIName("nico/internal"))
	})

	t.Run("surrounding slashes are trimmed", func(t *testing.T) {
		assert.Equal(t, "nico", NormalizeAPIName("/nico/"))
	})

	t.Run("valid name is returned as-is", func(t *testing.T) {
		assert.Equal(t, "nico", NormalizeAPIName("nico"))
	})
}

func TestCurrentAPINameRewriteTransport(t *testing.T) {
	t.Run("nil client returns false", func(t *testing.T) {
		_, ok := CurrentAPINameRewriteTransport(nil)
		assert.False(t, ok)
	})

	t.Run("client without rewrite transport returns false", func(t *testing.T) {
		_, ok := CurrentAPINameRewriteTransport(&http.Client{})
		assert.False(t, ok)
	})

	t.Run("client with rewrite transport returns it", func(t *testing.T) {
		transport := NewAPINameRewriteTransport("nico", nil)
		client := &http.Client{Transport: transport}
		rewriter, ok := CurrentAPINameRewriteTransport(client)
		assert.True(t, ok)
		assert.Equal(t, "nico", rewriter.APIName())
	})
}
