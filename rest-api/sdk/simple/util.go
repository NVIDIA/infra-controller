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

package simple

// Int32PtrToIntPtr converts a *int32 to a *int
func Int32PtrToIntPtr(i *int32) *int {
	if i == nil {
		return nil
	}
	ret := int(*i)
	return &ret
}

// IntPtrToInt32Ptr converts a *int to a *int32
func IntPtrToInt32Ptr(i *int) *int32 {
	if i == nil {
		return nil
	}
	ret := int32(*i)
	return &ret
}

// StringPtr returns a pointer to the provided string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the provided int
func IntPtr(i int) *int {
	return &i
}
