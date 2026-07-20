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

func TestTreeListModel(t *testing.T) {
	t.Run("nested tree list structure", func(t *testing.T) {
		task1 := &ListItem{ID: "t1", Title: "Task 1", Type: "task", Status: "ACTIVE", Progress: 0.5, ProgressText: "50%"}
		story1 := &ListItem{ID: "s1", Title: "Story 1", Type: "story", Status: "ACTIVE", Children: []*ListItem{task1}, Expanded: true}
		milestone1 := &ListItem{ID: "m1", Title: "Milestone 1", Type: "milestone", Status: "ACTIVE", Children: []*ListItem{story1}, Expanded: true}

		model := NewTreeList("Test Board", []*ListItem{milestone1})

		flatRows := model.GetFlatRows()
		if len(flatRows) != 3 {
			t.Errorf("expected 3 flat rows, got %d", len(flatRows))
		}

		if flatRows[0].Item.ID != "m1" || flatRows[1].Item.ID != "s1" || flatRows[2].Item.ID != "t1" {
			t.Error("expected correct flat row sequence matching depth-first expansion")
		}

		// Close story1 expansion
		story1.Expanded = false
		flatRows = model.GetFlatRows()
		if len(flatRows) != 2 {
			t.Errorf("expected 2 flat rows after collapsing story1, got %d", len(flatRows))
		}
	})

	t.Run("up/down navigation", func(t *testing.T) {
		a := &ListItem{ID: "a", Title: "A", Type: "item", Status: "TODO"}
		b := &ListItem{ID: "b", Title: "B", Type: "item", Status: "TODO"}
		model := NewTreeList("Test List", []*ListItem{a, b})

		if model.Cursor != 0 {
			t.Errorf("expected initial cursor=0, got %d", model.Cursor)
		}

		// Move down
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = updated.(TreeListModel)
		if model.Cursor != 1 {
			t.Errorf("expected cursor=1 after down arrow, got %d", model.Cursor)
		}

		// Move up
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
		model = updated.(TreeListModel)
		if model.Cursor != 0 {
			t.Errorf("expected cursor=0 after up arrow, got %d", model.Cursor)
		}
	})

	t.Run("enter toggles expansion", func(t *testing.T) {
		child := &ListItem{ID: "c", Title: "Child"}
		parent := &ListItem{ID: "p", Title: "Parent", Children: []*ListItem{child}, Expanded: false}
		model := NewTreeList("Test List", []*ListItem{parent})

		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		model = updated.(TreeListModel)

		if !parent.Expanded {
			t.Error("expected parent node to expand after enter key")
		}
	})
}
