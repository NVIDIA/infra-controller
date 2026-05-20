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

-- Add rack_id column to nvswitch table
-- This allows associating NV-Switch trays with physical racks for filtering and batch operations

ALTER TABLE nvswitch ADD COLUMN rack_id VARCHAR(64);

-- Index for efficient filtering by rack
CREATE INDEX nvswitch_rack_id_idx ON nvswitch(rack_id) WHERE rack_id IS NOT NULL;
