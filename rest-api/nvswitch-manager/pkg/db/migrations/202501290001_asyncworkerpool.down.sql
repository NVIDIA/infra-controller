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
-- Migration rollback: Remove async worker pool support
--

-- Drop the new index
DROP INDEX IF EXISTS firmware_update_last_checked_idx;

-- Restore the original active index (without CANCELLED)
DROP INDEX IF EXISTS firmware_update_active_idx;
CREATE INDEX firmware_update_active_idx 
    ON public.firmware_update (switch_uuid, component) 
    WHERE state NOT IN ('COMPLETED', 'FAILED');

-- Remove the new columns
ALTER TABLE public.firmware_update
    DROP COLUMN IF EXISTS last_checked_at;

ALTER TABLE public.firmware_update
    DROP COLUMN IF EXISTS exec_context;
