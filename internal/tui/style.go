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

	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Base styles (use GlobalTheme by default)
// ---------------------------------------------------------------------------

// Styles bundles all the styles for components.
type Styles struct {
	WindowStyle       lipgloss.Style
	TitleStyle        lipgloss.Style
	LabelStyle        lipgloss.Style
	FocusedLabelStyle lipgloss.Style
	HelperStyle       lipgloss.Style
	ErrorStyle        lipgloss.Style
	ButtonStyle       lipgloss.Style
	ActiveButtonStyle lipgloss.Style
	SuccessTitleStyle lipgloss.Style
	HeaderStyle       lipgloss.Style
}

// GenerateStyles creates a Styles struct initialized with the colors of the given Theme.
func GenerateStyles(t Theme) Styles {
	return Styles{
		WindowStyle: lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary).
			Background(t.Base),
		TitleStyle: lipgloss.NewStyle().
			Foreground(t.Base).
			Background(t.Primary).
			Padding(0, 2).
			Bold(true).
			MarginBottom(1),
		LabelStyle: lipgloss.NewStyle().
			Foreground(t.Text).
			Bold(true),
		FocusedLabelStyle: lipgloss.NewStyle().
			Foreground(t.Success).
			Bold(true),
		HelperStyle: lipgloss.NewStyle().
			Foreground(t.Overlay).
			Italic(true),
		ErrorStyle: lipgloss.NewStyle().
			Foreground(t.Error).
			Bold(true),
		ButtonStyle: lipgloss.NewStyle().
			Foreground(t.Text).
			Background(t.Overlay).
			Padding(0, 3).
			MarginTop(1),
		ActiveButtonStyle: lipgloss.NewStyle().
			Foreground(t.Base).
			Background(t.Success).
			Padding(0, 3).
			Bold(true).
			MarginTop(1),
		SuccessTitleStyle: lipgloss.NewStyle().
			Foreground(t.Base).
			Background(t.Success).
			Padding(0, 2).
			Bold(true).
			MarginBottom(1),
		HeaderStyle: lipgloss.NewStyle().
			Foreground(t.Base).
			Background(t.Primary).
			Bold(true),
	}
}

var (
	// GlobalTheme is the active theme, defaults to Catppuccin Mocha.
	GlobalTheme = DefaultTheme()

	// GlobalStyles is the active styles, generated from GlobalTheme.
	GlobalStyles = GenerateStyles(GlobalTheme)
)

// Active styles references (updated via SetActiveTheme)
var (
	WindowStyle       = GlobalStyles.WindowStyle
	TitleStyle        = GlobalStyles.TitleStyle
	LabelStyle        = GlobalStyles.LabelStyle
	FocusedLabelStyle = GlobalStyles.FocusedLabelStyle
	HelperStyle       = GlobalStyles.HelperStyle
	ErrorStyle        = GlobalStyles.ErrorStyle
	ButtonStyle       = GlobalStyles.ButtonStyle
	ActiveButtonStyle = GlobalStyles.ActiveButtonStyle
	SuccessTitleStyle = GlobalStyles.SuccessTitleStyle
	HeaderStyle       = GlobalStyles.HeaderStyle
)

// SetActiveTheme updates the global theme and all dependent global styles.
func SetActiveTheme(t Theme) {
	GlobalTheme = t
	GlobalStyles = GenerateStyles(t)

	WindowStyle = GlobalStyles.WindowStyle
	TitleStyle = GlobalStyles.TitleStyle
	LabelStyle = GlobalStyles.LabelStyle
	FocusedLabelStyle = GlobalStyles.FocusedLabelStyle
	HelperStyle = GlobalStyles.HelperStyle
	ErrorStyle = GlobalStyles.ErrorStyle
	ButtonStyle = GlobalStyles.ButtonStyle
	ActiveButtonStyle = GlobalStyles.ActiveButtonStyle
	SuccessTitleStyle = GlobalStyles.SuccessTitleStyle
	HeaderStyle = GlobalStyles.HeaderStyle
}

// ---------------------------------------------------------------------------
// Semantic helper constructors
// ---------------------------------------------------------------------------

// KeyValue returns a styled line like "Key: Value" using text color.
func KeyValue(key, value string) string {
	return LabelStyle.Render(key+":") + " " + TextStyle().Render(value)
}

// TextStyle returns a plain style colored with the current theme's text color.
func TextStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GlobalTheme.Text)
}

// CatMochaTextStyle is a deprecated alias for TextStyle.
func CatMochaTextStyle() lipgloss.Style {
	return TextStyle()
}

// Success returns a styled success label.
func Success(text string) string {
	return SuccessTitleStyle.Render(text)
}

// Error returns a styled error message.
func Error(text string) string {
	return ErrorStyle.Render(text)
}

// Warning returns a styled warning message.
func Warning(text string) string {
	return lipgloss.NewStyle().Foreground(GlobalTheme.Warning).Bold(true).Render(text)
}

// Header returns a full-width header bar.
func Header(title string, width int) string {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Background(GlobalTheme.Primary).
		Foreground(GlobalTheme.Base).
		Bold(true).
		Render(" " + strings.ToUpper(title) + " ")
}

// Code returns an inline code span.
func Code(text string) string {
	return lipgloss.NewStyle().
		Foreground(GlobalTheme.Accent).
		Background(GlobalTheme.Overlay).
		Padding(0, 1).
		Render(text)
}
