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

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gotunix.net/roster/internal/netbox"
	"gotunix.net/roster/internal/store"
	"gotunix.net/roster/internal/ui"
)

var (
	syncFilter         string
	syncConfigContexts bool
	syncHosts          bool
	syncInterfaces     bool
	syncIPs            bool
	syncVMDisks        bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync inventory from external sources",
}

var syncNetboxCmd = &cobra.Command{
	Use:   "netbox [url]",
	Short: "Sync hosts from NetBox API",
	Long:  `Fetch devices, VMs, interfaces, and contexts from NetBox and add them to the inventory. Uses NETBOX_TOKEN or value from roster.conf.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		confURL, confToken, confFilter := store.LoadRosterConf()

		baseURL := confURL
		if len(args) > 0 {
			baseURL = args[0]
		}
		if baseURL == "" {
			fmt.Fprintln(os.Stderr, ui.ErrorMsg("NetBox URL is required. Pass as argument or set in roster.conf"))
			return
		}

		token := os.Getenv("NETBOX_TOKEN")
		if token == "" {
			token = confToken
		}
		if token == "" {
			fmt.Fprintln(os.Stderr, ui.ErrorMsg("NETBOX_TOKEN environment variable or token in roster.conf is required"))
			return
		}

		filter := syncFilter
		if filter == "" {
			filter = confFilter
		}

		dir := inventoryPaths[0]
		if err := store.LockInventory(dir); err != nil {
			fmt.Fprintln(os.Stderr, ui.ErrorMsg("%v", err))
			return
		}
		defer store.UnlockInventory()

		opts := netbox.SyncOptions{
			SyncConfigContexts: syncConfigContexts,
			SyncHosts:          syncHosts,
			SyncInterfaces:     syncInterfaces,
			SyncIPs:            syncIPs,
			SyncVMDisks:        syncVMDisks,
		}
		if err := netbox.Sync(baseURL, token, dir, filter, opts); err != nil {
			fmt.Fprintln(os.Stderr, ui.ErrorMsg("Sync failed: %v", err))
		}
	},
}

func init() {
	syncNetboxCmd.Flags().StringVarP(&syncFilter, "filter", "f", "", "Filter query parameters (e.g. status=active,role=linux)")
	syncNetboxCmd.Flags().BoolVar(&syncConfigContexts, "config-contexts", true, "Sync group config contexts")
	syncNetboxCmd.Flags().BoolVar(&syncHosts, "hosts", true, "Sync hosts (devices & VMs)")
	syncNetboxCmd.Flags().BoolVar(&syncInterfaces, "interfaces", true, "Sync interfaces")
	syncNetboxCmd.Flags().BoolVar(&syncIPs, "ips", true, "Sync IP addresses")
	syncNetboxCmd.Flags().BoolVar(&syncVMDisks, "vm-disks", true, "Sync VM disks")
	syncCmd.AddCommand(syncNetboxCmd)
	rootCmd.AddCommand(syncCmd)
}
