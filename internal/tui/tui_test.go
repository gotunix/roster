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

	"github.com/charmbracelet/lipgloss"
)

func TestColors(t *testing.T) {
	tests := []struct {
		name string
		got  lipgloss.Color
		want string
	}{
		{"CatMochaBase", CatMochaBase, "#1e1e2e"},
		{"CatMochaText", CatMochaText, "#cdd6f4"},
		{"CatMochaSubtext", CatMochaSubtext, "#a6adc8"},
		{"CatMochaOverlay", CatMochaOverlay, "#6c7086"},
		{"CatMochaBlue", CatMochaBlue, "#89b4fa"},
		{"CatMochaGreen", CatMochaGreen, "#a6e3a1"},
		{"CatMochaRed", CatMochaRed, "#f38ba8"},
		{"CatMochaMauve", CatMochaMauve, "#cba6f7"},
		{"CatMochaPeach", CatMochaPeach, "#fab387"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Errorf("got %q, want %q", string(tt.got), tt.want)
			}
		})
	}
}

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()
	if string(theme.Base) != "#1e1e2e" {
		t.Errorf("Base = %q, want #1e1e2e", string(theme.Base))
	}
	if string(theme.Primary) != "#cba6f7" {
		t.Errorf("Primary = %q, want #cba6f7", string(theme.Primary))
	}
	if string(theme.Success) != "#a6e3a1" {
		t.Errorf("Success = %q, want #a6e3a1", string(theme.Success))
	}
	if string(theme.Error) != "#f38ba8" {
		t.Errorf("Error = %q, want #f38ba8", string(theme.Error))
	}
	if string(theme.Warning) != "#fab387" {
		t.Errorf("Warning = %q, want #fab387", string(theme.Warning))
	}
}

func TestStylesRender(t *testing.T) {
	rendered := WindowStyle.Render("hello")
	if rendered == "" {
		t.Error("WindowStyle rendered empty string")
	}
	if !strings.Contains(rendered, "hello") {
		t.Errorf("WindowStyle render should contain input text, got %q", rendered)
	}

	rendered = TitleStyle.Render("title")
	if rendered == "" {
		t.Error("TitleStyle rendered empty string")
	}

	rendered = LabelStyle.Render("label")
	if rendered == "" {
		t.Error("LabelStyle rendered empty string")
	}

	rendered = ErrorStyle.Render("error")
	if rendered == "" {
		t.Error("ErrorStyle rendered empty string")
	}

	rendered = HeaderStyle.Render("header")
	if rendered == "" {
		t.Error("HeaderStyle rendered empty string")
	}
}

func TestHelpers(t *testing.T) {
	t.Run("KeyValue", func(t *testing.T) {
		got := KeyValue("Version", "v0.1.0")
		if got == "" {
			t.Error("KeyValue returned empty string")
		}
		if !strings.Contains(got, "Version") || !strings.Contains(got, "v0.1.0") {
			t.Errorf("KeyValue missing expected content, got %q", got)
		}
	})

	t.Run("Success", func(t *testing.T) {
		got := Success("done")
		if got == "" {
			t.Error("Success returned empty string")
		}
	})

	t.Run("Error", func(t *testing.T) {
		got := Error("fail")
		if got == "" {
			t.Error("Error returned empty string")
		}
	})

	t.Run("Warning", func(t *testing.T) {
		got := Warning("caution")
		if got == "" {
			t.Error("Warning returned empty string")
		}
	})

	t.Run("Header", func(t *testing.T) {
		got := Header("test", 40)
		if got == "" {
			t.Error("Header returned empty string")
		}
		if !strings.Contains(got, "TEST") {
			t.Errorf("Header should uppercase, got %q", got)
		}
	})

	t.Run("Code", func(t *testing.T) {
		got := Code("fn()")
		if got == "" {
			t.Error("Code returned empty string")
		}
		if !strings.Contains(got, "fn()") {
			t.Errorf("Code should contain input, got %q", got)
		}
	})
}

func TestAdjustDimensions(t *testing.T) {
	tests := []struct {
		name           string
		width, height  int
		wantContentMin int
		wantTargetMin  int
		wantHeightMin  int
	}{
		{"small terminal", 40, 12, 20, 40, 12},
		{"large terminal", 200, 80, 182, 190, 76},
		{"tiny terminal", 20, 5, 20, 40, 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentW, targetW, targetH := AdjustDimensions(tt.width, tt.height)
			if contentW < tt.wantContentMin {
				t.Errorf("contentW = %d, want >= %d", contentW, tt.wantContentMin)
			}
			if targetW < tt.wantTargetMin {
				t.Errorf("targetW = %d, want >= %d", targetW, tt.wantTargetMin)
			}
			if targetH < tt.wantHeightMin {
				t.Errorf("targetH = %d, want >= %d", targetH, tt.wantHeightMin)
			}
		})
	}
}

func TestDivider(t *testing.T) {
	d := Divider(20)
	if len(d) < 20 {
		t.Errorf("Divider(20) length = %d, want >= 20", len(d))
	}

	_ = Divider(0) // should not panic
}
