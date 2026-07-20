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
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFormModel(t *testing.T) {
	t.Run("form field addition and retrieval", func(t *testing.T) {
		f := NewForm("Create Task")
		f.AddTextBox("title", "Task Title", "Enter task title", "Task title help text")
		f.AddTextArea("desc", "Description", "Enter task description", "")
		f.AddSelector("priority", "Priority", []string{"LOW", "MEDIUM", "HIGH"}, "Select priority")
		f.AddBoolean("changelog", "Write to Changelog?", "Check to record")
		f.AddButton("Submit", func(form *FormModel) tea.Cmd {
			form.Submitted = true
			return tea.Quit
		})

		if len(f.Fields) != 5 {
			t.Errorf("expected 5 fields, got %d", len(f.Fields))
		}

		// Initial states
		if f.GetString("title") != "" {
			t.Errorf("expected empty initial title, got %q", f.GetString("title"))
		}

		// Set value manually on textbox
		f.Fields[0].TextInput.SetValue("Implement feature A")
		if f.GetString("title") != "Implement feature A" {
			t.Errorf("expected value 'Implement feature A', got %q", f.GetString("title"))
		}

		// Set value on textarea
		f.Fields[1].TextArea.SetValue("Detailed description here")
		if f.GetString("desc") != "Detailed description here" {
			t.Errorf("expected value 'Detailed description here', got %q", f.GetString("desc"))
		}

		// Check selector initial value
		if f.GetString("priority") != "LOW" {
			t.Errorf("expected selector LOW, got %q", f.GetString("priority"))
		}
	})

	t.Run("size resize update", func(t *testing.T) {
		f := NewForm("test")
		f.AddTextBox("title", "Title", "", "")

		updated, _ := f.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
		form := updated.(*FormModel)

		if form.Width != 95 {
			t.Errorf("expected Width=95 based on default WidthPct=0.95, got %d", form.Width)
		}
	})

	t.Run("focus traversal with tab", func(t *testing.T) {
		f := NewForm("test")
		f.AddTextBox("a", "A", "", "")
		f.AddTextBox("b", "B", "", "")

		f.Init()
		if f.FocusIndex != 0 {
			t.Errorf("expected focus at index 0, got %d", f.FocusIndex)
		}

		updated, _ := f.Update(tea.KeyMsg{Type: tea.KeyTab})
		f = updated.(*FormModel)
		if f.FocusIndex != 1 {
			t.Errorf("expected focus at index 1 after tab, got %d", f.FocusIndex)
		}
	})

	t.Run("button action submit", func(t *testing.T) {
		f := NewForm("test")
		f.AddButton("Submit", func(form *FormModel) tea.Cmd {
			form.Submitted = true
			return tea.Quit
		})

		f.Init()
		_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected command on submit button enter press")
		}
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); !ok {
			t.Errorf("expected tea.QuitMsg, got %T", msg)
		}
		if !f.Submitted {
			t.Error("expected Submitted to be true after action execution")
		}
	})
}
