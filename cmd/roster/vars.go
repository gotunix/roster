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
	"strings"

	"github.com/spf13/cobra"
	"gotunix.net/roster/internal/interactive"
	"gotunix.net/roster/internal/store"
	"gotunix.net/roster/internal/ui"
)

var varsCmd = &cobra.Command{
	Use:   "vars",
	Short: "Manage Ansible variables",
}

var varsSetCmd = &cobra.Command{
	Use:   "set <type> <name> <key>=<value>",
	Short: "Set a variable for a host or group",
	Long:  `Set a variable. Type must be 'host' or 'group'.`,
	Args:  cobra.ExactArgs(3),
	Run: func(_ *cobra.Command, args []string) {
		vType := strings.ToLower(args[0])
		name := args[1]
		kv := args[2]
		dir := inventoryPaths[0]

		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			fmt.Println(ui.ErrorMsg("variable must be in key=value format"))
			return
		}
		key, value := parts[0], parts[1]

		var err error
		switch vType {
		case "host":
			err = store.SetHostVar(dir, name, key, value)
		case "group":
			err = store.SetGroupVar(dir, name, key, value)
		default:
			fmt.Println(ui.ErrorMsg("invalid type %q. Must be 'host' or 'group'", vType))
			return
		}

		if err != nil {
			fmt.Println(ui.ErrorMsg("%v", err))
		} else {
			fmt.Println(ui.SuccessMsg("Variable %s set for %s %s", key, vType, name))
		}
	},
}

var varsEditCmd = &cobra.Command{
	Use:   "edit <type> <name>",
	Short: "Edit variables interactively",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		vType := strings.ToLower(args[0])
		name := args[1]
		dir := inventoryPaths[0]

		var err error
		switch vType {
		case "host":
			err = interactive.EditHostInteractive(dir, name)
		case "group":
			err = interactive.EditGroupInteractive(dir, name)
		default:
			fmt.Println(ui.ErrorMsg("invalid type %q. Must be 'host' or 'group'", vType))
			return
		}

		if err != nil {
			fmt.Println(ui.ErrorMsg("%v", err))
		} else {
			fmt.Println(ui.SuccessMsg("Variables for %s %s updated", vType, name))
		}
	},
}

var varsSyncCmd = &cobra.Command{
	Use:   "sync <type> <name> <destination_dir>",
	Short: "Sync variables to another inventory",
	Args:  cobra.ExactArgs(3),
	Run: func(_ *cobra.Command, args []string) {
		vType := strings.ToLower(args[0])
		name := args[1]
		destDir := args[2]
		sourceDir := inventoryPaths[0]

		if err := store.SyncVars(sourceDir, destDir, vType, name); err != nil {
			fmt.Println(ui.ErrorMsg("syncing variables: %v", err))
		} else {
			fmt.Println(ui.SuccessMsg("Variables for %s %s synced to %s", vType, name, destDir))
		}
	},
}

func init() {
	varsCmd.AddCommand(varsSetCmd)
	varsCmd.AddCommand(varsEditCmd)
	varsCmd.AddCommand(varsSyncCmd)
	rootCmd.AddCommand(varsCmd)
}
