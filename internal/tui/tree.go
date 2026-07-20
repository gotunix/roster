package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gotunix.net/roster/internal/models"
	"gotunix.net/roster/internal/store"
)

func hasCycle(inv *models.Inventory) bool {
	visited := make(map[string]bool)
	temp := make(map[string]bool)

	var visit func(string) bool
	visit = func(node string) bool {
		if temp[node] {
			return true
		}
		if visited[node] {
			return false
		}
		temp[node] = true

		g, ok := inv.Groups[node]
		if ok {
			for _, child := range g.Children {
				if visit(child) {
					return true
				}
			}
		}

		temp[node] = false
		visited[node] = true
		return false
	}

	for name := range inv.Groups {
		if visit(name) {
			return true
		}
	}
	return false
}

func (m *MainTUIModel) getInventoryModTime() string {
	path := filepath.Join(m.inventoryPath, "main.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = filepath.Join(m.inventoryPath, "main.yml")
	}

	info, err := os.Stat(path)
	if err != nil {
		return "unknown"
	}
	return info.ModTime().Format("2006-01-02 15:04:05")
}

func (m *MainTUIModel) ensureInventoryLoaded() error {
	modTime := m.getInventoryModTime()
	if m.cachedInventory != nil && m.inventoryModTime == modTime && !m.rowsDirty {
		return nil
	}

	inv, err := store.LoadInventory(m.inventoryPath)
	if err != nil {
		return err
	}

	m.cachedInventory = inv
	m.inventoryModTime = modTime
	m.rowsDirty = true
	return nil
}

func (m *MainTUIModel) rebuildRows() {
	if m.cachedInventory == nil {
		m.cachedRows = nil
		return
	}

	inv := m.cachedInventory
	var gNames []string
	for name := range inv.Groups {
		gNames = append(gNames, name)
	}
	sort.Slice(gNames, func(i, j int) bool {
		if gNames[i] == "all" {
			return true
		}
		if gNames[j] == "all" {
			return false
		}
		return gNames[i] < gNames[j]
	})

	var rows []HostTreeRow
	for _, gName := range gNames {
		g := inv.Groups[gName]
		rows = append(rows, HostTreeRow{
			Name:      gName,
			IsGroup:   true,
			GroupName: gName,
		})

		if m.groupExpanded[gName] {
			hosts := g.Hosts
			sort.Strings(hosts)
			for j, hName := range hosts {
				rows = append(rows, HostTreeRow{
					Name:      hName,
					IsGroup:   false,
					GroupName: gName,
					IsLast:    j == len(hosts)-1,
				})
			}
		}
	}
	m.cachedRows = rows
	m.rowsDirty = false
}

func (m *MainTUIModel) getVisibleRows() []HostTreeRow {
	if err := m.ensureInventoryLoaded(); err != nil {
		return nil
	}
	if m.rowsDirty {
		m.rebuildRows()
	}
	if m.treeFilterText != "" {
		return m.treeFilteredRows
	}
	return m.cachedRows
}

func (m *MainTUIModel) applyTreeFilter() {
	if err := m.ensureInventoryLoaded(); err != nil {
		return
	}
	if m.rowsDirty {
		m.rebuildRows()
	}

	if m.treeFilterText == "" {
		m.treeFilteredRows = nil
		return
	}

	lowerFilter := strings.ToLower(m.treeFilterText)
	m.treeFilteredRows = nil

	matchingGroups := make(map[string]bool)

	for i, row := range m.cachedRows {
		if row.IsGroup {
			if strings.Contains(strings.ToLower(row.Name), lowerFilter) {
				matchingGroups[row.Name] = true
				m.treeFilteredRows = append(m.treeFilteredRows, row)
				if m.groupExpanded[row.Name] {
					for j := i + 1; j < len(m.cachedRows); j++ {
						child := m.cachedRows[j]
						if child.IsGroup {
							break
						}
						if child.GroupName == row.Name {
							m.treeFilteredRows = append(m.treeFilteredRows, child)
						}
					}
				}
			}
		} else {
			if strings.Contains(strings.ToLower(row.Name), lowerFilter) {
				m.treeFilteredRows = append(m.treeFilteredRows, row)
				matchingGroups[row.GroupName] = true
			}
		}
	}

	if len(m.treeFilteredRows) > 0 {
		var result []HostTreeRow
		seenGroups := make(map[string]bool)
		for _, row := range m.treeFilteredRows {
			if !row.IsGroup && !matchingGroups[row.GroupName] && !seenGroups[row.GroupName] {
				for _, cr := range m.cachedRows {
					if cr.IsGroup && cr.Name == row.GroupName {
						result = append(result, cr)
						break
					}
				}
				seenGroups[row.GroupName] = true
			}
			result = append(result, row)
		}
		m.treeFilteredRows = result
	}

	if m.treeCursor >= len(m.treeFilteredRows) {
		m.treeCursor = len(m.treeFilteredRows) - 1
	}
	if m.treeCursor < 0 {
		m.treeCursor = 0
	}
}
