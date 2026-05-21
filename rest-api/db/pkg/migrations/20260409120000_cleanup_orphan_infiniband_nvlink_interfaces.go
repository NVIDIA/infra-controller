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

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		tx, terr := db.BeginTx(ctx, &sql.TxOptions{})
		if terr != nil {
			handlePanic(terr, "failed to begin transaction")
		}

		// Soft-delete InfiniBand interfaces with no active instance (missing FK target or instance soft-deleted).
		_, err := tx.Exec(`
			UPDATE infiniband_interface ibi
			SET deleted = CURRENT_TIMESTAMP, updated = CURRENT_TIMESTAMP
			WHERE ibi.deleted IS NULL
			AND ibi.instance_id NOT IN (SELECT id FROM instance WHERE deleted IS NULL)`)
		handleError(tx, err)

		// Soft-delete NVLink interfaces with no active instance.
		_, err = tx.Exec(`
			UPDATE nvlink_interface nvli
			SET deleted = CURRENT_TIMESTAMP, updated = CURRENT_TIMESTAMP
			WHERE nvli.deleted IS NULL
			AND nvli.instance_id NOT IN (SELECT id FROM instance WHERE deleted IS NULL)`)
		handleError(tx, err)

		terr = tx.Commit()
		if terr != nil {
			handlePanic(terr, "failed to commit transaction")
		}

		fmt.Print(" [up migration] Soft-deleted orphan infiniband_interface and nvlink_interface rows. ")
		return nil
	}, func(_ context.Context, _ *bun.DB) error {
		fmt.Print(" [down migration] No-op (data cleanup cannot be reversed). ")
		return nil
	})
}
