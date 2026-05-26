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
package firmwaremanager

import (
	"fmt"
)

// firmwareVersion is a semantic version for PMC firmware (major.minor.patch) with comparison helpers.
type firmwareVersion struct {
	major, minor, patch int
}

// String returns the canonical string form of the version.
func (fv firmwareVersion) String() string {
	return fmt.Sprintf("r%d.%d.%d", fv.major, fv.minor, fv.patch)
}

// fwVersionFromStr parses a version string of the form r<major>.<minor>.<patch>.
func fwVersionFromStr(s string) (firmwareVersion, error) {
	var fw firmwareVersion

	_, err := fmt.Sscanf(s, "r%d.%d.%d",
		&fw.major, &fw.minor, &fw.patch)

	return fw, err
}

// cmp compares two versions; returns -1 if fv<other, 0 if equal, 1 if fv>other.
func (fv firmwareVersion) cmp(other firmwareVersion) int {
	if fv.major != other.major {
		if fv.major < other.major {
			return -1
		}
		return 1
	}
	if fv.minor != other.minor {
		if fv.minor < other.minor {
			return -1
		}
		return 1
	}
	if fv.patch != other.patch {
		if fv.patch < other.patch {
			return -1
		}
		return 1
	}
	return 0
}
