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
-- Migration: Add async worker pool support
-- Adds exec_context and last_checked_at for non-blocking firmware update workers
--

-- Add exec_context JSONB column for persisting async execution state
-- This stores TaskURI (Redfish), PID (Script/SSH), reachability tracking, etc.
ALTER TABLE public.firmware_update
    ADD COLUMN exec_context JSONB;

-- Add last_checked_at timestamp for poll timing
-- Workers use this to determine when an active update needs re-polling
ALTER TABLE public.firmware_update
    ADD COLUMN last_checked_at TIMESTAMP WITH TIME ZONE;

-- Index on last_checked_at for efficient polling queries
-- The ClaimNextWorkItem query orders by last_checked_at ASC NULLS FIRST
-- to prioritize updates that haven't been checked recently
CREATE INDEX firmware_update_last_checked_idx 
    ON public.firmware_update (last_checked_at ASC NULLS FIRST)
    WHERE state NOT IN ('QUEUED', 'COMPLETED', 'FAILED', 'CANCELLED');

-- Update the active index to include CANCELLED state
DROP INDEX IF EXISTS firmware_update_active_idx;
CREATE INDEX firmware_update_active_idx 
    ON public.firmware_update (switch_uuid, component) 
    WHERE state NOT IN ('COMPLETED', 'FAILED', 'CANCELLED');
