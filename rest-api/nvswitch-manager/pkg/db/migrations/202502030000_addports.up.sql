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
-- Add port columns to nvswitch table for custom BMC/NVOS ports
-- This allows connecting through tunnels or non-standard ports
--

-- Add BMC port (default 443 for Redfish HTTPS)
ALTER TABLE public.nvswitch
    ADD COLUMN bmc_port integer NOT NULL DEFAULT 443;

-- Add NVOS port (default 22 for SSH)
ALTER TABLE public.nvswitch
    ADD COLUMN nvos_port integer NOT NULL DEFAULT 22;

-- Update existing rows to have default ports (redundant but explicit)
UPDATE public.nvswitch SET bmc_port = 443 WHERE bmc_port IS NULL;
UPDATE public.nvswitch SET nvos_port = 22 WHERE nvos_port IS NULL;
