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
package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// humanReadableSize converts bytes to a human-readable format using binary prefixes
func HumanReadableSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// PrintPrettyResponse prints HTTP response status, headers, and attempts to pretty-print JSON bodies.
func PrintPrettyResponse(resp *http.Response) {
	// Print status
	fmt.Printf("Status: %s\n", resp.Status)

	// Print headers
	fmt.Println("Headers:")
	for key, value := range resp.Header {
		fmt.Printf("  %s: %s\n", key, value)
	}

	// Print body
	fmt.Println("Body:")
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %v\n", err)
	}
	defer resp.Body.Close()

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err != nil {
		// If the body is not JSON, just print it as a string
		fmt.Printf("Body is not JSON: %s\n", string(bodyBytes))
		return
	}

	fmt.Println(prettyJSON.String())
}

// GetGoroutineID returns the internal goroutine ID for the current goroutine.
func GetGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
