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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gotunix.net/roster/internal/netbox"
	"gotunix.net/roster/internal/store"
	"gotunix.net/roster/internal/ui"
)

var (
	syncFilter string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync inventory from external sources",
}

var syncNetboxCmd = &cobra.Command{
	Use:   "netbox <url>",
	Short: "Sync hosts from NetBox API",
	Long:  `Fetch devices, VMs, interfaces, and contexts from NetBox and add them to the inventory. Requires NETBOX_TOKEN env var.`,
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

		if err := netbox.Sync(baseURL, token, dir, syncFilter); err != nil {
			fmt.Fprintln(os.Stderr, ui.ErrorMsg("Sync failed: %v", err))
		}
	},
}

func init() {
	syncNetboxCmd.Flags().StringVarP(&syncFilter, "filter", "f", "", "Filter query parameters (e.g. status=active,role=linux)")
	syncCmd.AddCommand(syncNetboxCmd)
	rootCmd.AddCommand(syncCmd)
}
