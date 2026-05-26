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

package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCronTrigger(t *testing.T) {
	t.Run("valid 6-field expression succeeds", func(t *testing.T) {
		// "0 30 9 * * 1-5": fire at 09:30:00 on weekdays (Mon–Fri).
		// Requires WithSeconds() — the leading "0" is the seconds field.
		trig, err := NewCronTrigger("0 30 9 * * 1-5")
		require.NoError(t, err)
		assert.NotNil(t, trig)
		assert.Equal(t, "cron(0 30 9 * * 1-5)", trig.Description())
	})

	t.Run("5-field expression is rejected", func(t *testing.T) {
		// Standard 5-field cron (no seconds field) must be rejected because
		// WithSeconds() enforces exactly 6 fields; accepting a 5-field
		// expression would silently shift field semantics.
		_, err := NewCronTrigger("30 9 * * 1-5")
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "30 9 * * 1-5"),
			"error should quote the rejected expression, got: %s", err.Error())
	})
}
