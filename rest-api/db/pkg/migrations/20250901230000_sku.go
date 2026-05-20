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
		tx, terr := db.BeginTx(ctx, &sql.TxOptions{})
		if terr != nil {
			handlePanic(terr, "failed to begin transaction")
		}

		// create SKU table
		_, err := tx.NewCreateTable().Model((*model.SKU)(nil)).IfNotExists().Exec(ctx)
		handleError(tx, err)

		// index for sku_id
		_, err = tx.Exec("DROP INDEX IF EXISTS sku_id_idx")
		handleError(tx, err)
		_, err = tx.Exec("CREATE INDEX sku_id_idx ON sku(id)")
		handleError(tx, err)

		// index for site_id
		_, err = tx.Exec("DROP INDEX IF EXISTS sku_site_id_idx")
		handleError(tx, err)
		_, err = tx.Exec("CREATE INDEX sku_site_id_idx ON sku(site_id)")
		handleError(tx, err)

		// index for created
		_, err = tx.Exec("DROP INDEX IF EXISTS sku_created_idx")
		handleError(tx, err)
		_, err = tx.Exec("CREATE INDEX sku_created_idx ON sku(created)")
		handleError(tx, err)

		// index for updated
		_, err = tx.Exec("DROP INDEX IF EXISTS sku_updated_idx")
		handleError(tx, err)
		_, err = tx.Exec("CREATE INDEX sku_updated_idx ON sku(updated)")
		handleError(tx, err)

		terr = tx.Commit()
		if terr != nil {
			handlePanic(terr, "failed to commit transaction")
		}

		fmt.Print(" [up migration] ")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] No action taken")
		return nil
	})
}
