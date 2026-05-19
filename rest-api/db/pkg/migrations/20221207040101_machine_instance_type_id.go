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

func machineInstanceTypeIDUpMigration(ctx context.Context, db *bun.DB) error {
	// Start transactions
	tx, terr := db.BeginTx(ctx, &sql.TxOptions{})
	if terr != nil {
		handlePanic(terr, "failed to begin transaction")
	}

	// Add instance_type_id column to machine table
	_, err := tx.NewAddColumn().Model((*model.Machine)(nil)).IfNotExists().ColumnExpr("instance_type_id UUID NULL").Exec(ctx)
	handleError(tx, err)

	// Add InstanceType foreign key for machine
	// Drop if one exists (won't occur/harmless in dev/stage/prod but helps with test)
	_, err = tx.Exec("ALTER TABLE machine DROP CONSTRAINT IF EXISTS machine_instance_type_id_fkey")
	handleError(tx, err)

	_, err = tx.Exec("ALTER TABLE machine ADD CONSTRAINT machine_instance_type_id_fkey FOREIGN KEY (instance_type_id) REFERENCES public.instance_type(id)")
	handleError(tx, err)

	// Update all machine table entries with instance_type_id
	machineInstanceTypes := []model.MachineInstanceType{}
	err = tx.NewSelect().Model(&machineInstanceTypes).Scan(ctx)
	handleError(tx, err)

	for _, machineInstanceType := range machineInstanceTypes {
		_, err = tx.NewUpdate().Model(&model.Machine{ID: machineInstanceType.MachineID}).Set("instance_type_id = ?", machineInstanceType.InstanceTypeID).WherePK().Exec(ctx)
		handleError(tx, err)
	}

	terr = tx.Commit()
	if terr != nil {
		handlePanic(terr, "failed to commit transaction")
	}

	fmt.Print(" [up migration] ")
	return nil
}

func init() {
	Migrations.MustRegister(machineInstanceTypeIDUpMigration, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		return nil
	})
}
