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

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Start transactions
		tx, terr := db.BeginTx(ctx, &sql.TxOptions{})
		if terr != nil {
			handlePanic(terr, "failed to begin transaction")
		}

		/*
		  In step 1, we migrated the data from capabilities to config.
		  Now, in step 2, we should be able update the model and DB to remove the old column.
		  And we also need to default certain options to ON.
		*/

		// Add config column
		_, err := tx.NewAddColumn().Model((*model.Site)(nil)).IfNotExists().ColumnExpr("config jsonb DEFAULT '{}'::jsonb").Exec(ctx)
		handleError(tx, err)

		// Set some defaults for site config
		_, err = tx.Exec(`ALTER TABLE site ALTER COLUMN config SET DEFAULT '{"native_networking": true, "network_security_group": true}'::jsonb`)
		handleError(tx, err)

		// Drop the old column
		_, err = tx.Exec(`ALTER TABLE site DROP COLUMN IF EXISTS capabilities`)
		handleError(tx, err)

		terr = tx.Commit()
		if terr != nil {
			handlePanic(terr, "failed to commit transaction")
		}

		fmt.Print(" [up migration] Site capabilities migration step 2 completed successfully. ")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		// Intentionally no-op: this migration drops a legacy column and mutates config defaults/data.
		fmt.Print(" [down migration] no-op (irreversible) ")
		return nil
	})
}
