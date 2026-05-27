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

-- Connect as postgres user to template1 DB to run this script
-- Example: PGPASSWORD=postgres psql -U postgres -p 30432 -d template1 < scripts/setup.sql
-- Create extension for pg_trgm
CREATE EXTENSION IF NOT EXISTS pg_trgm;
-- Create Nico DB and user
CREATE DATABASE nico WITH ENCODING 'UTF8';
-- Password should be changed before use in environments deployed in Cloud
CREATE USER nico WITH PASSWORD 'nico';
-- Grant all privileges on Nico DB to Nico user
GRANT ALL PRIVILEGES ON DATABASE nico TO nico;
