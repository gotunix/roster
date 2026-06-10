// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors
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
	"runtime"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"gotunix.net/roster/internal/models"
	"gotunix.net/roster/internal/store"
	"gotunix.net/roster/internal/version"
)

const Logo = `
   /$$$$$$$                        /$$
  | $$__  $$                      | $$
  | $$  \ $$  /$$$$$$   /$$$$$$$ /$$$$$$    /$$$$$$   /$$$$$$
  | $$$$$$$/ /$$__  $$ /$$_____/|_  $$_/   /$$__  $$ /$$__  $$
  | $$__  $$| $$  \ $$|  $$$$$$   | $$    | $$$$$$$$| $$  \__/
  | $$  \ $$| $$  | $$ \____  $$  | $$ /$$| $$_____/| $$
  | $$  | $$|  $$$$$$/ /$$$$$$$/  |  $$$$/|  $$$$$$$| $$
  |__/  |__/ \______/ |_______/    \___/   \_______/|__/
`

func HandleHelp(cmd *cobra.Command, args []string) {
	cmd.Usage()
}

func HandleUsage(cmd *cobra.Command) error {
	fmt.Println(LogoStyle.Render(Logo))
	fmt.Println(HelpTitleStyle.Render(cmd.Short))
	fmt.Println(HelpDescStyle.Render(cmd.Long))

	fmt.Println(HelpSectionStyle.Render("USAGE"))
	fmt.Printf("  %s [command]\n", cmd.CommandPath())

	if len(cmd.Commands()) > 0 {
		fmt.Println(HelpSectionStyle.Render("AVAILABLE COMMANDS"))
		for _, c := range cmd.Commands() {
			if !c.Hidden {
				fmt.Printf("  %-15s %s\n", c.Name(), c.Short)
			}
		}
	}

	if cmd.Flags().HasFlags() {
		fmt.Println(HelpSectionStyle.Render("FLAGS"))
		fmt.Println(HelpFlagStyle.Render(cmd.Flags().FlagUsages()))
	}

	fmt.Println(HelpSectionStyle.Render("LEARN MORE"))
	fmt.Printf("  Use \"%s [command] --help\" for more information about a command.\n", cmd.CommandPath())
	return nil
}

func RenderVersion() error {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Errorf("failed to read build info")
	}

	totalWidth := GetTerminalWidth()
	var sb strings.Builder

	appVersion := fmt.Sprintf("%s %s", version.AppName, version.AppVersion)

	border := lipgloss.RoundedBorder()
	subStyle := lipgloss.NewStyle().Foreground(Subtle)
	labelStyle := LabelStyle.Padding(0, 1).Width(14)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1)

	// Header
	titleText := " SYSTEM INFORMATION "
	dLeft := (totalWidth - 2 - len(titleText)) / 2
	dRight := totalWidth - 2 - len(titleText) - dLeft
	sb.WriteString("\n" + subStyle.Render(border.TopLeft+strings.Repeat(border.Top, dLeft)) +
		TitleStyle.Render("SYSTEM INFORMATION") +
		subStyle.Render(strings.Repeat(border.Top, dRight)+border.TopRight) + "\n")

	// Row Calculations
	contentWidth := totalWidth - 2
	availForSplit := contentWidth - 28 - 3
	vW1 := availForSplit / 2
	vW2 := availForSplit - vW1

	renderSplitRow := func(l1, v1, l2, v2 string, isLast bool) {
		sb.WriteString(subStyle.Render("│") + labelStyle.Render(l1) + subStyle.Render("│") + valStyle.Width(vW1).Render(v1) +
			subStyle.Render("│") + labelStyle.Render(l2) + subStyle.Render("│") + valStyle.Width(vW2).Render(v2) +
			subStyle.Render("│") + "\n")
		if !isLast {
			sb.WriteString(subStyle.Render("├──────────────┼"+strings.Repeat("─", vW1)+"┼──────────────┼"+strings.Repeat("─", vW2)+"┤") + "\n")
		}
	}

	renderSplitRow("APP:", appVersion, "GO:", info.GoVersion, false)
	renderSplitRow("OS:", runtime.GOOS, "ARCH:", runtime.GOARCH, true)

	sb.WriteString(subStyle.Render(border.BottomLeft+strings.Repeat(border.Bottom, totalWidth-2)+border.BottomRight) + "\n")
	sb.WriteString("\n")

	// Dependencies Window
	depContentWidth := totalWidth - 10
	dotStyle := lipgloss.NewStyle().Foreground(Subtle)

	var depLines []string
	for _, dep := range info.Deps {
		path := lipgloss.NewStyle().Foreground(Cyan).Render(dep.Path)
		version := lipgloss.NewStyle().Foreground(Green).Render(dep.Version)

		label := "• " + path + " "
		repeat := depContentWidth - lipgloss.Width(label) - lipgloss.Width(dep.Version) - 1
		if repeat < 0 {
			repeat = 0
		}

		depLines = append(depLines, label+dotStyle.Render(strings.Repeat(".", repeat))+" "+version)
	}
	sort.Strings(depLines)

	RenderWindow(&sb, "DEPENDENCIES", strings.Join(depLines, "\n"), totalWidth)

	fmt.Print(sb.String())
	return nil
}

// RenderHostList renders a sorted list of hosts, optionally filtered by group
func RenderHostList(inv *models.Inventory, groupFilter string) string {
	title := "HOSTS"
	var hosts []string

	if groupFilter != "" {
		group, ok := inv.Groups[groupFilter]
		if !ok {
			return ErrorMsg("group %q not found", groupFilter)
		}
		title = fmt.Sprintf("HOSTS IN GROUP: %s", strings.ToUpper(groupFilter))
		hosts = append([]string{}, group.Hosts...)
	} else {
		for name := range inv.Hosts {
			hosts = append(hosts, name)
		}
	}

	if len(hosts) == 0 {
		var sb strings.Builder
		msg := "No hosts found."
		if groupFilter != "" {
			msg = fmt.Sprintf("No hosts found in group %q.", groupFilter)
		}
		RenderWindow(&sb, title, "  "+DescriptionStyle.Render(msg), GetTerminalWidth())
		return sb.String()
	}

	sort.Strings(hosts)

	var listSb strings.Builder
	for _, h := range hosts {
		listSb.WriteString(fmt.Sprintf("• %s\n", BoldStyle.Foreground(Green).Render(h)))
	}

	var sb strings.Builder
	RenderWindow(&sb, title, strings.TrimSpace(listSb.String()), GetTerminalWidth())
	return sb.String()
}

// RenderDashboard renders a hierarchical tree view of the inventory
func RenderDashboard(inv *models.Inventory) string {
	// Content Area
	var contentSb strings.Builder

	// Summary Header
	statsStyle := lipgloss.NewStyle().
		Foreground(Subtle).
		Italic(true).
		PaddingBottom(1)

	contentSb.WriteString(statsStyle.Render(fmt.Sprintf("Inventory contains %d groups and %d hosts", len(inv.Groups), len(inv.Hosts))) + "\n")

	// Sort groups
	var groups []string
	for gName := range inv.Groups {
		groups = append(groups, gName)
	}
	sort.Strings(groups)

	subtleStyle := lipgloss.NewStyle().Foreground(Subtle)

	for _, gName := range groups {
		g := inv.Groups[gName]
		contentSb.WriteString(fmt.Sprintf("%s %s\n",
			lipgloss.NewStyle().Foreground(Magenta).Render("📂"),
			BoldStyle.Foreground(Magenta).Render(strings.ToUpper(gName))))

		// Hosts in group
		sort.Strings(g.Hosts)
		for i, hName := range g.Hosts {
			branch := subtleStyle.Render("├─")
			if i == len(g.Hosts)-1 && len(g.Children) == 0 {
				branch = subtleStyle.Render("└─")
			}

			hostDisplay := BoldStyle.Foreground(Green).Render(hName)
			if h, ok := inv.Hosts[hName]; ok && h.Vars != nil {
				if desc, ok := h.Vars["description"].(string); ok && desc != "" {
					hostDisplay = fmt.Sprintf("%s %s", hostDisplay, DescriptionStyle.Render("("+desc+")"))
				}
			}

			contentSb.WriteString(fmt.Sprintf("  %s %s %s\n",
				branch,
				lipgloss.NewStyle().Foreground(Green).Render("🖥 "),
				hostDisplay))
		}

		// Children groups
		sort.Strings(g.Children)
		for i, cName := range g.Children {
			branch := subtleStyle.Render("├─")
			if i == len(g.Children)-1 {
				branch = subtleStyle.Render("└─")
			}
			contentSb.WriteString(fmt.Sprintf("  %s %s %s\n",
				branch,
				lipgloss.NewStyle().Foreground(Cyan).Render("📂"),
				BoldStyle.Foreground(Cyan).Render(cName)))
		}
		contentSb.WriteString("\n")
	}

	var sb strings.Builder
	RenderWindow(&sb, "ANSIBLE INVENTORY", strings.TrimSpace(contentSb.String()), GetTerminalWidth())
	return sb.String()
}

// RenderGroupList renders a sorted list of all groups in the inventory
func RenderGroupList(inv *models.Inventory) string {
	if len(inv.Groups) == 0 {
		return "  " + DescriptionStyle.Render("No groups found.")
	}

	var groups []string
	for name := range inv.Groups {
		groups = append(groups, name)
	}
	sort.Strings(groups)

	var listSb strings.Builder
	for _, g := range groups {
		listSb.WriteString(fmt.Sprintf("• %s\n", BoldStyle.Foreground(Magenta).Render(g)))
	}

	var sb strings.Builder
	RenderWindow(&sb, "GROUPS", strings.TrimSpace(listSb.String()), GetTerminalWidth())
	return sb.String()
}

// RenderGroupView aggregates data and renders the view for a single group
func RenderGroupView(groupname string, dir string, inv *models.Inventory) (string, error) {
	group, ok := inv.Groups[groupname]
	if !ok {
		return "", fmt.Errorf("group %q not found", groupname)
	}

	totalWidth := GetTerminalWidth()
	var sb strings.Builder

	// Header
	titleText := fmt.Sprintf(" GROUP: %s ", groupname)
	dLeft := (totalWidth - 2 - len(titleText)) / 2
	dRight := totalWidth - 2 - len(titleText) - dLeft
	border := lipgloss.RoundedBorder()
	subStyle := lipgloss.NewStyle().Foreground(Subtle)

	sb.WriteString("\n" + subStyle.Render(border.TopLeft+strings.Repeat(border.Top, dLeft)) +
		TitleStyle.Render(fmt.Sprintf("GROUP: %s", groupname)) +
		subStyle.Render(strings.Repeat(border.Top, dRight)+border.TopRight) + "\n")

	// Members Row
	labelStyle := LabelStyle.Padding(0, 1).Width(14)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1)
	fullValW := totalWidth - 17

	members := strings.Join(group.Hosts, ", ")
	if members == "" {
		members = "(none)"
	}

	sb.WriteString(subStyle.Render("│") + labelStyle.Render("HOSTS:") + subStyle.Render("│") +
		valStyle.Width(fullValW).Render(members) + subStyle.Render("│") + "\n")

	if len(group.Children) > 0 {
		sb.WriteString(subStyle.Render("├──────────────┼"+strings.Repeat("─", fullValW)+"┤") + "\n")
		sb.WriteString(subStyle.Render("│") + labelStyle.Render("CHILDREN:") + subStyle.Render("│") +
			valStyle.Width(fullValW).Render(strings.Join(group.Children, ", ")) + subStyle.Render("│") + "\n")
	}

	sb.WriteString(subStyle.Render(border.BottomLeft+strings.Repeat(border.Bottom, totalWidth-2)+border.BottomRight) + "\n")

	// --- Variable Boxes ---
	// 1. Direct Group Vars
	gvStr := FormatVars(group.Vars)
	RenderWindow(&sb, "DIRECT GROUP VARS", gvStr, totalWidth)

	// 2. Inherited from 'all'
	if groupname != "all" {
		av, _ := store.GetGroupVars(dir, "all")
		avStr := FormatVars(av)
		RenderWindow(&sb, "INHERITED VARS (ALL)", avStr, totalWidth)
	}

	return sb.String(), nil
}

// RenderHostView aggregates data and renders the view for a single host
func RenderHostView(hostname string, dir string, inv *models.Inventory) (string, error) {
	totalWidth := GetTerminalWidth()
	var sb strings.Builder

	// Header
	titleText := fmt.Sprintf(" HOST: %s ", hostname)
	dLeft := (totalWidth - 2 - len(titleText)) / 2
	dRight := totalWidth - 2 - len(titleText) - dLeft
	border := lipgloss.RoundedBorder()
	subStyle := lipgloss.NewStyle().Foreground(Subtle)

	sb.WriteString("\n" + subStyle.Render(border.TopLeft+strings.Repeat(border.Top, dLeft)) +
		TitleStyle.Render(fmt.Sprintf("HOST: %s", hostname)) +
		subStyle.Render(strings.Repeat(border.Top, dRight)+border.TopRight) + "\n")

	// Groups Row
	var groups []string
	for gName, g := range inv.Groups {
		for _, hName := range g.Hosts {
			if hName == hostname {
				groups = append(groups, gName)
				break
			}
		}
	}

	labelStyle := LabelStyle.Padding(0, 1).Width(14)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1)
	fullValW := totalWidth - 17

	sb.WriteString(subStyle.Render("│") + labelStyle.Render("GROUPS:") + subStyle.Render("│") +
		valStyle.Width(fullValW).Render(strings.Join(groups, ", ")) + subStyle.Render("│") + "\n")
	sb.WriteString(subStyle.Render(border.BottomLeft+strings.Repeat(border.Bottom, totalWidth-2)+border.BottomRight) + "\n")

	// --- Variable Boxes ---
	// 1. Direct host_vars
	hv, _ := store.GetHostVars(dir, hostname)
	hvStr := FormatVars(hv)
	RenderWindow(&sb, "DIRECT HOST VARS", hvStr, totalWidth)

	// 2. Inherited Vars
	inherited := make(map[string]interface{})
	av, _ := store.GetGroupVars(dir, "all")
	for k, v := range av {
		inherited[k] = v
	}
	for _, gName := range groups {
		gv, _ := store.GetGroupVars(dir, gName)
		for k, v := range gv {
			inherited[k] = v
		}
	}

	ivStr := FormatVars(inherited)
	RenderWindow(&sb, "INHERITED VARS", ivStr, totalWidth)

	return sb.String(), nil
}

// Helper to render a titled window box with precise border alignment
func RenderWindow(sb *strings.Builder, title, content string, width int) {
	if content == "" {
		content = "(empty)"
	}

	border := lipgloss.RoundedBorder()
	subStyle := lipgloss.NewStyle().Foreground(Subtle)
	// 1. Top Border with Integrated Header
	tText := fmt.Sprintf(" %s ", title)
	dL := (width - 2 - len(tText)) / 2
	dR := width - 2 - len(tText) - dL

	sb.WriteString(subStyle.Render(border.TopLeft+strings.Repeat(border.Top, dL)) +
		TitleStyle.Render(title) +
		subStyle.Render(strings.Repeat(border.Top, dR)+border.TopRight) + "\n")

	// 2. Content with Lipgloss managed width and padding
	contentStyle := lipgloss.NewStyle().
		Width(width-2). // Width between bars
		Padding(0, 2)   // 2 cells of padding on each side

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		renderedLine := contentStyle.Render(line)
		sb.WriteString(subStyle.Render("│") + renderedLine + subStyle.Render("│") + "\n")
	}

	// 3. Bottom Border
	sb.WriteString(subStyle.Render(border.BottomLeft+strings.Repeat(border.Bottom, width-2)+border.BottomRight) + "\n")
}

// RenderSplitRow renders a row with two columns and a middle border, precisely aligned to width
func RenderSplitRow(sb *strings.Builder, l1, v1, l2, v2 string, width int, subStyle, labelStyle, valStyle lipgloss.Style, isLast bool) {
	availForSplit := width - 33
	if availForSplit < 0 {
		availForSplit = 0
	}
	vW1 := availForSplit / 2
	vW2 := availForSplit - vW1

	sb.WriteString(subStyle.Render("│") + labelStyle.Render(l1) + subStyle.Render("│") + valStyle.Width(vW1).Render(v1) +
		subStyle.Render("│") + labelStyle.Render(l2) + subStyle.Render("│") + valStyle.Width(vW2).Render(v2) +
		subStyle.Render("│") + "\n")
	if !isLast {
		sb.WriteString(subStyle.Render("├──────────────┼"+strings.Repeat("─", vW1)+"┼──────────────┼"+strings.Repeat("─", vW2)+"┤") + "\n")
	}
}

// RenderFullRow renders a single label/value row spanning the full width
func RenderFullRow(sb *strings.Builder, l, v string, width int, subStyle, labelStyle, valStyle lipgloss.Style, isLast bool) {
	fullValW := width - 17
	if fullValW < 0 {
		fullValW = 0
	}

	sb.WriteString(subStyle.Render("│") + labelStyle.Render(l) + subStyle.Render("│") + valStyle.Width(fullValW).Render(v) + subStyle.Render("│") + "\n")
	if !isLast {
		sb.WriteString(subStyle.Render("├──────────────┼"+strings.Repeat("─", fullValW)+"┤") + "\n")
	}
}

// FormatVars formats a map of variables into a YAML-like string with coloring
func FormatVars(vars map[string]interface{}) string {
	if len(vars) == 0 {
		return ""
	}
	bytes, err := yaml.Marshal(vars)
	if err != nil {
		return fmt.Sprintf("Error formatting vars: %v", err)
	}

	rawYaml := string(bytes)
	lines := strings.Split(rawYaml, "\n")
	var coloredLines []string

	keyStyle := lipgloss.NewStyle().Foreground(Cyan)

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Handle keys in YAML (including nested ones)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			keyPart := parts[0]
			valPart := parts[1]

			// Handle indentation and coloring
			trimmedKey := strings.TrimLeft(keyPart, " -")
			indent := keyPart[:len(keyPart)-len(trimmedKey)]

			coloredLines = append(coloredLines, fmt.Sprintf("%s%s:%s", indent, keyStyle.Render(trimmedKey), valPart))
		} else {
			// Lines without colons (e.g., list items with only values, or comments)
			coloredLines = append(coloredLines, line)
		}
	}

	return strings.Join(coloredLines, "\n")
}
