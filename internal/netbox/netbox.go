// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The Roster Authors
// =============================================================================================== //
//                                                                                                 //
//                   /$$$$$$$                        /$$                                           //
//                  | $$__  $$                      | $$                                           //
//                  | $$  \ $$  /$$$$$$   /$$$$$$$ /$$$$$$    /$$$$$$   /$$$$$$                    //
//                  | $$$$$$$/ /$$__  $$ /$$_____/|_  $$_/   /$$__  $$ /$$__  $$                   //
//                  | $$__  $$| $$  \ $$|  $$$$$$   | $$    | $$$$$$$$| $$  \__/                   //
//                  | $$  \ $$| $$  | $$ \____  $$  | $$ /$$| $$_____/| $$                         //
//                  | $$  | $$|  $$$$$$/ /$$$$$$$/  |  $$$$/|  $$$$$$$| $$                         //
//                  |__/  |__/ \______/ |_______/    \___/   \_______/|__/                         //
//                                                                                                 //
// =============================================================================================== //
//              This program is free software: you can redistribute it and/or modify               //
//              it under the terms of the GNU General Public License as published by               //
//              the Free Software Foundation, either version 3 of the License, or                  //
//              (at your option) any later version.                                                //
//                                                                                                 //
//              This program is distributed in the hope that it will be useful,                    //
//              but WITHOUT ANY WARRANTY; without even the implied warranty of                     //
//              MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the                      //
//              GNU General Public License for more details.                                       //
//                                                                                                 //
//              You should have received a copy of the GNU General Public License                  //
//              along with this program.  If not, see <https://www.gnu.org/licenses/>.             //
// =============================================================================================== //

package netbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"github.com/charmbracelet/lipgloss"
	"gotunix.net/roster/internal/models"
	"gotunix.net/roster/internal/store"
	"gotunix.net/roster/internal/ui"
)

// LogWriter is where NetBox synchronization logs are written (defaults to os.Stderr)
var LogWriter io.Writer = os.Stderr

type NetBoxResponse struct {
	Count   int            `json:"count"`
	Next    *string        `json:"next"`
	Results []NetBoxObject `json:"results"`
}

type NetBoxObject struct {
	ID           int                    `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	LocalContext map[string]interface{} `json:"local_context_data"`
	PrimaryIP    *struct {
		Address string `json:"address"`
	} `json:"primary_ip"`
	Tags []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	} `json:"tags"`
	Status     interface{} `json:"status"`
	Site       interface{} `json:"site"`
	DeviceRole interface{} `json:"device_role"`
	Role       interface{} `json:"role"`
	Platform   interface{} `json:"platform"`
	DeviceType interface{} `json:"device_type"`
	Cluster    interface{} `json:"cluster"`
}

// SyncOptions controls which NetBox resources to sync
type SyncOptions struct {
	SyncConfigContexts bool
	SyncHosts          bool
	SyncInterfaces     bool
	SyncIPs            bool
	SyncVMDisks        bool
}

// DefaultSyncOptions returns options with all sync features enabled
func DefaultSyncOptions() SyncOptions {
	return SyncOptions{
		SyncConfigContexts: true,
		SyncHosts:          true,
		SyncInterfaces:     true,
		SyncIPs:            true,
		SyncVMDisks:        true,
	}
}

func Sync(baseURL, token, dir, filter string, opts SyncOptions) error {
	baseURL = strings.TrimSuffix(baseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/api")

	fmt.Fprintln(LogWriter, ui.BoldStyle.Foreground(ui.Cyan).Render("🔄 Syncing from NetBox: "+baseURL))

	s, err := NewInMemoryStore(dir)
	if err != nil {
		return fmt.Errorf("failed to initialize batch store: %v", err)
	}

	// 1. Sync Config Contexts (Groups, optional)
	if opts.SyncConfigContexts {
		if err := syncConfigContexts(baseURL, token, s); err != nil {
			return err
		}
	}

	// 2. Sync Hosts (Devices & VMs, optional)
	var deviceIDs, vmIDs map[int]string
	totalSynced := 0
	if opts.SyncHosts {
		fmt.Fprintln(LogWriter, ui.BoldStyle.Foreground(ui.Cyan).Render("🖥  Syncing Hosts..."))
		deviceIDs, vmIDs, totalSynced, err = syncHosts(baseURL, token, s, filter)
		if err != nil {
			return err
		}
	}

	// 3. Sync Interfaces (optional)
	if opts.SyncInterfaces {
		fmt.Fprintln(LogWriter, ui.BoldStyle.Foreground(ui.Cyan).Render("🔌 Syncing Interfaces..."))
		intfToHost, intfToName, isVMIntf, err := syncInterfaces(baseURL, token, s, deviceIDs, vmIDs)
		if err != nil {
			return err
		}

		// 4. Sync IP Addresses (optional, depends on interfaces)
		if opts.SyncIPs {
			fmt.Fprintln(LogWriter, ui.BoldStyle.Foreground(ui.Cyan).Render("🌐 Syncing IP Addresses..."))
			if err := syncIPAddresses(baseURL, token, s, intfToHost, intfToName, isVMIntf); err != nil {
				return err
			}
		}
	}

	// 5. Sync VM Disks (optional)
	if opts.SyncVMDisks && len(vmIDs) > 0 {
		fmt.Fprintln(LogWriter, ui.BoldStyle.Foreground(ui.Cyan).Render("💾 Syncing VM Disks..."))
		if err := syncVMDisks(baseURL, token, s, vmIDs); err != nil {
			return err
		}
	}

	// Flush all changes to disk
	fmt.Fprintln(LogWriter, ui.BoldStyle.Foreground(ui.Cyan).Render("💾 Saving changes to disk..."))
	if err := s.Flush(); err != nil {
		return fmt.Errorf("failed to save changes to disk: %v", err)
	}

	newCount := len(s.NewHosts)
	updatedCount := len(s.UpdatedHosts)
	parts := []string{}
	if newCount > 0 {
		parts = append(parts, fmt.Sprintf("%d new", newCount))
	}
	if updatedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d updated", updatedCount))
	}
	summary := ""
	if len(parts) > 0 {
		summary = " (" + strings.Join(parts, ", ") + ")"
	}
	fmt.Fprintln(LogWriter, ui.SuccessMsg("Successfully synced %d hosts from NetBox%s", totalSynced, summary))
	return nil
}

func syncConfigContexts(baseURL, token string, s *InMemoryStore) error {
	fmt.Fprintln(LogWriter, ui.BoldStyle.Foreground(ui.Cyan).Render("📦 Syncing Group Config Contexts..."))
	contextURL, _ := url.Parse(baseURL)
	contextURL.Path = strings.TrimSuffix(contextURL.Path, "/") + "/api/extras/config-contexts/"

	cURL := contextURL.String()
	for cURL != "" {
		resp, err := authenticatedGet(cURL, token)
		if err != nil {
			return fmt.Errorf("context request failed: %v", err)
		}
		if resp.StatusCode == http.StatusNotFound {
			_ = resp.Body.Close()
			fmt.Fprintln(LogWriter, ui.WarningMsg("Group config contexts endpoint not found (404), skipping..."))
			cURL = ""
			continue
		}
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			return fmt.Errorf("context request failed: API returned status %s", resp.Status)
		}

		var ctxResp struct {
			Next    *string `json:"next"`
			Results []struct {
				Name         string                 `json:"name"`
				Data         map[string]interface{} `json:"data"`
				Roles        []interface{}          `json:"roles"`
				DeviceGroups []interface{}          `json:"device_groups"`
				Tags         []interface{}          `json:"tags"`
				Platforms    []interface{}          `json:"platforms"`
				Sites        []interface{}          `json:"sites"`
			} `json:"results"`
		}
		err = json.NewDecoder(resp.Body).Decode(&ctxResp)
		_ = resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to decode contexts: %v", err)
		}

		for i := range ctxResp.Results {
			ctx := &ctxResp.Results[i]
			targetGroups := make(map[string]bool)
			for _, slug := range extractSlugs(ctx.Roles) {
				targetGroups[sanitizeName(slug)] = true
			}
			for _, slug := range extractSlugs(ctx.DeviceGroups) {
				targetGroups[sanitizeName(slug)] = true
			}
			for _, slug := range extractSlugs(ctx.Tags) {
				targetGroups[sanitizeName(slug)] = true
			}
			for _, slug := range extractSlugs(ctx.Platforms) {
				targetGroups[sanitizeName(slug)] = true
			}
			for _, slug := range extractSlugs(ctx.Sites) {
				targetGroups[sanitizeName(slug)] = true
			}

			if len(targetGroups) == 0 {
				continue
			}

			fmt.Fprintln(LogWriter, ui.DescriptionStyle.Render(fmt.Sprintf("  • Context: %s (mapping to %d groups)", ctx.Name, len(targetGroups))))

			for gName := range targetGroups {
				s.EnsureGroupExists(gName)
				s.MergeGroupVars(gName, ctx.Data)
			}
		}
		if ctxResp.Next != nil {
			cURL = *ctxResp.Next
		} else {
			cURL = ""
		}
	}
	return nil
}

func syncHosts(baseURL, token string, s *InMemoryStore, filter string) (map[int]string, map[int]string, int, error) {
	endpoints := []string{"/api/dcim/devices/", "/api/virtualization/virtual-machines/"}
	totalSynced := 0
	deviceIDs := make(map[int]string)
	vmIDs := make(map[int]string)

	for _, endpoint := range endpoints {
		isVM := strings.Contains(endpoint, "virtual-machines")
		apiURL, err := url.Parse(baseURL)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("invalid baseURL: %v", err)
		}
		apiURL.Path = strings.TrimSuffix(apiURL.Path, "/") + endpoint

		q := apiURL.Query()
		q.Add("exclude", "config_context")
		if filter != "" {
			filters := strings.Split(filter, ",")
			for _, f := range filters {
				parts := strings.SplitN(f, "=", 2)
				if len(parts) == 2 {
					q.Add(parts[0], parts[1])
				}
			}
		}
		apiURL.RawQuery = q.Encode()
		currentURL := apiURL.String()

		for currentURL != "" {
			resp, err := authenticatedGet(currentURL, token)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("host request failed: %v", err)
			}

			if resp.StatusCode != http.StatusOK {
				_ = resp.Body.Close()
				return nil, nil, 0, fmt.Errorf("API returned status %s for %s", resp.Status, currentURL)
			}

			var nbResp NetBoxResponse
			err = json.NewDecoder(resp.Body).Decode(&nbResp)
			_ = resp.Body.Close()
			if err != nil {
				return nil, nil, 0, fmt.Errorf("failed to decode response: %v", err)
			}

			for i := range nbResp.Results {
				obj := &nbResp.Results[i]
				name := obj.Name
				if name == "" {
					continue
				}

				if isVM {
					vmIDs[obj.ID] = name
				} else {
					deviceIDs[obj.ID] = name
				}

				s.AddHostToMain(name)

				if obj.PrimaryIP != nil && obj.PrimaryIP.Address != "" {
					ip := strings.Split(obj.PrimaryIP.Address, "/")[0]
					s.SetHostVar(name, "ansible_host", ip)
				}

				if obj.Description != "" {
					s.SetHostVar(name, "description", obj.Description)
				}

				if len(obj.LocalContext) > 0 {
					s.MergeHostVars(name, obj.LocalContext)
				}

				netboxVars := make(map[string]interface{})
				if v := extractVal(obj.Status); v != nil {
					netboxVars["status"] = v
				}
				if v := extractVal(obj.Site); v != nil {
					netboxVars["site"] = v
				}
				if v := extractVal(obj.Platform); v != nil {
					netboxVars["platform"] = v
				}
				if v := extractVal(obj.DeviceType); v != nil {
					netboxVars["device_type"] = v
				}
				if v := extractVal(obj.Cluster); v != nil {
					netboxVars["cluster"] = v
				}

				role := extractVal(obj.DeviceRole)
				if role == nil {
					role = extractVal(obj.Role)
				}
				if role != nil {
					netboxVars["role"] = role
				}

				if len(netboxVars) > 0 {
					s.MergeHostVars(name, map[string]interface{}{"netbox": netboxVars})
				}

			// Collect tag slugs for group membership sync
			var tagSlugs []string
			for _, tag := range obj.Tags {
				if tag.Slug != "" {
					tagSlugs = append(tagSlugs, sanitizeName(tag.Slug))
				}
			}

			// Sync group memberships to match NetBox tags (adds missing, removes stale)
			s.SyncHostGroupMembership(name, tagSlugs)

			// Assign to a status group (e.g. status_active, status_offline)
			// so Ansible patterns like hosts: status_active or hosts: all:!status_offline work
			if statusVal := extractVal(obj.Status); statusVal != nil {
				statusGroup := fmt.Sprintf("status_%v", sanitizeName(fmt.Sprintf("%v", statusVal)))
				s.EnsureGroupExists(statusGroup)
				s.AssignHostToGroup(name, statusGroup)
			}

			// Assign to a platform group (e.g. platform_linux, platform_windows)
			// If platform changed, remove from old group before adding to new one
			var oldPlatGroup string
			for gName, g := range s.Inventory.Groups {
				if strings.HasPrefix(gName, "platform_") {
					for _, h := range g.Hosts {
						if h == name {
							oldPlatGroup = gName
							break
						}
					}
				}
				if oldPlatGroup != "" {
					break
				}
			}
			platGroup := ""
			if platVal := extractVal(obj.Platform); platVal != nil {
				platGroup = fmt.Sprintf("platform_%v", sanitizeName(fmt.Sprintf("%v", platVal)))
			}
			if oldPlatGroup != platGroup {
				if oldPlatGroup != "" {
					s.RemoveHostFromGroup(name, oldPlatGroup)
				}
				if platGroup != "" {
					s.EnsureGroupExists(platGroup)
					s.AssignHostToGroup(name, platGroup)
					s.NetboxManagedGroups[platGroup] = true
					s.SeenNetboxGroups[platGroup] = true
					if g, ok := s.Inventory.Groups[platGroup]; ok {
						if g.Vars == nil {
							g.Vars = make(map[string]interface{})
						}
						g.Vars["roster_netbox_managed"] = true
					}
				}
			}

			// Print host status immediately — skip unchanged hosts entirely
				if s.NewHosts[name] {
					fmt.Fprintln(LogWriter, lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true).Render("  [NEW]     "+name))
				} else if s.UpdatedHosts[name] {
					fmt.Fprintln(LogWriter, lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387")).Bold(true).Render("  [UPDATED] "+name))
				}

				totalSynced++
			}
			if nbResp.Next != nil {
				currentURL = *nbResp.Next
			} else {
				currentURL = ""
			}
		}
	}

	// Remove stale group memberships now that all hosts' tags are known
	s.CleanupStaleGroupMemberships()

	// Remove orphaned NetBox-managed groups (deleted tags in NetBox)
	s.RemoveOrphanedGroups()

	return deviceIDs, vmIDs, totalSynced, nil
}

func syncInterfaces(baseURL, token string, s *InMemoryStore, deviceIDs, vmIDs map[int]string) (map[int]string, map[int]string, map[int]bool, error) {
	intfToHost := make(map[int]string)
	intfToName := make(map[int]string)
	isVMIntf := make(map[int]bool)

	interfaceEndpoints := []struct {
		path string
		ids  map[int]string
		isVM bool
	}{
		{"/api/dcim/interfaces/", deviceIDs, false},
		{"/api/virtualization/interfaces/", vmIDs, true},
	}

	for _, ie := range interfaceEndpoints {
		if len(ie.ids) == 0 {
			continue
		}

		apiURL, _ := url.Parse(baseURL)
		apiURL.Path = strings.TrimSuffix(apiURL.Path, "/") + ie.path
		currentURL := apiURL.String()

		for currentURL != "" {
			resp, err := authenticatedGet(currentURL, token)
			if err != nil {
				return nil, nil, nil, err
			}
			if resp.StatusCode == http.StatusNotFound {
				_ = resp.Body.Close()
				fmt.Fprintln(LogWriter, ui.WarningMsg("Interfaces endpoint not found (404) for %s, skipping...", ie.path))
				currentURL = ""
				continue
			}
			if resp.StatusCode != http.StatusOK {
				_ = resp.Body.Close()
				return nil, nil, nil, fmt.Errorf("interfaces request failed: API returned status %s", resp.Status)
			}

			var intfResp struct {
				Next    *string `json:"next"`
				Results []struct {
					ID      int              `json:"id"`
					Name    string           `json:"name"`
					MAC     string           `json:"mac_address"`
					MTU     int              `json:"mtu"`
					Enabled bool             `json:"enabled"`
					Device  struct{ ID int } `json:"device"`
					VM      struct{ ID int } `json:"virtual_machine"`
				} `json:"results"`
			}
			err = json.NewDecoder(resp.Body).Decode(&intfResp)
			_ = resp.Body.Close()
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to decode interfaces response: %v", err)
			}

			for _, intf := range intfResp.Results {
				hostID := intf.Device.ID
				if hostID == 0 {
					hostID = intf.VM.ID
				}

				hName, ok := ie.ids[hostID]
				if !ok {
					continue
				}

				intfToHost[intf.ID] = hName
				intfToName[intf.ID] = intf.Name
				isVMIntf[intf.ID] = ie.isVM

				intfData := map[string]interface{}{
					"name":    intf.Name,
					"enabled": intf.Enabled,
				}
				if intf.MAC != "" {
					intfData["mac"] = intf.MAC
				}
				if intf.MTU != 0 {
					intfData["mtu"] = intf.MTU
				}

				interfaceMap := map[string]interface{}{
					"interfaces": map[string]interface{}{
						intf.Name: intfData,
					},
				}
				s.MergeHostVars(hName, map[string]interface{}{"netbox": interfaceMap})
			}
			if intfResp.Next != nil {
				currentURL = *intfResp.Next
			} else {
				currentURL = ""
			}
		}
	}
	return intfToHost, intfToName, isVMIntf, nil
}

func syncIPAddresses(baseURL, token string, s *InMemoryStore, intfToHost, intfToName map[int]string, isVMIntf map[int]bool) error {
	ipURL, _ := url.Parse(baseURL)
	ipURL.Path = strings.TrimSuffix(ipURL.Path, "/") + "/api/ipam/ip-addresses/"
	currentIPURL := ipURL.String()

	for currentIPURL != "" {
		resp, err := authenticatedGet(currentIPURL, token)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusNotFound {
			_ = resp.Body.Close()
			fmt.Fprintln(LogWriter, ui.WarningMsg("IP addresses endpoint not found (404), skipping..."))
			currentIPURL = ""
			continue
		}
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			return fmt.Errorf("IP addresses request failed: API returned status %s", resp.Status)
		}

		var ipResp struct {
			Next    *string `json:"next"`
			Results []struct {
				Address            string                 `json:"address"`
				Status             struct{ Value string } `json:"status"`
				AssignedObjectID   int                    `json:"assigned_object_id"`
				AssignedObjectType string                 `json:"assigned_object_type"`
			} `json:"results"`
		}
		err = json.NewDecoder(resp.Body).Decode(&ipResp)
		_ = resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to decode ip response: %v", err)
		}

		for _, ip := range ipResp.Results {
			if ip.AssignedObjectID == 0 {
				continue
			}

			hName, ok := intfToHost[ip.AssignedObjectID]
			if !ok {
				continue
			}

			isVM := strings.Contains(ip.AssignedObjectType, "vminterface")
			if isVM != isVMIntf[ip.AssignedObjectID] {
				continue
			}

			intfName := intfToName[ip.AssignedObjectID]
			ipMap := map[string]interface{}{
				"interfaces": map[string]interface{}{
					intfName: map[string]interface{}{
						"ips": map[string]interface{}{
							ip.Address: map[string]interface{}{
								"status": ip.Status.Value,
							},
						},
					},
				},
			}
			s.MergeHostVars(hName, map[string]interface{}{"netbox": ipMap})
		}
		if ipResp.Next != nil {
			currentIPURL = *ipResp.Next
		} else {
			currentIPURL = ""
		}
	}
	return nil
}

func syncVMDisks(baseURL, token string, s *InMemoryStore, vmIDs map[int]string) error {
	apiURL, _ := url.Parse(baseURL)
	apiURL.Path = strings.TrimSuffix(apiURL.Path, "/") + "/api/virtualization/virtual-disks/"
	currentURL := apiURL.String()

	for currentURL != "" {
		resp, err := authenticatedGet(currentURL, token)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusNotFound {
			_ = resp.Body.Close()
			fmt.Fprintln(LogWriter, ui.WarningMsg("Virtualization disks endpoint not found (404), skipping..."))
			currentURL = ""
			continue
		}
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			return fmt.Errorf("disks request failed: API returned status %s", resp.Status)
		}

		var diskResp struct {
			Next    *string `json:"next"`
			Results []struct {
				Name string           `json:"name"`
				Size int              `json:"size"`
				VM   struct{ ID int } `json:"virtual_machine"`
			} `json:"results"`
		}
		err = json.NewDecoder(resp.Body).Decode(&diskResp)
		_ = resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to decode disk response: %v", err)
		}

		for _, disk := range diskResp.Results {
			hName, ok := vmIDs[disk.VM.ID]
			if !ok {
				continue
			}
			diskData := map[string]interface{}{
				"name": disk.Name,
				"size": disk.Size,
			}
			diskMap := map[string]interface{}{
				"disks": map[string]interface{}{
					disk.Name: diskData,
				},
			}
			s.MergeHostVars(hName, map[string]interface{}{"netbox": diskMap})
		}
		if diskResp.Next != nil {
			currentURL = *diskResp.Next
		} else {
			currentURL = ""
		}
	}
	return nil
}

func authenticatedGet(url, token string) (*http.Response, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Token "+token)
	req.Header.Add("Accept", "application/json")
	return client.Do(req)
}

// sanitizeName replaces characters that Ansible cannot handle in group names
func sanitizeName(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

func extractSlugs(items []interface{}) []string {
	var slugs []string
	for _, item := range items {
		switch v := item.(type) {
		case string:
			slugs = append(slugs, v)
		case map[string]interface{}:
			if slug, ok := v["slug"].(string); ok {
				slugs = append(slugs, slug)
			}
		}
	}
	return slugs
}

func extractVal(item interface{}) interface{} {
	if item == nil {
		return nil
	}
	if m, ok := item.(map[string]interface{}); ok {
		if slug, ok := m["slug"].(string); ok {
			return slug
		}
		if val, ok := m["value"].(string); ok {
			return val
		}
	}
	return item
}

// InMemoryStore tracks changes in memory and flushes them to disk at the end of the synchronization
type InMemoryStore struct {
	Dir              string
	Inventory        *models.Inventory
	GroupVars        map[string]map[string]interface{}
	HostVars         map[string]map[string]interface{}
	MainYaml         map[string]interface{}
	NewHosts         map[string]bool
	UpdatedHosts     map[string]bool
	SyncedHosts      map[string]bool
	NetboxManagedGroups  map[string]bool   // Groups created/managed by NetBox tags (persisted)
	SeenNetboxGroups     map[string]bool   // Groups seen during current sync run
	HostTags             map[string][]string // Tags per host from current sync
}

func NewInMemoryStore(dir string) (*InMemoryStore, error) {
	inv, err := store.LoadInventory(dir)
	if err != nil {
		return nil, err
	}

	s := &InMemoryStore{
		Dir:                 dir,
		Inventory:           inv,
		GroupVars:           make(map[string]map[string]interface{}),
		HostVars:            make(map[string]map[string]interface{}),
		NewHosts:            make(map[string]bool),
		UpdatedHosts:        make(map[string]bool),
		SyncedHosts:         make(map[string]bool),
		NetboxManagedGroups:  make(map[string]bool),
		SeenNetboxGroups:     make(map[string]bool),
		HostTags:             make(map[string][]string),
	}

	// Load main.yaml if it exists
	mainPath := filepath.Join(dir, "main.yaml")
	if data, err := os.ReadFile(mainPath); err == nil {
		var raw map[string]interface{}
		if err := yaml.Unmarshal(data, &raw); err == nil {
			s.MainYaml = raw
		}
	}
	if s.MainYaml == nil {
		s.MainYaml = map[string]interface{}{
			"all": map[string]interface{}{
				"hosts": make(map[string]interface{}),
			},
		}
	}

	// Load persisted NetBox-managed groups from main.yaml
	if rawGroups, ok := s.MainYaml["netbox_managed_groups"].([]interface{}); ok {
		for _, g := range rawGroups {
			if gName, ok := g.(string); ok {
				s.NetboxManagedGroups[gName] = true
			}
		}
	}

	// Also load from group var markers (handles first-run migration and
	// groups whose tags were removed from all hosts)
	for gName, g := range inv.Groups {
		if g.Vars != nil {
			if managed, ok := g.Vars["roster_netbox_managed"].(bool); ok && managed {
				s.NetboxManagedGroups[gName] = true
			}
		}
	}

	return s, nil
}

func (s *InMemoryStore) AddHostToMain(hostname string) {
	all, ok := s.MainYaml["all"].(map[string]interface{})
	if !ok || all == nil {
		all = make(map[string]interface{})
		s.MainYaml["all"] = all
	}
	hosts, ok := all["hosts"].(map[string]interface{})
	if !ok || hosts == nil {
		hosts = make(map[string]interface{})
		all["hosts"] = hosts
	}
	hosts[hostname] = make(map[string]interface{})

	// Ensure in models.Inventory
	if _, ok := s.Inventory.Hosts[hostname]; !ok {
		s.Inventory.Hosts[hostname] = &models.Host{Name: hostname}
		s.NewHosts[hostname] = true
	}
	s.SyncedHosts[hostname] = true
}

func (s *InMemoryStore) MergeGroupVars(groupName string, newVars map[string]interface{}) {
	vars, ok := s.GroupVars[groupName]
	if !ok {
		var err error
		vars, err = store.GetGroupVars(s.Dir, groupName)
		if err != nil || vars == nil {
			vars = make(map[string]interface{})
		}
		s.GroupVars[groupName] = vars
	}
	MergeNetboxVars(vars, newVars)

	// Update models.Inventory Group
	g, ok := s.Inventory.Groups[groupName]
	if !ok {
		g = &models.Group{Name: groupName}
		s.Inventory.Groups[groupName] = g
	}
	if g.Vars == nil {
		g.Vars = make(map[string]interface{})
	}
	MergeNetboxVars(g.Vars, newVars)
}

func (s *InMemoryStore) MergeHostVars(hostname string, newVars map[string]interface{}) {
	vars, ok := s.HostVars[hostname]
	if !ok {
		var err error
		vars, err = store.GetHostVars(s.Dir, hostname)
		if err != nil || vars == nil {
			vars = make(map[string]interface{})
		}
		s.HostVars[hostname] = vars
	}
	if MergeNetboxVars(vars, newVars) {
		s.UpdatedHosts[hostname] = true
	}

	// Update models.Inventory Host
	h, ok := s.Inventory.Hosts[hostname]
	if !ok {
		h = &models.Host{Name: hostname}
		s.Inventory.Hosts[hostname] = h
	}
	if h.Vars == nil {
		h.Vars = make(map[string]interface{})
	}
	MergeNetboxVars(h.Vars, newVars)
}

func (s *InMemoryStore) SetHostVar(hostname string, key string, value interface{}) {
	s.MergeHostVars(hostname, map[string]interface{}{key: value})
}

func (s *InMemoryStore) EnsureGroupExists(groupName string) {
	if _, ok := s.Inventory.Groups[groupName]; !ok {
		s.Inventory.Groups[groupName] = &models.Group{Name: groupName}
	}
}

func (s *InMemoryStore) AssignHostToGroup(hostname string, groupName string) {
	s.EnsureGroupExists(groupName)
	g := s.Inventory.Groups[groupName]

	found := false
	for _, h := range g.Hosts {
		if h == hostname {
			found = true
			break
		}
	}
	if !found {
		g.Hosts = append(g.Hosts, hostname)
		s.UpdatedHosts[hostname] = true
	}
}

// RemoveHostFromGroup removes a host from a group
func (s *InMemoryStore) RemoveHostFromGroup(hostname string, groupName string) {
	g, ok := s.Inventory.Groups[groupName]
	if !ok {
		return
	}
	newHosts := make([]string, 0, len(g.Hosts))
	for _, h := range g.Hosts {
		if h != hostname {
			newHosts = append(newHosts, h)
		}
	}
	if len(newHosts) != len(g.Hosts) {
		g.Hosts = newHosts
		s.UpdatedHosts[hostname] = true
	}
}

// SyncHostGroupMembership records a host's NetBox tags and adds the host
// to the corresponding groups. Removal of stale memberships is done later
// by CleanupStaleGroupMemberships after all hosts have been processed.
func (s *InMemoryStore) SyncHostGroupMembership(hostname string, netboxTags []string) {
	// Record this host's tags for later cleanup
	s.HostTags[hostname] = netboxTags

	for _, tag := range netboxTags {
		if tag == "" {
			continue
		}
		s.NetboxManagedGroups[tag] = true
		s.SeenNetboxGroups[tag] = true
		s.AssignHostToGroup(hostname, tag)
		// Mark the group as NetBox-managed so cleanup works even
		// if the tag is removed from all hosts in a future sync
		if g, ok := s.Inventory.Groups[tag]; ok {
			if g.Vars == nil {
				g.Vars = make(map[string]interface{})
			}
			g.Vars["roster_netbox_managed"] = true
		}
	}
}

// CleanupStaleGroupMemberships removes hosts from groups whose NetBox tags
// are no longer present. It uses SeenNetboxGroups (all tags from all hosts
// in this sync) and NetboxManagedGroups (persisted from previous syncs) to
// determine which groups are NetBox-managed, preserving local-only groups.
func (s *InMemoryStore) CleanupStaleGroupMemberships() {
	// knownNetboxGroups = groups seen in this sync + groups from persistence
	knownNetboxGroups := make(map[string]bool)
	for g := range s.SeenNetboxGroups {
		knownNetboxGroups[g] = true
	}
	for g := range s.NetboxManagedGroups {
		knownNetboxGroups[g] = true
	}

	for hostname, tags := range s.HostTags {
		tagSet := make(map[string]bool)
		for _, t := range tags {
			if t != "" {
				tagSet[t] = true
			}
		}

		for gName, g := range s.Inventory.Groups {
			if gName == "all" || strings.HasPrefix(gName, "status_") || strings.HasPrefix(gName, "platform_") {
				continue
			}
			if !knownNetboxGroups[gName] {
				continue // local group, don't touch
			}
			if tagSet[gName] {
				continue // host still has this tag
			}
			// Host is in a NetBox-managed group but no longer has the tag
			for _, h := range g.Hosts {
				if h == hostname {
					s.RemoveHostFromGroup(hostname, gName)
					fmt.Fprintln(LogWriter, lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387")).Bold(true).Render("  [UPDATED] "+hostname+" (removed from group <"+gName+">)"))
					break
				}
			}
		}
	}
}

// RemoveOrphanedGroups deletes group definition files for NetBox-managed
// groups that had all hosts removed (e.g. tag deleted in NetBox). Group
// variables (group_vars/<name>.yaml) are preserved for manual reference.
func (s *InMemoryStore) RemoveOrphanedGroups() {
	knownNetboxGroups := make(map[string]bool)
	for g := range s.SeenNetboxGroups {
		knownNetboxGroups[g] = true
	}
	for g := range s.NetboxManagedGroups {
		knownNetboxGroups[g] = true
	}

	for gName, g := range s.Inventory.Groups {
		if gName == "all" || strings.HasPrefix(gName, "status_") || strings.HasPrefix(gName, "platform_") {
			continue
		}
		if !knownNetboxGroups[gName] {
			continue // local group
		}
		if s.SeenNetboxGroups[gName] {
			continue // still active this sync
		}
		if len(g.Hosts) > 0 || len(g.Children) > 0 {
			continue // still has members
		}

		groupPath := filepath.Join(s.Dir, gName+".yaml")
		if err := os.Remove(groupPath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintln(LogWriter, lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387")).Render("  Warning: failed to remove orphaned group file "+groupPath+": "+err.Error()))
		}

		delete(s.Inventory.Groups, gName)
		delete(s.NetboxManagedGroups, gName)

		fmt.Fprintln(LogWriter, lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387")).Bold(true).Render("  [DELETED] group <"+gName+"> (tag no longer in NetBox)"))
	}
}

func (s *InMemoryStore) saveGroupClean(gName string, g *models.Group) error {
	// Read existing inline vars first so we don't accidentally write external variables inline
	path := filepath.Join(s.Dir, gName+".yaml")

	var inlineVars map[string]interface{}
	if data, err := os.ReadFile(path); err == nil {
		var raw map[string]interface{}
		if err := yaml.Unmarshal(data, &raw); err == nil {
			if gContent, ok := raw[gName].(map[string]interface{}); ok && gContent != nil {
				if v, ok := gContent["vars"].(map[string]interface{}); ok {
					inlineVars = v
				}
			}
		}
	}

	// Ensure roster_netbox_managed marker is preserved in saved group vars
	if g.Vars != nil {
		if managed, ok := g.Vars["roster_netbox_managed"].(bool); ok && managed {
			if inlineVars == nil {
				inlineVars = make(map[string]interface{})
			}
			inlineVars["roster_netbox_managed"] = true
		}
	}

	cloned := &models.Group{
		Name:     g.Name,
		Hosts:    g.Hosts,
		Children: g.Children,
		Vars:     inlineVars,
	}
	return store.SaveGroup(s.Dir, gName, cloned)
}

func (s *InMemoryStore) Flush() error {
	// Persist NetBox-managed groups that were actually seen in this sync
	var activeGroups []string
	for gName := range s.SeenNetboxGroups {
		activeGroups = append(activeGroups, gName)
	}
	if len(activeGroups) > 0 {
		sort.Strings(activeGroups)
		s.MainYaml["netbox_managed_groups"] = activeGroups
	} else {
		delete(s.MainYaml, "netbox_managed_groups")
	}

	// 1. Save main.yaml
	mainPath := filepath.Join(s.Dir, "main.yaml")
	bytes, err := yaml.Marshal(s.MainYaml)
	if err != nil {
		return fmt.Errorf("failed to marshal main.yaml: %v", err)
	}
	if err := os.WriteFile(mainPath, bytes, 0644); err != nil {
		return fmt.Errorf("failed to write main.yaml: %v", err)
	}

	// 2. Save group definitions
	for gName, g := range s.Inventory.Groups {
		if gName == "all" {
			continue
		}
		if err := s.saveGroupClean(gName, g); err != nil {
			return fmt.Errorf("failed to save group definition %s: %v", gName, err)
		}
	}

	// 3. Save group variables
	for gName, vars := range s.GroupVars {
		if err := store.SaveGroupVars(s.Dir, gName, vars); err != nil {
			return fmt.Errorf("failed to save group vars for %s: %v", gName, err)
		}
	}

	// 4. Save host variables
	for hName, vars := range s.HostVars {
		if err := store.SaveHostVars(s.Dir, hName, vars); err != nil {
			return fmt.Errorf("failed to save host vars for %s: %v", hName, err)
		}
	}

	return nil
}

// MergeNetboxVars merges NetBox-synced variables into local variables map without overwriting existing local values.
// The "netbox" nested structure is always deep-merged to update metadata.
func MergeNetboxVars(localVars map[string]interface{}, netboxVars map[string]interface{}) bool {
	netboxChanged := false
	for k, v := range netboxVars {
		if k == "netbox" {
			existingNetbox, ok := localVars["netbox"].(map[string]interface{})
			if !ok {
				existingNetbox = make(map[string]interface{})
				localVars["netbox"] = existingNetbox
			}
			newNetbox, ok := v.(map[string]interface{})
			if ok {
				if store.DeepMerge(existingNetbox, newNetbox) {
					netboxChanged = true
				}
			}
		} else {
			// Only write to the local file if the variable key does not exist yet.
			// This preserves all local user modifications (e.g. customized ansible_host or ansible_user).
			if _, exists := localVars[k]; !exists {
				localVars[k] = v
			}
		}
	}
	return netboxChanged
}

// cloneMap recursively deep-copies a map[string]interface{}
func cloneMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	res := make(map[string]interface{})
	for k, v := range m {
		if subMap, ok := v.(map[string]interface{}); ok {
			res[k] = cloneMap(subMap)
		} else {
			res[k] = v
		}
	}
	return res
}
