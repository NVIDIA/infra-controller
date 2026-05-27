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

		// Add infrastructure_provider_id column to vpc_peering table
		_, err := tx.NewAddColumn().
			Model((*model.VpcPeering)(nil)).
			IfNotExists().
			ColumnExpr("infrastructure_provider_id uuid NULL").
			Exec(ctx)
		handleError(tx, err)

		// Add tenant_id column to vpc_peering table
		_, err = tx.NewAddColumn().
			Model((*model.VpcPeering)(nil)).
			IfNotExists().
			ColumnExpr("tenant_id uuid NULL").
			Exec(ctx)
		handleError(tx, err)

		// Drop constraint if it exists
		_, err = tx.Exec("ALTER TABLE vpc_peering DROP CONSTRAINT IF EXISTS vpc_peering_infrastructure_provider_id_fkey")
		handleError(tx, err)

		// Add foreign key constraint for infrastructure_provider_id
		_, err = tx.Exec("ALTER TABLE vpc_peering ADD CONSTRAINT vpc_peering_infrastructure_provider_id_fkey FOREIGN KEY (infrastructure_provider_id) REFERENCES public.infrastructure_provider(id)")
		handleError(tx, err)

		// Drop constraint if it exists
		_, err = tx.Exec("ALTER TABLE vpc_peering DROP CONSTRAINT IF EXISTS vpc_peering_tenant_id_fkey")
		handleError(tx, err)

		// Add foreign key constraint for tenant_id
		_, err = tx.Exec("ALTER TABLE vpc_peering ADD CONSTRAINT vpc_peering_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenant(id)")
		handleError(tx, err)

		// Drop index if it exists
		_, err = tx.Exec("DROP INDEX IF EXISTS idx_vpc_peering_infrastructure_provider_id")
		handleError(tx, err)

		// Add index for infrastructure_provider_id
		_, err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_vpc_peering_infrastructure_provider_id ON vpc_peering(infrastructure_provider_id)")
		handleError(tx, err)

		// Drop index if it exists
		_, err = tx.Exec("DROP INDEX IF EXISTS idx_vpc_peering_tenant_id")
		handleError(tx, err)

		// Add index for tenant_id
		_, err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_vpc_peering_tenant_id ON vpc_peering(tenant_id)")
		handleError(tx, err)

		// Drop canonical ordering uniqueness constraint if exists
		_, err = tx.Exec("DROP INDEX IF EXISTS uq_vpc_peering_canonical_vpc_pair")
		handleError(tx, err)

		// Add canonical ordering uniqueness constraint
		_, err = tx.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS uq_vpc_peering_canonical_vpc_pair
			ON vpc_peering (LEAST(vpc1_id, vpc2_id), GREATEST(vpc1_id, vpc2_id))
			WHERE deleted IS NULL
		`)
		handleError(tx, err)

		terr = tx.Commit()
		if terr != nil {
			handlePanic(terr, "failed to commit transaction")
		}

		fmt.Print(" [up migration] Added optional infrastructure_provider_id and tenant_id columns and enforced canonical VPC pair uniqueness. ")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] No action taken")
		return nil
	})
}
