// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The Roster Authors
// =============================================================================================== //
//                                                                                                 //
//            /$$$$$$                                                                              //
//           /$$__  $$                                                                             //
//          | $$  \__/  /$$$$$$  /$$   /$$  /$$$$$$   /$$$$$$$  /$$$$$$                            //
//          |  $$$$$$  /$$__  $$| $$  | $$ /$$__  $$ /$$_____/ /$$__  $$                           //
//           \____  $$| $$  \ $$| $$  | $$| $$  \__/| $$      | $$$$$$$$                           //
//           /$$  \ $$| $$  | $$| $$  | $$| $$      | $$      | $$_____/                           //
//          |  $$$$$$/|  $$$$$$/|  $$$$$$/| $$      |  $$$$$$$|  $$$$$$$                           //
//           \______/  \______/  \______/ |__/       \_______/ \_______/                           //
//                                                                                                 //
//                                             /$$    /$$                    /$$   /$$             //
//                                            | $$   | $$                   | $$  | $$             //
//                                            | $$   | $$ /$$$$$$  /$$   /$$| $$ /$$$$$$           //
//                                            |  $$ / $$/|____  $$| $$  | $$| $$|_  $$_/           //
//                                             \  $$ $$/  /$$$$$$$| $$  | $$| $$  | $$             //
//                                              \  $$$/  /$$__  $$| $$  | $$| $$  | $$ /$$         //
//                                               \  $/  |  $$$$$$$|  $$$$$$/| $$  |  $$$$/         //
//                                                \_/    \_______/ \______/ |__/   \___/           //
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

// Package tui provides a reusable lipgloss-based styling toolkit for all SourceVault
// client binaries (svctl, svboards, svsh). It defines the Catppuccin Mocha color
// palette, semantic theme, layout helpers, and common style primitives so that
// every CLI tool renders with a consistent visual identity.
package tui

import "github.com/charmbracelet/lipgloss"

// ---------------------------------------------------------------------------
// Catppuccin Mocha palette
// ---------------------------------------------------------------------------

var (
	CatMochaBase    = lipgloss.Color("#1e1e2e")
	CatMochaText    = lipgloss.Color("#cdd6f4")
	CatMochaSubtext = lipgloss.Color("#a6adc8")
	CatMochaOverlay = lipgloss.Color("#6c7086")
	CatMochaBlue    = lipgloss.Color("#89b4fa")
	CatMochaGreen   = lipgloss.Color("#a6e3a1")
	CatMochaRed     = lipgloss.Color("#f38ba8")
	CatMochaMauve   = lipgloss.Color("#cba6f7")
	CatMochaPeach   = lipgloss.Color("#fab387")
)

// Theme bundles colors into semantic roles. When every style reads from a
// Theme, switching the look of an entire application is a single assignment.
type Theme struct {
	Base    lipgloss.Color
	Text    lipgloss.Color
	Subtext lipgloss.Color
	Overlay lipgloss.Color

	Primary lipgloss.Color
	Success lipgloss.Color
	Error   lipgloss.Color
	Warning lipgloss.Color
	Accent  lipgloss.Color
}

// DefaultTheme returns a Theme initialised with the Catppuccin Mocha palette.
func DefaultTheme() Theme {
	return Theme{
		Base:    CatMochaBase,
		Text:    CatMochaText,
		Subtext: CatMochaSubtext,
		Overlay: CatMochaOverlay,
		Primary: CatMochaMauve,
		Success: CatMochaGreen,
		Error:   CatMochaRed,
		Warning: CatMochaPeach,
		Accent:  CatMochaBlue,
	}
}
