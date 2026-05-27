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

package Query

import (
	"errors"
)

const (
	OrderAscending  OrderDirection = "ASC"
	OrderDescending OrderDirection = "DESC"
)

type OrderDirection string

const DefaultPaginationLimit = 100

type Pagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
}

// DefaultPagination returns a Pagination with offset 0 and the default limit.
func DefaultPagination() *Pagination {
	return &Pagination{Offset: 0, Limit: DefaultPaginationLimit}
}

func (p *Pagination) Validate() error {
	if p == nil {
		return nil
	}

	if p.Offset < 0 {
		return errors.New("offset must not be negative")
	}

	if p.Limit <= 0 {
		return errors.New("limit must be greater than 0")
	}

	return nil
}

type OrderBy struct {
	Column    string         `json:"column"`
	Direction OrderDirection `json:"direction"`
}

func (ob *OrderBy) Validate() error {
	if ob == nil {
		return nil
	}

	if ob.Column == "" {
		return errors.New("column is required")
	}

	if ob.Direction != OrderAscending && ob.Direction != OrderDescending {
		return errors.New("direction must be ASC or DESC")
	}

	return nil
}

func (ob *OrderBy) String() string {
	if ob.Direction == "" {
		return ob.Column
	}
	return ob.Column + " " + string(ob.Direction)
}
