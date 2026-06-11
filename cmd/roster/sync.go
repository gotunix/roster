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

		for _, endpoint := range endpoints {
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

		fmt.Fprintln(os.Stderr, ui.SuccessMsg("Successfully synced %d hosts from NetBox", totalSynced))
	},
}

func init() {
	syncNetboxCmd.Flags().StringVarP(&syncFilter, "filter", "f", "", "Filter query parameters (e.g. status=active,role=linux)")
	syncCmd.AddCommand(syncNetboxCmd)
	rootCmd.AddCommand(syncCmd)
}
