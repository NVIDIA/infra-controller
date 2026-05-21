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
package inventorymanager

import (
	"context"
	"fmt"
	"github.com/NVIDIA/infra-controller-rest/powershelf-manager/pkg/common/runner"
	"github.com/NVIDIA/infra-controller-rest/powershelf-manager/pkg/objects/powershelf"
	"github.com/NVIDIA/infra-controller-rest/powershelf-manager/pkg/pmcmanager"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"

	"time"
)

var (
	inventory            sync.Map
	collector            *runner.Runner
	collectorWaiterSleep = time.Second * 30
)

func getPowershelf(mac net.HardwareAddr) (*powershelf.PowerShelf, error) {
	value, exists := inventory.Load(mac.String())
	if !exists {
		return nil, fmt.Errorf("could not find an entry with PMC MAC %s in the powershelf inventory", mac)
	}

	powershelf, ok := value.(*powershelf.PowerShelf)
	if !ok {
		return nil, fmt.Errorf("found an entry with PMC MAC %s but it is not a powershelf", mac)
	}

	return powershelf, nil
}

func GetPowershelves(macs []net.HardwareAddr) []*powershelf.PowerShelf {
	var shelves []*powershelf.PowerShelf
	for _, mac := range macs {
		ps, err := getPowershelf(mac)
		if err != nil {
			log.Printf("failed to get powershelf for PMC MAC %s: %v", mac.String(), err)
			continue
		}
		shelves = append(shelves, ps)
	}
	return shelves
}

func GetAllPowershelves() []*powershelf.PowerShelf {
	var shelves []*powershelf.PowerShelf
	inventory.Range(func(key, value interface{}) bool {
		if ps, ok := value.(*powershelf.PowerShelf); ok {
			shelves = append(shelves, ps)
		}
		return true
	})
	return shelves
}

func Start(registry *pmcmanager.PmcManager) error {
	collector = runner.New("inventory collector", func() interface{} { return registry }, collectorWaiter, collectorRunner)
	return nil
}

func Stop() error {
	collector.Stop()
	return nil
}

func collectorWaiter(ctx interface{}) interface{} {
	log.Println("Inventory Collector: Waiter")
	time.Sleep(collectorWaiterSleep)
	return nil
}

func collectorRunner(ctx interface{}, task interface{}) {
	start := time.Now()
	log.Println("Inventory Collector: Runner")

	pmcManager := ctx.(*pmcmanager.PmcManager)
	pmcs, err := pmcManager.GetAllPmcs(context.Background())
	if err != nil {
		log.Printf("failed to query PMC registry: %v\n", err)
		return
	}

	total := len(pmcs)
	successCount := 0
	failureCount := 0
	for i, pmc := range pmcs {
		pmcMAC := pmc.MAC.String()
		powershelf, err := pmcManager.QueryPowerShelf(context.Background(), pmc)
		if err != nil {
			log.Printf("failed to query powershelf for pmc %s: %v", pmcMAC, err)
			failureCount++
			continue
		}

		inventory.Store(pmcMAC, powershelf)
		successCount++
		if (i+1)%10 == 0 || i == total-1 {
			log.Printf("Inventory Collector: processed %d/%d PMCs", i+1, total)
		}
	}

	log.Printf("Inventory Collector: finished collecting (Success: %d; Failure: %d) in %s", successCount, failureCount, time.Since(start))
}
