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

package bmc

import "net"

// MACAddress wraps net.HardwareAddr and implements encoding.TextMarshaler and
// encoding.TextUnmarshaler so that JSON serialization produces the human-readable
// colon-separated MAC address string (e.g. "aa:bb:cc:dd:ee:ff") rather than a
// base64-encoded byte array.
type MACAddress struct {
	net.HardwareAddr
}

// MarshalText implements encoding.TextMarshaler.
func (m MACAddress) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (m *MACAddress) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		m.HardwareAddr = nil
		return nil
	}

	addr, err := net.ParseMAC(string(text))
	if err != nil {
		return err
	}

	m.HardwareAddr = addr
	return nil
}
