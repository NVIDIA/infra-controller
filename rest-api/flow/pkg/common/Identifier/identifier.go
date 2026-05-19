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

package identifier

import (
	"errors"

	"github.com/google/uuid"
)

type Identifier struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func New(id uuid.UUID, name string) *Identifier {
	return &Identifier{
		ID:   id,
		Name: name,
	}
}

func (id *Identifier) Validate() error {
	if id == nil {
		return errors.New("identifier is not specfied")
	}

	if id.Name == "" {
		return errors.New("identifier name is not specfied")
	}

	if id.ID == uuid.Nil {
		return errors.New("identifier id is not specfied")
	}

	return nil
}

func (id *Identifier) ValidateAtLeastOne() bool {
	return id != nil && (id.ID != uuid.Nil || id.Name != "")
}
