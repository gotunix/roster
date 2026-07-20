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

func TestVersionHeader(t *testing.T) {
	s := lipgloss.NewStyle().Bold(true)
	h := VersionHeader(" VERSION DETAILS: TEST ", 40, s)
	if h == "" {
		t.Error("VersionHeader returned empty string")
	}
	if !strings.Contains(h, "TEST") {
		t.Errorf("VersionHeader should contain text, got %q", h)
	}
}

func TestSplitDivider(t *testing.T) {
	s := lipgloss.NewStyle()
	d := SplitDivider(80, 40, s)
	if d == "" {
		t.Error("SplitDivider returned empty string")
	}
	if !strings.Contains(d, "┬") {
		t.Errorf("SplitDivider should contain ┬, got %q", d)
	}
}

func TestTwoColumnRow(t *testing.T) {
	labelStyle := lipgloss.NewStyle().Bold(true)
	valStyle := lipgloss.NewStyle()
	bgStyle := lipgloss.NewStyle()
	sepStyle := lipgloss.NewStyle()

	r := TwoColumnRow("Left:", "a", "Right:", "b", 30, 30, labelStyle, valStyle, bgStyle, sepStyle)
	if r == "" {
		t.Error("TwoColumnRow returned empty string")
	}
	if !strings.Contains(r, "Left:") || !strings.Contains(r, "Right:") {
		t.Errorf("TwoColumnRow should contain both labels, got %q", r)
	}
	if !strings.Contains(r, "│") {
		t.Errorf("TwoColumnRow should contain │ separator, got %q", r)
	}
}

func TestRenderVersion(t *testing.T) {
	theme := DefaultTheme()
	vs := VersionStyle{
		Title: lipgloss.NewStyle().
			Background(theme.Primary).
			Foreground(theme.Base).
			Padding(0, 1).
			Bold(true),
		Sep:   lipgloss.NewStyle().Foreground(theme.Base),
		Label: lipgloss.NewStyle().Foreground(theme.Text).Bold(true),
		Value: lipgloss.NewStyle().Foreground(theme.Text),
		Bg:    lipgloss.NewStyle(),
		Deps:  lipgloss.NewStyle().Foreground(theme.Primary).Bold(true),
	}
	result := RenderVersion(RenderVersionConfig{
		AppName: "TestApp",
		Version: "v1.0.0",
		Width:   80,
		Style:   vs,
	})
	if result == "" {
		t.Fatal("RenderVersion returned empty string")
	}
	if strings.Contains(result, "Error") {
		t.Skip("debug.ReadBuildInfo not available in test mode")
	}
	if !strings.Contains(result, "TestApp") {
		t.Errorf("RenderVersion should contain app name, got %q", result)
	}
	if !strings.Contains(result, "App:") || !strings.Contains(result, "Build:") {
		t.Errorf("RenderVersion should contain grid labels, got %q", result)
	}
}
