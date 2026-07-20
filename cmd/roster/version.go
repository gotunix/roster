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

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the version information of roster",
	Long:  `Displays application details, compile-time Go versions, system architecture, and dependencies. Supports a console mode and an interactive TUI window.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tuiFlag, _ := cmd.Flags().GetBool("tui")

		if tuiFlag {
			p := tea.NewProgram(tui.NewVersionWindow(version.AppName, version.AppVersion, version.Commit, version.Date), tea.WithAltScreen())
			_, err := p.Run()
			return err
		}

		cmd.Print(tui.RenderVersionStatic(version.AppName, version.AppVersion, version.Commit, version.Date))
		return nil
	},
}

func init() {
	versionCmd.SetHelpFunc(ui.HandleHelp)
	versionCmd.SetUsageFunc(ui.HandleUsage)
	versionCmd.Flags().BoolP("tui", "t", false, "display version inside an interactive TUI window")
	rootCmd.AddCommand(versionCmd)
}
