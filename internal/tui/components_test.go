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
)

func TestSpinner(t *testing.T) {
	m := NewSpinner("loading...")
	if m.View() == "" {
		t.Error("Spinner view should not be empty")
	}
	if !strings.Contains(m.View(), "loading...") {
		t.Errorf("Spinner view should contain label, got %q", m.View())
	}

	updated, _ := m.Update(tea.KeyMsg{})
	if updated.View() == "" {
		t.Error("Spinner view after update should not be empty")
	}
}

func TestConfirm(t *testing.T) {
	t.Run("initial view", func(t *testing.T) {
		m := NewConfirm("Are you sure?")
		if m.View() == "" {
			t.Error("Confirm view should not be empty")
		}
		if !strings.Contains(m.View(), "Are you sure?") {
			t.Errorf("Confirm view should contain prompt, got %q", m.View())
		}
	})

	t.Run("tab switches focus", func(t *testing.T) {
		m := NewConfirm("test")
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
		// focused should now be 1 (cancel)
		if m.focused != 1 {
			t.Errorf("expected focus=1 after tab, got %d", m.focused)
		}
	})

	t.Run("enter confirms", func(t *testing.T) {
		m := NewConfirm("test")
		result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if result.done != true {
			t.Error("expected done after enter")
		}
		if cmd == nil {
			t.Error("expected a cmd (ConfirmResult) after enter")
		} else {
			msg := cmd()
			if r, ok := msg.(ConfirmResult); ok {
				if !bool(r) {
					t.Error("expected ConfirmResult(true) when focused=0 (yes)")
				}
			} else {
				t.Errorf("expected ConfirmResult message, got %T", msg)
			}
		}
	})

	t.Run("escape cancels", func(t *testing.T) {
		m := NewConfirm("test")
		result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if result.done != true {
			t.Error("expected done after esc")
		}
		if cmd == nil {
			t.Error("expected a cmd (ConfirmResult) after esc")
		} else {
			msg := cmd()
			if r, ok := msg.(ConfirmResult); ok {
				if bool(r) {
					t.Error("expected ConfirmResult(false) on esc")
				}
			} else {
				t.Errorf("expected ConfirmResult message, got %T", msg)
			}
		}
	})

	t.Run("result getter", func(t *testing.T) {
		m := NewConfirm("test")
		// confirm (focused=0) and press enter
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if !m.Result() {
			t.Error("expected Result()=true after confirming")
		}
	})
}

func TestInput(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		m := NewInput("Name", "Enter name")
		if m.View() == "" {
			t.Error("Input view should not be empty")
		}
		if !strings.Contains(m.View(), "Name") {
			t.Errorf("Input view should contain label, got %q", m.View())
		}
		if m.Value() != "" {
			t.Errorf("expected empty value, got %q", m.Value())
		}
	})

	t.Run("set value", func(t *testing.T) {
		m := NewInput("Name", "Enter name")
		m = m.SetValue("hello")
		if m.Value() != "hello" {
			t.Errorf("expected value 'hello', got %q", m.Value())
		}
	})

	t.Run("focus styling", func(t *testing.T) {
		m := NewInput("Name", "Enter name")
		m = m.Focus()
		if !m.focus {
			t.Error("expected focus=true after Focus()")
		}
		m = m.Blur()
		if m.focus {
			t.Error("expected focus=false after Blur()")
		}
	})
}

func TestStatusBar(t *testing.T) {
	t.Run("initial view", func(t *testing.T) {
		m := NewStatusBar("connected", "v0.1.0")
		if m.View() == "" {
			t.Error("StatusBar view should not be empty")
		}
		if !strings.Contains(m.View(), "connected") || !strings.Contains(m.View(), "v0.1.0") {
			t.Errorf("StatusBar should contain both sections, got %q", m.View())
		}
	})

	t.Run("window resize", func(t *testing.T) {
		m := NewStatusBar("left", "right")
		updated, _ := m.Update(tea.WindowSizeMsg{Width: 120})
		if updated.Width != 120 {
			t.Errorf("expected Width=120 after resize, got %d", updated.Width)
		}
	})
}
