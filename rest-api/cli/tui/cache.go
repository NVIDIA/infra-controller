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

package tui

import "time"

// NamedItem is a generic resource with a human-readable name and UUID.
type NamedItem struct {
	Name   string
	ID     string
	Status string
	Labels map[string]string
	Extra  map[string]string
	Raw    interface{}
}

// Cache holds lazy-loaded resources for the interactive session with a TTL.
type Cache struct {
	items   map[string][]NamedItem
	fetched map[string]time.Time
	ttl     time.Duration
}

func NewCache() *Cache {
	return &Cache{
		items:   make(map[string][]NamedItem),
		fetched: make(map[string]time.Time),
		ttl:     30 * time.Second,
	}
}

func (c *Cache) Get(resourceType string) []NamedItem {
	fetched, ok := c.fetched[resourceType]
	if !ok || time.Since(fetched) > c.ttl {
		return nil
	}
	return c.items[resourceType]
}

func (c *Cache) Set(resourceType string, items []NamedItem) {
	c.items[resourceType] = items
	c.fetched[resourceType] = time.Now()
}

func (c *Cache) Invalidate(resourceType string) {
	delete(c.items, resourceType)
	delete(c.fetched, resourceType)
}

func (c *Cache) InvalidateAll() {
	c.items = make(map[string][]NamedItem)
	c.fetched = make(map[string]time.Time)
}

func (c *Cache) InvalidateFiltered() {
	for _, rt := range []string{"vpc", "subnet", "instance", "instance-type",
		"allocation", "machine", "ip-block", "operating-system",
		"ssh-key-group", "network-security-group",
		"vpc-prefix", "rack", "expected-machine", "sku",
		"dpu-extension-service", "infiniband-partition", "nvlink-logical-partition"} {
		delete(c.items, rt)
		delete(c.fetched, rt)
	}
}

func (c *Cache) LookupByName(resourceType, name string) *NamedItem {
	items := c.Get(resourceType)
	if items == nil {
		return nil
	}
	for _, item := range items {
		if equalFold(item.Name, name) {
			return &item
		}
	}
	return nil
}

func (c *Cache) LookupByID(resourceType, id string) *NamedItem {
	items := c.Get(resourceType)
	if items == nil {
		return nil
	}
	for _, item := range items {
		if item.ID == id {
			return &item
		}
	}
	return nil
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}
