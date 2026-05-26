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
	"errors"
	"math"
)

// A convenience function for converting a pointer to
// a native Go integer to a pointer to a uint32 for
// use with a protobuf message. Accepts a pointer to
// an int and returns a uint32 pointer.
//
// If the input is nil, nil will be returned.
// If a pointer to a value greater than
// uint32 max is submitted, an error will be returned.
func GetIntPtrToUint32Ptr(i *int) (*uint32, error) {
	if i == nil {
		return nil, nil
	}

	if *i > math.MaxUint32 {
		return nil, errors.New("conversion to uint32 pointer would exceed uint32 max")
	}

	i32 := uint32(*i)

	return &i32, nil
}

// A convenience function for converting a pointer to
// a uint32 to a pointer to a an int.
//
// If the input is nil, nil will be returned.
func GetUint32PtrToIntPtr(u32 *uint32) *int {
	if u32 == nil {
		return nil
	}

	i := int(*u32)

	return &i
}

func GetUint32Ptr(i uint32) *uint32 {
	return (&i)
}
