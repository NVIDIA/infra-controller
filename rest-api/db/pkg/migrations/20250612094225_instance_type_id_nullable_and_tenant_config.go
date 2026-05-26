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

package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
)

func tenantConfigUpMigration(ctx context.Context, db *bun.DB) error {
	// Start transactions
	tx, terr := db.BeginTx(ctx, &sql.TxOptions{})
	if terr != nil {
		handlePanic(terr, "failed to begin transaction")
	}

	// Remove NOT NULL constraint from instance_type_id in instance table
	_, err := tx.Exec("ALTER TABLE instance ALTER COLUMN instance_type_id DROP NOT NULL;")
	handleError(tx, err)

	// Ensure existing column will get an empty JSON as default value
	_, err = tx.Exec("ALTER TABLE tenant ALTER COLUMN config SET DEFAULT '{}'::jsonb")
	handleError(tx, err)

	// Set the config column in tenant table to {}::jsonb
	_, err = tx.Exec("UPDATE tenant SET config='{}'::jsonb WHERE config IS NULL")
	handleError(tx, err)

	// Set the config column in tenant table to not null
	_, err = tx.Exec("ALTER TABLE tenant ALTER COLUMN config SET NOT NULL")
	handleError(tx, err)

	terr = tx.Commit()
	if terr != nil {
		handlePanic(terr, "failed to commit transaction")
	}

	fmt.Print(" [up migration] ")
	return nil
}

func init() {
	Migrations.MustRegister(tenantConfigUpMigration, func(_ context.Context, _ *bun.DB) error {
		fmt.Print(" [down migration] ")
		return nil
	})
}
