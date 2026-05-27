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

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

// Command represents a registered interactive command.
type Command struct {
	Name        string
	Description string
	Run         func(s *Session, args []string) error
}

// AllCommands returns all available commands.
func AllCommands() []Command {
	return []Command{
		{Name: "site list", Description: "List all sites", Run: cmdSiteList},
		{Name: "site get", Description: "Get site details", Run: cmdSiteGet},
		{Name: "site create", Description: "Create a site", Run: cmdSiteCreate},
		{Name: "site update", Description: "Update a site", Run: cmdSiteUpdate},
		{Name: "site delete", Description: "Delete a site", Run: cmdSiteDelete},

		{Name: "vpc list", Description: "List all VPCs", Run: cmdVPCList},
		{Name: "vpc get", Description: "Get VPC details", Run: cmdVPCGet},
		{Name: "vpc create", Description: "Create a VPC", Run: cmdVPCCreate},
		{Name: "vpc update", Description: "Update a VPC", Run: cmdVPCUpdate},
		{Name: "vpc virtualization update", Description: "Update VPC virtualization", Run: cmdVPCVirtualizationUpdate},
		{Name: "vpc delete", Description: "Delete a VPC", Run: cmdVPCDelete},

		{Name: "subnet list", Description: "List all subnets", Run: cmdSubnetList},
		{Name: "subnet get", Description: "Get subnet details", Run: cmdSubnetGet},
		{Name: "subnet create", Description: "Create a subnet", Run: cmdSubnetCreate},
		{Name: "subnet update", Description: "Update a subnet", Run: cmdSubnetUpdate},
		{Name: "subnet delete", Description: "Delete a subnet", Run: cmdSubnetDelete},

		{Name: "instance-type list", Description: "List instance types", Run: cmdInstanceTypeList},
		{Name: "instance-type get", Description: "Get instance type details", Run: cmdInstanceTypeGet},

		{Name: "instance list", Description: "List all instances", Run: cmdInstanceList},
		{Name: "instance get", Description: "Get instance details", Run: cmdInstanceGet},
		{Name: "instance create", Description: "Create an instance on a machine", Run: cmdInstanceCreate},
		{Name: "instance update", Description: "Update an instance (rename, change OS, rotate ssh key groups, trigger reboot)", Run: cmdInstanceUpdate},
		{Name: "instance reboot", Description: "Reboot an instance, optionally with custom iPXE / pending updates", Run: cmdInstanceReboot},
		{Name: "instance delete", Description: "Delete an instance", Run: cmdInstanceDelete},

		{Name: "machine list", Description: "List machines", Run: cmdMachineList},
		{Name: "machine get", Description: "Get machine details", Run: cmdMachineGet},

		{Name: "operating-system list", Description: "List operating systems", Run: cmdOSList},
		{Name: "operating-system get", Description: "Get operating system details", Run: cmdOSGet},
		{Name: "operating-system create", Description: "Create an operating system", Run: cmdOSCreate},
		{Name: "operating-system update", Description: "Update an operating system", Run: cmdOSUpdate},
		{Name: "operating-system delete", Description: "Delete an operating system", Run: cmdOSDelete},

		{Name: "ssh-key-group list", Description: "List SSH key groups", Run: cmdSSHKeyGroupList},
		{Name: "ssh-key-group get", Description: "Get SSH key group details", Run: cmdSSHKeyGroupGet},
		{Name: "ssh-key-group create", Description: "Create an SSH key group", Run: cmdSSHKeyGroupCreate},
		{Name: "ssh-key-group update", Description: "Update an SSH key group", Run: cmdSSHKeyGroupUpdate},
		{Name: "ssh-key-group delete", Description: "Delete an SSH key group", Run: cmdSSHKeyGroupDelete},

		{Name: "ssh-key list", Description: "List SSH keys", Run: cmdSSHKeyList},
		{Name: "ssh-key get", Description: "Get SSH key details", Run: cmdSSHKeyGet},
		{Name: "ssh-key create", Description: "Create an SSH key", Run: cmdSSHKeyCreate},
		{Name: "ssh-key update", Description: "Update an SSH key", Run: cmdSSHKeyUpdate},
		{Name: "ssh-key delete", Description: "Delete an SSH key", Run: cmdSSHKeyDelete},

		{Name: "allocation list", Description: "List allocations", Run: cmdAllocationList},
		{Name: "allocation get", Description: "Get allocation details", Run: cmdAllocationGet},
		{Name: "allocation create", Description: "Create an allocation", Run: cmdAllocationCreate},
		{Name: "allocation update", Description: "Update an allocation", Run: cmdAllocationUpdate},
		{Name: "allocation delete", Description: "Delete an allocation", Run: cmdAllocationDelete},

		{Name: "ip-block list", Description: "List IP blocks", Run: cmdIPBlockList},
		{Name: "ip-block get", Description: "Get IP block details", Run: cmdIPBlockGet},
		{Name: "ip-block create", Description: "Create an IP block", Run: cmdIPBlockCreate},
		{Name: "ip-block update", Description: "Update an IP block", Run: cmdIPBlockUpdate},
		{Name: "ip-block delete", Description: "Delete an IP block", Run: cmdIPBlockDelete},

		{Name: "network-security-group list", Description: "List network security groups", Run: cmdNSGList},
		{Name: "network-security-group get", Description: "Get network security group details", Run: cmdNSGGet},
		{Name: "network-security-group create", Description: "Create a network security group", Run: cmdNSGCreate},
		{Name: "network-security-group update", Description: "Update a network security group", Run: cmdNSGUpdate},
		{Name: "network-security-group delete", Description: "Delete a network security group", Run: cmdNSGDelete},

		{Name: "sku list", Description: "List SKUs", Run: cmdSKUList},
		{Name: "sku get", Description: "Get SKU details", Run: cmdSKUGet},

		{Name: "rack list", Description: "List racks", Run: cmdRackList},
		{Name: "rack get", Description: "Get rack details", Run: cmdRackGet},

		{Name: "vpc-prefix list", Description: "List VPC prefixes", Run: cmdVPCPrefixList},
		{Name: "vpc-prefix get", Description: "Get VPC prefix details", Run: cmdVPCPrefixGet},
		{Name: "vpc-prefix create", Description: "Create a VPC prefix", Run: cmdVPCPrefixCreate},
		{Name: "vpc-prefix update", Description: "Update a VPC prefix", Run: cmdVPCPrefixUpdate},
		{Name: "vpc-prefix delete", Description: "Delete a VPC prefix", Run: cmdVPCPrefixDelete},

		{Name: "tenant-account list", Description: "List tenant accounts", Run: cmdTenantAccountList},
		{Name: "tenant-account get", Description: "Get tenant account details", Run: cmdTenantAccountGet},
		{Name: "tenant-account create", Description: "Create a tenant account", Run: cmdTenantAccountCreate},
		{Name: "tenant-account update", Description: "Update a tenant account", Run: cmdTenantAccountUpdate},
		{Name: "tenant-account delete", Description: "Delete a tenant account", Run: cmdTenantAccountDelete},

		{Name: "expected-machine list", Description: "List expected machines", Run: cmdExpectedMachineList},
		{Name: "expected-machine get", Description: "Get expected machine details", Run: cmdExpectedMachineGet},

		{Name: "infiniband-partition list", Description: "List InfiniBand partitions", Run: cmdInfiniBandPartitionList},
		{Name: "infiniband-partition get", Description: "Get InfiniBand partition details", Run: cmdInfiniBandPartitionGet},

		{Name: "nvlink-logical-partition list", Description: "List NVLink logical partitions", Run: cmdNVLinkLogicalPartitionList},
		{Name: "nvlink-logical-partition get", Description: "Get NVLink logical partition details", Run: cmdNVLinkLogicalPartitionGet},

		{Name: "dpu-extension-service list", Description: "List DPU extension services", Run: cmdDPUExtensionServiceList},
		{Name: "dpu-extension-service get", Description: "Get DPU extension service details", Run: cmdDPUExtensionServiceGet},

		{Name: "audit list", Description: "List audit log entries", Run: cmdAuditList},
		{Name: "audit get", Description: "Get audit log entry details", Run: cmdAuditGet},

		{Name: "metadata get", Description: "Get API metadata", Run: cmdMetadataGet},
		{Name: "user current", Description: "Get current user", Run: cmdUserCurrent},
		{Name: "tenant current", Description: "Get current tenant", Run: cmdTenantCurrent},
		{Name: "tenant stats", Description: "Get tenant stats", Run: cmdTenantStats},
		{Name: "infrastructure-provider current", Description: "Get current infrastructure provider", Run: cmdInfraProviderCurrent},
		{Name: "infrastructure-provider stats", Description: "Get infrastructure provider stats", Run: cmdInfraProviderStats},

		{Name: "login", Description: "Login / refresh auth token", Run: cmdLogin},
		{Name: "help", Description: "Show available commands", Run: cmdHelp},
	}
}

// LogCmd prints the equivalent cli one-liner for reference.
func LogCmd(s *Session, parts ...string) {
	cmdParts := []string{"cli"}
	if s != nil && strings.TrimSpace(s.ConfigPath) != "" {
		cmdParts = append(cmdParts, "--config", s.ConfigPath)
	}
	cmdParts = append(cmdParts, appendScopeFlags(s, parts)...)
	fmt.Printf("%s %s\n", Dim("INFO:"), strings.Join(cmdParts, " "))
}

func appendScopeFlags(s *Session, parts []string) []string {
	out := append([]string(nil), parts...)
	if s == nil || len(parts) < 2 {
		return out
	}
	resource := strings.TrimSpace(parts[0])
	action := strings.TrimSpace(parts[1])
	if action != "list" {
		return out
	}
	scopeSiteID := strings.TrimSpace(s.Scope.SiteID)
	scopeVpcID := strings.TrimSpace(s.Scope.VpcID)
	switch resource {
	case "vpc", "allocation", "ip-block", "operating-system", "ssh-key-group",
		"network-security-group", "sku", "rack", "expected-machine", "instance-type",
		"dpu-extension-service", "infiniband-partition", "nvlink-logical-partition":
		if scopeSiteID != "" {
			out = append(out, "--site-id", scopeSiteID)
		}
	case "subnet", "vpc-prefix", "instance", "machine":
		if scopeSiteID != "" {
			out = append(out, "--site-id", scopeSiteID)
		}
		if scopeVpcID != "" {
			out = append(out, "--vpc-id", scopeVpcID)
		}
	}
	return out
}

func fetchMachinesWithSiteFallback(s *Session, missingSitePrompt string) ([]NamedItem, error) {
	savedSiteID := s.Scope.SiteID
	savedVpcID, savedVpcName := s.Scope.VpcID, s.Scope.VpcName

	s.Scope.VpcID = ""
	s.Scope.VpcName = ""
	s.Cache.InvalidateFiltered()
	defer func() {
		if s.Scope.SiteID == savedSiteID {
			s.Scope.VpcID = savedVpcID
			s.Scope.VpcName = savedVpcName
		}
		s.Cache.InvalidateFiltered()
	}()

	items, err := s.Resolver.Fetch(context.Background(), "machine")
	if err == nil {
		return items, nil
	}
	if s.Scope.SiteID == "" && strings.Contains(err.Error(), "400") {
		fmt.Printf("%s %s\n", Dim("Note:"), missingSitePrompt)
		site, resolveErr := s.Resolver.Resolve(context.Background(), "site", "Site")
		if resolveErr != nil {
			return nil, resolveErr
		}
		s.Scope.SiteID = site.ID
		s.Scope.SiteName = site.Name
		s.Cache.InvalidateFiltered()
		return s.Resolver.Fetch(context.Background(), "machine")
	}
	return nil, err
}

func setSiteScopeFromID(s *Session, siteID string) {
	siteID = strings.TrimSpace(siteID)
	if siteID == "" || s.Scope.SiteID == siteID {
		return
	}
	s.Scope.SiteID = siteID
	s.Scope.SiteName = s.Resolver.ResolveID("site", siteID)
	s.Scope.VpcID = ""
	s.Scope.VpcName = ""
	s.Cache.InvalidateFiltered()
}

func readyMachineItemsForSite(machines []NamedItem, siteID string) []SelectItem {
	siteID = strings.TrimSpace(siteID)
	readyItems := make([]SelectItem, 0, len(machines))
	for _, m := range machines {
		if siteID != "" && strings.TrimSpace(m.Extra["siteId"]) != siteID {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(m.Status), "Ready") {
			readyItems = append(readyItems, SelectItem{Label: machineSelectLabel(m), ID: m.ID})
		}
	}
	return readyItems
}

// machineSelectLabel formats a machine for the interactive select list. It
// always includes the resolved display name (which may be a serial number when
// no friendly labels are set) plus the full machine ID, so reviewers and
// scripts that already track machines by ID can find them without having to
// memorize serial-to-id mappings.
func machineSelectLabel(m NamedItem) string {
	name := strings.TrimSpace(m.Name)
	id := strings.TrimSpace(m.ID)
	if name == "" {
		return id
	}
	if id == "" {
		return name
	}
	return name + "  " + Dim(id)
}

// -- List commands --

func cmdSiteList(s *Session, _ []string) error {
	LogCmd(s, "site", "list")
	items, err := s.Resolver.Fetch(context.Background(), "site")
	if err != nil {
		return err
	}
	return printResourceTable(os.Stdout, "NAME", "STATUS", "ID", items)
}

func cmdSiteCreate(s *Session, _ []string) error {
	name, err := PromptText("Site name", true)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	serialConsoleHostname, err := PromptText("Serial console hostname (optional)", false)
	if err != nil {
		return err
	}
	city, err := PromptText("Location city (optional)", false)
	if err != nil {
		return err
	}
	state, err := PromptText("Location state (optional)", false)
	if err != nil {
		return err
	}
	country, err := PromptText("Location country (optional)", false)
	if err != nil {
		return err
	}
	contactEmail, err := PromptText("Contact email (optional)", false)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"name": name,
	}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	if strings.TrimSpace(serialConsoleHostname) != "" {
		body["serialConsoleHostname"] = strings.TrimSpace(serialConsoleHostname)
	}
	location := map[string]interface{}{}
	if strings.TrimSpace(city) != "" {
		location["city"] = strings.TrimSpace(city)
	}
	if strings.TrimSpace(state) != "" {
		location["state"] = strings.TrimSpace(state)
	}
	if strings.TrimSpace(country) != "" {
		location["country"] = strings.TrimSpace(country)
	}
	if len(location) > 0 {
		body["location"] = location
	}
	if strings.TrimSpace(contactEmail) != "" {
		body["contact"] = map[string]interface{}{"email": strings.TrimSpace(contactEmail)}
	}

	LogCmd(s, "site", "create", "--name", name)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("POST", apiPath(s, "site"), nil, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("creating site: %w", err)
	}
	s.Cache.Invalidate("site")
	s.Cache.InvalidateFiltered()
	var created map[string]interface{}
	json.Unmarshal(resp, &created)
	fmt.Printf("%s Site created: %s (%s)\n", Green("OK"), str(created, "name"), str(created, "id"))
	return nil
}

func cmdSiteUpdate(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "site", "Site to update", args)
	if err != nil {
		return err
	}
	name, err := PromptText("Site name (optional)", false)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	serialConsoleHostname, err := PromptText("Serial console hostname (optional)", false)
	if err != nil {
		return err
	}
	renewTokenText, err := PromptText("Renew registration token? (true/false, blank to keep)", false)
	if err != nil {
		return err
	}
	serialEnabledText, err := PromptText("Serial console enabled? (true/false, blank to keep)", false)
	if err != nil {
		return err
	}
	sshKeysEnabledText, err := PromptText("Serial console SSH keys enabled? (true/false, blank to keep)", false)
	if err != nil {
		return err
	}
	idleTimeoutText, err := PromptText("Serial console idle timeout seconds (optional)", false)
	if err != nil {
		return err
	}
	maxSessionText, err := PromptText("Serial console max session length seconds (optional)", false)
	if err != nil {
		return err
	}
	city, err := PromptText("Location city (optional)", false)
	if err != nil {
		return err
	}
	state, err := PromptText("Location state (optional)", false)
	if err != nil {
		return err
	}
	country, err := PromptText("Location country (optional)", false)
	if err != nil {
		return err
	}
	contactEmail, err := PromptText("Contact email (optional)", false)
	if err != nil {
		return err
	}

	body := map[string]interface{}{}
	if strings.TrimSpace(name) != "" {
		body["name"] = strings.TrimSpace(name)
	}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	if strings.TrimSpace(serialConsoleHostname) != "" {
		body["serialConsoleHostname"] = strings.TrimSpace(serialConsoleHostname)
	}
	if v, ok := parseOptionalBool(renewTokenText); ok {
		body["renewRegistrationToken"] = v
	}
	if v, ok := parseOptionalBool(serialEnabledText); ok {
		body["isSerialConsoleEnabled"] = v
	}
	if v, ok := parseOptionalBool(sshKeysEnabledText); ok {
		body["isSerialConsoleSSHKeysEnabled"] = v
	}
	if v, ok := parseOptionalInt(idleTimeoutText); ok {
		body["serialConsoleIdleTimeout"] = v
	}
	if v, ok := parseOptionalInt(maxSessionText); ok {
		body["serialConsoleMaxSessionLength"] = v
	}
	location := map[string]interface{}{}
	if strings.TrimSpace(city) != "" {
		location["city"] = strings.TrimSpace(city)
	}
	if strings.TrimSpace(state) != "" {
		location["state"] = strings.TrimSpace(state)
	}
	if strings.TrimSpace(country) != "" {
		location["country"] = strings.TrimSpace(country)
	}
	if len(location) > 0 {
		body["location"] = location
	}
	if strings.TrimSpace(contactEmail) != "" {
		body["contact"] = map[string]interface{}{"email": strings.TrimSpace(contactEmail)}
	}
	if len(body) == 0 {
		return fmt.Errorf("no updates provided")
	}

	LogCmd(s, "site", "update", item.ID)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("PATCH", apiPath(s, "site/{id}"), map[string]string{"id": item.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("updating site: %w", err)
	}
	s.Cache.Invalidate("site")
	s.Cache.InvalidateFiltered()
	var updated map[string]interface{}
	json.Unmarshal(resp, &updated)
	fmt.Printf("%s Site updated: %s (%s)\n", Green("OK"), str(updated, "name"), str(updated, "id"))
	return nil
}

func cmdSiteDelete(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "site", "Site to delete", args)
	if err != nil {
		return err
	}
	ok, err := PromptConfirm(fmt.Sprintf("Delete site %s (%s)?", item.Name, item.ID))
	if err != nil || !ok {
		return err
	}
	purgeMachines, err := PromptConfirm("Purge machine data as part of delete?")
	if err != nil {
		return err
	}
	query := map[string]string{}
	if purgeMachines {
		query["purgeMachines"] = "true"
	}
	LogCmd(s, "site", "delete", item.ID)
	_, _, err = s.Client.Do("DELETE", apiPath(s, "site/{id}"), map[string]string{"id": item.ID}, query, nil)
	if err != nil {
		return fmt.Errorf("deleting site: %w", err)
	}
	s.Cache.Invalidate("site")
	s.Cache.InvalidateFiltered()
	fmt.Printf("%s Site delete requested: %s\n", Green("OK"), item.Name)
	return nil
}

func cmdVPCList(s *Session, args []string) error {
	LogCmd(s, "vpc", "list")
	items, err := s.Resolver.Fetch(context.Background(), "vpc")
	if err != nil {
		return err
	}
	_, cmdLabels, sortKey, err := parseLabelArgs(args)
	if err != nil {
		return err
	}
	merged, mergeErr := mergeLabels(s.Scope.LabelFilters, cmdLabels)
	if mergeErr != nil {
		return mergeErr
	}
	items = filterByLabels(items, merged)
	if sortKey != "" {
		items = sortByLabelKey(items, sortKey)
	}
	siteNameByID := map[string]string{}
	if sites, err := s.Resolver.Fetch(context.Background(), "site"); err == nil {
		for _, site := range sites {
			siteNameByID[site.ID] = site.Name
		}
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	defer printLabelHint(os.Stderr, items, merged)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tSITE\tLABELS\tID")
	for _, item := range items {
		siteID := item.Extra["siteId"]
		siteName := strings.TrimSpace(siteNameByID[siteID])
		if siteName == "" {
			siteName = siteID
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", item.Name, item.Status, siteName, formatLabels(item.Labels, 60), item.ID)
	}
	return tw.Flush()
}

func cmdVPCCreate(s *Session, _ []string) error {
	site, err := s.Resolver.Resolve(context.Background(), "site", "Site")
	if err != nil {
		return err
	}
	name, err := PromptText("VPC name", true)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	body := map[string]interface{}{"name": name, "siteId": site.ID}
	if strings.TrimSpace(desc) != "" {
		body["description"] = desc
	}
	LogCmd(s, "vpc", "create", "--name", name, "--site-id", site.ID)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("POST", apiPath(s, "vpc"), nil, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("creating VPC: %w", err)
	}
	s.Cache.Invalidate("vpc")
	var created map[string]interface{}
	json.Unmarshal(resp, &created)
	fmt.Printf("%s VPC created: %s (%s)\n", Green("OK"), str(created, "name"), str(created, "id"))
	return nil
}

func cmdVPCUpdate(s *Session, args []string) error {
	vpc, err := s.Resolver.ResolveWithArgs(context.Background(), "vpc", "VPC to update", args)
	if err != nil {
		return err
	}
	name, err := PromptText("VPC name (optional)", false)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	body := map[string]interface{}{}
	if strings.TrimSpace(name) != "" {
		body["name"] = strings.TrimSpace(name)
	}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	if len(body) == 0 {
		return fmt.Errorf("no updates provided")
	}
	LogCmd(s, "vpc", "update", vpc.ID)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("PATCH", apiPath(s, "vpc/{id}"), map[string]string{"id": vpc.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("updating VPC: %w", err)
	}
	s.Cache.Invalidate("vpc")
	s.Cache.InvalidateFiltered()
	var updated map[string]interface{}
	json.Unmarshal(resp, &updated)
	fmt.Printf("%s VPC updated: %s (%s)\n", Green("OK"), str(updated, "name"), str(updated, "id"))
	return nil
}

func cmdVPCVirtualizationUpdate(s *Session, args []string) error {
	vpc, err := s.Resolver.ResolveWithArgs(context.Background(), "vpc", "VPC to update virtualization", args)
	if err != nil {
		return err
	}
	virtType, err := PromptText("Network virtualization type (ETHERNET_VIRTUALIZER or FNN)", true)
	if err != nil {
		return err
	}
	virtType = strings.ToUpper(strings.TrimSpace(virtType))
	if virtType != "ETHERNET_VIRTUALIZER" && virtType != "FNN" {
		return fmt.Errorf("network virtualization type must be ETHERNET_VIRTUALIZER or FNN")
	}
	body := map[string]interface{}{
		"networkVirtualizationType": virtType,
	}
	if virtType == "FNN" {
		useNVLink, err := PromptConfirm("Select NVLink logical partition?")
		if err != nil {
			return err
		}
		if useNVLink {
			item, err := s.Resolver.Resolve(context.Background(), "nvlink-logical-partition", "NVLink Logical Partition")
			if err != nil {
				return err
			}
			body["nvLinkLogicalPartitionId"] = item.ID
		}
	}
	LogCmd(s, "vpc", "virtualization", "update", vpc.ID, "--network-virtualization-type", virtType)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("PATCH", apiPath(s, "vpc/{id}/virtualization"), map[string]string{"id": vpc.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("updating VPC virtualization: %w", err)
	}
	s.Cache.Invalidate("vpc")
	s.Cache.InvalidateFiltered()
	var updated map[string]interface{}
	json.Unmarshal(resp, &updated)
	fmt.Printf("%s VPC virtualization update submitted: %s (%s)\n", Green("OK"), str(updated, "name"), str(updated, "id"))
	return nil
}

func cmdVPCDelete(s *Session, args []string) error {
	vpc, err := s.Resolver.ResolveWithArgs(context.Background(), "vpc", "VPC to delete", args)
	if err != nil {
		return err
	}
	ok, err := PromptConfirm(fmt.Sprintf("Delete VPC %s (%s)?", vpc.Name, vpc.ID))
	if err != nil || !ok {
		return err
	}
	LogCmd(s, "vpc", "delete", vpc.ID)
	_, _, err = s.Client.Do("DELETE", apiPath(s, "vpc/{id}"), map[string]string{"id": vpc.ID}, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting VPC: %w", err)
	}
	s.Cache.Invalidate("vpc")
	fmt.Printf("%s VPC deleted: %s\n", Green("OK"), vpc.Name)
	return nil
}

func cmdSubnetList(s *Session, _ []string) error {
	LogCmd(s, "subnet", "list")
	items, err := s.Resolver.Fetch(context.Background(), "subnet")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tVPC\tID")
	for _, item := range items {
		vpcName := s.Resolver.ResolveID("vpc", item.Extra["vpcId"])
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.Name, item.Status, vpcName, item.ID)
	}
	return tw.Flush()
}

func cmdSubnetCreate(s *Session, _ []string) error {
	vpc, err := s.Resolver.Resolve(context.Background(), "vpc", "VPC")
	if err != nil {
		return err
	}
	vpcSiteID := strings.TrimSpace(vpc.Extra["siteId"])
	setSiteScopeFromID(s, vpcSiteID)

	name, err := PromptText("Subnet name", true)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	prefixLenText, err := PromptText("Prefix length (1-32)", true)
	if err != nil {
		return err
	}
	var prefixLen int
	fmt.Sscanf(prefixLenText, "%d", &prefixLen)
	if prefixLen < 1 || prefixLen > 32 {
		return fmt.Errorf("prefix length must be between 1 and 32")
	}

	ipBlocks, err := s.Resolver.Fetch(context.Background(), "ip-block")
	if err != nil {
		return fmt.Errorf("fetching IP blocks: %w", err)
	}
	blockItems := make([]SelectItem, 0, len(ipBlocks))
	for _, block := range ipBlocks {
		if vpcSiteID != "" && strings.TrimSpace(block.Extra["siteId"]) != vpcSiteID {
			continue
		}
		blockItems = append(blockItems, SelectItem{Label: block.Name, ID: block.ID})
	}
	if len(blockItems) == 0 {
		if vpcSiteID != "" {
			return fmt.Errorf("no IP blocks available for selected VPC site")
		}
		return fmt.Errorf("no IP blocks available")
	}
	block, err := Select("IPv4 Block", blockItems)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"name":         name,
		"vpcId":        vpc.ID,
		"ipv4BlockId":  block.ID,
		"prefixLength": prefixLen,
	}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	LogCmd(s, "subnet", "create", "--name", name, "--vpc-id", vpc.ID, "--ipv4-block-id", block.ID, "--prefix-length", prefixLenText)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("POST", apiPath(s, "subnet"), nil, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("creating subnet: %w", err)
	}
	s.Cache.Invalidate("subnet")
	s.Cache.InvalidateFiltered()
	var created map[string]interface{}
	json.Unmarshal(resp, &created)
	fmt.Printf("%s Subnet created: %s (%s)\n", Green("OK"), str(created, "name"), str(created, "id"))
	return nil
}

func cmdSubnetUpdate(s *Session, args []string) error {
	subnet, err := s.Resolver.ResolveWithArgs(context.Background(), "subnet", "Subnet to update", args)
	if err != nil {
		return err
	}
	name, err := PromptText("Subnet name (optional)", false)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	body := map[string]interface{}{}
	if strings.TrimSpace(name) != "" {
		body["name"] = strings.TrimSpace(name)
	}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	if len(body) == 0 {
		return fmt.Errorf("no updates provided")
	}
	LogCmd(s, "subnet", "update", subnet.ID)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("PATCH", apiPath(s, "subnet/{id}"), map[string]string{"id": subnet.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("updating subnet: %w", err)
	}
	s.Cache.Invalidate("subnet")
	s.Cache.InvalidateFiltered()
	var updated map[string]interface{}
	json.Unmarshal(resp, &updated)
	fmt.Printf("%s Subnet updated: %s (%s)\n", Green("OK"), str(updated, "name"), str(updated, "id"))
	return nil
}

func cmdSubnetDelete(s *Session, args []string) error {
	subnet, err := s.Resolver.ResolveWithArgs(context.Background(), "subnet", "Subnet to delete", args)
	if err != nil {
		return err
	}
	ok, err := PromptConfirm(fmt.Sprintf("Delete subnet %s (%s)?", subnet.Name, subnet.ID))
	if err != nil || !ok {
		return err
	}
	LogCmd(s, "subnet", "delete", subnet.ID)
	_, _, err = s.Client.Do("DELETE", apiPath(s, "subnet/{id}"), map[string]string{"id": subnet.ID}, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting subnet: %w", err)
	}
	s.Cache.Invalidate("subnet")
	s.Cache.InvalidateFiltered()
	fmt.Printf("%s Subnet deleted: %s\n", Green("OK"), subnet.Name)
	return nil
}

func cmdInstanceTypeList(s *Session, args []string) error {
	LogCmd(s, "instance-type", "list")
	items, err := s.Resolver.Fetch(context.Background(), "instance-type")
	if err != nil {
		return err
	}
	_, cmdLabels, sortKey, err := parseLabelArgs(args)
	if err != nil {
		return err
	}
	merged, mergeErr := mergeLabels(s.Scope.LabelFilters, cmdLabels)
	if mergeErr != nil {
		return mergeErr
	}
	items = filterByLabels(items, merged)
	if sortKey != "" {
		items = sortByLabelKey(items, sortKey)
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	defer printLabelHint(os.Stderr, items, merged)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tLABELS\tID")
	for _, item := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.Name, item.Status, formatLabels(item.Labels, 60), item.ID)
	}
	return tw.Flush()
}

func cmdInstanceTypeGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "instance-type", "Instance Type", args)
	if err != nil {
		return err
	}
	LogCmd(s, "instance-type", "get", item.ID)
	return getAndPrint(s, apiPath(s, "instance/type/{id}"), item.ID)
}

func cmdInstanceList(s *Session, args []string) error {
	LogCmd(s, "instance", "list")
	_, _ = s.Resolver.Fetch(context.Background(), "vpc")
	_, _ = s.Resolver.Fetch(context.Background(), "site")
	items, err := s.Resolver.Fetch(context.Background(), "instance")
	if err != nil {
		return err
	}
	_, cmdLabels, sortKey, err := parseLabelArgs(args)
	if err != nil {
		return err
	}
	merged, mergeErr := mergeLabels(s.Scope.LabelFilters, cmdLabels)
	if mergeErr != nil {
		return mergeErr
	}
	items = filterByLabels(items, merged)
	if sortKey != "" {
		items = sortByLabelKey(items, sortKey)
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	defer printLabelHint(os.Stderr, items, merged)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tVPC\tSITE\tLABELS\tID")
	for _, item := range items {
		vpcName := s.Resolver.ResolveID("vpc", item.Extra["vpcId"])
		siteName := s.Resolver.ResolveID("site", item.Extra["siteId"])
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", item.Name, item.Status, vpcName, siteName, formatLabels(item.Labels, 60), item.ID)
	}
	return tw.Flush()
}

func cmdMachineList(s *Session, args []string) error {
	LogCmd(s, "machine", "list")
	items, err := fetchMachinesWithSiteFallback(s, "Machine listing requires a site filter. Select a site.")
	if err != nil {
		return err
	}

	_, _ = s.Resolver.Fetch(context.Background(), "vpc")
	vpcNamesByMachineID := s.buildMachineVPCNames(context.Background())

	if s.Scope.VpcID != "" {
		filtered := make([]NamedItem, 0, len(items))
		for _, item := range items {
			if _, ok := vpcNamesByMachineID[item.ID]; ok {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	_, cmdLabels, sortKey, err := parseLabelArgs(args)
	if err != nil {
		return err
	}
	merged, mergeErr := mergeLabels(s.Scope.LabelFilters, cmdLabels)
	if mergeErr != nil {
		return mergeErr
	}
	items = filterByLabels(items, merged)
	if sortKey != "" {
		items = sortByLabelKey(items, sortKey)
	}

	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	defer printLabelHint(os.Stderr, items, merged)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tSITE\tVPC\tLABELS\tID")
	for _, item := range items {
		siteName := s.Resolver.ResolveID("site", item.Extra["siteId"])
		vpcNames := strings.TrimSpace(vpcNamesByMachineID[item.ID])
		if vpcNames == "" {
			vpcNames = "-"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", item.Name, item.Status, siteName, vpcNames, formatLabels(item.Labels, 60), item.ID)
	}
	return tw.Flush()
}

func cmdOSList(s *Session, _ []string) error {
	LogCmd(s, "operating-system", "list")
	items, err := s.Resolver.Fetch(context.Background(), "operating-system")
	if err != nil {
		return err
	}
	return printResourceTable(os.Stdout, "NAME", "STATUS", "ID", items)
}

func cmdOSCreate(s *Session, _ []string) error {
	name, err := PromptText("Operating system name", true)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	tenantID, err := s.getTenantID(context.Background())
	if err != nil {
		return fmt.Errorf("resolving tenant id: %w", err)
	}
	ipxeScript, err := PromptText("iPXE script or URL", true)
	if err != nil {
		return err
	}
	userData, err := PromptText("User data (optional)", false)
	if err != nil {
		return err
	}
	isCloudInit, err := PromptConfirm("Cloud-init enabled?")
	if err != nil {
		return err
	}
	allowOverride, err := PromptConfirm("Allow override at instance creation?")
	if err != nil {
		return err
	}
	phoneHomeEnabled, err := PromptConfirm("Enable phone home?")
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"name":             name,
		"tenantId":         tenantID,
		"ipxeScript":       ipxeScript,
		"isCloudInit":      isCloudInit,
		"allowOverride":    allowOverride,
		"phoneHomeEnabled": phoneHomeEnabled,
	}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	if strings.TrimSpace(userData) != "" {
		body["userData"] = strings.TrimSpace(userData)
	}
	LogCmd(s, "operating-system", "create", "--name", name)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("POST", apiPath(s, "operating-system"), nil, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("creating operating system: %w", err)
	}
	s.Cache.Invalidate("operating-system")
	s.Cache.InvalidateFiltered()
	var created map[string]interface{}
	json.Unmarshal(resp, &created)
	fmt.Printf("%s Operating system created: %s (%s)\n", Green("OK"), str(created, "name"), str(created, "id"))
	return nil
}

func cmdOSUpdate(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "operating-system", "Operating System to update", args)
	if err != nil {
		return err
	}
	name, err := PromptText("Operating system name (optional)", false)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	ipxeScript, err := PromptText("iPXE script or URL (optional)", false)
	if err != nil {
		return err
	}
	userData, err := PromptText("User data (optional)", false)
	if err != nil {
		return err
	}
	allowOverrideText, err := PromptText("Allow override? (true/false, blank to keep)", false)
	if err != nil {
		return err
	}
	phoneHomeText, err := PromptText("Phone home enabled? (true/false, blank to keep)", false)
	if err != nil {
		return err
	}
	activeText, err := PromptText("Set active? (true/false, blank to keep)", false)
	if err != nil {
		return err
	}

	body := map[string]interface{}{}
	if strings.TrimSpace(name) != "" {
		body["name"] = strings.TrimSpace(name)
	}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	if strings.TrimSpace(ipxeScript) != "" {
		body["ipxeScript"] = strings.TrimSpace(ipxeScript)
	}
	if strings.TrimSpace(userData) != "" {
		body["userData"] = strings.TrimSpace(userData)
	}
	if v, ok := parseOptionalBool(allowOverrideText); ok {
		body["allowOverride"] = v
	}
	if v, ok := parseOptionalBool(phoneHomeText); ok {
		body["phoneHomeEnabled"] = v
	}
	if v, ok := parseOptionalBool(activeText); ok {
		body["isActive"] = v
		if !v {
			note, err := PromptText("Deactivation note (optional)", false)
			if err != nil {
				return err
			}
			if strings.TrimSpace(note) != "" {
				body["deactivationNote"] = strings.TrimSpace(note)
			}
		}
	}
	if len(body) == 0 {
		return fmt.Errorf("no updates provided")
	}
	LogCmd(s, "operating-system", "update", item.ID)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("PATCH", apiPath(s, "operating-system/{id}"), map[string]string{"id": item.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("updating operating system: %w", err)
	}
	s.Cache.Invalidate("operating-system")
	s.Cache.InvalidateFiltered()
	var updated map[string]interface{}
	json.Unmarshal(resp, &updated)
	fmt.Printf("%s Operating system updated: %s (%s)\n", Green("OK"), str(updated, "name"), str(updated, "id"))
	return nil
}

func cmdOSDelete(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "operating-system", "Operating System to delete", args)
	if err != nil {
		return err
	}
	ok, err := PromptConfirm(fmt.Sprintf("Delete operating system %s (%s)?", item.Name, item.ID))
	if err != nil || !ok {
		return err
	}
	LogCmd(s, "operating-system", "delete", item.ID)
	_, _, err = s.Client.Do("DELETE", apiPath(s, "operating-system/{id}"), map[string]string{"id": item.ID}, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting operating system: %w", err)
	}
	s.Cache.Invalidate("operating-system")
	s.Cache.InvalidateFiltered()
	fmt.Printf("%s Operating system deleted: %s\n", Green("OK"), item.Name)
	return nil
}

func cmdSSHKeyGroupList(s *Session, _ []string) error {
	LogCmd(s, "ssh-key-group", "list")
	items, err := s.Resolver.Fetch(context.Background(), "ssh-key-group")
	if err != nil {
		return err
	}
	return printResourceTable(os.Stdout, "NAME", "STATUS", "ID", items)
}

func cmdSSHKeyGroupCreate(s *Session, _ []string) error {
	ctx := context.Background()
	name, err := PromptText("SSH key group name", true)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}

	siteIDs, err := promptOptionalResourceIDs(s, ctx, "site", "site")
	if err != nil {
		return err
	}
	sshKeyIDs, err := promptOptionalResourceIDs(s, ctx, "ssh-key", "SSH key")
	if err != nil {
		return err
	}

	body := map[string]interface{}{"name": name}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	if len(siteIDs) > 0 {
		body["siteIds"] = siteIDs
	}
	if len(sshKeyIDs) > 0 {
		body["sshKeyIds"] = sshKeyIDs
	}

	LogCmd(s, "ssh-key-group", "create", "--name", name)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("POST", apiPath(s, "sshkeygroup"), nil, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("creating SSH key group: %w", err)
	}
	s.Cache.Invalidate("ssh-key-group")
	s.Cache.InvalidateFiltered()
	var created map[string]interface{}
	json.Unmarshal(resp, &created)
	fmt.Printf("%s SSH key group created: %s (%s)\n", Green("OK"), str(created, "name"), str(created, "id"))
	return nil
}

func cmdSSHKeyGroupUpdate(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "ssh-key-group", "SSH Key Group to update", args)
	if err != nil {
		return err
	}
	name, err := PromptText("SSH key group name (optional)", false)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	siteIDsText, err := PromptText("Replace site IDs (comma-separated, blank to keep)", false)
	if err != nil {
		return err
	}
	sshKeyIDsText, err := PromptText("Replace SSH key IDs (comma-separated, blank to keep)", false)
	if err != nil {
		return err
	}

	version := rawFieldString(item.Raw, "version")
	if strings.TrimSpace(version) == "" {
		version, err = PromptText("Version", true)
		if err != nil {
			return err
		}
	}
	body := map[string]interface{}{
		"version": version,
	}
	if strings.TrimSpace(name) != "" {
		body["name"] = strings.TrimSpace(name)
	}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	if strings.TrimSpace(siteIDsText) != "" {
		body["siteIds"] = splitCommaSeparated(siteIDsText)
	}
	if strings.TrimSpace(sshKeyIDsText) != "" {
		body["sshKeyIds"] = splitCommaSeparated(sshKeyIDsText)
	}

	LogCmd(s, "ssh-key-group", "update", item.ID)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("PATCH", apiPath(s, "sshkeygroup/{id}"), map[string]string{"id": item.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("updating SSH key group: %w", err)
	}
	s.Cache.Invalidate("ssh-key-group")
	s.Cache.InvalidateFiltered()
	var updated map[string]interface{}
	json.Unmarshal(resp, &updated)
	fmt.Printf("%s SSH key group updated: %s (%s)\n", Green("OK"), str(updated, "name"), str(updated, "id"))
	return nil
}

func cmdSSHKeyGroupDelete(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "ssh-key-group", "SSH Key Group to delete", args)
	if err != nil {
		return err
	}
	ok, err := PromptConfirm(fmt.Sprintf("Delete SSH key group %s (%s)?", item.Name, item.ID))
	if err != nil || !ok {
		return err
	}
	LogCmd(s, "ssh-key-group", "delete", item.ID)
	_, _, err = s.Client.Do("DELETE", apiPath(s, "sshkeygroup/{id}"), map[string]string{"id": item.ID}, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting SSH key group: %w", err)
	}
	s.Cache.Invalidate("ssh-key-group")
	s.Cache.InvalidateFiltered()
	fmt.Printf("%s SSH key group deleted: %s\n", Green("OK"), item.Name)
	return nil
}

func cmdSSHKeyList(s *Session, _ []string) error {
	LogCmd(s, "ssh-key", "list")
	items, err := s.Resolver.Fetch(context.Background(), "ssh-key")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tFINGERPRINT\tID")
	for _, item := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", item.Name, item.Extra["fingerprint"], item.ID)
	}
	return tw.Flush()
}

func cmdSSHKeyCreate(s *Session, _ []string) error {
	name, err := PromptText("SSH key name", true)
	if err != nil {
		return err
	}
	publicKey, err := PromptText("Public key", true)
	if err != nil {
		return err
	}
	sshKeyGroupID, err := PromptText("SSH key group ID (optional)", false)
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"name":      name,
		"publicKey": publicKey,
	}
	if strings.TrimSpace(sshKeyGroupID) != "" {
		body["sshKeyGroupId"] = strings.TrimSpace(sshKeyGroupID)
	}
	LogCmd(s, "ssh-key", "create", "--name", name)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("POST", apiPath(s, "sshkey"), nil, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("creating SSH key: %w", err)
	}
	s.Cache.Invalidate("ssh-key")
	s.Cache.Invalidate("ssh-key-group")
	s.Cache.InvalidateFiltered()
	var created map[string]interface{}
	json.Unmarshal(resp, &created)
	fmt.Printf("%s SSH key created: %s (%s)\n", Green("OK"), str(created, "name"), str(created, "id"))
	return nil
}

func cmdSSHKeyUpdate(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "ssh-key", "SSH Key to update", args)
	if err != nil {
		return err
	}
	name, err := PromptText("SSH key name", true)
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"name": strings.TrimSpace(name),
	}
	LogCmd(s, "ssh-key", "update", item.ID, "--name", strings.TrimSpace(name))
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("PATCH", apiPath(s, "sshkey/{id}"), map[string]string{"id": item.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("updating SSH key: %w", err)
	}
	s.Cache.Invalidate("ssh-key")
	s.Cache.Invalidate("ssh-key-group")
	s.Cache.InvalidateFiltered()
	var updated map[string]interface{}
	json.Unmarshal(resp, &updated)
	fmt.Printf("%s SSH key updated: %s (%s)\n", Green("OK"), str(updated, "name"), str(updated, "id"))
	return nil
}

func cmdSSHKeyDelete(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "ssh-key", "SSH Key to delete", args)
	if err != nil {
		return err
	}
	ok, err := PromptConfirm(fmt.Sprintf("Delete SSH key %s (%s)?", item.Name, item.ID))
	if err != nil || !ok {
		return err
	}
	LogCmd(s, "ssh-key", "delete", item.ID)
	_, _, err = s.Client.Do("DELETE", apiPath(s, "sshkey/{id}"), map[string]string{"id": item.ID}, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting SSH key: %w", err)
	}
	s.Cache.Invalidate("ssh-key")
	s.Cache.Invalidate("ssh-key-group")
	s.Cache.InvalidateFiltered()
	fmt.Printf("%s SSH key deleted: %s\n", Green("OK"), item.Name)
	return nil
}

func cmdAllocationList(s *Session, _ []string) error {
	LogCmd(s, "allocation", "list")
	items, err := s.Resolver.Fetch(context.Background(), "allocation")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tSITE\tID")
	for _, item := range items {
		siteName := s.Resolver.ResolveID("site", item.Extra["siteId"])
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.Name, item.Status, siteName, item.ID)
	}
	return tw.Flush()
}

func cmdAllocationCreate(s *Session, _ []string) error {
	ctx := context.Background()
	site, err := s.Resolver.Resolve(ctx, "site", "Site")
	if err != nil {
		return err
	}
	// Scope subsequent resolver lookups (ip-block, instance-type) to the
	// allocation site so the constraint resource belongs to the same site.
	// This mutation is reverted on every exit path (success, error, or
	// interactive cancel) so the user's interactive scope is unchanged
	// after the command returns.
	savedScope := s.Scope
	setSiteScopeFromID(s, site.ID)
	defer func() {
		s.Scope = savedScope
		s.Cache.InvalidateFiltered()
	}()
	name, err := PromptText("Allocation name", true)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	tenantID, err := promptAllocationTenantID(s, ctx)
	if err != nil {
		return err
	}
	constraints, err := promptAllocationConstraints(s, ctx)
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"name":                  name,
		"siteId":                site.ID,
		"tenantId":              tenantID,
		"allocationConstraints": constraints,
	}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	LogCmd(s, "allocation", "create", "--name", name, "--site-id", site.ID, "--tenant-id", tenantID)
	fmt.Fprintf(os.Stderr, "%s allocation constraints are passed via JSON body, not flags; see --debug for the full request\n", Dim("note:"))
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("POST", apiPath(s, "allocation"), nil, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("creating allocation: %w", err)
	}
	s.Cache.Invalidate("allocation")
	s.Cache.InvalidateFiltered()
	var created map[string]interface{}
	json.Unmarshal(resp, &created)
	fmt.Printf("%s Allocation created: %s (%s)\n", Green("OK"), str(created, "name"), str(created, "id"))
	return nil
}

const tenantManualEntrySentinel = "__manual__"

// promptAllocationTenantID prompts the user to pick a tenant for the allocation.
// It lists tenants derived from tenant-accounts (the allocation API expects a
// tenant ID, not a tenant-account ID) and falls back to manual entry if no
// tenant accounts are visible or if the user explicitly opts out of the list.
func promptAllocationTenantID(s *Session, ctx context.Context) (string, error) {
	accounts, err := s.Resolver.Fetch(ctx, "tenant-account")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s could not list tenant accounts (%v); falling back to manual entry\n", Dim("note:"), err)
		accounts = nil
	}
	// Dev orgs (and any org that is both Provider Admin and Tenant Admin)
	// have an implicit self-tenant that does not appear in tenant-accounts.
	// Surface it as a first-class option so the operator does not have to
	// paste a raw UUID just to allocate to themselves.
	selfTenantID, _ := s.getTenantID(ctx)
	items := buildTenantSelectItems(accounts, selfTenantID, s.Org)
	if len(items) == 0 {
		fmt.Fprintf(os.Stderr, "%s no tenants found via tenant-account or current-tenant; falling back to manual entry\n", Dim("note:"))
		return promptTenantIDRaw()
	}
	selected, err := Select("Tenant:", items)
	if err != nil {
		return "", err
	}
	if selected.ID == tenantManualEntrySentinel {
		return promptTenantIDRaw()
	}
	return selected.ID, nil
}

func promptTenantIDRaw() (string, error) {
	raw, err := PromptText("Tenant ID", true)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(raw), nil
}

// buildTenantSelectItems builds the tenant picker for allocation create.
//
// Sources (all optional):
//   - accounts: tenant-account rows the provider has established. Each row's
//     `Extra["tenantId"]` is the selector ID; duplicate tenantIds across rows
//     are collapsed.
//   - selfTenantID / selfOrg: the caller's own tenant, surfaced first with a
//     "(self)" suffix when the caller also holds a Tenant Admin role. Common
//     for dev orgs where provider and tenant are the same entity.
//
// Returns nil when no source yields a tenantId so the caller can fall back
// to raw manual entry. When any items exist, a trailing manual-entry
// sentinel is always appended so the user can still type a raw UUID.
// Distinct tenants that share a display name are disambiguated with a
// short tenant-id suffix so the picker is never ambiguous.
func buildTenantSelectItems(accounts []NamedItem, selfTenantID, selfOrg string) []SelectItem {
	items := make([]SelectItem, 0, len(accounts)+2)
	seen := make(map[string]struct{}, len(accounts)+1)
	labelCounts := make(map[string]int, len(accounts)+1)

	selfTenantID = strings.TrimSpace(selfTenantID)
	if selfTenantID != "" {
		selfLabel := strings.TrimSpace(selfOrg)
		if selfLabel == "" {
			selfLabel = selfTenantID
		}
		selfLabel += " (self)"
		seen[selfTenantID] = struct{}{}
		labelCounts[selfLabel]++
		items = append(items, SelectItem{Label: selfLabel, ID: selfTenantID})
	}

	for _, acc := range accounts {
		tenantID := ""
		if acc.Extra != nil {
			tenantID = acc.Extra["tenantId"]
		}
		if tenantID == "" {
			continue
		}
		if _, dup := seen[tenantID]; dup {
			continue
		}
		seen[tenantID] = struct{}{}
		label := acc.Name
		if strings.TrimSpace(label) == "" {
			label = tenantID
		}
		if acc.Status != "" {
			label += "  " + acc.Status
		}
		labelCounts[label]++
		items = append(items, SelectItem{Label: label, ID: tenantID})
	}
	if len(items) == 0 {
		return nil
	}
	// Disambiguate any items that share a label: distinct tenants with the
	// same display name would otherwise route the request to whichever the
	// user happens to highlight, which is silently wrong. Append a short
	// tenant id suffix so the picker is always unambiguous.
	for i := range items {
		if labelCounts[items[i].Label] > 1 {
			items[i].Label = fmt.Sprintf("%s (%s)", items[i].Label, shortTenantID(items[i].ID))
		}
	}
	// Sort everything except the self entry (kept at the top so the
	// operator does not have to hunt for the common case).
	sortStart := 0
	if selfTenantID != "" {
		sortStart = 1
	}
	if sortStart < len(items) {
		tail := items[sortStart:]
		sort.SliceStable(tail, func(i, j int) bool {
			if tail[i].Label == tail[j].Label {
				return tail[i].ID < tail[j].ID
			}
			return tail[i].Label < tail[j].Label
		})
	}
	items = append(items, SelectItem{Label: "Enter Tenant ID manually...", ID: tenantManualEntrySentinel})
	return items
}

// shortTenantID returns up to the last 8 chars of a tenant UUID for use
// inside disambiguating picker labels. UUID strings shorter than 8 chars
// are returned as-is.
func shortTenantID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[len(id)-8:]
}

// promptAllocationConstraints collects exactly one allocation constraint.
// The REST API enforces len(allocationConstraints) == 1 in
// APIAllocationCreateRequest.Validate, so multiple entries are rejected
// server-side and a zero-length list is rejected as well.
func promptAllocationConstraints(s *Session, ctx context.Context) ([]map[string]interface{}, error) {
	fmt.Fprintf(os.Stderr, "%s allocation requires exactly one constraint\n", Dim("note:"))
	c, err := promptSingleAllocationConstraint(s, ctx)
	if err != nil {
		return nil, err
	}
	return []map[string]interface{}{c}, nil
}

func promptSingleAllocationConstraint(s *Session, ctx context.Context) (map[string]interface{}, error) {
	rt, err := Select("Resource type:", allocationConstraintResourceTypes())
	if err != nil {
		return nil, err
	}
	resolverKey, resolverLabel, ok := resolverResourceForAllocationResourceType(rt.ID)
	if !ok {
		return nil, fmt.Errorf("unsupported resource type %q", rt.ID)
	}
	item, err := s.Resolver.Resolve(ctx, resolverKey, resolverLabel)
	if err != nil {
		return nil, err
	}
	ct, err := Select("Constraint type:", allocationConstraintTypes())
	if err != nil {
		return nil, err
	}
	valueText, err := PromptText(fmt.Sprintf("Constraint value (%s)", allocationConstraintValueHint(rt.ID)), true)
	if err != nil {
		return nil, err
	}
	return buildAllocationConstraint(rt.ID, item.ID, ct.ID, valueText)
}

// allocationConstraintResourceTypes lists the supported resource types for an
// allocation constraint. These mirror the values accepted by the REST API
// (see APIAllocationConstraintCreateRequest.Validate).
func allocationConstraintResourceTypes() []SelectItem {
	return []SelectItem{
		{Label: "IPBlock (IP address allocation; value = prefix length)", ID: "IPBlock"},
		{Label: "InstanceType (machine allocation; value = machine count)", ID: "InstanceType"},
	}
}

// allocationConstraintTypes lists the constraint types offered to the user.
// Only Reserved is exposed: the API validator in
// api/pkg/api/model/allocationconstraint.go accepts OnDemand and Preemptible
// as well, but those two are documented as "not supported by current
// implementation" in the SDK and would turn a normal create flow into a
// server-side failure path.
//
// TODO(reenable-on-demand-preemptible): re-enable OnDemand and Preemptible
// SelectItems once the backend implements them end-to-end (track via the
// constraint-type validator in api/pkg/api/model/allocationconstraint.go).
func allocationConstraintTypes() []SelectItem {
	return []SelectItem{
		{Label: "Reserved", ID: "Reserved"},
	}
}

// resolverResourceForAllocationResourceType maps an allocation constraint
// resource type (as accepted by the API) to the TUI resolver key and a
// human-readable label to use when prompting.
func resolverResourceForAllocationResourceType(resourceType string) (resolverKey, label string, ok bool) {
	switch resourceType {
	case "IPBlock":
		return "ip-block", "IP Block", true
	case "InstanceType":
		return "instance-type", "Instance Type", true
	}
	return "", "", false
}

// allocationConstraintValueHint returns a short hint describing what the
// constraint value represents for a given resource type.
func allocationConstraintValueHint(resourceType string) string {
	switch resourceType {
	case "IPBlock":
		return "prefix length, e.g. 28"
	case "InstanceType":
		return "machine count, e.g. 4"
	}
	return "integer"
}

// buildAllocationConstraint assembles an API-shaped constraint body from its
// prompted fields. valueText is parsed as an integer and range-checked against
// the resource type so the user gets immediate feedback rather than waiting
// for a server-side rejection.
func buildAllocationConstraint(resourceType, resourceTypeID, constraintType, valueText string) (map[string]interface{}, error) {
	value, err := strconv.Atoi(strings.TrimSpace(valueText))
	if err != nil {
		return nil, fmt.Errorf("constraint value must be an integer: %w", err)
	}
	if err := validateAllocationConstraintValue(resourceType, value); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"resourceType":    resourceType,
		"resourceTypeId":  resourceTypeID,
		"constraintType":  constraintType,
		"constraintValue": value,
	}, nil
}

// validateAllocationConstraintValue enforces per-resource value ranges that
// the REST API itself currently validates only loosely (ConstraintValue is
// required but not range-checked, see the TODO in
// api/pkg/api/model/allocationconstraint.go).
func validateAllocationConstraintValue(resourceType string, value int) error {
	switch resourceType {
	case "IPBlock":
		if value < 1 || value > 32 {
			return fmt.Errorf("IPBlock constraint value must be an IPv4 prefix length between 1 and 32, got %d", value)
		}
	case "InstanceType":
		if value < 1 {
			return fmt.Errorf("InstanceType constraint value (machine count) must be at least 1, got %d", value)
		}
	default:
		if value <= 0 {
			return fmt.Errorf("constraint value must be a positive integer, got %d", value)
		}
	}
	return nil
}

func cmdAllocationUpdate(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "allocation", "Allocation to update", args)
	if err != nil {
		return err
	}
	name, err := PromptText("Allocation name (optional)", false)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	body := map[string]interface{}{}
	if strings.TrimSpace(name) != "" {
		body["name"] = strings.TrimSpace(name)
	}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	if len(body) == 0 {
		return fmt.Errorf("no updates provided")
	}
	LogCmd(s, "allocation", "update", item.ID)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("PATCH", apiPath(s, "allocation/{id}"), map[string]string{"id": item.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("updating allocation: %w", err)
	}
	s.Cache.Invalidate("allocation")
	s.Cache.InvalidateFiltered()
	var updated map[string]interface{}
	json.Unmarshal(resp, &updated)
	fmt.Printf("%s Allocation updated: %s (%s)\n", Green("OK"), str(updated, "name"), str(updated, "id"))
	return nil
}

func cmdIPBlockList(s *Session, _ []string) error {
	LogCmd(s, "ip-block", "list")
	items, err := s.Resolver.Fetch(context.Background(), "ip-block")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tSITE\tID")
	for _, item := range items {
		siteName := s.Resolver.ResolveID("site", item.Extra["siteId"])
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.Name, item.Status, siteName, item.ID)
	}
	return tw.Flush()
}

// ipBlockProtocolVersions are the valid values for the protocolVersion field on
// the ipblock create API. Order matters: the first option is used as the
// default in interactive prompts.
var ipBlockProtocolVersions = []string{"IPv4", "IPv6"}

// ipBlockRoutingTypes are the valid values for the routingType field on the
// ipblock create API. Order matters: the first option is used as the default
// in interactive prompts.
var ipBlockRoutingTypes = []string{"DatacenterOnly", "Public"}

// validateIPBlockPrefixLength returns an error when the prefix length is out
// of range for the selected protocol version. IPv4 allows 1-32, IPv6 allows
// 1-128.
func validateIPBlockPrefixLength(protocolVersion string, prefixLength int) error {
	switch protocolVersion {
	case "IPv4":
		if prefixLength < 1 || prefixLength > 32 {
			return fmt.Errorf("prefix length must be between 1 and 32 for IPv4")
		}
	case "IPv6":
		if prefixLength < 1 || prefixLength > 128 {
			return fmt.Errorf("prefix length must be between 1 and 128 for IPv6")
		}
	default:
		return fmt.Errorf("unsupported protocol version: %s", protocolVersion)
	}
	return nil
}

// buildIPBlockCreateBody constructs the JSON request body sent to the ipblock
// create API. The field names must match the API's APIIPBlockCreateRequest
// struct tags or the server will reject the request with a 400.
func buildIPBlockCreateBody(name, siteID, prefix string, prefixLength int, protocolVersion, routingType string) map[string]interface{} {
	return map[string]interface{}{
		"name":            name,
		"siteId":          siteID,
		"prefix":          prefix,
		"prefixLength":    prefixLength,
		"protocolVersion": protocolVersion,
		"routingType":     routingType,
	}
}

func cmdIPBlockCreate(s *Session, _ []string) error {
	site, err := s.Resolver.Resolve(context.Background(), "site", "Site")
	if err != nil {
		return err
	}
	name, err := PromptText("IP block name", true)
	if err != nil {
		return err
	}
	protocolVersion, err := PromptChoice("Protocol version", ipBlockProtocolVersions, ipBlockProtocolVersions[0])
	if err != nil {
		return err
	}
	routingType, err := PromptChoice("Routing type", ipBlockRoutingTypes, ipBlockRoutingTypes[0])
	if err != nil {
		return err
	}
	prefix, err := PromptText("Prefix (e.g. 10.0.0.0)", true)
	if err != nil {
		return err
	}
	prefixLen, err := PromptText("Prefix length (e.g. 16)", true)
	if err != nil {
		return err
	}
	pl, err := strconv.Atoi(strings.TrimSpace(prefixLen))
	if err != nil {
		return fmt.Errorf("prefix length must be a number, got %q", prefixLen)
	}
	if err := validateIPBlockPrefixLength(protocolVersion, pl); err != nil {
		return err
	}
	LogCmd(s, "ip-block", "create",
		"--name", name,
		"--site-id", site.ID,
		"--protocol-version", protocolVersion,
		"--routing-type", routingType,
		"--prefix", prefix,
		"--prefix-length", prefixLen,
	)
	body := buildIPBlockCreateBody(name, site.ID, prefix, pl, protocolVersion, routingType)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("POST", apiPath(s, "ipblock"), nil, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("creating IP block: %w", err)
	}
	s.Cache.Invalidate("ip-block")
	var created map[string]interface{}
	json.Unmarshal(resp, &created)
	fmt.Printf("%s IP block created: %s (%s)\n", Green("OK"), str(created, "name"), str(created, "id"))
	return nil
}

func cmdIPBlockUpdate(s *Session, args []string) error {
	block, err := s.Resolver.ResolveWithArgs(context.Background(), "ip-block", "IP Block to update", args)
	if err != nil {
		return err
	}
	name, err := PromptText("IP block name (optional)", false)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	body := map[string]interface{}{}
	if strings.TrimSpace(name) != "" {
		body["name"] = strings.TrimSpace(name)
	}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	if len(body) == 0 {
		return fmt.Errorf("no updates provided")
	}
	LogCmd(s, "ip-block", "update", block.ID)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("PATCH", apiPath(s, "ipblock/{id}"), map[string]string{"id": block.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("updating IP block: %w", err)
	}
	s.Cache.Invalidate("ip-block")
	s.Cache.InvalidateFiltered()
	var updated map[string]interface{}
	json.Unmarshal(resp, &updated)
	fmt.Printf("%s IP block updated: %s (%s)\n", Green("OK"), str(updated, "name"), str(updated, "id"))
	return nil
}

func cmdIPBlockDelete(s *Session, args []string) error {
	block, err := s.Resolver.ResolveWithArgs(context.Background(), "ip-block", "IP Block to delete", args)
	if err != nil {
		return err
	}
	ok, err := PromptConfirm(fmt.Sprintf("Delete IP block %s (%s)?", block.Name, block.ID))
	if err != nil || !ok {
		return err
	}
	LogCmd(s, "ip-block", "delete", block.ID)
	_, _, err = s.Client.Do("DELETE", apiPath(s, "ipblock/{id}"), map[string]string{"id": block.ID}, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting IP block: %w", err)
	}
	s.Cache.Invalidate("ip-block")
	fmt.Printf("%s IP block deleted: %s\n", Green("OK"), block.Name)
	return nil
}

func cmdNSGList(s *Session, args []string) error {
	LogCmd(s, "network-security-group", "list")
	items, err := s.Resolver.Fetch(context.Background(), "network-security-group")
	if err != nil {
		return err
	}
	_, cmdLabels, sortKey, err := parseLabelArgs(args)
	if err != nil {
		return err
	}
	merged, mergeErr := mergeLabels(s.Scope.LabelFilters, cmdLabels)
	if mergeErr != nil {
		return mergeErr
	}
	items = filterByLabels(items, merged)
	if sortKey != "" {
		items = sortByLabelKey(items, sortKey)
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	defer printLabelHint(os.Stderr, items, merged)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tLABELS\tID")
	for _, item := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.Name, item.Status, formatLabels(item.Labels, 60), item.ID)
	}
	return tw.Flush()
}

func cmdNSGCreate(s *Session, _ []string) error {
	site, err := s.Resolver.Resolve(context.Background(), "site", "Site")
	if err != nil {
		return err
	}
	name, err := PromptText("Network security group name", true)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	body := map[string]interface{}{"name": name, "siteId": site.ID}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	LogCmd(s, "network-security-group", "create", "--name", name, "--site-id", site.ID)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("POST", apiPath(s, "network-security-group"), nil, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("creating network security group: %w", err)
	}
	s.Cache.Invalidate("network-security-group")
	s.Cache.InvalidateFiltered()
	var created map[string]interface{}
	json.Unmarshal(resp, &created)
	fmt.Printf("%s Network security group created: %s (%s)\n", Green("OK"), str(created, "name"), str(created, "id"))
	return nil
}

func cmdNSGUpdate(s *Session, args []string) error {
	nsg, err := s.Resolver.ResolveWithArgs(context.Background(), "network-security-group", "Network Security Group to update", args)
	if err != nil {
		return err
	}
	name, err := PromptText("Network security group name (optional)", false)
	if err != nil {
		return err
	}
	desc, err := PromptText("Description (optional)", false)
	if err != nil {
		return err
	}
	body := map[string]interface{}{}
	if strings.TrimSpace(name) != "" {
		body["name"] = strings.TrimSpace(name)
	}
	if strings.TrimSpace(desc) != "" {
		body["description"] = strings.TrimSpace(desc)
	}
	if len(body) == 0 {
		return fmt.Errorf("no updates provided")
	}
	LogCmd(s, "network-security-group", "update", nsg.ID)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("PATCH", apiPath(s, "network-security-group/{id}"), map[string]string{"id": nsg.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("updating network security group: %w", err)
	}
	s.Cache.Invalidate("network-security-group")
	s.Cache.InvalidateFiltered()
	var updated map[string]interface{}
	json.Unmarshal(resp, &updated)
	fmt.Printf("%s Network security group updated: %s (%s)\n", Green("OK"), str(updated, "name"), str(updated, "id"))
	return nil
}

func cmdNSGDelete(s *Session, args []string) error {
	nsg, err := s.Resolver.ResolveWithArgs(context.Background(), "network-security-group", "Network Security Group to delete", args)
	if err != nil {
		return err
	}
	ok, err := PromptConfirm(fmt.Sprintf("Delete network security group %s (%s)?", nsg.Name, nsg.ID))
	if err != nil || !ok {
		return err
	}
	LogCmd(s, "network-security-group", "delete", nsg.ID)
	_, _, err = s.Client.Do("DELETE", apiPath(s, "network-security-group/{id}"), map[string]string{"id": nsg.ID}, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting network security group: %w", err)
	}
	s.Cache.Invalidate("network-security-group")
	s.Cache.InvalidateFiltered()
	fmt.Printf("%s Network security group deleted: %s\n", Green("OK"), nsg.Name)
	return nil
}

func cmdSKUList(s *Session, _ []string) error {
	LogCmd(s, "sku", "list")
	items, err := s.Resolver.Fetch(context.Background(), "sku")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "DEVICE TYPE\tSITE\tID")
	for _, item := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", item.Extra["deviceType"], item.Extra["siteId"], item.ID)
	}
	return tw.Flush()
}

func cmdRackList(s *Session, _ []string) error {
	LogCmd(s, "rack", "list")
	items, err := s.Resolver.Fetch(context.Background(), "rack")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tMANUFACTURER\tMODEL\tID")
	for _, item := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.Name, item.Extra["manufacturer"], item.Extra["model"], item.ID)
	}
	return tw.Flush()
}

func cmdVPCPrefixList(s *Session, _ []string) error {
	LogCmd(s, "vpc-prefix", "list")
	items, err := s.Resolver.Fetch(context.Background(), "vpc-prefix")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tVPC\tID")
	for _, item := range items {
		vpcName := s.Resolver.ResolveID("vpc", item.Extra["vpcId"])
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.Name, item.Status, vpcName, item.ID)
	}
	return tw.Flush()
}

func cmdVPCPrefixCreate(s *Session, _ []string) error {
	vpc, err := s.Resolver.Resolve(context.Background(), "vpc", "VPC")
	if err != nil {
		return err
	}
	vpcSiteID := strings.TrimSpace(vpc.Extra["siteId"])
	setSiteScopeFromID(s, vpcSiteID)

	name, err := PromptText("VPC prefix name", true)
	if err != nil {
		return err
	}
	prefixLenText, err := PromptText("Prefix length (8-31)", true)
	if err != nil {
		return err
	}
	var prefixLen int
	fmt.Sscanf(prefixLenText, "%d", &prefixLen)
	if prefixLen < 8 || prefixLen > 31 {
		return fmt.Errorf("prefix length must be between 8 and 31")
	}
	ipBlockID, err := PromptText("IP block ID", true)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"name":         name,
		"vpcId":        vpc.ID,
		"ipBlockId":    strings.TrimSpace(ipBlockID),
		"prefixLength": prefixLen,
	}
	LogCmd(s, "vpc-prefix", "create", "--name", name, "--vpc-id", vpc.ID, "--ip-block-id", strings.TrimSpace(ipBlockID), "--prefix-length", prefixLenText)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("POST", apiPath(s, "vpc-prefix"), nil, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("creating VPC prefix: %w", err)
	}
	s.Cache.Invalidate("vpc-prefix")
	s.Cache.InvalidateFiltered()
	var created map[string]interface{}
	json.Unmarshal(resp, &created)
	fmt.Printf("%s VPC prefix created: %s (%s)\n", Green("OK"), str(created, "name"), str(created, "id"))
	return nil
}

func cmdVPCPrefixUpdate(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "vpc-prefix", "VPC Prefix to update", args)
	if err != nil {
		return err
	}
	name, err := PromptText("VPC prefix name", true)
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"name": strings.TrimSpace(name),
	}
	LogCmd(s, "vpc-prefix", "update", item.ID, "--name", strings.TrimSpace(name))
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("PATCH", apiPath(s, "vpc-prefix/{id}"), map[string]string{"id": item.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("updating VPC prefix: %w", err)
	}
	s.Cache.Invalidate("vpc-prefix")
	s.Cache.InvalidateFiltered()
	var updated map[string]interface{}
	json.Unmarshal(resp, &updated)
	fmt.Printf("%s VPC prefix updated: %s (%s)\n", Green("OK"), str(updated, "name"), str(updated, "id"))
	return nil
}

func cmdVPCPrefixDelete(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "vpc-prefix", "VPC Prefix to delete", args)
	if err != nil {
		return err
	}
	ok, err := PromptConfirm(fmt.Sprintf("Delete VPC prefix %s (%s)?", item.Name, item.ID))
	if err != nil || !ok {
		return err
	}
	LogCmd(s, "vpc-prefix", "delete", item.ID)
	_, _, err = s.Client.Do("DELETE", apiPath(s, "vpc-prefix/{id}"), map[string]string{"id": item.ID}, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting VPC prefix: %w", err)
	}
	s.Cache.Invalidate("vpc-prefix")
	s.Cache.InvalidateFiltered()
	fmt.Printf("%s VPC prefix deleted: %s\n", Green("OK"), item.Name)
	return nil
}

func cmdTenantAccountList(s *Session, _ []string) error {
	ctx := context.Background()
	var extraFlags []string
	if id, err := s.getInfrastructureProviderID(ctx); err == nil {
		extraFlags = append(extraFlags, "--infrastructure-provider-id", id)
	} else if id, err := s.getTenantID(ctx); err == nil {
		extraFlags = append(extraFlags, "--tenant-id", id)
	}
	LogCmd(s, append([]string{"tenant-account", "list"}, extraFlags...)...)
	items, err := s.Resolver.Fetch(ctx, "tenant-account")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "TENANT ORG\tSTATUS\tINFRA PROVIDER ID\tID")
	for _, item := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.Name, item.Status, item.Extra["infrastructureProviderId"], item.ID)
	}
	return tw.Flush()
}

func cmdTenantAccountCreate(s *Session, _ []string) error {
	infrastructureProviderID, err := PromptText("Infrastructure provider ID", true)
	if err != nil {
		return err
	}
	tenantOrg, err := PromptText("Tenant org", true)
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"infrastructureProviderId": strings.TrimSpace(infrastructureProviderID),
		"tenantOrg":                strings.TrimSpace(tenantOrg),
	}
	LogCmd(s, "tenant-account", "create", "--tenant-org", strings.TrimSpace(tenantOrg))
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("POST", apiPath(s, "tenant/account"), nil, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("creating tenant account: %w", err)
	}
	s.Cache.Invalidate("tenant-account")
	var created map[string]interface{}
	json.Unmarshal(resp, &created)
	fmt.Printf("%s Tenant account created: %s (%s)\n", Green("OK"), str(created, "tenantOrg"), str(created, "id"))
	return nil
}

func cmdTenantAccountUpdate(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "tenant-account", "Tenant Account to accept", args)
	if err != nil {
		return err
	}
	ok, err := PromptConfirm(fmt.Sprintf("Accept invitation for tenant account %s (%s)?", item.Name, item.ID))
	if err != nil || !ok {
		return err
	}
	LogCmd(s, "tenant-account", "update", item.ID)
	bodyJSON, _ := json.Marshal(map[string]interface{}{})
	resp, _, err := s.Client.Do("PATCH", apiPath(s, "tenant/account/{id}"), map[string]string{"id": item.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("accepting tenant account invitation: %w", err)
	}
	s.Cache.Invalidate("tenant-account")
	var updated map[string]interface{}
	json.Unmarshal(resp, &updated)
	fmt.Printf("%s Tenant account accepted: %s (%s)\n", Green("OK"), str(updated, "tenantOrg"), str(updated, "id"))
	return nil
}

func cmdTenantAccountDelete(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "tenant-account", "Tenant Account to delete", args)
	if err != nil {
		return err
	}
	ok, err := PromptConfirm(fmt.Sprintf("Delete tenant account %s (%s)?", item.Name, item.ID))
	if err != nil || !ok {
		return err
	}
	LogCmd(s, "tenant-account", "delete", item.ID)
	_, _, err = s.Client.Do("DELETE", apiPath(s, "tenant/account/{id}"), map[string]string{"id": item.ID}, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting tenant account: %w", err)
	}
	s.Cache.Invalidate("tenant-account")
	fmt.Printf("%s Tenant account deleted: %s\n", Green("OK"), item.Name)
	return nil
}

func cmdExpectedMachineList(s *Session, args []string) error {
	LogCmd(s, "expected-machine", "list")
	items, err := s.Resolver.Fetch(context.Background(), "expected-machine")
	if err != nil {
		return err
	}
	_, cmdLabels, sortKey, err := parseLabelArgs(args)
	if err != nil {
		return err
	}
	merged, mergeErr := mergeLabels(s.Scope.LabelFilters, cmdLabels)
	if mergeErr != nil {
		return mergeErr
	}
	items = filterByLabels(items, merged)
	if sortKey != "" {
		items = sortByLabelKey(items, sortKey)
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	defer printLabelHint(os.Stderr, items, merged)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "SITE ID\tBMC MAC\tCHASSIS SN\tLABELS\tID")
	for _, item := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", item.Extra["siteId"], item.Extra["bmcMacAddress"], item.Extra["chassisSerialNumber"], formatLabels(item.Labels, 60), item.ID)
	}
	return tw.Flush()
}

func cmdInfiniBandPartitionList(s *Session, args []string) error {
	LogCmd(s, "infiniband-partition", "list")
	items, err := s.Resolver.Fetch(context.Background(), "infiniband-partition")
	if err != nil {
		return err
	}
	_, cmdLabels, sortKey, err := parseLabelArgs(args)
	if err != nil {
		return err
	}
	merged, mergeErr := mergeLabels(s.Scope.LabelFilters, cmdLabels)
	if mergeErr != nil {
		return mergeErr
	}
	items = filterByLabels(items, merged)
	if sortKey != "" {
		items = sortByLabelKey(items, sortKey)
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	defer printLabelHint(os.Stderr, items, merged)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tLABELS\tID")
	for _, item := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.Name, item.Status, formatLabels(item.Labels, 60), item.ID)
	}
	return tw.Flush()
}

func cmdNVLinkLogicalPartitionList(s *Session, _ []string) error {
	LogCmd(s, "nvlink-logical-partition", "list")
	items, err := s.Resolver.Fetch(context.Background(), "nvlink-logical-partition")
	if err != nil {
		return err
	}
	return printResourceTable(os.Stdout, "NAME", "STATUS", "ID", items)
}

func cmdDPUExtensionServiceList(s *Session, _ []string) error {
	LogCmd(s, "dpu-extension-service", "list")
	items, err := s.Resolver.Fetch(context.Background(), "dpu-extension-service")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tTYPE\tSITE ID\tID")
	for _, item := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.Name, item.Extra["serviceType"], item.Extra["siteId"], item.ID)
	}
	return tw.Flush()
}

func cmdAuditList(s *Session, _ []string) error {
	LogCmd(s, "audit", "list")
	items, err := s.Resolver.Fetch(context.Background(), "audit")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "METHOD\tENDPOINT\tSTATUS CODE\tID")
	for _, item := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.Extra["method"], item.Extra["endpoint"], item.Status, item.ID)
	}
	return tw.Flush()
}

// -- Get commands (raw JSON detail) --

func cmdSiteGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "site", "Site", args)
	if err != nil {
		return err
	}
	LogCmd(s, "site", "get", item.ID)
	return getAndPrint(s, apiPath(s, "site/{id}"), item.ID)
}

func cmdVPCGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "vpc", "VPC", args)
	if err != nil {
		return err
	}
	LogCmd(s, "vpc", "get", item.ID)
	return getAndPrint(s, apiPath(s, "vpc/{id}"), item.ID)
}

func cmdSubnetGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "subnet", "Subnet", args)
	if err != nil {
		return err
	}
	LogCmd(s, "subnet", "get", item.ID)
	return getAndPrint(s, apiPath(s, "subnet/{id}"), item.ID)
}

func cmdInstanceGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "instance", "Instance", args)
	if err != nil {
		return err
	}
	LogCmd(s, "instance", "get", item.ID)
	return getAndPrint(s, apiPath(s, "instance/{id}"), item.ID)
}

func cmdInstanceCreate(s *Session, _ []string) error {
	ctx := context.Background()
	vpc, err := s.Resolver.Resolve(ctx, "vpc", "VPC")
	if err != nil {
		return err
	}
	vpcSiteID := strings.TrimSpace(vpc.Extra["siteId"])
	setSiteScopeFromID(s, vpcSiteID)

	// Temporarily clear VPC scope so fetchMachines returns all site machines
	// rather than filtering to machines already assigned to a prior VPC.
	savedVpcID, savedVpcName := s.Scope.VpcID, s.Scope.VpcName
	s.Scope.VpcID, s.Scope.VpcName = "", ""
	machines, err := fetchMachinesWithSiteFallback(s, "Machine listing requires a site filter. Select a site.")
	s.Scope.VpcID, s.Scope.VpcName = savedVpcID, savedVpcName
	if err != nil {
		return fmt.Errorf("fetching machines: %w", err)
	}
	readyItems := readyMachineItemsForSite(machines, vpcSiteID)
	if len(readyItems) == 0 {
		if vpcSiteID != "" {
			return fmt.Errorf("no machines in Ready state available for selected VPC site")
		}
		return fmt.Errorf("no machines in Ready state available")
	}
	machine, err := Select("Machine", readyItems)
	if err != nil {
		return err
	}
	name, err := PromptText("Instance name", true)
	if err != nil {
		return err
	}

	var osID *string
	osList, osErr := s.Resolver.Fetch(ctx, "operating-system")
	if osErr == nil && len(osList) > 0 {
		useOS, confirmErr := PromptConfirm("Select an operating system?")
		if confirmErr != nil {
			return confirmErr
		}
		if useOS {
			osItem, selectErr := s.Resolver.Resolve(ctx, "operating-system", "Operating System")
			if selectErr != nil {
				return selectErr
			}
			osID = &osItem.ID
		}
	}

	// Scope vpc-prefix lookups to the selected VPC so the picker only offers
	// prefixes that are actually attachable to this instance.
	savedVpcID2, savedVpcName2 := s.Scope.VpcID, s.Scope.VpcName
	s.Scope.VpcID, s.Scope.VpcName = vpc.ID, vpc.Name
	s.Cache.InvalidateFiltered()
	defer func() {
		s.Scope.VpcID, s.Scope.VpcName = savedVpcID2, savedVpcName2
		s.Cache.InvalidateFiltered()
	}()

	interfaces, err := promptInstanceInterfaces(s, ctx)
	if err != nil {
		return err
	}

	sshKeyGroupIDs, err := promptOptionalResourceIDs(s, ctx, "ssh-key-group", "SSH key group")
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"name":      name,
		"machineId": machine.ID,
		"vpcId":     vpc.ID,
	}
	if osID != nil {
		body["operatingSystemId"] = *osID
	}
	if len(interfaces) > 0 {
		body["interfaces"] = interfaces
	}
	if len(sshKeyGroupIDs) > 0 {
		body["sshKeyGroupIds"] = sshKeyGroupIDs
	}
	LogCmd(s, "instance", "create", "--name", name, "--machine-id", machine.ID, "--vpc-id", vpc.ID)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("POST", apiPath(s, "instance"), nil, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	s.Cache.Invalidate("instance")
	s.Cache.InvalidateFiltered()
	var created map[string]interface{}
	json.Unmarshal(resp, &created)
	fmt.Printf("%s Instance created: %s (%s)\n", Green("OK"), str(created, "name"), str(created, "id"))
	return nil
}

// promptInstanceInterfaces builds the interfaces[] array for an instance
// create request by walking the operator through one VPC-prefix-backed
// interface at a time. The OpenAPI schema requires at least one entry, so
// the first interface is always prompted; subsequent interfaces are opt-in.
// Returns nil (not error) if no vpc-prefixes exist for the current VPC scope
// so cmdInstanceCreate can still attempt the API call and surface the
// server-side validation error instead of silently sending an empty array.
func promptInstanceInterfaces(s *Session, ctx context.Context) ([]map[string]interface{}, error) {
	prefixes, err := s.Resolver.Fetch(ctx, "vpc-prefix")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s could not list vpc-prefixes (%v); the API may reject this create if interfaces are required\n", Dim("note:"), err)
		return nil, nil
	}
	if len(prefixes) == 0 {
		fmt.Fprintf(os.Stderr, "%s no vpc-prefixes available for the selected VPC; the API may reject this create if interfaces are required\n", Dim("note:"))
		return nil, nil
	}
	var ifaces []map[string]interface{}
	usedPrefixes := make(map[string]bool)
	for {
		label := "VPC prefix for interface"
		if len(ifaces) > 0 {
			confirmLabel := fmt.Sprintf("Add another interface (have %d)?", len(ifaces))
			more, confirmErr := PromptConfirm(confirmLabel)
			if confirmErr != nil {
				return ifaces, confirmErr
			}
			if !more {
				return ifaces, nil
			}
		}
		available := make([]NamedItem, 0, len(prefixes))
		for _, p := range prefixes {
			if !usedPrefixes[p.ID] {
				available = append(available, p)
			}
		}
		if len(available) == 0 {
			fmt.Fprintf(os.Stderr, "%s no more vpc-prefixes to attach\n", Dim("note:"))
			return ifaces, nil
		}
		picked, err := s.Resolver.SelectFromItems(label, available)
		if err != nil {
			return ifaces, err
		}
		usedPrefixes[picked.ID] = true
		ifaces = append(ifaces, map[string]interface{}{
			"vpcPrefixId": picked.ID,
			"isPhysical":  true,
		})
	}
}

// promptOptionalResourceIDs offers a series of single-select pickers on the
// given resource type, accumulating IDs until the user declines to add
// another. Returns nil if the resource type has no items at all (so callers
// can omit the field entirely from the request body).
func promptOptionalResourceIDs(s *Session, ctx context.Context, resourceType, singular string) ([]string, error) {
	items, err := s.Resolver.Fetch(ctx, resourceType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s could not list %s (%v); skipping\n", Dim("note:"), resourceType, err)
		return nil, nil
	}
	if len(items) == 0 {
		return nil, nil
	}
	var ids []string
	picked := make(map[string]bool)
	for {
		prompt := fmt.Sprintf("Add %s?", singular)
		if len(ids) > 0 {
			prompt = fmt.Sprintf("Add another %s (have %d)?", singular, len(ids))
		}
		more, err := PromptConfirm(prompt)
		if err != nil {
			return ids, err
		}
		if !more {
			return ids, nil
		}
		available := make([]NamedItem, 0, len(items))
		for _, item := range items {
			if !picked[item.ID] {
				available = append(available, item)
			}
		}
		if len(available) == 0 {
			fmt.Fprintf(os.Stderr, "%s no more %s available\n", Dim("note:"), resourceType)
			return ids, nil
		}
		selected, err := s.Resolver.SelectFromItems(singular, available)
		if err != nil {
			return ids, err
		}
		ids = append(ids, selected.ID)
		picked[selected.ID] = true
	}
}

// instanceUpdateInputs collects the optional fields exposed by the TUI
// instance update form. Extracted so cmdInstanceUpdate stays linear and
// cmdInstanceReboot can drive a stripped-down version of the same flow.
type instanceUpdateInputs struct {
	name                 string
	description          string
	osID                 string
	sshKeyGroupIDs       []string
	triggerReboot        bool
	rebootWithCustomIpxe bool
	applyUpdatesOnReboot bool
}

func (u instanceUpdateInputs) toBody() map[string]interface{} {
	body := map[string]interface{}{}
	if strings.TrimSpace(u.name) != "" {
		body["name"] = strings.TrimSpace(u.name)
	}
	if strings.TrimSpace(u.description) != "" {
		body["description"] = strings.TrimSpace(u.description)
	}
	if strings.TrimSpace(u.osID) != "" {
		body["operatingSystemId"] = strings.TrimSpace(u.osID)
	}
	if len(u.sshKeyGroupIDs) > 0 {
		body["sshKeyGroupIds"] = u.sshKeyGroupIDs
	}
	if u.triggerReboot {
		body["triggerReboot"] = true
		if u.rebootWithCustomIpxe {
			body["rebootWithCustomIpxe"] = true
		}
		if u.applyUpdatesOnReboot {
			body["applyUpdatesOnReboot"] = true
		}
	}
	return body
}

func cmdInstanceUpdate(s *Session, args []string) error {
	ctx := context.Background()
	item, err := s.Resolver.ResolveWithArgs(ctx, "instance", "Instance to update", args)
	if err != nil {
		return err
	}
	inputs := instanceUpdateInputs{}
	inputs.name, err = PromptText("New name (optional)", false)
	if err != nil {
		return err
	}
	inputs.description, err = PromptText("New description (optional)", false)
	if err != nil {
		return err
	}

	if osList, osErr := s.Resolver.Fetch(ctx, "operating-system"); osErr == nil && len(osList) > 0 {
		changeOS, confirmErr := PromptConfirm("Change operating system?")
		if confirmErr != nil {
			return confirmErr
		}
		if changeOS {
			osItem, selectErr := s.Resolver.Resolve(ctx, "operating-system", "Operating System")
			if selectErr != nil {
				return selectErr
			}
			inputs.osID = osItem.ID
		}
	}

	rotateGroups, err := PromptConfirm("Replace ssh key groups (this overwrites the existing list)?")
	if err != nil {
		return err
	}
	if rotateGroups {
		inputs.sshKeyGroupIDs, err = promptOptionalResourceIDs(s, ctx, "ssh-key-group", "SSH key group")
		if err != nil {
			return err
		}
	}

	inputs.triggerReboot, err = PromptConfirm("Trigger reboot now?")
	if err != nil {
		return err
	}
	if inputs.triggerReboot {
		inputs.rebootWithCustomIpxe, err = PromptConfirm("Reboot with custom iPXE (one-time)?")
		if err != nil {
			return err
		}
		inputs.applyUpdatesOnReboot, err = PromptConfirm("Apply pending updates on reboot?")
		if err != nil {
			return err
		}
	}

	body := inputs.toBody()
	if len(body) == 0 {
		return fmt.Errorf("no updates provided")
	}

	LogCmd(s, "instance", "update", item.ID)
	bodyJSON, _ := json.Marshal(body)
	resp, _, err := s.Client.Do("PATCH", apiPath(s, "instance/{id}"), map[string]string{"id": item.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	s.Cache.Invalidate("instance")
	s.Cache.InvalidateFiltered()
	var updated map[string]interface{}
	json.Unmarshal(resp, &updated)
	fmt.Printf("%s Instance updated: %s (%s)\n", Green("OK"), str(updated, "name"), str(updated, "id"))
	return nil
}

func cmdInstanceReboot(s *Session, args []string) error {
	ctx := context.Background()
	item, err := s.Resolver.ResolveWithArgs(ctx, "instance", "Instance to reboot", args)
	if err != nil {
		return err
	}
	rebootWithCustomIpxe, err := PromptConfirm("Reboot with custom iPXE (one-time)?")
	if err != nil {
		return err
	}
	applyUpdatesOnReboot, err := PromptConfirm("Apply pending updates on reboot?")
	if err != nil {
		return err
	}
	confirm, err := PromptConfirm(fmt.Sprintf("Reboot instance %s (%s) now?", item.Name, item.ID))
	if err != nil || !confirm {
		return err
	}

	body := instanceUpdateInputs{
		triggerReboot:        true,
		rebootWithCustomIpxe: rebootWithCustomIpxe,
		applyUpdatesOnReboot: applyUpdatesOnReboot,
	}.toBody()

	LogCmd(s, "instance", "update", item.ID, "--trigger-reboot=true")
	bodyJSON, _ := json.Marshal(body)
	_, _, err = s.Client.Do("PATCH", apiPath(s, "instance/{id}"), map[string]string{"id": item.ID}, nil, bodyJSON)
	if err != nil {
		return fmt.Errorf("rebooting instance: %w", err)
	}
	s.Cache.Invalidate("instance")
	s.Cache.InvalidateFiltered()
	fmt.Printf("%s Reboot requested for instance %s (%s)\n", Green("OK"), item.Name, item.ID)
	return nil
}

func cmdInstanceDelete(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "instance", "Instance to delete", args)
	if err != nil {
		return err
	}
	ok, err := PromptConfirm(fmt.Sprintf("Delete instance %s (%s)?", item.Name, item.ID))
	if err != nil || !ok {
		return err
	}
	LogCmd(s, "instance", "delete", item.ID)
	_, _, err = s.Client.Do("DELETE", apiPath(s, "instance/{id}"), map[string]string{"id": item.ID}, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	s.Cache.Invalidate("instance")
	s.Cache.InvalidateFiltered()
	fmt.Printf("%s Instance deleted: %s\n", Green("OK"), item.Name)
	return nil
}

func cmdMachineGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "machine", "Machine", args)
	if err != nil {
		return err
	}
	LogCmd(s, "machine", "get", item.ID)
	return getAndPrint(s, apiPath(s, "machine/{id}"), item.ID)
}

func cmdOSGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "operating-system", "Operating System", args)
	if err != nil {
		return err
	}
	LogCmd(s, "operating-system", "get", item.ID)
	return getAndPrint(s, apiPath(s, "operating-system/{id}"), item.ID)
}

func cmdSSHKeyGroupGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "ssh-key-group", "SSH Key Group", args)
	if err != nil {
		return err
	}
	LogCmd(s, "ssh-key-group", "get", item.ID)
	return getAndPrint(s, apiPath(s, "sshkeygroup/{id}"), item.ID)
}

func cmdSSHKeyGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "ssh-key", "SSH Key", args)
	if err != nil {
		return err
	}
	LogCmd(s, "ssh-key", "get", item.ID)
	return getAndPrint(s, apiPath(s, "sshkey/{id}"), item.ID)
}

func cmdAllocationGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "allocation", "Allocation", args)
	if err != nil {
		return err
	}
	LogCmd(s, "allocation", "get", item.ID)
	return getAndPrint(s, apiPath(s, "allocation/{id}"), item.ID)
}

func cmdAllocationDelete(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "allocation", "Allocation to delete", args)
	if err != nil {
		return err
	}
	ok, err := PromptConfirm(fmt.Sprintf("Delete allocation %s (%s)?", item.Name, item.ID))
	if err != nil || !ok {
		return err
	}
	LogCmd(s, "allocation", "delete", item.ID)
	_, _, err = s.Client.Do("DELETE", apiPath(s, "allocation/{id}"), map[string]string{"id": item.ID}, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting allocation: %w", err)
	}
	s.Cache.Invalidate("allocation")
	fmt.Printf("%s Allocation deleted: %s\n", Green("OK"), item.Name)
	return nil
}

func cmdIPBlockGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "ip-block", "IP Block", args)
	if err != nil {
		return err
	}
	LogCmd(s, "ip-block", "get", item.ID)
	return getAndPrint(s, apiPath(s, "ipblock/{id}"), item.ID)
}

func cmdNSGGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "network-security-group", "Network Security Group", args)
	if err != nil {
		return err
	}
	LogCmd(s, "network-security-group", "get", item.ID)
	return getAndPrint(s, apiPath(s, "network-security-group/{id}"), item.ID)
}

func cmdSKUGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "sku", "SKU", args)
	if err != nil {
		return err
	}
	LogCmd(s, "sku", "get", item.ID)
	return getAndPrint(s, apiPath(s, "sku/{id}"), item.ID)
}

func cmdRackGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "rack", "Rack", args)
	if err != nil {
		return err
	}
	LogCmd(s, "rack", "get", item.ID)
	return getAndPrint(s, apiPath(s, "rack/{id}"), item.ID)
}

func cmdVPCPrefixGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "vpc-prefix", "VPC Prefix", args)
	if err != nil {
		return err
	}
	LogCmd(s, "vpc-prefix", "get", item.ID)
	return getAndPrint(s, apiPath(s, "vpc-prefix/{id}"), item.ID)
}

func cmdTenantAccountGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "tenant-account", "Tenant Account", args)
	if err != nil {
		return err
	}
	LogCmd(s, "tenant-account", "get", item.ID)
	return getAndPrint(s, apiPath(s, "tenant/account/{id}"), item.ID)
}

func cmdExpectedMachineGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "expected-machine", "Expected Machine", args)
	if err != nil {
		return err
	}
	LogCmd(s, "expected-machine", "get", item.ID)
	return getAndPrint(s, apiPath(s, "expected-machine/{id}"), item.ID)
}

func cmdInfiniBandPartitionGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "infiniband-partition", "InfiniBand Partition", args)
	if err != nil {
		return err
	}
	LogCmd(s, "infiniband-partition", "get", item.ID)
	return getAndPrint(s, apiPath(s, "infiniband-partition/{id}"), item.ID)
}

func cmdNVLinkLogicalPartitionGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "nvlink-logical-partition", "NVLink Logical Partition", args)
	if err != nil {
		return err
	}
	LogCmd(s, "nvlink-logical-partition", "get", item.ID)
	return getAndPrint(s, apiPath(s, "nvlink-logical-partition/{id}"), item.ID)
}

func cmdDPUExtensionServiceGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "dpu-extension-service", "DPU Extension Service", args)
	if err != nil {
		return err
	}
	LogCmd(s, "dpu-extension-service", "get", item.ID)
	return getAndPrint(s, apiPath(s, "dpu-extension-service/{id}"), item.ID)
}

func cmdAuditGet(s *Session, args []string) error {
	item, err := s.Resolver.ResolveWithArgs(context.Background(), "audit", "Audit Entry", args)
	if err != nil {
		return err
	}
	LogCmd(s, "audit", "get", item.ID)
	return getAndPrint(s, apiPath(s, "audit/{id}"), item.ID)
}

// -- Singleton / info commands --

func cmdMetadataGet(s *Session, _ []string) error {
	LogCmd(s, "metadata", "get")
	body, _, err := s.Client.Do("GET", apiPath(s, "metadata"), nil, nil, nil)
	if err != nil {
		return fmt.Errorf("getting metadata: %w", err)
	}
	return printDetailJSON(os.Stdout, body)
}

func cmdUserCurrent(s *Session, _ []string) error {
	LogCmd(s, "user", "current")
	body, _, err := s.Client.Do("GET", apiPath(s, "user/current"), nil, nil, nil)
	if err != nil {
		return fmt.Errorf("getting current user: %w", err)
	}
	return printDetailJSON(os.Stdout, body)
}

func cmdTenantCurrent(s *Session, _ []string) error {
	LogCmd(s, "tenant", "current")
	body, _, err := s.Client.Do("GET", apiPath(s, "tenant/current"), nil, nil, nil)
	if err != nil {
		return fmt.Errorf("getting current tenant: %w", err)
	}
	return printDetailJSON(os.Stdout, body)
}

func cmdTenantStats(s *Session, _ []string) error {
	LogCmd(s, "tenant", "stats")
	body, _, err := s.Client.Do("GET", apiPath(s, "tenant/current/stats"), nil, nil, nil)
	if err != nil {
		return fmt.Errorf("getting tenant stats: %w", err)
	}
	return printDetailJSON(os.Stdout, body)
}

func cmdInfraProviderCurrent(s *Session, _ []string) error {
	LogCmd(s, "infrastructure-provider", "current")
	body, _, err := s.Client.Do("GET", apiPath(s, "infrastructure-provider/current"), nil, nil, nil)
	if err != nil {
		return fmt.Errorf("getting infrastructure provider: %w", err)
	}
	return printDetailJSON(os.Stdout, body)
}

func cmdInfraProviderStats(s *Session, _ []string) error {
	LogCmd(s, "infrastructure-provider", "stats")
	body, _, err := s.Client.Do("GET", apiPath(s, "infrastructure-provider/current/stats"), nil, nil, nil)
	if err != nil {
		return fmt.Errorf("getting infrastructure provider stats: %w", err)
	}
	return printDetailJSON(os.Stdout, body)
}

func cmdLogin(s *Session, _ []string) error {
	if s.LoginFn == nil {
		return fmt.Errorf("login not available (no auth method configured)")
	}
	fmt.Println("Logging in...")
	token, err := s.LoginFn()
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	s.RefreshClient(token)
	fmt.Printf("%s Logged in successfully.\n", Green("OK"))
	return nil
}

func cmdHelp(_ *Session, _ []string) error {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "\nCOMMAND\tDESCRIPTION")
	fmt.Fprintln(tw, "-------\t-----------")
	for _, cmd := range AllCommands() {
		fmt.Fprintf(tw, "%s\t%s\n", cmd.Name, cmd.Description)
	}
	fmt.Fprintln(tw, "org\tShow current org")
	fmt.Fprintln(tw, "org list\tList available orgs (from JWT claims)")
	fmt.Fprintln(tw, "org set <name>\tSwitch to a different org")
	fmt.Fprintln(tw, "scope\tShow current scope filters")
	fmt.Fprintln(tw, "scope site [name]\tSet site scope (filters lists)")
	fmt.Fprintln(tw, "scope vpc [name]\tSet VPC scope (filters lists)")
	fmt.Fprintln(tw, "scope label key=value\tAdd a label filter (persists across commands)")
	fmt.Fprintln(tw, "scope label clear\tClear all label filters")
	fmt.Fprintln(tw, "scope clear\tClear all scope filters")
	fmt.Fprintln(tw, "<list> --label key=value\tFilter a single label-capable list by a label (no persist)")
	fmt.Fprintln(tw, "<list> --sort-label key\tSort a single label-capable list by a label key")
	fmt.Fprintln(tw, "\t  label-capable lists: "+strings.Join(labelCapableListCommands, ", "))
	fmt.Fprintln(tw, "exit\tExit interactive mode")
	tw.Flush()
	fmt.Printf("\n%s\n", Bold("KEYBINDINGS"))
	fmt.Println("  Ctrl+C    Clear current line")
	fmt.Println("  Ctrl+D    Quit interactive mode")
	fmt.Println("  Esc       Cancel current selection")
	fmt.Println("  Tab       Accept suggestion")
	fmt.Println("  Up/Down   Navigate suggestions or history")
	fmt.Println()
	return nil
}

// -- Helpers --

func splitCommaSeparated(input string) []string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func parseOptionalBool(input string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "":
		return false, false
	case "true", "t", "yes", "y", "1":
		return true, true
	case "false", "f", "no", "n", "0":
		return false, true
	default:
		return false, false
	}
}

func parseOptionalInt(input string) (int, bool) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return 0, false
	}
	var out int
	if _, err := fmt.Sscanf(trimmed, "%d", &out); err != nil {
		return 0, false
	}
	return out, true
}

func rawFieldString(raw interface{}, key string) string {
	m, ok := raw.(map[string]interface{})
	if !ok {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

func getAndPrint(s *Session, path, id string) error {
	body, _, err := s.Client.Do("GET", path, map[string]string{"id": id}, nil, nil)
	if err != nil {
		return err
	}
	return printDetailJSON(os.Stdout, body)
}

func printDetailJSON(w io.Writer, data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Fprintln(w, string(data))
		return nil
	}
	pretty, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintln(w, string(data))
		return nil
	}
	fmt.Fprintln(w, string(pretty))
	return nil
}

func formatLabels(labels map[string]string, maxWidth int) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+labels[k])
	}
	s := strings.Join(parts, ", ")
	if maxWidth > 3 && len(s) > maxWidth {
		return s[:maxWidth-3] + "..."
	}
	return s
}

// labelCapableListCommands enumerates the list commands that render a LABELS
// column and accept --label / --sort-label flags. Kept here next to the hint
// helpers so adding a new label-capable list in one place keeps the hint
// behavior, the help text, and any future label-aware tooling in sync.
var labelCapableListCommands = []string{
	"vpc list",
	"instance-type list",
	"instance list",
	"machine list",
	"network-security-group list",
	"expected-machine list",
	"infiniband-partition list",
}

// printLabelHint writes a one-line hint about how to filter, sort, and persist
// label filters. Uses generic <key>/<value> placeholders -- the hint is the
// same regardless of what is in the current result set, so no per-call sample
// or key listing is needed (and no max-keys cap to tune). Suppressed when a
// label filter is already active or when no items carry labels, so the hint
// only surfaces when it is actionable.
func printLabelHint(w io.Writer, items []NamedItem, activeFilters map[string]string) {
	if len(activeFilters) > 0 {
		return
	}
	if !anyItemHasLabels(items) {
		return
	}
	fmt.Fprintln(w, Dim(
		"Hint: --label <key>=<value> filter | --sort-label <key> sort | scope label <key>=<value> persist",
	))
}

// anyItemHasLabels reports whether at least one item carries a non-empty
// label key. Used to gate printLabelHint so the hint only appears on lists
// where labels are actually in play.
func anyItemHasLabels(items []NamedItem) bool {
	for _, item := range items {
		for k := range item.Labels {
			if strings.TrimSpace(k) != "" {
				return true
			}
		}
	}
	return false
}

func filterByLabels(items []NamedItem, filters map[string]string) []NamedItem {
	if len(filters) == 0 {
		return items
	}
	result := make([]NamedItem, 0, len(items))
	for _, item := range items {
		match := true
		for k, v := range filters {
			if item.Labels[k] != v {
				match = false
				break
			}
		}
		if match {
			result = append(result, item)
		}
	}
	return result
}

func sortByLabelKey(items []NamedItem, key string) []NamedItem {
	sorted := make([]NamedItem, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(i, j int) bool {
		vi, oki := sorted[i].Labels[key]
		vj, okj := sorted[j].Labels[key]
		if !oki && !okj {
			return false
		}
		if !oki {
			return false
		}
		if !okj {
			return true
		}
		return vi < vj
	})
	return sorted
}

// parseLabelArgs extracts --label key=value and --sort-label key from args.
// Returns the remaining args, label filters, sort-label key, and an error
// if a --label value is missing "=" or --sort-label has no following token.
func parseLabelArgs(args []string) (remaining []string, labels map[string]string, sortKey string, err error) {
	labels = map[string]string{}
	for i := 0; i < len(args); i++ {
		if args[i] == "--label" {
			if i+1 >= len(args) {
				return nil, nil, "", fmt.Errorf("--label requires a key=value argument")
			}
			i++
			if k, v, ok := strings.Cut(args[i], "="); ok {
				if prev, exists := labels[k]; exists && prev != v {
					return nil, nil, "", fmt.Errorf("conflicting --label filters for %q", k)
				}
				labels[k] = v
			} else {
				return nil, nil, "", fmt.Errorf("--label value %q must contain '='", args[i])
			}
		} else if args[i] == "--sort-label" {
			if i+1 >= len(args) {
				return nil, nil, "", fmt.Errorf("--sort-label requires a key argument")
			}
			i++
			sortKey = args[i]
		} else {
			remaining = append(remaining, args[i])
		}
	}
	return remaining, labels, sortKey, nil
}

// mergeLabels combines scope label filters with per-command label filters.
// Returns an error if a command-level label conflicts with a scope label.
func mergeLabels(scope, cmd map[string]string) (map[string]string, error) {
	if len(scope) == 0 && len(cmd) == 0 {
		return nil, nil
	}
	merged := make(map[string]string, len(scope)+len(cmd))
	for k, v := range scope {
		merged[k] = v
	}
	for k, v := range cmd {
		if prev, exists := merged[k]; exists && prev != v {
			return nil, fmt.Errorf("--label %s=%s conflicts with scope label %s=%s", k, v, k, prev)
		}
		merged[k] = v
	}
	return merged, nil
}

func printResourceTable(w io.Writer, col1, col2, col3 string, items []NamedItem) error {
	fmt.Fprintf(os.Stderr, "%d items\n", len(items))
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\n", col1, col2, col3)
	for _, item := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", item.Name, item.Status, item.ID)
	}
	return tw.Flush()
}
