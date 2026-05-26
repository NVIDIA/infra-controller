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

package health

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Check captures the API response for workflow service health check
type Check struct {
	IsHealthy bool    `json:"is_healthy"`
	Error     *string `json:"error"`
}

// StatusHandler is an API handler to return health status of the workflow service
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	check := Check{
		IsHealthy: true,
	}
	bytes, err := json.Marshal(check)
	if err != nil {
		log.Error().Err(err).Msg("error converting health check object into JSON")
		http.Error(w, "failed to construct health check response", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		log.Error().Err(err).Msg("failed to return health check response")
	}
}
