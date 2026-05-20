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

func siteSshHostnameRenameUpMigration(ctx context.Context, db *bun.DB) error {
	// Start transaction
	tx, terr := db.BeginTx(ctx, &sql.TxOptions{})
	if terr != nil {
		handlePanic(terr, "failed to begin transaction")
	}

	// Rename site sshHostname column to serialConsoleHostname if that column exists
	res, err := tx.Exec("SELECT column_name FROM information_schema.columns WHERE table_name = 'site' AND column_name = 'ssh_hostname'")
	handleError(tx, err)
	sshHostnameRowsAffected, err := res.RowsAffected()
	res, err = tx.Exec("SELECT column_name FROM information_schema.columns WHERE table_name = 'site' AND column_name = 'serial_console_hostname'")
	handleError(tx, err)
	serialConsoleHostnameRowsAffected, err := res.RowsAffected()
	handleError(tx, err)
	if sshHostnameRowsAffected > 0 && serialConsoleHostnameRowsAffected == 0 {
		_, err := tx.Exec("ALTER TABLE site RENAME COLUMN ssh_hostname TO serial_console_hostname")
		handleError(tx, err)
	} else {
		fmt.Println("sshHostname rename to serialConsoleHostname: Migration skipped. Either the column does not exist or already renamed")
	}

	terr = tx.Commit()
	if terr != nil {
		handlePanic(terr, "failed to commit transaction")
	}

	fmt.Print(" [up migration] ")
	return nil
}

func init() {
	Migrations.MustRegister(siteSshHostnameRenameUpMigration, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		return nil
	})
}
