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

-- Drop firmware_update table (must drop first due to FK constraint)
DROP INDEX IF EXISTS public.firmware_update_predecessor_idx;
DROP INDEX IF EXISTS public.firmware_update_bundle_idx;
DROP INDEX IF EXISTS public.firmware_update_active_idx;
DROP INDEX IF EXISTS public.firmware_update_switch_uuid_idx;
DROP INDEX IF EXISTS public.firmware_update_state_created_idx;
DROP INDEX IF EXISTS public.firmware_update_created_at_idx;
DROP INDEX IF EXISTS public.firmware_update_state_idx;
DROP TABLE IF EXISTS public.firmware_update;

-- Drop nvswitch table
DROP INDEX IF EXISTS public.nvswitch_vendor_idx;
DROP TABLE IF EXISTS public.nvswitch;
