// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors
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

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gotunix.net/roster/internal/models"
	"gotunix.net/roster/internal/store"
	"gotunix.net/roster/internal/ui"
)

var (
	syncFilter string
)

type NetBoxResponse struct {
	Count    int             `json:"count"`
	Next     *string         `json:"next"`
	Results  []NetBoxObject  `json:"results"`
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

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync inventory from external sources",
}

var syncNetboxCmd = &cobra.Command{
	Use:   "netbox <url>",
	Short: "Sync hosts from NetBox API",
	Long:  `Fetch devices and VMs from NetBox and add them to the inventory. Requires NETBOX_TOKEN env var.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		baseURL := args[0]
		token := os.Getenv("NETBOX_TOKEN")
		if token == "" {
			fmt.Fprintln(os.Stderr, ui.ErrorMsg("NETBOX_TOKEN environment variable is not set"))
			return
		}

		dir := inventoryPaths[0]
		if err := store.LockInventory(dir); err != nil {
			fmt.Fprintln(os.Stderr, ui.ErrorMsg("%v", err))
			return
		}
		defer store.UnlockInventory()

		fmt.Fprintln(os.Stderr, ui.BoldStyle.Foreground(ui.Cyan).Render("🔄 Syncing from NetBox: "+baseURL))

		extractSlugs := func(items []interface{}) []string {
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

		extractVal := func(item interface{}) interface{} {
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

		// 1. Sync Config Contexts (Groups)
		fmt.Fprintln(os.Stderr, ui.BoldStyle.Foreground(ui.Cyan).Render("📦 Syncing Group Config Contexts..."))
		contextURL, _ := url.Parse(baseURL)
		contextURL.Path = strings.TrimSuffix(contextURL.Path, "/") + "/api/extras/config-contexts/"
		
		cURL := contextURL.String()
		for cURL != "" {
			client := &http.Client{Timeout: 30 * time.Second}
			req, _ := http.NewRequest("GET", cURL, nil)
			req.Header.Add("Authorization", "Token "+token)
			req.Header.Add("Accept", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("Context request failed: %v", err))
				break
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
			if err := json.NewDecoder(resp.Body).Decode(&ctxResp); err != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("Failed to decode contexts: %v", err))
				resp.Body.Close()
				break
			}
			resp.Body.Close()

			for _, ctx := range ctxResp.Results {
				// Map this context data to each assigned group type
				targetGroups := make(map[string]bool)
				for _, s := range extractSlugs(ctx.Roles) { targetGroups[s] = true }
				for _, s := range extractSlugs(ctx.DeviceGroups) { targetGroups[s] = true }
				for _, s := range extractSlugs(ctx.Tags) { targetGroups[s] = true }
				for _, s := range extractSlugs(ctx.Platforms) { targetGroups[s] = true }
				for _, s := range extractSlugs(ctx.Sites) { targetGroups[s] = true }

				if len(targetGroups) == 0 {
					continue
				}

				fmt.Fprintln(os.Stderr, ui.DescriptionStyle.Render(fmt.Sprintf("  • Context: %s (mapping to %d groups)", ctx.Name, len(targetGroups))))

				for gName := range targetGroups {
					// Ensure group exists
					inv, _ := store.LoadInventory(dir)
					if _, ok := inv.Groups[gName]; !ok {
						if err := store.SaveGroup(dir, gName, &models.Group{Name: gName}); err != nil {
							fmt.Fprintln(os.Stderr, ui.ErrorMsg("  ! Failed to create group %s: %v", gName, err))
							continue
						}
					}
					// Merge data
					if err := store.MergeGroupVars(dir, gName, ctx.Data); err != nil {
						fmt.Fprintln(os.Stderr, ui.ErrorMsg("  ! Merging context to group %s: %v", gName, err))
					}
				}
			}

			if ctxResp.Next != nil { cURL = *ctxResp.Next } else { cURL = "" }
		}

		// 2. Sync Hosts (Devices & VMs)
		fmt.Fprintln(os.Stderr, ui.BoldStyle.Foreground(ui.Cyan).Render("🖥  Syncing Hosts..."))
		endpoints := []string{"/api/dcim/devices/", "/api/virtualization/virtual-machines/"}
		totalSynced := 0

		// Keep track of synced IDs to map interfaces/disks later
		deviceIDs := make(map[int]string) // id -> name
		vmIDs := make(map[int]string)     // id -> name

		for _, endpoint := range endpoints {
			isVM := strings.Contains(endpoint, "virtual-machines")
			apiURL, err := url.Parse(baseURL)
			if err != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("Invalid baseURL: %v", err))
				return
			}
			apiURL.Path = strings.TrimSuffix(apiURL.Path, "/") + endpoint

			// Apply filters and exclude the merged config_context to get local_context_data instead
			q := apiURL.Query()
			q.Add("exclude", "config_context")
			if syncFilter != "" {
				filters := strings.Split(syncFilter, ",")
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
				client := &http.Client{Timeout: 30 * time.Second}
				req, err := http.NewRequest("GET", currentURL, nil)
				if err != nil {
					fmt.Fprintln(os.Stderr, ui.ErrorMsg("Failed to create request: %v", err))
					break
				}
				req.Header.Add("Authorization", "Token "+token)
				req.Header.Add("Accept", "application/json")

				resp, err := client.Do(req)
				if err != nil {
					fmt.Fprintln(os.Stderr, ui.ErrorMsg("Request failed: %v", err))
					break
				}

				if resp.StatusCode != http.StatusOK {
					fmt.Fprintln(os.Stderr, ui.ErrorMsg("API returned status %s for %s", resp.Status, currentURL))
					resp.Body.Close()
					break
				}

				var nbResp NetBoxResponse
				if err := json.NewDecoder(resp.Body).Decode(&nbResp); err != nil {
					fmt.Fprintln(os.Stderr, ui.ErrorMsg("Failed to decode response: %v", err))
					resp.Body.Close()
					break
				}
				resp.Body.Close()

				for _, obj := range nbResp.Results {
					name := obj.Name
					if name == "" {
						continue
					}

					if isVM {
						vmIDs[obj.ID] = name
					} else {
						deviceIDs[obj.ID] = name
					}

					// Add host
					if err := store.AddHostToMain(dir, name); err != nil {
						fmt.Fprintln(os.Stderr, ui.ErrorMsg("Adding host %s: %v", name, err))
						continue
					}

					// Map primary IP to ansible_host
					if obj.PrimaryIP != nil && obj.PrimaryIP.Address != "" {
						ip := strings.Split(obj.PrimaryIP.Address, "/")[0] // Strip CIDR
						if err := store.SetHostVar(dir, name, "ansible_host", ip); err != nil {
							fmt.Fprintln(os.Stderr, ui.ErrorMsg("Setting ansible_host for %s: %v", name, err))
						}
					}

					// Map description
					if obj.Description != "" {
						if err := store.SetHostVar(dir, name, "description", obj.Description); err != nil {
							fmt.Fprintln(os.Stderr, ui.ErrorMsg("Setting description for %s: %v", name, err))
						}
					}

					// Map local config context
					if len(obj.LocalContext) > 0 {
						if err := store.MergeHostVars(dir, name, obj.LocalContext); err != nil {
							fmt.Fprintln(os.Stderr, ui.ErrorMsg("Merging local_context for %s: %v", name, err))
						}
					}

					// Map NetBox Metadata
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

					// Role (DeviceRole for Devices, Role for VMs)
					role := extractVal(obj.DeviceRole)
					if role == nil {
						role = extractVal(obj.Role)
					}
					if role != nil {
						netboxVars["role"] = role
					}

					if len(netboxVars) > 0 {
						if err := store.MergeHostVars(dir, name, map[string]interface{}{"netbox": netboxVars}); err != nil {
							fmt.Fprintln(os.Stderr, ui.ErrorMsg("Merging netbox metadata for %s: %v", name, err))
						}
					}

					// Handle Tags -> Groups
					for _, tag := range obj.Tags {
						groupName := tag.Slug
						if groupName == "" {
							continue
						}

						// Ensure group exists
						inv, _ := store.LoadInventory(dir)
						if _, ok := inv.Groups[groupName]; !ok {
							if err := store.SaveGroup(dir, groupName, &models.Group{Name: groupName}); err != nil {
								fmt.Fprintln(os.Stderr, ui.ErrorMsg("Creating group %s from tag: %v", groupName, err))
								continue
							}
						}

						// Assign host to group
						if err := store.AssignHostToGroup(dir, name, groupName); err != nil {
							fmt.Fprintln(os.Stderr, ui.ErrorMsg("Assigning %s to group %s: %v", name, groupName, err))
						}
					}

					totalSynced++
				}

				// Move to next page
				if nbResp.Next != nil {
					currentURL = *nbResp.Next
				} else {
					currentURL = ""
				}
			}
		}

		// 3. Sync Interfaces
		fmt.Fprintln(os.Stderr, ui.BoldStyle.Foreground(ui.Cyan).Render("🔌 Syncing Interfaces..."))
		interfaceEndpoints := []struct {
			path string
			key  string
			ids  map[int]string
		}{
			{"/api/dcim/interfaces/", "device_id", deviceIDs},
			{"/api/virtualization/interfaces/", "virtual_machine_id", vmIDs},
		}

		for _, ie := range interfaceEndpoints {
			if len(ie.ids) == 0 {
				continue
			}

			apiURL, _ := url.Parse(baseURL)
			apiURL.Path = strings.TrimSuffix(apiURL.Path, "/") + ie.path
			currentURL := apiURL.String()

			for currentURL != "" {
				client := &http.Client{Timeout: 30 * time.Second}
				req, _ := http.NewRequest("GET", currentURL, nil)
				req.Header.Add("Authorization", "Token "+token)
				req.Header.Add("Accept", "application/json")

				resp, err := client.Do(req)
				if err != nil {
					break
				}

				var intfResp struct {
					Next    *string `json:"next"`
					Results []struct {
						Name       string      `json:"name"`
						MAC        string      `json:"mac_address"`
						MTU        int         `json:"mtu"`
						Enabled    bool        `json:"enabled"`
						Device     struct{ ID int } `json:"device"`
						VM         struct{ ID int } `json:"virtual_machine"`
					} `json:"results"`
				}
				json.NewDecoder(resp.Body).Decode(&intfResp)
				resp.Body.Close()

				for _, intf := range intfResp.Results {
					hostID := intf.Device.ID
					if hostID == 0 {
						hostID = intf.VM.ID
					}

					hName, ok := ie.ids[hostID]
					if !ok {
						continue
					}

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

					// Merge into netbox.interfaces (list)
					// Note: Since MergeHostVars handles maps, we'll store interfaces as a map keyed by name
					// for easier deep merging without duplication
					interfaceMap := map[string]interface{}{
						"interfaces": map[string]interface{}{
							intf.Name: intfData,
						},
					}
					store.MergeHostVars(dir, hName, map[string]interface{}{"netbox": interfaceMap})
				}

				if intfResp.Next != nil {
					currentURL = *intfResp.Next
				} else {
					currentURL = ""
				}
			}
		}

		// 4. Sync VM Disks
		if len(vmIDs) > 0 {
			fmt.Fprintln(os.Stderr, ui.BoldStyle.Foreground(ui.Cyan).Render("💾 Syncing VM Disks..."))
			apiURL, _ := url.Parse(baseURL)
			apiURL.Path = strings.TrimSuffix(apiURL.Path, "/") + "/api/virtualization/disks/"
			currentURL := apiURL.String()

			for currentURL != "" {
				client := &http.Client{Timeout: 30 * time.Second}
				req, _ := http.NewRequest("GET", currentURL, nil)
				req.Header.Add("Authorization", "Token "+token)
				req.Header.Add("Accept", "application/json")

				resp, err := client.Do(req)
				if err != nil {
					break
				}

				var diskResp struct {
					Next    *string `json:"next"`
					Results []struct {
						Name string      `json:"name"`
						Size int         `json:"size"`
						VM   struct{ ID int } `json:"virtual_machine"`
					} `json:"results"`
				}
				json.NewDecoder(resp.Body).Decode(&diskResp)
				resp.Body.Close()

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
					store.MergeHostVars(dir, hName, map[string]interface{}{"netbox": diskMap})
				}

				if diskResp.Next != nil {
					currentURL = *diskResp.Next
				} else {
					currentURL = ""
				}
			}
		}

		fmt.Fprintln(os.Stderr, ui.SuccessMsg("Successfully synced %d hosts from NetBox", totalSynced))
	},
}

func init() {
	syncNetboxCmd.Flags().StringVarP(&syncFilter, "filter", "f", "", "Filter query parameters (e.g. status=active,role=linux)")
	syncCmd.AddCommand(syncNetboxCmd)
	rootCmd.AddCommand(syncCmd)
}
