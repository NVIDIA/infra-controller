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
		// Start transactions
		tx, terr := db.BeginTx(ctx, &sql.TxOptions{})
		if terr != nil {
			handlePanic(terr, "failed to begin transaction")
		}

		// Drop if the index exists (won't occur/harmless in dev/stage/prod but helps with test)
		_, err := tx.Exec("DROP INDEX IF EXISTS allocation_gin_idx")
		handleError(tx, err)

		// Add GIN index for allocation table
		_, err = tx.Exec("CREATE INDEX allocation_gin_idx ON public.allocation USING GIN (name gin_trgm_ops, description gin_trgm_ops, status gin_trgm_ops)")
		handleError(tx, err)

		// Drop if the index exists (won't occur/harmless in dev/stage/prod but helps with test)
		_, err = tx.Exec("DROP INDEX IF EXISTS instance_gin_idx")
		handleError(tx, err)

		// Add GIN index for instance table
		_, err = tx.Exec("CREATE INDEX instance_gin_idx ON public.instance USING GIN (name gin_trgm_ops, status gin_trgm_ops)")
		handleError(tx, err)

		// Drop if the index exists (won't occur/harmless in dev/stage/prod but helps with test)
		_, err = tx.Exec("DROP INDEX IF EXISTS instance_type_gin_idx")
		handleError(tx, err)

		// Add GIN index for instance_type table
		_, err = tx.Exec("CREATE INDEX instance_type_gin_idx ON public.instance_type USING GIN (name gin_trgm_ops, display_name gin_trgm_ops, description gin_trgm_ops, status gin_trgm_ops)")
		handleError(tx, err)

		// Drop if the index exists (won't occur/harmless in dev/stage/prod but helps with test)
		_, err = tx.Exec("DROP INDEX IF EXISTS ip_block_gin_idx")
		handleError(tx, err)

		// Add GIN index for ip_block table
		_, err = tx.Exec("CREATE INDEX ip_block_gin_idx ON public.ip_block USING GIN (name gin_trgm_ops, description gin_trgm_ops, status gin_trgm_ops)")
		handleError(tx, err)

		// Drop if the index exists (won't occur/harmless in dev/stage/prod but helps with test)
		_, err = tx.Exec("DROP INDEX IF EXISTS operating_system_gin_idx")
		handleError(tx, err)

		// Add GIN index for operating_system table
		_, err = tx.Exec("CREATE INDEX operating_system_gin_idx ON public.operating_system USING GIN (name gin_trgm_ops, description gin_trgm_ops, status gin_trgm_ops)")
		handleError(tx, err)

		// Drop if the index exists (won't occur/harmless in dev/stage/prod but helps with test)
		_, err = tx.Exec("DROP INDEX IF EXISTS site_gin_idx")
		handleError(tx, err)

		// Add GIN index for site table
		_, err = tx.Exec("CREATE INDEX site_gin_idx ON public.site USING GIN (name gin_trgm_ops, description gin_trgm_ops, status gin_trgm_ops)")
		handleError(tx, err)

		// Drop if the index exists (won't occur/harmless in dev/stage/prod but helps with test)
		_, err = tx.Exec("DROP INDEX IF EXISTS ssh_key_gin_idx")
		handleError(tx, err)

		// Add GIN index for ssh_key table
		_, err = tx.Exec("CREATE INDEX ssh_key_gin_idx ON public.ssh_key USING GIN (name gin_trgm_ops)")
		handleError(tx, err)

		// Drop if the index exists (won't occur/harmless in dev/stage/prod but helps with test)
		_, err = tx.Exec("DROP INDEX IF EXISTS sshkey_group_gin_idx")
		handleError(tx, err)

		// Add GIN index for ssh_key table
		_, err = tx.Exec("CREATE INDEX sshkey_group_gin_idx ON public.sshkey_group USING GIN (name gin_trgm_ops)")
		handleError(tx, err)

		// Drop if the index exists (won't occur/harmless in dev/stage/prod but helps with test)
		_, err = tx.Exec("DROP INDEX IF EXISTS subnet_gin_idx")
		handleError(tx, err)

		// Add GIN index for subnet table
		_, err = tx.Exec("CREATE INDEX subnet_gin_idx ON public.subnet USING GIN (name gin_trgm_ops, description gin_trgm_ops, status gin_trgm_ops)")
		handleError(tx, err)

		// Drop if the index exists (won't occur/harmless in dev/stage/prod but helps with test)
		_, err = tx.Exec("DROP INDEX IF EXISTS vpc_gin_idx")
		handleError(tx, err)

		// Add GIN index for vpc table
		_, err = tx.Exec("CREATE INDEX vpc_gin_idx ON public.vpc USING GIN (name gin_trgm_ops, description gin_trgm_ops, status gin_trgm_ops)")
		handleError(tx, err)

		terr = tx.Commit()
		if terr != nil {
			handlePanic(terr, "failed to commit transaction")
		}

		fmt.Print(" [up migration] ")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		return nil
	})
}
