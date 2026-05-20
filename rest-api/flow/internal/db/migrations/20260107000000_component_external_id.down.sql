-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
-- http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

-- Restore associated_id column
ALTER TABLE component ADD COLUMN IF NOT EXISTS associated_id character varying;

-- Drop unique index on (type, external_id)
DROP INDEX IF EXISTS component_type_external_id_idx;

-- Drop new index
DROP INDEX IF EXISTS component_external_id_idx;

-- Rename external_id back to machine_id
ALTER TABLE component RENAME COLUMN external_id TO machine_id;

-- Recreate old index
CREATE INDEX component_machine_id ON component (machine_id);
