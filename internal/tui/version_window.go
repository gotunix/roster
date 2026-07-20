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
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// VersionWindowModel is a Bubble Tea component that displays application version info
// and scrollable compile-time module dependencies in a structured box.
// It is fully theme-agnostic and customizable.
type VersionWindowModel struct {
	AppName  string
	Version  string
	Commit   string
	Date     string
	Styles   Styles
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

// NewVersionWindow creates an initialized VersionWindowModel with default global styles.
func NewVersionWindow(appName, version, commit, date string) VersionWindowModel {
	return VersionWindowModel{
		AppName: appName,
		Version: version,
		Commit:  commit,
		Date:    date,
		Styles:  GlobalStyles,
	}
}

// WithStyles sets a custom Styles configuration for the version window, ensuring dynamic theme support.
func (m VersionWindowModel) WithStyles(s Styles) VersionWindowModel {
	m.Styles = s
	return m
}

// Init initializes the viewport.
func (m VersionWindowModel) Init() tea.Cmd {
	return nil
}

// Update handles resizing and viewport scroll key inputs.
func (m VersionWindowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Get frame height with explicit Padding left/right = 0
		winStyle := m.Styles.WindowStyle.Copy().PaddingTop(0).PaddingBottom(1).PaddingLeft(0).PaddingRight(0)
		borderWidth, borderHeight := winStyle.GetFrameSize()

		// Chrome height calculation:
		// 1 (top header) + 1 (footer) + borderHeight (window box borders/padding)
		// + 4 (two-column layout metadata rows) + 1 (divider) + 1 (dependencies title) + 1 (safety buffer)
		chromeHeight := 9 + borderHeight

		viewportHeight := msg.Height - chromeHeight
		if viewportHeight < 3 {
			viewportHeight = 3 // Minimum scroll viewport height
		}

		// Available width inside the window (borders + padding left/right)
		// Box outer width is msg.Width - 2 to prevent right border cutoff
		boxWidth := msg.Width - 2
		// We leave 2 spaces padding on left and right for dependencies inside the viewport
		viewportWidth := boxWidth - borderWidth - 4
		if viewportWidth < 20 {
			viewportWidth = 20
		}

		if !m.ready {
			m.viewport = viewport.New(viewportWidth, viewportHeight)
			m.ready = true
		} else {
			m.viewport.Width = viewportWidth
			m.viewport.Height = viewportHeight
		}

		// Populate dependencies content
		m.viewport.SetContent(m.getDependenciesList())
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the complete TUI box.
func (m VersionWindowModel) View() string {
	if !m.ready {
		return "Initializing version display..."
	}

	winStyle := m.Styles.WindowStyle.Copy().PaddingTop(0).PaddingBottom(1).PaddingLeft(0).PaddingRight(0)
	borderWidth, borderHeight := winStyle.GetFrameSize()

	// Outer width is m.width - 2 to prevent right border cutoff
	boxWidth := m.width - 2
	if boxWidth < 38 {
		boxWidth = 38
	}

	// Inner width is the space inside the borders and padding
	innerWidth := boxWidth - borderWidth
	if innerWidth < 30 {
		innerWidth = 30
	}

	// Inner height of the box content area (excluding borders/padding)
	// Outer height of the box is m.height - 3 (leaving room for header, footer, and safety line)
	innerHeight := m.height - borderHeight - 3
	if innerHeight < 10 {
		innerHeight = 10
	}

	// 1. Render Top Header
	header := m.Styles.HeaderStyle.Copy().Width(m.width).Align(lipgloss.Center).Render(" " + strings.ToUpper(m.AppName) + " ")

	// Fetch module & Go runtime properties
	modPath := "gotunix.net/roster"
	goVer := "unknown"
	if bi, ok := debug.ReadBuildInfo(); ok {
		if bi.Main.Path != "" {
			modPath = bi.Main.Path
		}
		if bi.GoVersion != "" {
			goVer = bi.GoVersion
		}
	}

	// Calculate column widths
	col1W := (innerWidth - 3) / 2
	col2W := innerWidth - col1W - 3

	// Helper function to format a single column cell
	formatCell := func(label string, labelWidth int, value string, colWidth int) string {
		// Render styled label padded to a fixed width
		labelText := label + strings.Repeat(" ", labelWidth-len(label)) + "  :  "
		styledLabel := m.Styles.LabelStyle.Render(labelText)
		
		// Visible width available for value: colWidth - labelTextLen
		valLimit := colWidth - len(labelText)
		if valLimit < 0 {
			valLimit = 0
		}

		visibleVal := value
		if len(visibleVal) > valLimit {
			visibleVal = visibleVal[:valLimit]
		}

		styledVal := m.Styles.HelperStyle.Render(visibleVal)
		padding := colWidth - len(labelText) - len(visibleVal)
		if padding < 0 {
			padding = 0
		}

		return styledLabel + styledVal + strings.Repeat(" ", padding)
	}

	// Style for inner dividers to match the window borders exactly
	borderStyle := lipgloss.NewStyle().Foreground(m.Styles.HeaderStyle.GetBackground())

	// 2. Build Box Content rows
	var lines []string
	// Top padding line with vertical separator touching the top border
	lines = append(lines, "  "+strings.Repeat(" ", col1W-2)+" "+borderStyle.Render("│")+" "+strings.Repeat(" ", col2W-2)+"  ")
	lines = append(lines, "  "+formatCell("Application", 11, m.AppName, col1W-2)+" "+borderStyle.Render("│")+" "+formatCell("Commit", 6, m.Commit, col2W-2)+"  ")
	lines = append(lines, "  "+formatCell("Version", 11, m.Version, col1W-2)+" "+borderStyle.Render("│")+" "+formatCell("Module", 6, modPath, col2W-2)+"  ")
	lines = append(lines, "  "+formatCell("Built", 11, m.Date, col1W-2)+" "+borderStyle.Render("│")+" "+formatCell("OS", 6, runtime.GOOS, col2W-2)+"  ")
	lines = append(lines, "  "+formatCell("GO", 11, goVer, col1W-2)+" "+borderStyle.Render("│")+" "+formatCell("Arch", 6, runtime.GOARCH, col2W-2)+"  ")
	// Bottom divider with junction character T-junction pointing up (┴)
	dividerLine := strings.Repeat("─", col1W+1) + "┴" + strings.Repeat("─", col2W+1)
	lines = append(lines, borderStyle.Render(dividerLine))
	lines = append(lines, "  "+m.Styles.LabelStyle.Render("Dependencies:"))

	headerContent := strings.Join(lines, "\n")

	// Viewport content with left margin of 2
	viewportContent := m.viewport.View()
	viewportContent = lipgloss.NewStyle().MarginLeft(2).Render(viewportContent)

	// Assemble complete body to go inside the box
	dialogBody := headerContent + "\n" + viewportContent

	// Render window specifying inner width and height
	windowContent := winStyle.Width(innerWidth).Height(innerHeight).Render(dialogBody)

	// Dynamically center the box horizontally
	marginLeft := (m.width - boxWidth) / 2
	if marginLeft < 0 {
		marginLeft = 0
	}
	windowContent = lipgloss.NewStyle().MarginLeft(marginLeft).Render(windowContent)

	// 3. Render Footer
	helpText := " ↑/↓ or k/j: scroll dependencies • q/esc: exit "
	footer := m.Styles.HeaderStyle.Copy().Width(m.width).Align(lipgloss.Center).Render(helpText)

	// 4. Assemble everything in vertical sequence
	rawView := header + "\n" + windowContent + "\n" + footer

	// Post-process the top border line to inject a ┬ junction aligned with the vertical separator
	viewLines := strings.Split(rawView, "\n")
	if len(viewLines) > 1 {
		targetIdx := marginLeft + 1 + col1W + 1
		viewLines[1] = replaceVisibleChar(viewLines[1], targetIdx, '┬')
	}

	return strings.Join(viewLines, "\n")
}

// getDependenciesList reads modules info at runtime and returns a formatted list.
func (m VersionWindowModel) getDependenciesList() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok || len(bi.Deps) == 0 {
		return "    (no dependencies or build info unavailable)"
	}

	var builder strings.Builder
	for _, dep := range bi.Deps {
		builder.WriteString(fmt.Sprintf("    • %s %s\n", dep.Path, dep.Version))
	}
	return builder.String()
}

// replaceVisibleChar replaces a character at a specific visible (non-ANSI) column index.
func replaceVisibleChar(s string, targetVisibleIdx int, newChar rune) string {
	var builder strings.Builder
	visibleIdx := 0
	inEscape := false
	
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\x1b' {
			inEscape = true
			builder.WriteRune(r)
			continue
		}
		if inEscape {
			builder.WriteRune(r)
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		
		if visibleIdx == targetVisibleIdx {
			builder.WriteRune(newChar)
		} else {
			builder.WriteRune(r)
		}
		visibleIdx++
	}
	return builder.String()
}

// RenderVersionStatic renders a static version window styled box for printing to stdout.
func RenderVersionStatic(appName, version, commit, date string) string {
	winStyle := GlobalStyles.WindowStyle.Copy().PaddingTop(0).PaddingBottom(1).PaddingLeft(0).PaddingRight(0)
	borderWidth, _ := winStyle.GetFrameSize()

	boxWidth := 80
	innerWidth := boxWidth - borderWidth

	// Calculate column widths
	col1W := (innerWidth - 3) / 2
	col2W := innerWidth - col1W - 3

	// Fetch module & Go runtime properties
	modPath := "gotunix.net/roster"
	goVer := "unknown"
	if bi, ok := debug.ReadBuildInfo(); ok {
		if bi.Main.Path != "" {
			modPath = bi.Main.Path
		}
		if bi.GoVersion != "" {
			goVer = bi.GoVersion
		}
	}

	// Helper function to format a single column cell
	formatCell := func(label string, labelWidth int, value string, colWidth int) string {
		labelText := label + strings.Repeat(" ", labelWidth-len(label)) + "  :  "
		styledLabel := GlobalStyles.LabelStyle.Render(labelText)
		
		valLimit := colWidth - len(labelText)
		if valLimit < 0 {
			valLimit = 0
		}

		visibleVal := value
		if len(visibleVal) > valLimit {
			visibleVal = visibleVal[:valLimit]
		}

		styledVal := GlobalStyles.HelperStyle.Render(visibleVal)
		padding := colWidth - len(labelText) - len(visibleVal)
		if padding < 0 {
			padding = 0
		}

		return styledLabel + styledVal + strings.Repeat(" ", padding)
	}

	// Style for inner dividers to match the window borders exactly
	borderStyle := lipgloss.NewStyle().Foreground(GlobalStyles.HeaderStyle.GetBackground())

	// 2. Build Box Content rows
	var lines []string
	// Top padding line with vertical separator touching the top border
	lines = append(lines, "  "+strings.Repeat(" ", col1W-2)+" "+borderStyle.Render("│")+" "+strings.Repeat(" ", col2W-2)+"  ")
	lines = append(lines, "  "+formatCell("Application", 11, appName, col1W-2)+" "+borderStyle.Render("│")+" "+formatCell("Commit", 6, commit, col2W-2)+"  ")
	lines = append(lines, "  "+formatCell("Version", 11, version, col1W-2)+" "+borderStyle.Render("│")+" "+formatCell("Module", 6, modPath, col2W-2)+"  ")
	lines = append(lines, "  "+formatCell("Built", 11, date, col1W-2)+" "+borderStyle.Render("│")+" "+formatCell("OS", 6, runtime.GOOS, col2W-2)+"  ")
	lines = append(lines, "  "+formatCell("GO", 11, goVer, col1W-2)+" "+borderStyle.Render("│")+" "+formatCell("Arch", 6, runtime.GOARCH, col2W-2)+"  ")
	// Bottom divider with junction character T-junction pointing up (┴)
	dividerLine := strings.Repeat("─", col1W+1) + "┴" + strings.Repeat("─", col2W+1)
	lines = append(lines, borderStyle.Render(dividerLine))
	lines = append(lines, "  "+GlobalStyles.LabelStyle.Render("Dependencies:"))

	// Read all dependencies untruncated
	var depsBuilder strings.Builder
	if bi, ok := debug.ReadBuildInfo(); ok && len(bi.Deps) > 0 {
		for _, dep := range bi.Deps {
			depsBuilder.WriteString(fmt.Sprintf("    • %s %s\n", dep.Path, dep.Version))
		}
	} else {
		depsBuilder.WriteString("    (no dependencies or build info unavailable)\n")
	}

	headerContent := strings.Join(lines, "\n")
	viewportContent := depsBuilder.String()
	
	// Add left margin of 2 to the dependencies list
	viewportContent = lipgloss.NewStyle().MarginLeft(2).Render(viewportContent)

	dialogBody := headerContent + "\n" + viewportContent

	// Render window (unconstrained height)
	windowContent := winStyle.Width(innerWidth).Render(dialogBody)

	// Add 1 space left margin to center the box
	marginLeft := 1
	windowContent = lipgloss.NewStyle().MarginLeft(marginLeft).Render(windowContent)

	// 1. Render Top Header
	header := GlobalStyles.HeaderStyle.Copy().Width(boxWidth).Align(lipgloss.Center).Render(" " + strings.ToUpper(appName) + " ")
	header = lipgloss.NewStyle().MarginLeft(marginLeft).Render(header)

	rawView := header + "\n" + windowContent

	// Post-process the top border line to inject a ┬ junction aligned with the vertical separator
	viewLines := strings.Split(rawView, "\n")
	if len(viewLines) > 1 {
		targetIdx := marginLeft + 1 + col1W + 1
		viewLines[1] = replaceVisibleChar(viewLines[1], targetIdx, '┬')
	}

	return strings.Join(viewLines, "\n") + "\n"
}
