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
	"strings"

	"github.com/spf13/cobra"
	"gotunix.net/roster/internal/interactive"
	"gotunix.net/roster/internal/store"
	"gotunix.net/roster/internal/ui"
)

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Manage Ansible hosts",
}

var hostListCmd = &cobra.Command{
	Use:   "list [groups|group_name]",
	Short: "List hosts in an inventory. Use 'groups' to show hierarchy.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := inventoryPaths[0]
		inv, err := store.LoadInventory(dir)
		if err != nil {
			fmt.Println(ui.ErrorMsg("Loading inventory: %v", err))
			return
		}

		groupFilter := ""
		showGroups := false

		if len(args) > 0 {
			if args[0] == "groups" {
				showGroups = true
			} else {
				groupFilter = args[0]
			}
		}

		fmt.Print(ui.RenderHostList(inv, groupFilter, showGroups))
	},
}

var hostViewCmd = &cobra.Command{
	Use:   "view <hostname>",
	Short: "View host details and variables",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hostname := args[0]
		dir := inventoryPaths[0]

		inv, err := store.LoadInventory(dir)
		if err != nil {
			fmt.Println(ui.ErrorMsg("Loading inventory: %v", err))
			return
		}

		if _, ok := inv.Hosts[hostname]; !ok {
			fmt.Println(ui.ErrorMsg("Host %q not found", hostname))
			return
		}

		output, err := ui.RenderHostView(hostname, dir, inv)
		if err != nil {
			fmt.Println(ui.ErrorMsg("Rendering view: %v", err))
			return
		}
		fmt.Print(output)
	},
}

var hostAddCmd = &cobra.Command{
	Use:   "add <hostname1,host2,...>",
	Short: "Add one or more hosts to the inventory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hostList := args[0]
		dir := inventoryPaths[0]

		hosts := strings.Split(hostList, ",")
		for _, hostname := range hosts {
			hostname = strings.TrimSpace(hostname)
			if hostname == "" {
				continue
			}

			if err := store.AddHostToMain(dir, hostname); err != nil {
				fmt.Println(ui.ErrorMsg("Adding host %s: %v", hostname, err))
			} else {
				fmt.Println(ui.SuccessMsg("Host %s added successfully", hostname))
			}
		}
	},
}

var hostRemoveCmd = &cobra.Command{
	Use:   "remove <hostname>",
	Short: "Remove a host from the inventory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hostname := args[0]
		dir := inventoryPaths[0]

		if err := store.RemoveHost(dir, hostname); err != nil {
			fmt.Println(ui.ErrorMsg("%v", err))
		} else {
			fmt.Println(ui.SuccessMsg("Host %s removed successfully", hostname))
		}
	},
}

var hostEditUseEditor bool

var hostEditCmd = &cobra.Command{
	Use:   "edit <hostname>",
	Short: "Edit host details and variables",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hostname := args[0]
		dir := inventoryPaths[0]

		var err error
		if hostEditUseEditor {
			err = interactive.EditHostExternal(dir, hostname)
		} else {
			err = interactive.EditHostInteractive(dir, hostname)
		}

		if err != nil {
			fmt.Println(ui.ErrorMsg("%v", err))
		} else {
			fmt.Println(ui.SuccessMsg("Host %s updated", hostname))
		}
	},
}

var hostMoveCmd = &cobra.Command{
	Use:   "move <hostname> <destination_dir>",
	Short: "Move a host to another inventory",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		hostname := args[0]
		destDir := args[1]
		sourceDir := inventoryPaths[0]

		if err := store.MoveHost(sourceDir, destDir, hostname); err != nil {
			fmt.Println(ui.ErrorMsg("Moving host %s: %v", hostname, err))
		} else {
			fmt.Println(ui.SuccessMsg("Host %s migrated successfully from %s to %s", hostname, sourceDir, destDir))
		}
	},
}

func init() {
	hostCmd.AddCommand(hostAddCmd)
	hostCmd.AddCommand(hostRemoveCmd)
	hostEditCmd.Flags().BoolVarP(&hostEditUseEditor, "editor", "e", false, "Use external $EDITOR instead of built-in form")
	hostCmd.AddCommand(hostEditCmd)
	hostCmd.AddCommand(hostMoveCmd)
	hostCmd.AddCommand(hostListCmd)
	hostCmd.AddCommand(hostViewCmd)
	rootCmd.AddCommand(hostCmd)
}
