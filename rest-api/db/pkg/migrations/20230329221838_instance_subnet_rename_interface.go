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

func renameInstanceSubnetToInterfaceUpMigration(ctx context.Context, db *bun.DB) error {
	// Start transaction
	tx, terr := db.BeginTx(ctx, &sql.TxOptions{})
	if terr != nil {
		handlePanic(terr, "failed to begin transaction")
		return terr
	}

	// Rename instance_subnet table to interface if it exists
	_, err := tx.Exec("ALTER TABLE IF EXISTS instance_subnet RENAME TO interface")
	handleError(tx, err)

	// Drop the older index if exists
	_, err = tx.Exec("DROP INDEX IF EXISTS instance_subnet_status_idx")
	handleError(tx, err)

	// Drop if the index exists (won't occur/harmless in dev/stage/prod but helps with test)
	_, err = tx.Exec("DROP INDEX IF EXISTS interface_status_idx")
	handleError(tx, err)

	// Add status index for interface (formerly instance_subnet) model
	_, err = tx.Exec("CREATE INDEX interface_status_idx ON public.interface(status) WHERE deleted IS NULL")
	handleError(tx, err)

	terr = tx.Commit()
	if terr != nil {
		handlePanic(terr, "failed to commit transaction")
		return terr
	}

	fmt.Print(" [up migration] ")
	return nil
}

func init() {
	Migrations.MustRegister(renameInstanceSubnetToInterfaceUpMigration, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		return nil
	})
}
