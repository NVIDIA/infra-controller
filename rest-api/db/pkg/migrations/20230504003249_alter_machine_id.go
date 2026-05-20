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

func alterMachineIDUpMigration(ctx context.Context, db *bun.DB) error {
	// Start transactions
	tx, terr := db.BeginTx(ctx, &sql.TxOptions{})
	if terr != nil {
		handlePanic(terr, "failed to begin transaction")
	}

	// Drop foreign key constraints
	_, err := tx.Exec("ALTER TABLE machine_capability DROP CONSTRAINT machine_capability_machine_id_fkey")
	handleError(tx, err)

	_, err = tx.Exec("ALTER TABLE machine_instance_type DROP CONSTRAINT machine_instance_type_machine_id_fkey")
	handleError(tx, err)

	_, err = tx.Exec("ALTER TABLE machine_interface DROP CONSTRAINT machine_interface_machine_id_fkey")
	handleError(tx, err)

	_, err = tx.Exec("ALTER TABLE instance DROP CONSTRAINT instance_machine_id_fkey")
	handleError(tx, err)

	// Alter the type of id field in the machine table from uuid to string
	_, err = tx.Exec("ALTER TABLE machine ALTER COLUMN id TYPE TEXT USING id::text")
	handleError(tx, err)

	// Alter the type of machine_id field in the machine_capability table from uuid to string
	_, err = tx.Exec("ALTER TABLE machine_capability ALTER COLUMN machine_id TYPE TEXT USING machine_id::text")
	handleError(tx, err)

	// Alter the type of machine_id field in the machine_instance_type table from uuid to string
	_, err = tx.Exec("ALTER TABLE machine_instance_type ALTER COLUMN machine_id TYPE TEXT USING machine_id::text")
	handleError(tx, err)

	// Alter the type of machine_id field in the machine_interface table from uuid to string
	_, err = tx.Exec("ALTER TABLE machine_interface ALTER COLUMN machine_id TYPE TEXT USING machine_id::text")
	handleError(tx, err)

	// Alter the type of machine_id field in the instance table from uuid to string
	_, err = tx.Exec("ALTER TABLE instance ALTER COLUMN machine_id TYPE TEXT USING machine_id::text")
	handleError(tx, err)

	// Copy Controller Machine ID to Machine ID
	machines := []model.Machine{}
	err = tx.NewSelect().Model(&machines).Scan(ctx)
	handleError(tx, err)

	count := 0
	for _, machine := range machines {
		curMachine := machine

		if machine.Deleted != nil {
			fmt.Printf("Machine: %v is deleted", machine.ID)
			continue
		}

		// Update Machine Capability records for each Machine
		_, err = tx.Exec("UPDATE machine_capability SET machine_id = ? WHERE machine_id = ?", curMachine.ControllerMachineID, curMachine.ID)
		handleError(tx, err)

		// Update Machine Instance Type records for each Machine
		_, err = tx.Exec("UPDATE machine_instance_type SET machine_id = ? WHERE machine_id = ?", curMachine.ControllerMachineID, curMachine.ID)
		handleError(tx, err)

		// Update Machine Interface records for each Machine
		_, err = tx.Exec("UPDATE machine_interface SET machine_id = ? WHERE machine_id = ?", curMachine.ControllerMachineID, curMachine.ID)
		handleError(tx, err)

		// Update Instance records for each Machine
		_, err = tx.Exec("UPDATE instance SET machine_id = ? WHERE machine_id = ?", curMachine.ControllerMachineID, curMachine.ID)
		handleError(tx, err)

		// Update Machine ID
		_, err = tx.Exec("UPDATE machine SET id = ? WHERE id = ?", curMachine.ControllerMachineID, curMachine.ID)
		if err != nil {
			fmt.Printf("Failed to update machine id for machine: %v", curMachine.ID)
		}
		handleError(tx, err)

		count++
	}

	fmt.Printf("Updated %v machines\n", count)

	// Add back foreign key constraint
	_, err = tx.Exec("ALTER TABLE machine_capability ADD CONSTRAINT machine_capability_machine_id_fkey FOREIGN KEY (machine_id) REFERENCES public.machine(id)")
	handleError(tx, err)

	_, err = tx.Exec("ALTER TABLE machine_instance_type ADD CONSTRAINT machine_instance_type_machine_id_fkey FOREIGN KEY (machine_id) REFERENCES public.machine(id)")
	handleError(tx, err)

	_, err = tx.Exec("ALTER TABLE machine_interface ADD CONSTRAINT machine_interface_machine_id_fkey FOREIGN KEY (machine_id) REFERENCES public.machine(id)")
	handleError(tx, err)

	_, err = tx.Exec("ALTER TABLE instance ADD CONSTRAINT instance_machine_id_fkey FOREIGN KEY (machine_id) REFERENCES public.machine(id)")
	handleError(tx, err)

	terr = tx.Commit()
	if terr != nil {
		handlePanic(terr, "failed to commit transaction")
	}

	fmt.Print(" [up migration] ")
	return nil
}

func init() {
	Migrations.MustRegister(alterMachineIDUpMigration, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		return nil
	})
}
