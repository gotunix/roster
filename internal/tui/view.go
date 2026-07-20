package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"gotunix.net/roster/internal/store"
)

func (m MainTUIModel) renderErrorWindow(title, errorMsg string) string {
	header := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(" " + strings.ToUpper(title) + " ")

	boxWidth := m.terminalWidth - 2
	if boxWidth < 38 {
		boxWidth = 38
	}

	winStyle := WindowStyle.Copy().PaddingTop(0).PaddingBottom(1).PaddingLeft(0).PaddingRight(0)
	borderWidth, borderHeight := winStyle.GetFrameSize()

	innerWidth := boxWidth - borderWidth
	innerHeight := m.terminalHeight - borderHeight - 3
	if innerHeight < 8 {
		innerHeight = 8
	}

	dialogBody := "\n\n  " + ErrorStyle.Render(errorMsg) + "\n\n"
	boxContent := lipgloss.Place(innerWidth, innerHeight, lipgloss.Center, lipgloss.Center, dialogBody, lipgloss.WithWhitespaceBackground(GlobalTheme.Base))
	windowContent := winStyle.Render(boxContent)

	marginLeft := (m.terminalWidth - boxWidth) / 2
	if marginLeft < 0 {
		marginLeft = 0
	}
	windowContent = lipgloss.NewStyle().MarginLeft(marginLeft).Render(windowContent)

	helpText := " q/esc: back to main menu "
	footer := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(helpText)

	return header + "\n" + windowContent + "\n" + footer
}

func (m MainTUIModel) View() string {
	if m.state == StateVersion {
		return m.versionModel.View()
	}

	if m.state == StateVarsForm {
		return m.varsForm.View()
	}

	if m.state == StateTUIForm {
		return m.tuiForm.View()
	}

	if m.state == StateSyncMenu {
		return m.tuiForm.View()
	}

	if m.state == StateExportMenu {
		return m.tuiForm.View()
	}

	if m.state == StateSyncing {
		header := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(" SYNCING NETBOX INVENTORY ")
		winStyle := WindowStyle.Copy().PaddingTop(0).PaddingBottom(1).PaddingLeft(0).PaddingRight(0)
		borderWidth, borderHeight := winStyle.GetFrameSize()

		boxWidth := m.terminalWidth - 2
		if boxWidth < 38 {
			boxWidth = 38
		}
		innerWidth := boxWidth - borderWidth
		innerHeight := m.terminalHeight - borderHeight - 3
		if innerHeight < 8 {
			innerHeight = 8
		}

		logContent := ""
		if m.syncLogBuffer != nil {
			rawLogs := m.syncLogBuffer.String()
			lines := strings.Split(rawLogs, "\n")
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1]
			}
			var filteredLines []string
			for _, line := range lines {
				stripped := strings.TrimSpace(line)
				if strings.Contains(stripped, "[NEW]") || strings.Contains(stripped, "[UPDATED]") {
					filteredLines = append(filteredLines, "  "+line)
				}
			}
			maxLines := innerHeight - 4
			if maxLines < 3 {
				maxLines = 3
			}
			if len(filteredLines) > maxLines {
				filteredLines = filteredLines[len(filteredLines)-maxLines:]
			}
			logContent = strings.Join(filteredLines, "\n")
		}

		body := fmt.Sprintf("\n  Syncing NetBox... %s\n\n%s\n", m.syncSpinner.View(), logContent)
		boxContent := lipgloss.Place(innerWidth, innerHeight, lipgloss.Left, lipgloss.Top, body, lipgloss.WithWhitespaceBackground(GlobalTheme.Base))
		windowContent := winStyle.Render(boxContent)

		marginLeft := (m.terminalWidth - boxWidth) / 2
		if marginLeft < 0 {
			marginLeft = 0
		}
		windowContent = lipgloss.NewStyle().MarginLeft(marginLeft).Render(windowContent)

		footer := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(" Synchronizing background process... ")
		return header + "\n" + windowContent + "\n" + footer
	}

	if m.state == StateSyncResult {
		header := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(" NETBOX SYNC RESULT ")
		winStyle := WindowStyle.Copy().PaddingTop(0).PaddingBottom(1).PaddingLeft(0).PaddingRight(0)
		borderWidth, borderHeight := winStyle.GetFrameSize()

		boxWidth := m.terminalWidth - 2
		if boxWidth < 38 {
			boxWidth = 38
		}
		innerWidth := boxWidth - borderWidth
		innerHeight := m.terminalHeight - borderHeight - 3
		if innerHeight < 8 {
			innerHeight = 8
		}

		var body string
		var alignX lipgloss.Position = lipgloss.Center
		var alignY lipgloss.Position = lipgloss.Center

		if m.syncError != nil {
			body = "\n\n  " + ErrorStyle.Render("Sync failed: "+m.syncError.Error()) + "\n\n"
		} else {
			alignX = lipgloss.Left
			alignY = lipgloss.Top
			var hostLines []string
			if m.syncLogBuffer != nil {
				for _, line := range strings.Split(m.syncLogBuffer.String(), "\n") {
					stripped := strings.TrimSpace(line)
					if strings.Contains(stripped, "[NEW]") || strings.Contains(stripped, "[UPDATED]") {
						hostLines = append(hostLines, "  "+line)
					}
				}
			}
			newCount := 0
			updatedCount := 0
			for _, l := range hostLines {
				if strings.Contains(l, "[NEW]") {
					newCount++
				} else {
					updatedCount++
				}
			}
			summaryLine := ""
			if newCount > 0 || updatedCount > 0 {
				summaryLine = "\n  " +
					lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true).Render(fmt.Sprintf("%d new", newCount)) +
					"  " +
					lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387")).Bold(true).Render(fmt.Sprintf("%d updated", updatedCount)) +
					"\n"
			} else {
				summaryLine = "\n  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#7f849c")).Render("No changes detected.") + "\n"
			}
			hostContent := strings.Join(hostLines, "\n")
			body = fmt.Sprintf("\n  %s%s\n%s\n",
				SuccessTitleStyle.Render("Sync completed successfully!"),
				summaryLine,
				hostContent)
		}

		boxContent := lipgloss.Place(innerWidth, innerHeight, alignX, alignY, body, lipgloss.WithWhitespaceBackground(GlobalTheme.Base))
		windowContent := winStyle.Render(boxContent)

		marginLeft := (m.terminalWidth - boxWidth) / 2
		if marginLeft < 0 {
			marginLeft = 0
		}
		windowContent = lipgloss.NewStyle().MarginLeft(marginLeft).Render(windowContent)

		footer := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(" q/esc: back to main menu ")
		return header + "\n" + windowContent + "\n" + footer
	}

	if m.state == StateExportResult {
		header := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(" CSV EXPORT RESULT ")
		winStyle := WindowStyle.Copy().PaddingTop(0).PaddingBottom(1).PaddingLeft(0).PaddingRight(0)
		borderWidth, borderHeight := winStyle.GetFrameSize()

		boxWidth := m.terminalWidth - 2
		if boxWidth < 38 {
			boxWidth = 38
		}
		innerWidth := boxWidth - borderWidth
		innerHeight := m.terminalHeight - borderHeight - 3
		if innerHeight < 8 {
			innerHeight = 8
		}

		var body string
		if m.exportError != nil {
			body = "\n\n  " + ErrorStyle.Render("Export failed: "+m.exportError.Error()) + "\n\n"
		} else {
			msgStr := fmt.Sprintf("Export completed successfully!\n  Exported %d hosts to %s", m.exportNumHosts, m.exportOutFile)
			body = "\n\n  " + SuccessTitleStyle.Render(msgStr) + "\n\n"
		}

		boxContent := lipgloss.Place(innerWidth, innerHeight, lipgloss.Center, lipgloss.Center, body, lipgloss.WithWhitespaceBackground(GlobalTheme.Base))
		windowContent := winStyle.Render(boxContent)

		marginLeft := (m.terminalWidth - boxWidth) / 2
		if marginLeft < 0 {
			marginLeft = 0
		}
		windowContent = lipgloss.NewStyle().MarginLeft(marginLeft).Render(windowContent)

		footer := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(" q/esc: back to main menu ")
		return header + "\n" + windowContent + "\n" + footer
	}

	if m.state == StateManageHostsMenu || m.state == StateManageGroupsMenu {
		title := "MANAGE HOSTS"
		if m.state == StateManageGroupsMenu {
			title = "MANAGE GROUPS"
		}
		header := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(" " + title + " ")

		boxWidth := m.terminalWidth - 2
		if boxWidth < 38 {
			boxWidth = 38
		}

		winStyle := WindowStyle.Copy().PaddingTop(0).PaddingBottom(1).PaddingLeft(0).PaddingRight(0)
		borderWidth, borderHeight := winStyle.GetFrameSize()

		innerWidth := boxWidth - borderWidth
		if innerWidth < 30 {
			innerWidth = 30
		}

		var menuLines []string
		menuLines = append(menuLines, FocusedLabelStyle.Copy().Render(title+" OPTIONS"))
		menuLines = append(menuLines, "")

		for i, item := range m.subMenuItems {
			if i == m.subMenuCursor {
				styled := ActiveButtonStyle.Render(item)
				menuLines = append(menuLines, styled)
			} else {
				styled := ButtonStyle.Render(item)
				menuLines = append(menuLines, styled)
			}
		}

		var centeredLines []string
		for _, line := range menuLines {
			centeredLines = append(centeredLines, lipgloss.NewStyle().
				Width(innerWidth).
				Align(lipgloss.Center).
				Background(GlobalTheme.Base).
				Render(line))
		}
		dialogBody := strings.Join(centeredLines, "\n")

		innerHeight := m.terminalHeight - borderHeight - 3
		if innerHeight < 8 {
			innerHeight = 8
		}

		boxContent := lipgloss.Place(innerWidth, innerHeight, lipgloss.Center, lipgloss.Center, dialogBody, lipgloss.WithWhitespaceBackground(GlobalTheme.Base))
		windowContent := winStyle.Render(boxContent)

		marginLeft := (m.terminalWidth - boxWidth) / 2
		if marginLeft < 0 {
			marginLeft = 0
		}
		windowContent = lipgloss.NewStyle().MarginLeft(marginLeft).Render(windowContent)

		helpText := " ↑/↓ or j/k: navigate • enter: select • q/esc: back to main menu "
		footer := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(helpText)

		return header + "\n" + windowContent + "\n" + footer
	}

	if m.state == StateVars {
		targetType := "Host"
		if m.varsTargetIsGroup {
			targetType = "Group"
		}
		header := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(fmt.Sprintf(" VARIABLES FOR %s: %s ", strings.ToUpper(targetType), strings.ToUpper(m.varsTargetName)))

		boxWidth := m.terminalWidth - 2
		if boxWidth < 38 {
			boxWidth = 38
		}

		winStyle := WindowStyle.Copy().PaddingTop(0).PaddingBottom(1).PaddingLeft(0).PaddingRight(0)
		borderWidth, borderHeight := winStyle.GetFrameSize()

		innerWidth := boxWidth - borderWidth
		if innerWidth < 30 {
			innerWidth = 30
		}

		innerHeight := m.terminalHeight - borderHeight - 3
		if innerHeight < 8 {
			innerHeight = 8
		}

		items := m.buildVarsItems()

		type lineSlot struct {
			text   string
			itemID int
		}
		var displayLines []lineSlot
		for itemID, item := range items {
			for _, l := range item.lines {
				displayLines = append(displayLines, lineSlot{text: l, itemID: itemID})
			}
		}

		totalLines := len(displayLines)

		cursorToLine := make([]int, len(items))
		dl := 0
		for itemID, item := range items {
			cursorToLine[itemID] = dl
			dl += len(item.lines)
		}

		cursorIdx := m.varsCursor
		if cursorIdx >= len(cursorToLine) {
			cursorIdx = len(cursorToLine) - 1
		}
		if cursorIdx < 0 {
			cursorIdx = 0
		}
		cursorLine := cursorToLine[cursorIdx]
		start := cursorLine - (innerHeight / 2)
		if start < 0 {
			start = 0
		}
		end := start + innerHeight
		if end > totalLines {
			end = totalLines
			start = end - innerHeight
			if start < 0 {
				start = 0
			}
		}

		cursorItemID := -1
		if cursorIdx >= 0 && cursorIdx < len(items) {
			cursorItemID = cursorIdx
		}
		isCursorLine := make([]bool, len(displayLines))
		if cursorItemID >= 0 && items[cursorItemID].cursorSlots > 0 {
			cStart := cursorToLine[cursorItemID]
			cEnd := cStart + len(items[cursorItemID].lines)
			for i := cStart; i < cEnd && i < len(isCursorLine); i++ {
				isCursorLine[i] = true
			}
		}

		var renderedLines []string
		for li := start; li < end; li++ {
			sl := displayLines[li]
			if isCursorLine[li] {
				styled := lipgloss.NewStyle().
					Foreground(GlobalTheme.Base).
					Background(GlobalTheme.Primary).
					Bold(true).
					Render(sl.text)
				renderedLines = append(renderedLines, styled)
			} else {
				item := items[sl.itemID]
				var line string
				switch {
				case sl.itemID == 0:
					line = lipgloss.NewStyle().Foreground(GlobalTheme.Success).Bold(true).Render(sl.text)
				case item.isHeader:
					line = lipgloss.NewStyle().Foreground(GlobalTheme.Subtext).Italic(true).Render(sl.text)
				case item.dimmed:
					line = lipgloss.NewStyle().Foreground(GlobalTheme.Subtext).Italic(true).Render(sl.text)
				default:
					line = lipgloss.NewStyle().Foreground(GlobalTheme.Text).Bold(true).Render(sl.text)
				}
				renderedLines = append(renderedLines, line)
			}
		}

		renderedStr := strings.Join(renderedLines, "\n")
		if len(renderedLines) < innerHeight {
			renderedStr += strings.Repeat("\n", innerHeight-len(renderedLines))
		}

		dialogBody := renderedStr
		boxContent := lipgloss.Place(innerWidth, innerHeight, lipgloss.Left, lipgloss.Top, dialogBody, lipgloss.WithWhitespaceBackground(GlobalTheme.Base))
		windowContent := winStyle.Render(boxContent)

		marginLeft := (m.terminalWidth - boxWidth) / 2
		if marginLeft < 0 {
			marginLeft = 0
		}
		windowContent = lipgloss.NewStyle().MarginLeft(marginLeft).Render(windowContent)

		helpText := " ↑/↓: navigate • enter: edit/add • d: delete • e: open in editor • q/esc: back "
		footer := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(helpText)

		return header + "\n" + windowContent + "\n" + footer
	}

	if m.state == StateDashboard {
		inv, err := store.LoadInventory(m.inventoryPath)
		if err != nil {
			return m.renderErrorWindow("INVENTORY DASHBOARD", "Error loading inventory: "+err.Error())
		}

		header := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(" INVENTORY DASHBOARD ")

		boxWidth := m.terminalWidth - 2
		if boxWidth < 38 {
			boxWidth = 38
		}

		winStyle := WindowStyle.Copy().PaddingTop(0).PaddingBottom(1).PaddingLeft(0).PaddingRight(0)
		borderWidth, borderHeight := winStyle.GetFrameSize()

		innerWidth := boxWidth - borderWidth
		if innerWidth < 30 {
			innerWidth = 30
		}

		innerHeight := m.terminalHeight - borderHeight - 3
		if innerHeight < 8 {
			innerHeight = 8
		}

		totalVars := 0
		for _, g := range inv.Groups {
			totalVars += len(g.Vars)
		}
		for _, h := range inv.Hosts {
			totalVars += len(h.Vars)
		}

		health := lipgloss.NewStyle().Foreground(GlobalTheme.Success).Bold(true).Render("OK")
		if hasCycle(inv) {
			health = lipgloss.NewStyle().Foreground(GlobalTheme.Error).Bold(true).Render("DEGRADED (Cyclic Dependency)")
		}

		modTime := m.getInventoryModTime()

		var dbLines []string
		dbLines = append(dbLines, "")
		dbLines = append(dbLines, "  "+FocusedLabelStyle.Render("INVENTORY SUMMARY"))
		dbLines = append(dbLines, "")

		formatRow := func(label, value string) string {
			lbl := LabelStyle.Render("  " + label + ": ")
			val := HelperStyle.Render(value)
			return lbl + val
		}

		dbLines = append(dbLines, formatRow("Inventory Path", m.inventoryPath))
		dbLines = append(dbLines, formatRow("Total Hosts", fmt.Sprintf("%d", len(inv.Hosts))))
		dbLines = append(dbLines, formatRow("Total Groups", fmt.Sprintf("%d", len(inv.Groups))))
		dbLines = append(dbLines, formatRow("Total Variables", fmt.Sprintf("%d", totalVars)))
		dbLines = append(dbLines, "  "+LabelStyle.Render("  Precedence Health: ")+health)
		dbLines = append(dbLines, formatRow("Last Modified", modTime))
		dbLines = append(dbLines, "")

		dialogBody := strings.Join(dbLines, "\n")
		boxContent := lipgloss.Place(innerWidth, innerHeight, lipgloss.Left, lipgloss.Center, dialogBody, lipgloss.WithWhitespaceBackground(GlobalTheme.Base))
		windowContent := winStyle.Render(boxContent)

		marginLeft := (m.terminalWidth - boxWidth) / 2
		if marginLeft < 0 {
			marginLeft = 0
		}
		windowContent = lipgloss.NewStyle().MarginLeft(marginLeft).Render(windowContent)

		helpText := " q/esc: back to main menu "
		footer := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(helpText)

		return header + "\n" + windowContent + "\n" + footer
	}

	if m.state == StateHosts {
		header := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(" GROUPED HOSTS INVENTORY ")

		boxWidth := m.terminalWidth - 2
		if boxWidth < 38 {
			boxWidth = 38
		}

		winStyle := WindowStyle.Copy().PaddingTop(0).PaddingBottom(1).PaddingLeft(0).PaddingRight(0)
		borderWidth, borderHeight := winStyle.GetFrameSize()

		innerWidth := boxWidth - borderWidth
		if innerWidth < 30 {
			innerWidth = 30
		}

		innerHeight := m.terminalHeight - borderHeight - 3
		if innerHeight < 8 {
			innerHeight = 8
		}

		visibleRows := m.getVisibleRows()
		var dialogBody string

		if len(visibleRows) == 0 {
			if m.treeFilterActive && m.treeFilterText != "" {
				dialogBody = fmt.Sprintf("  No results match filter %q.", m.treeFilterText)
			} else {
				dialogBody = "  No hosts or groups found in inventory."
			}
		} else {
			if m.treeCursor >= len(visibleRows) {
				m.treeCursor = len(visibleRows) - 1
			}
			if m.treeCursor < 0 {
				m.treeCursor = 0
			}

			start := 0
			end := len(visibleRows)
			if len(visibleRows) > innerHeight {
				start = m.treeCursor - (innerHeight / 2)
				if start < 0 {
					start = 0
				}
				end = start + innerHeight
				if end > len(visibleRows) {
					end = len(visibleRows)
					start = end - innerHeight
				}
			}

			slicedRows := visibleRows[start:end]
			var lines []string

			for i, row := range slicedRows {
				actualIdx := start + i
				isCursor := actualIdx == m.treeCursor
				var line string

				if row.IsGroup {
					expanded := m.groupExpanded[row.GroupName]
					icon := "▶ "
					if expanded {
						icon = "▼ "
					}

					groupText := fmt.Sprintf("• %s%s", icon, row.Name)
					if isCursor {
						line = lipgloss.NewStyle().
							Foreground(GlobalTheme.Base).
							Background(GlobalTheme.Primary).
							Bold(true).
							Render("  "+groupText+"  ")
					} else {
						line = "  " + lipgloss.NewStyle().Foreground(GlobalTheme.Accent).Bold(true).Render(groupText)
					}
				} else {
					branch := "├─ "
					if row.IsLast {
						branch = "└─ "
					}
					branchText := "    " + branch + row.Name

					if isCursor {
						line = lipgloss.NewStyle().
							Foreground(GlobalTheme.Base).
							Background(GlobalTheme.Primary).
							Bold(true).
							Render("  "+branchText+"  ")
					} else {
						line = "    " + lipgloss.NewStyle().Foreground(GlobalTheme.Overlay).Render(branch) + lipgloss.NewStyle().Foreground(GlobalTheme.Text).Render(row.Name)
					}
				}
				lines = append(lines, line)
			}
			dialogBody = strings.Join(lines, "\n")
		}

		boxContent := lipgloss.Place(innerWidth, innerHeight, lipgloss.Left, lipgloss.Top, dialogBody, lipgloss.WithWhitespaceBackground(GlobalTheme.Base))
		windowContent := winStyle.Render(boxContent)

		marginLeft := (m.terminalWidth - boxWidth) / 2
		if marginLeft < 0 {
			marginLeft = 0
		}
		windowContent = lipgloss.NewStyle().MarginLeft(marginLeft).Render(windowContent)

		var filterBar string
		if m.treeFilterActive {
			filterMargin := lipgloss.NewStyle().MarginLeft(marginLeft)
			filterInput := m.treeFilterInput.View()
			if filterInput == "" {
				filterInput = "/ "
			}
			filterBar = filterMargin.Render(filterInput) + "\n"
		}

		helpText := " ↑/↓: navigate • pgup/pgdn: scroll page • /: filter • enter: toggle group • q/esc: back "
		footer := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(helpText)

		return header + "\n" + filterBar + windowContent + "\n" + footer
	}

	if m.terminalWidth == 0 || m.terminalHeight == 0 {
		return "Initializing menu..."
	}

	header := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(" ROSTER INVENTORY MANAGER ")

	boxWidth := m.terminalWidth - 2
	if boxWidth < 38 {
		boxWidth = 38
	}

	winStyle := WindowStyle.Copy().PaddingTop(0).PaddingBottom(1).PaddingLeft(0).PaddingRight(0)
	borderWidth, borderHeight := winStyle.GetFrameSize()

	innerWidth := boxWidth - borderWidth
	if innerWidth < 30 {
		innerWidth = 30
	}

	innerHeight := m.terminalHeight - borderHeight - 3
	if innerHeight < 8 {
		innerHeight = 8
	}

	var menuLines []string
	menuLines = append(menuLines, FocusedLabelStyle.Copy().Render("MAIN MENU"))
	menuLines = append(menuLines, "")

	for i, item := range m.menuItems {
		if i == m.menuCursor {
			styled := ActiveButtonStyle.Render(item)
			menuLines = append(menuLines, styled)
		} else {
			styled := ButtonStyle.Render(item)
			menuLines = append(menuLines, styled)
		}
	}

	var centeredLines []string
	for _, line := range menuLines {
		centeredLines = append(centeredLines, lipgloss.NewStyle().
			Width(innerWidth).
			Align(lipgloss.Center).
			Background(GlobalTheme.Base).
			Render(line))
	}
	dialogBody := strings.Join(centeredLines, "\n")

	boxContent := lipgloss.Place(innerWidth, innerHeight, lipgloss.Center, lipgloss.Center, dialogBody, lipgloss.WithWhitespaceBackground(GlobalTheme.Base))
	windowContent := winStyle.Render(boxContent)

	marginLeft := (m.terminalWidth - boxWidth) / 2
	if marginLeft < 0 {
		marginLeft = 0
	}
	windowContent = lipgloss.NewStyle().MarginLeft(marginLeft).Render(windowContent)

	helpText := " ↑/↓ or j/k: navigate • enter: select • q/esc: exit "
	footer := HeaderStyle.Copy().Width(m.terminalWidth).Align(lipgloss.Center).Render(helpText)

	return header + "\n" + windowContent + "\n" + footer
}
