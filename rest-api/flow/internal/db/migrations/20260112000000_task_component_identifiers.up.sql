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

--
-- Migration: Update task table to rename component_ids to component_uuids
-- (1 task = 1 rack principle, rack_id already exists)
--

-- Add component_uuids column
ALTER TABLE public.task
    ADD COLUMN component_uuids jsonb;

-- Migrate existing data: copy component_ids to component_uuids
UPDATE public.task
SET component_uuids = component_ids
WHERE component_ids IS NOT NULL;

-- Drop old component_ids column
ALTER TABLE public.task
    DROP COLUMN component_ids;
