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

package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// ErrorChecker abstracts database error classification.
type ErrorChecker interface {
	IsErrNoRows(err error) bool
	IsUniqueConstraintError(err error) bool
}

// PostgresErrorChecker classifies common Postgres errors such as
// no rows and unique constraint violations.
type PostgresErrorChecker struct{}

func (checker *PostgresErrorChecker) IsErrNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func (checker *PostgresErrorChecker) IsUniqueConstraintError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}

// CurTime returns the current UTC time rounded to microseconds
// (useful for DB timestamps).
func CurTime() time.Time {
	return time.Now().UTC().Round(time.Microsecond)
}
