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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusBarModel is a persistent footer bar with left and right aligned sections.
// Uses the reference's HeaderStyle pattern (CatMochaBase fg, CatMochaMauve bg, bold).
type StatusBarModel struct {
	Left  string
	Right string
	Width int
}

// NewStatusBar creates a new status bar.
func NewStatusBar(left, right string) StatusBarModel {
	return StatusBarModel{Left: left, Right: right, Width: 80}
}

func (m StatusBarModel) Init() tea.Cmd { return nil }

func (m StatusBarModel) Update(msg tea.Msg) (StatusBarModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
	}
	return m, nil
}

func (m StatusBarModel) View() string {
	left := HeaderStyle.Render(" " + m.Left)
	right := HeaderStyle.Render(m.Right + " ")

	filler := m.Width - lipgloss.Width(left) - lipgloss.Width(right)
	if filler < 1 {
		filler = 1
	}

	middle := HeaderStyle.Render(strings.Repeat(" ", filler))
	return lipgloss.JoinHorizontal(lipgloss.Top, left, middle, right)
}
