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

package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"gotunix.net/roster/internal/tui"
	"gotunix.net/roster/internal/ui"
	"gotunix.net/roster/internal/version"
)

var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "Launch the interactive TUI menu for roster",
	Long:  `Launches an interactive Terminal User Interface (TUI) menu built with Bubble Tea, showcasing styled controls and version info.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(inventoryPaths) > 0 {
			path = inventoryPaths[0]
		}
		p := tea.NewProgram(tui.NewMainTUIModel(version.AppName, version.AppVersion, version.Commit, version.Date, path), tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}

func init() {
	menuCmd.SetHelpFunc(ui.HandleHelp)
	menuCmd.SetUsageFunc(ui.HandleUsage)
	rootCmd.AddCommand(menuCmd)
}
