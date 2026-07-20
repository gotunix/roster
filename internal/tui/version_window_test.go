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
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestVersionWindowModel(t *testing.T) {
	t.Run("initialization and custom styling", func(t *testing.T) {
		m := NewVersionWindow("TestApp", "v1.2.3", "abcdef", "2026-07-07")
		if m.AppName != "TestApp" {
			t.Errorf("expected AppName='TestApp', got %q", m.AppName)
		}

		customTheme := DefaultTheme()
		customTheme.Primary = "#ff00ff"
		customStyles := GenerateStyles(customTheme)
		m = m.WithStyles(customStyles)

		bg, ok := m.Styles.HeaderStyle.GetBackground().(lipgloss.Color)
		if !ok || string(bg) != "#ff00ff" {
			t.Errorf("expected header style background to be custom color #ff00ff, got %v", m.Styles.HeaderStyle.GetBackground())
		}
	})

	t.Run("view before initialization", func(t *testing.T) {
		m := NewVersionWindow("TestApp", "v1.2.3", "abcdef", "2026-07-07")
		view := m.View()
		if !strings.Contains(view, "Initializing") {
			t.Errorf("expected initializing view, got %q", view)
		}
	})

	t.Run("update size and rendering", func(t *testing.T) {
		m := NewVersionWindow("TestApp", "v1.2.3", "abcdef", "2026-07-07")

		// Send terminal resize message
		updatedModel, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		m = updatedModel.(VersionWindowModel)

		if cmd == nil {
			// Init viewport command is empty but we return it in batch
		}

		if !m.ready {
			t.Error("expected model to be ready after size message")
		}

		view := m.View()
		if view == "" {
			t.Error("expected non-empty view")
		}

		// Check for static elements in view
		for _, term := range []string{"TESTAPP", "Version", "v1.2.3", "Commit", "abcdef", "Built", "2026-07-07", "Module", "GO", "OS", "Arch", "Dependencies:"} {
			if !strings.Contains(view, term) {
				t.Errorf("expected view to contain %q, got:\n%s", term, view)
			}
		}
	})

	t.Run("quit keypress", func(t *testing.T) {
		m := NewVersionWindow("TestApp", "v1.2.3", "abcdef", "2026-07-07")
		m.ready = true

		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd == nil {
			t.Error("expected quit command when pressing q")
		} else {
			msg := cmd()
			if _, ok := msg.(tea.QuitMsg); !ok {
				t.Errorf("expected tea.QuitMsg, got %T", msg)
			}
		}
	})
}
