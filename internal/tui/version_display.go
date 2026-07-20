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
//                     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the               //
//              GNU General Public License for more details.                                       //
//                                                                                                 //
//              You should have received a copy of the GNU General Public License                  //
//              along with this program.  If not, see <https://www.gnu.org/licenses/>.             //
// =============================================================================================== //

package tui

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// VersionStyle bundles all styling for version display rendering.
// Each field is a pre-configured lipgloss.Style; callers should use
// .Copy() to avoid mutating shared styles.
type VersionStyle struct {
	Title lipgloss.Style
	Sep   lipgloss.Style
	Label lipgloss.Style
	Value lipgloss.Style
	Bg    lipgloss.Style
	Deps  lipgloss.Style
}

// RenderVersionConfig configures a complete version display.
type RenderVersionConfig struct {
	AppName string
	Version string
	Width   int
	Style   VersionStyle
}

// VersionHeader renders a styled title bar (e.g. "VERSION DETAILS: APP V0.1.0").
func VersionHeader(text string, width int, s lipgloss.Style) string {
	return s.Copy().Width(width).Render(text)
}

// SplitDivider renders a horizontal divider with a ┬ junction at the given split point.
func SplitDivider(width, splitAt int, s lipgloss.Style) string {
	left := splitAt
	right := width - left - 1
	if left < 0 {
		left = 0
	}
	if right < 0 {
		right = 0
	}
	return s.Render(strings.Repeat("─", left) + "┬" + strings.Repeat("─", right))
}

// TwoColumnRow renders a two-column grid row with a │ separator.
// To align values vertically, callers should set a fixed Width on labelStyle
// (e.g. lipgloss.NewStyle().Width(7)) so every label renders at the same visual width.
func TwoColumnRow(label1, val1, label2, val2 string, col1W, col2W int, labelStyle, valStyle, bgStyle, sepStyle lipgloss.Style) string {
	formatField := func(label, val string, colW int) string {
		renderedLabel := labelStyle.Render(label)
		lW := lipgloss.Width(renderedLabel)
		vW := colW - lW - 1
		if vW < 0 {
			vW = 0
		}
		if runewidth.StringWidth(val) > vW {
			val = runewidth.Truncate(val, vW, "")
		}
		valW := runewidth.StringWidth(val)
		return renderedLabel + bgStyle.Render(" ") + valStyle.Render(val) + bgStyle.Render(strings.Repeat(" ", vW-valW))
	}
	return formatField(label1, val1, col1W) + sepStyle.Render(" │ ") + formatField(label2, val2, col2W)
}

// RenderVersion renders a complete version display string including build info
// and dependencies. It calls debug.ReadBuildInfo internally.
func RenderVersion(cfg RenderVersionConfig) string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "Error loading version information"
	}

	col1W := (cfg.Width - 3) / 2
	col2W := cfg.Width - col1W - 3
	if col1W < 30 {
		col1W = 30
	}
	if col2W < 30 {
		col2W = 30
	}

	appVer := fmt.Sprintf("%s %s", cfg.AppName, cfg.Version)
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(VersionHeader(fmt.Sprintf("VERSION DETAILS: %s", strings.ToUpper(appVer)), cfg.Width, cfg.Style.Title))
	sb.WriteString("\n")
	sb.WriteString(SplitDivider(cfg.Width, col1W+1, cfg.Style.Sep))
	sb.WriteString("\n")
	sb.WriteString(TwoColumnRow("App:", appVer, "Build:", bi.Main.Version, col1W, col2W, cfg.Style.Label, cfg.Style.Value, cfg.Style.Bg, cfg.Style.Sep))
	sb.WriteString("\n")
	sb.WriteString(TwoColumnRow("Go:", bi.GoVersion, "OS:", runtime.GOOS, col1W, col2W, cfg.Style.Label, cfg.Style.Value, cfg.Style.Bg, cfg.Style.Sep))
	sb.WriteString("\n")
	sb.WriteString(TwoColumnRow("Arch:", runtime.GOARCH, "Module:", bi.Main.Path, col1W, col2W, cfg.Style.Label, cfg.Style.Value, cfg.Style.Bg, cfg.Style.Sep))
	sb.WriteString("\n\n")

	sb.WriteString(cfg.Style.Deps.Render("Dependencies:"))
	sb.WriteString("\n")
	if len(bi.Deps) > 0 {
		for _, dep := range bi.Deps {
			sb.WriteString(fmt.Sprintf("  • %s %s\n", dep.Path, dep.Version))
		}
	} else {
		sb.WriteString("  (no dependencies)\n")
	}

	return sb.String()
}
