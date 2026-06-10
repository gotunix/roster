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
	"gotunix.net/roster/internal/models"
	"gotunix.net/roster/internal/store"
	"gotunix.net/roster/internal/ui"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage Ansible groups",
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups in an inventory",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		dir := inventoryPaths[0]
		inv, err := store.LoadInventory(dir)
		if err != nil {
			fmt.Println(ui.ErrorMsg("Loading inventory: %v", err))
			return
		}
		fmt.Print(ui.RenderGroupList(inv))
	},
}

var groupViewCmd = &cobra.Command{
	Use:   "view <groupname>",
	Short: "View group details and variables",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupname := args[0]
		dir := inventoryPaths[0]

		inv, err := store.LoadInventory(dir)
		if err != nil {
			fmt.Println(ui.ErrorMsg("Loading inventory: %v", err))
			return
		}

		output, err := ui.RenderGroupView(groupname, dir, inv)
		if err != nil {
			fmt.Println(ui.ErrorMsg("%v", err))
			return
		}
		fmt.Print(output)
	},
}

var groupAddCmd = &cobra.Command{
	Use:   "add <group1,group2,...>",
	Short: "Add one or more groups to the inventory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupList := args[0]
		dir := inventoryPaths[0]

		groups := strings.Split(groupList, ",")
		for _, groupname := range groups {
			groupname = strings.TrimSpace(groupname)
			if groupname == "" {
				continue
			}

			group := &models.Group{Name: groupname}
			if err := store.SaveGroup(dir, groupname, group); err != nil {
				fmt.Println(ui.ErrorMsg("Adding group %s: %v", groupname, err))
			} else {
				fmt.Println(ui.SuccessMsg("Group %s added successfully", groupname))
			}
		}
	},
}

var groupAssignCmd = &cobra.Command{
	Use:   "assign <hostname> <groupname1,group2,...>",
	Short: "Assign a host to one or more groups",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		hostname := args[0]
		groupList := args[1]
		dir := inventoryPaths[0]

		inv, err := store.LoadInventory(dir)
		if err != nil {
			fmt.Println(ui.ErrorMsg("Loading inventory: %v", err))
			return
		}
		if _, ok := inv.Hosts[hostname]; !ok {
			fmt.Println(ui.ErrorMsg("Host %q not found in inventory", hostname))
			return
		}

		groups := strings.Split(groupList, ",")
		for _, groupname := range groups {
			groupname = strings.TrimSpace(groupname)
			if groupname == "" {
				continue
			}
			if err := store.AssignHostToGroup(dir, hostname, groupname); err != nil {
				fmt.Println(ui.ErrorMsg("Assigning %s to %s: %v", hostname, groupname, err))
			} else {
				fmt.Println(ui.SuccessMsg("Host %s assigned to group %s", hostname, groupname))
			}
		}
	},
}

var groupRemoveCmd = &cobra.Command{
	Use:   "remove <groupname>",
	Short: "Remove a group from the inventory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupname := args[0]
		dir := inventoryPaths[0]

		if err := store.RemoveGroup(dir, groupname); err != nil {
			fmt.Println(ui.ErrorMsg("%v", err))
		} else {
			fmt.Println(ui.SuccessMsg("Group %s removed successfully", groupname))
		}
	},
}

var groupEditCmd = &cobra.Command{
	Use:   "edit <groupname>",
	Short: "Edit group details and variables",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupname := args[0]
		dir := inventoryPaths[0]

		if err := interactive.EditGroupInteractive(dir, groupname); err != nil {
			fmt.Println(ui.ErrorMsg("%v", err))
		} else {
			fmt.Println(ui.SuccessMsg("Group %s updated", groupname))
		}
	},
}

var groupCopyCmd = &cobra.Command{
	Use:   "copy <source_group> <dest_group>",
	Short: "Clone an existing group to a new name",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		dest := args[1]
		dir := inventoryPaths[0]

		if err := store.CopyGroup(dir, source, dest); err != nil {
			fmt.Println(ui.ErrorMsg("Cloning group %s: %v", source, err))
		} else {
			fmt.Println(ui.SuccessMsg("Group %s cloned to %s successfully", source, dest))
		}
	},
}

func init() {
	groupCmd.AddCommand(groupAddCmd)
	groupCmd.AddCommand(groupAssignCmd)
	groupCmd.AddCommand(groupRemoveCmd)
	groupCmd.AddCommand(groupCopyCmd)
	groupCmd.AddCommand(groupEditCmd)
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupViewCmd)
	rootCmd.AddCommand(groupCmd)
}
