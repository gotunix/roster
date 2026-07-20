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
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputModel is a styled text input field matching the reference project's
// AddTextBox() styling (CatMochaBlue prompt, CatMochaText value, CatMochaMauve cursor).
type InputModel struct {
	inner textinput.Model
	label string
	focus bool
}

// NewInput creates a new styled input with the given label and placeholder.
func NewInput(label, placeholder string) InputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = "  "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(CatMochaBlue)
	ti.TextStyle = lipgloss.NewStyle().Foreground(CatMochaText)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(CatMochaMauve)
	ti.CharLimit = 256

	return InputModel{inner: ti, label: label, focus: false}
}

func (m InputModel) Init() tea.Cmd {
	if m.focus {
		return textinput.Blink
	}
	return nil
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.inner, cmd = m.inner.Update(msg)
	return m, cmd
}

func (m InputModel) View() string {
	label := LabelStyle
	if m.focus {
		label = FocusedLabelStyle
	}
	return label.Render(m.label) + "\n" + m.inner.View()
}

// Focus sets the focus state of the input.
func (m InputModel) Focus() InputModel {
	m.focus = true
	m.inner.Focus()
	return m
}

// Blur removes focus from the input.
func (m InputModel) Blur() InputModel {
	m.focus = false
	m.inner.Blur()
	return m
}

// Value returns the current input value.
func (m InputModel) Value() string {
	return m.inner.Value()
}

// SetValue sets the input value.
func (m InputModel) SetValue(v string) InputModel {
	m.inner.SetValue(v)
	return m
}
