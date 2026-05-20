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

	"github.com/NVIDIA/infra-controller-rest/db/pkg/db/model"
	"github.com/uptrace/bun"
)

func operatingSystemImageAttributeUpMigration(ctx context.Context, db *bun.DB) error {
	// Start transactions
	tx, terr := db.BeginTx(ctx, &sql.TxOptions{})
	if terr != nil {
		handlePanic(terr, "failed to begin transaction")
	}

	// Add type column to OperatingSystem table
	_, err := tx.NewAddColumn().Model((*model.OperatingSystem)(nil)).IfNotExists().ColumnExpr("type varchar").Exec(ctx)
	handleError(tx, err)

	// Add image attributes columns to OperatingSystem table
	_, err = tx.NewAddColumn().Model((*model.OperatingSystem)(nil)).IfNotExists().ColumnExpr("image_sha varchar").Exec(ctx)
	handleError(tx, err)

	_, err = tx.NewAddColumn().Model((*model.OperatingSystem)(nil)).IfNotExists().ColumnExpr("image_auth_type varchar").Exec(ctx)
	handleError(tx, err)

	_, err = tx.NewAddColumn().Model((*model.OperatingSystem)(nil)).IfNotExists().ColumnExpr("image_auth_token varchar").Exec(ctx)
	handleError(tx, err)

	_, err = tx.NewAddColumn().Model((*model.OperatingSystem)(nil)).IfNotExists().ColumnExpr("image_disk varchar").Exec(ctx)
	handleError(tx, err)

	_, err = tx.NewAddColumn().Model((*model.OperatingSystem)(nil)).IfNotExists().ColumnExpr("root_fs_id varchar").Exec(ctx)
	handleError(tx, err)

	// Update Type record for each Operating System
	res, err := tx.Exec("UPDATE operating_system SET type = 'iPXE'")
	handleError(tx, err)

	osRowAffected, _ := res.RowsAffected()
	fmt.Printf("Updated %v operating systems \n", osRowAffected)

	// Make type column not nullable after we updated each row to default value
	_, err = tx.Exec("ALTER TABLE operating_system ALTER COLUMN type SET NOT NULL")
	handleError(tx, err)

	terr = tx.Commit()
	if terr != nil {
		handlePanic(terr, "failed to commit transaction")
	}

	fmt.Print(" [up migration] ")
	return nil
}

func init() {
	Migrations.MustRegister(operatingSystemImageAttributeUpMigration, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		return nil
	})
}
