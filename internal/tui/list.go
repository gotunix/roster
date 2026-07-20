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
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListItem represents a node in the dynamic tree view.
type ListItem struct {
	ID           string
	Title        string
	Type         string
	TypeColor    lipgloss.Color
	Status       string
	Progress     float64 // 0.0 to 1.0
	ProgressText string
	Expanded     bool
	Children     []*ListItem
}

// FlatRow represents a flattened node prepared for scrolling and cursor selection.
type FlatRow struct {
	Depth       int
	Expanded    bool
	HasChildren bool
	Item        *ListItem
}

// TreeListModel represents a generic, interactive tree-grid layout.
type TreeListModel struct {
	Title          string
	Items          []*ListItem
	Cursor         int
	WidthPct       float64
	HeightPct      float64
	terminalWidth  int
	terminalHeight int
	Styles         Styles
}

// NewTreeList creates an initialized TreeListModel with default global styles.
func NewTreeList(title string, items []*ListItem) TreeListModel {
	return TreeListModel{
		Title:     title,
		Items:     items,
		Cursor:    0,
		WidthPct:  0.95,
		HeightPct: 0.95,
		Styles:    GlobalStyles,
	}
}

// WithStyles sets a custom Styles configuration for the tree list.
func (m TreeListModel) WithStyles(s Styles) TreeListModel {
	m.Styles = s
	return m
}

// GetFlatRows flattens the nested tree structure according to expansion states.
func (m *TreeListModel) GetFlatRows() []FlatRow {
	var rows []FlatRow
	var flatten func(items []*ListItem, depth int)
	flatten = func(items []*ListItem, depth int) {
		for _, item := range items {
			hasChildren := len(item.Children) > 0
			rows = append(rows, FlatRow{
				Depth:       depth,
				Expanded:    item.Expanded,
				HasChildren: hasChildren,
				Item:        item,
			})
			if item.Expanded && hasChildren {
				flatten(item.Children, depth+1)
			}
		}
	}
	flatten(m.Items, 0)
	return rows
}

// Init initializes the tree list.
func (m TreeListModel) Init() tea.Cmd {
	return nil
}

// Update processes key events and size messages.
func (m TreeListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			flatRows := m.GetFlatRows()
			if m.Cursor < len(flatRows)-1 {
				m.Cursor++
			}
		case "enter":
			flatRows := m.GetFlatRows()
			if len(flatRows) > 0 {
				row := flatRows[m.Cursor]
				row.Item.Expanded = !row.Item.Expanded
			}
		}
	}
	return m, nil
}

func (m TreeListModel) renderHeader(widths []int) string {
	cells := []string{}
	headers := []string{"NAME", "TYPE", "STATUS", "PROGRESS"}

	headerBg := GlobalTheme.Primary
	headerFg := GlobalTheme.Base

	for i, col := range headers {
		w := widths[i]
		var styled string
		if len(col) > w {
			styled = m.Styles.HeaderStyle.Render(col[:w-3] + "...")
		} else {
			styled = m.Styles.HeaderStyle.Render(col) + lipgloss.NewStyle().Background(headerBg).Render(strings.Repeat(" ", w-len(col)))
		}
		cells = append(cells, styled)
	}

	prefix := lipgloss.NewStyle().Background(headerBg).Render("  ")
	separator := lipgloss.NewStyle().Foreground(headerFg).Background(headerBg).Render(" │ ")
	middle := prefix + strings.Join(cells, separator)

	var parts []string
	for _, w := range widths {
		parts = append(parts, strings.Repeat("─", w))
	}
	borderStyle := lipgloss.NewStyle().Foreground(headerFg).Background(headerBg)

	topBorder := borderStyle.Render("──" + strings.Join(parts, "─┬─"))
	botBorder := borderStyle.Render("──" + strings.Join(parts, "─┼─"))

	return topBorder + "\n" + middle + "\n" + botBorder
}

func (m TreeListModel) formatStatusCell(status string, width int, isCursor bool) string {
	var bg lipgloss.Color
	var fg lipgloss.Color = GlobalTheme.Base

	switch status {
	case "ACTIVE", "In Progress", "IN-PROGRESS":
		bg = GlobalTheme.Accent
	case "COMPLETED", "Done", "SUCCESS":
		bg = GlobalTheme.Success
	case "CANCELLED", "ERROR":
		bg = GlobalTheme.Error
	case "BACKLOG", "Todo":
		bg = GlobalTheme.Overlay
	default:
		bg = GlobalTheme.Overlay
	}

	badge := lipgloss.NewStyle().Foreground(fg).Background(bg).Bold(true).Padding(0, 1).Render(status)
	visibleLen := lipgloss.Width(badge)

	if visibleLen > width {
		return badge[:width]
	}

	var style lipgloss.Style
	rowStyle := m.Styles.LabelStyle.Copy()
	if isCursor {
		rowStyle = m.Styles.ActiveButtonStyle.Copy()
	}

	style = rowStyle.Width(width).Align(lipgloss.Center)
	return style.Render(badge)
}

func (m TreeListModel) renderProgressBar(pct float64, text string, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}

	filledLen := int(float64(width) * pct)
	emptyLen := width - filledLen

	filled := lipgloss.NewStyle().Foreground(GlobalTheme.Success).Render(strings.Repeat("█", filledLen))
	empty := lipgloss.NewStyle().Foreground(GlobalTheme.Overlay).Render(strings.Repeat("░", emptyLen))

	bar := filled + empty

	if text != "" {
		textStyle := m.Styles.HelperStyle
		return fmt.Sprintf("%s %s", bar, textStyle.Render(text))
	}
	return bar
}

func getRowTitle(row FlatRow) string {
	icon := ""
	if row.HasChildren {
		if row.Expanded {
			icon = "▼ "
		} else {
			icon = "▶ "
		}
	} else {
		icon = "  "
	}
	indent := strings.Repeat("  ", row.Depth)
	return indent + icon + row.Item.Title
}

func (m TreeListModel) renderRow(title, typeStr, statusCell, progress string, widths []int, isCursor bool, rowStyle lipgloss.Style) string {
	formatCol := func(txt string, w int, style lipgloss.Style) string {
		var styled string
		var spaces string

		visibleLen := lipgloss.Width(txt)

		if visibleLen > w {
			styled = style.Render(txt[:w-3] + "...")
		} else {
			spaces = strings.Repeat(" ", w-visibleLen)
			styled = style.Render(txt) + style.Render(spaces)
		}
		return styled
	}

	if isCursor {
		rowStyle = lipgloss.NewStyle().Foreground(GlobalTheme.Text).Background(GlobalTheme.Overlay).Bold(true)
	}

	c1 := formatCol(title, widths[0], rowStyle)
	c2 := formatCol(typeStr, widths[1], rowStyle)
	c3 := statusCell
	c4 := formatCol(progress, widths[3], rowStyle)

	prefix := "  "
	if isCursor {
		prefix = lipgloss.NewStyle().Foreground(GlobalTheme.Success).Bold(true).Render("▶ ")
	}
	prefix = rowStyle.Render(prefix)
	separator := rowStyle.Render(" │ ")

	return prefix + c1 + separator + c2 + separator + c3 + separator + c4
}

// View outputs the complete visual layout of the tree list.
func (m TreeListModel) View() string {
	if m.terminalWidth == 0 || m.terminalHeight == 0 {
		return ""
	}

	targetWidth := int(float64(m.terminalWidth) * m.WidthPct)
	targetHeight := int(float64(m.terminalHeight) * m.HeightPct)

	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.HiddenBorder()).
		Height(targetHeight)

	headerWidth := targetWidth
	widths := []int{
		int(float64(headerWidth) * 0.4),
		int(float64(headerWidth) * 0.2),
		int(float64(headerWidth) * 0.2),
		int(float64(headerWidth)*0.2) - 6,
	}

	flatRows := m.GetFlatRows()
	var listSections []string

	for i, row := range flatRows {
		isCursor := i == m.Cursor

		titleCell := getRowTitle(row)

		// Row style varies by depth
		rowStyle := m.Styles.LabelStyle
		if row.Depth > 0 {
			rowStyle = m.Styles.HelperStyle
		}

		typeColor := row.Item.TypeColor
		if typeColor == "" {
			typeColor = GlobalTheme.Accent
		}
		badgeStyle := lipgloss.NewStyle().Foreground(GlobalTheme.Base).Background(typeColor).Bold(true).Padding(0, 1)
		typeCell := badgeStyle.Width(widths[1]).Align(lipgloss.Center).Render(strings.ToUpper(row.Item.Type))

		statusCell := m.formatStatusCell(row.Item.Status, widths[2], isCursor)

		var progressCell string
		if row.Item.ProgressText != "" || row.Item.Progress > 0 {
			progressCell = m.renderProgressBar(row.Item.Progress, row.Item.ProgressText, widths[3])
		} else {
			progressCell = ""
		}

		listSections = append(listSections, m.renderRow(titleCell, typeCell, statusCell, progressCell, widths, isCursor, rowStyle))
	}

	listSections = append(listSections, lipgloss.NewStyle().Background(GlobalTheme.Base).Render(""))

	paddedList := lipgloss.NewStyle().
		Background(GlobalTheme.Base).
		Width(headerWidth).
		Height(targetHeight - 8).
		Render(strings.Join(listSections, "\n"))

	cursorIndicator := ""
	if len(flatRows) > 0 {
		cursorIndicator = fmt.Sprintf(" [%d/%d]", m.Cursor+1, len(flatRows))
	}
	headerFull := m.Styles.HeaderStyle.Copy().Width(headerWidth).Align(lipgloss.Center).Render(" " + strings.ToUpper(m.Title) + cursorIndicator + " ")
	headerRow := lipgloss.NewStyle().Padding(0, 2).Background(GlobalTheme.Primary).Render(m.renderHeader(widths))

	helpStr := " ↑/↓ or j/k: move • enter: toggle expansion • q/esc: quit "
	footer := m.Styles.HeaderStyle.Copy().Width(headerWidth).Align(lipgloss.Center).Render(helpStr)

	finalContent := headerFull + "\n" + headerRow + "\n" + paddedList + "\n" + footer

	activeWindow := windowStyle.Copy().Width(targetWidth)
	result := activeWindow.Render(finalContent)

	return lipgloss.Place(m.terminalWidth, m.terminalHeight, lipgloss.Center, lipgloss.Top, result, lipgloss.WithWhitespaceBackground(GlobalTheme.Base))
}
