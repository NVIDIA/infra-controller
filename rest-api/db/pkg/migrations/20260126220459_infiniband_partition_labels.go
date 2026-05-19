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

		// Add labels column to infiniband_partition table
		_, err := tx.NewAddColumn().Model((*model.InfiniBandPartition)(nil)).IfNotExists().ColumnExpr("labels JSONB NOT NULL DEFAULT ('{}')").Exec(ctx)
		handleError(tx, err)

		// Drop if the existing infiniband_partition_tsv_idx exists
		_, err = tx.Exec("DROP INDEX IF EXISTS infiniband_partition_tsv_idx")
		handleError(tx, err)

		// Add tsv index which includes labels for infiniband_partition table
		_, err = tx.Exec("CREATE INDEX infiniband_partition_tsv_idx ON infiniband_partition USING gin(to_tsvector('english', name || ' ' || description || ' ' || partition_key || ' ' || partition_name || ' ' || status || ' ' || labels::text))")
		handleError(tx, err)

		terr = tx.Commit()
		if terr != nil {
			handlePanic(terr, "failed to commit transaction")
		}

		fmt.Print(" [up migration] Added 'labels' column to 'infiniband_partition' table successfully. ")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		return nil
	})
}
