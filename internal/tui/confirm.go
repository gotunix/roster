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
)

// ConfirmResult is sent as a tea.Msg when the user confirms or cancels.
type ConfirmResult bool

// ConfirmModel is a yes/no confirmation dialog rendered using the reference's
// WindowStyle and ButtonStyle/ActiveButtonStyle patterns.
type ConfirmModel struct {
	Prompt  string
	focused int // 0 = confirm, 1 = cancel
	done    bool
	result  bool
}

// NewConfirm creates a new confirmation dialog with the given prompt.
func NewConfirm(prompt string) ConfirmModel {
	return ConfirmModel{Prompt: prompt, focused: 0}
}

func (m ConfirmModel) Init() tea.Cmd { return nil }

func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.done = true
			m.result = false
			return m, func() tea.Msg { return ConfirmResult(false) }
		case "tab", "left", "right":
			m.focused = (m.focused + 1) % 2
		case "enter":
			m.done = true
			m.result = m.focused == 0
			return m, func() tea.Msg { return ConfirmResult(m.result) }
		}
	}
	return m, nil
}

func (m ConfirmModel) View() string {
	confirmBtn := ButtonStyle.Render(" YES ")
	cancelBtn := ButtonStyle.Render(" NO ")
	if m.focused == 0 {
		confirmBtn = ActiveButtonStyle.Render(" YES ")
	} else {
		cancelBtn = ActiveButtonStyle.Render(" NO ")
	}

	content := strings.Join([]string{
		"",
		"  " + m.Prompt,
		"",
		"  " + confirmBtn + "  " + cancelBtn,
		"",
	}, "\n")

	return WindowStyle.Render(content)
}

// Result returns the user's choice. Only valid after the model has finished.
func (m ConfirmModel) Result() bool {
	return m.result
}
