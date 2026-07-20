package store

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
	"strings"

	"gotunix.net/roster/internal/models"
)

type colDef struct {
	label string
	path  string
}

// ExportInventory generates CSV-encoded export data for a single inventory.
// Returns the CSV bytes, the number of host rows, and any error.
func ExportInventory(inv *models.Inventory, inventoryDir string, vars []string, exclude []string, groupFilter string) ([]byte, int, error) {
	cols := buildColumns(vars)
	header := make([]string, len(cols))
	for i, c := range cols {
		header[i] = c.label
	}

	excludeMap := make(map[string]bool)
	for _, g := range exclude {
		excludeMap[g] = true
	}

	var hostNames []string
	for name := range inv.Hosts {
		hostNames = append(hostNames, name)
	}
	sort.Strings(hostNames)

	allRows := [][]string{header}

	for _, hName := range hostNames {
		var groups []string
		excluded := false
		inGroup := groupFilter == ""
		for gName, g := range inv.Groups {
			for _, member := range g.Hosts {
				if member == hName {
					if excludeMap[gName] {
						excluded = true
					}
					groups = append(groups, gName)
					if gName == groupFilter {
						inGroup = true
					}
					break
				}
			}
		}

		if excluded || !inGroup {
			continue
		}

		sort.Strings(groups)
		effectiveVars := GetEffectiveHostVars(inv, hName)

		row := make([]string, len(cols))
		for i, c := range cols {
			switch {
			case c.path == "" && c.label == "Host":
				row[i] = hName
			case c.path == "" && c.label == "Inventory":
				row[i] = inventoryDir
			case c.path == "" && c.label == "Groups":
				row[i] = strings.Join(groups, ", ")
			case c.path == "":
				row[i] = hName
			default:
				if v := ResolveNestedVar(effectiveVars, c.path); v != nil {
					row[i] = fmt.Sprintf("%v", v)
				}
			}
		}
		allRows = append(allRows, row)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.WriteAll(allRows); err != nil {
		return nil, 0, err
	}

	return buf.Bytes(), len(allRows) - 1, nil
}

func buildColumns(vars []string) []colDef {
	if len(vars) == 0 {
		return []colDef{
			{label: "Inventory"},
			{label: "Host"},
			{label: "Groups"},
		}
	}

	var cols []colDef
	for _, v := range vars {
		parts := strings.SplitN(v, ":", 2)
		label := parts[0]
		path := parts[0]
		if len(parts) == 2 {
			label = parts[1]
		}
		if strings.EqualFold(path, "Hostname") {
			cols = append(cols, colDef{label: label})
		} else {
			cols = append(cols, colDef{label: label, path: path})
		}
	}
	return cols
}
