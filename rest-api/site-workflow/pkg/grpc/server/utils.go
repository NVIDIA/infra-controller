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

package server

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	mrand "math/rand"
)

func generateMacAddress() string {
	buf := make([]byte, 6)
	rand.Read(buf)

	// Set the local bit
	buf[0] |= 2
	maca := fmt.Sprintf("Random MAC address: %02x:%02x:%02x:%02x:%02x:%02x\n", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])

	return maca
}

func generateInteger(max int) int {
	s := mrand.NewSource(time.Now().UnixNano())
	r := mrand.New(s)

	return r.Intn(max)
}

func generateIPAddress() string {
	buf := make([]byte, 4)

	ip := mrand.Uint32()
	binary.LittleEndian.PutUint32(buf, ip)

	return fmt.Sprintf("%s\n", net.IP(buf))
}

func getStrPtr(s string) *string {
	sp := s
	return &sp
}
