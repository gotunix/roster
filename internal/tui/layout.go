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

package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// AdjustDimensions calculates layout sizes based on terminal dimensions.
// Returns contentWidth, targetWidth, targetHeight.
func AdjustDimensions(width, height int) (contentWidth, targetWidth, targetHeight int) {
	targetWidth = int(float64(width) * 0.95)
	if targetWidth < 40 {
		targetWidth = 40
	}
	targetHeight = int(float64(height) * 0.95)
	if targetHeight < 12 {
		targetHeight = 12
	}
	contentWidth = targetWidth - 8
	if contentWidth < 20 {
		contentWidth = 20
	}
	return
}

// Divider renders a horizontal rule using the overlay colour.
func Divider(width int) string {
	return lipgloss.NewStyle().
		Foreground(CatMochaOverlay).
		Render(strings.Repeat("─", width))
}
