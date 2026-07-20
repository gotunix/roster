// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The Roster Authors
// =============================================================================================== //
//                                                                                                 //
//                   /$$$$$$$                        /$$                                           //
//                  | $$__  $$                      | $$                                           //
//                  | $$  \ $$  /$$$$$$   /$$$$$$$ /$$$$$$    /$$$$$$   /$$$$$$                    //
//                  | $$$$$$$/ /$$__  $$ /$$_____/|_  $$_/   /$$__  $$ /$$__  $$                   //
//                  | $$__  $$| $$  \ $$|  $$$$$$   | $$    | $$$$$$$$| $$  \__/                   //
//                  | $$  \ $$| $$  | $$ \____  $$  | $$ /$$| $$_____/| $$                         //
//                  | $$  | $$|  $$$$$$/ /$$$$$$$/  |  $$$$/|  $$$$$$$| $$                         //
//                  |__/  |__/ \______/ |_______/    \___/   \_______/|__/                         //
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

package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

func GetTerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 100 // Fallback
	}
	// Ensure a minimum width for readability
	if w < 80 {
		return 80
	}
	return w
}

var (
	Version = "v0.1.0"

	Subtle  = lipgloss.AdaptiveColor{Light: "#D9D9D9", Dark: "#383838"}
	Magenta = lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}
	Cyan    = lipgloss.AdaptiveColor{Light: "#02BAEF", Dark: "#02BAEF"}
	Green   = lipgloss.AdaptiveColor{Light: "#02BA59", Dark: "#02BA59"}
	Yellow  = lipgloss.AdaptiveColor{Light: "#EAD027", Dark: "#EAD027"}
	Red     = lipgloss.AdaptiveColor{Light: "#FF4672", Dark: "#FF4672"}
	Gray    = lipgloss.AdaptiveColor{Light: "#828282", Dark: "#828282"}

	BoldStyle  = lipgloss.NewStyle().Bold(true)
	TitleStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFFFF")).
			Bold(true)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(Magenta).
			Bold(true).
			Underline(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	StatusStyle = func(status string) lipgloss.Style {
		s := lipgloss.NewStyle().Bold(true)
		switch strings.ToUpper(status) {
		case "ACTIVE":
			return s.Foreground(Green)
		case "BACKLOG":
			return s.Foreground(Cyan)
		case "COMPLETED":
			return s.Foreground(Magenta)
		case "IN-PROGRESS":
			return s.Foreground(Yellow)
		case "CANCELLED":
			return s.Foreground(Gray)
		default:
			return s.Foreground(Gray)
		}
	}

	BorderStyle = lipgloss.NewStyle().
			Foreground(Subtle).
			Border(lipgloss.NormalBorder(), true, false, true, false)

		//	UsageStyle       = lipgloss.NewStyle().Padding(1, 2)
		//	CommandStyle     = lipgloss.NewStyle().Foreground(Magenta).Bold(true).Width(10)
		//	DescriptionStyle = lipgloss.NewStyle().Foreground(Gray)
	UsageStyle       = lipgloss.NewStyle().Padding(1, 2)
	CommandStyle     = lipgloss.NewStyle().Foreground(Magenta).Bold(true).Width(10)
	DescriptionStyle = lipgloss.NewStyle().Foreground(Gray)

	LogoStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)
	HelpTitleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#FFFFFF")).Padding(0, 1).Bold(true).MarginBottom(1)
	HelpDescStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#AFB1B6")).Italic(true).MarginBottom(1)
	HelpSectionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true).MarginTop(1)
	HelpFlagStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#AFB1B6"))
)

// SuccessMsg returns a formatted success message
func SuccessMsg(format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)
	return BoldStyle.Foreground(Green).Render("✔ " + msg)
}

// ErrorMsg returns a formatted error message
func ErrorMsg(format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)
	return BoldStyle.Foreground(Red).Render("✘ Error: " + msg)
}

// WarningMsg returns a formatted warning message
func WarningMsg(format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)
	return BoldStyle.Foreground(lipgloss.Color("#FFB200")).Render("⚠ Warning: " + msg)
}
