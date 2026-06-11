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
	Name      string `json:"name"`
	PrimaryIP *struct {
		Address string `json:"address"`
	} `json:"primary_ip"`
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

		endpoints := []string{"/api/dcim/devices/", "/api/virtualization/virtual-machines/"}
		totalSynced := 0

		for _, endpoint := range endpoints {
			apiURL, err := url.Parse(baseURL)
			if err != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("Invalid baseURL: %v", err))
				return
			}
			apiURL.Path = strings.TrimSuffix(apiURL.Path, "/") + endpoint
			
			// Apply filters
			q := apiURL.Query()
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

			client := &http.Client{Timeout: 30 * time.Second}
			req, err := http.NewRequest("GET", apiURL.String(), nil)
			if err != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("Failed to create request: %v", err))
				continue
			}
			req.Header.Add("Authorization", "Token "+token)
			req.Header.Add("Accept", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("Request failed: %v", err))
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("API returned status: %s", resp.Status))
				continue
			}

			var nbResp NetBoxResponse
			if err := json.NewDecoder(resp.Body).Decode(&nbResp); err != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("Failed to decode response: %v", err))
				continue
			}

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

				totalSynced++
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
