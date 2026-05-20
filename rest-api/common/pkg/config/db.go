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

package config

import "fmt"

// DBConfig holds configuration for database access
type DBConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

// NewDBConfig initializes and returns a configuration object for managing database access
func NewDBConfig(host string, port int, name string, user string, password string) *DBConfig {
	return &DBConfig{
		Host:     host,
		Port:     port,
		Name:     name,
		User:     user,
		Password: password,
	}
}

// GetHostPort returns the concatenated host & port.
func (dbcfg *DBConfig) GetHostPort() string {
	return fmt.Sprintf("%v:%v", dbcfg.Host, dbcfg.Port)
}
