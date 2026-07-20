package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"

	"gotunix.net/roster/internal/store"
)

type dispItem struct {
	lines       []string
	cursorSlots int
	key         string
	dimmed      bool
	isHeader    bool
}

func formatVarValue(val interface{}) []string {
	data, err := yaml.Marshal(val)
	if err != nil {
		return []string{fmt.Sprintf("%v", val)}
	}
	s := strings.TrimRight(string(data), "\n")
	return strings.Split(s, "\n")
}

func indentLines(lines []string, indent string) []string {
	res := make([]string, len(lines))
	for i, l := range lines {
		res[i] = indent + l
	}
	return res
}

func renderVarItem(key string, val interface{}, dimmed bool) dispItem {
	valLines := formatVarValue(val)

	if len(valLines) == 1 {
		line := "  " + key + ": " + valLines[0]
		return dispItem{lines: []string{line}, dimmed: dimmed}
	}

	firstLine := "  " + key + ":"
	restLines := indentLines(valLines, "    ")
	allLines := append([]string{firstLine}, restLines...)
	return dispItem{lines: allLines, dimmed: dimmed}
}

func (m *MainTUIModel) buildVarsItems() []dispItem {
	var items []dispItem

	items = append(items, dispItem{lines: []string{"  [ Edit Variables ]"}, cursorSlots: 1})

	targetType := "Host"
	if m.varsTargetIsGroup {
		targetType = "Group"
	}
	hostHeader := fmt.Sprintf("  ─── %s Variables (%s) ───", targetType, m.varsTargetName)
	items = append(items, dispItem{lines: []string{hostHeader}, cursorSlots: 0, isHeader: true})

	if m.varsNested != nil {
		var nestedKeys []string
		for k := range m.varsNested {
			nestedKeys = append(nestedKeys, k)
		}
		sort.Strings(nestedKeys)
		for _, nk := range nestedKeys {
			val := m.varsNested[nk]
			ri := renderVarItem(nk, val, false)
			ri.cursorSlots = 1
			ri.key = nk
			items = append(items, ri)
		}
	}

	hasInherited := false
	for _, gName := range m.inheritedGroups {
		if keys, ok := m.inheritedKeysByGroup[gName]; ok && len(keys) > 0 {
			hasInherited = true
			break
		}
	}
	if hasInherited {
		items = append(items, dispItem{lines: []string{""}, cursorSlots: 0})
	}

	firstSection := true
	for _, gName := range m.inheritedGroups {
		keys := m.inheritedKeysByGroup[gName]
		if len(keys) == 0 {
			continue
		}
		if !firstSection {
			items = append(items, dispItem{lines: []string{""}, cursorSlots: 0})
		}
		firstSection = false
		inheritLabel := fmt.Sprintf("  ─── Inherited (%s) ───", gName)
		items = append(items, dispItem{lines: []string{inheritLabel}, cursorSlots: 0, isHeader: true})
		for _, ikey := range keys {
			ival := m.inheritedByGroup[gName][ikey]
			ri := renderVarItem(ikey, ival, true)
			ri.cursorSlots = 1
			items = append(items, ri)
		}
	}

	return items
}

func stripInternalVars(flat map[string]interface{}) {
	delete(flat, "roster_netbox_managed")
}

func (m *MainTUIModel) loadVars() {
	m.inheritedGroups = nil
	m.inheritedByGroup = nil
	m.inheritedKeysByGroup = nil

	if m.varsTargetIsGroup {
		vars, err := store.GetGroupVars(m.inventoryPath, m.varsTargetName)
		if err != nil || vars == nil {
			m.varsValues = make(map[string]interface{})
			m.varsKeys = nil
			m.varsNested = nil
			return
		}
		stripInternalVars(vars)
		m.varsNested = vars
		flat := make(map[string]interface{})
		store.FlattenMap(vars, "", flat)
		stripInternalVars(flat)
		m.varsValues = flat
		m.varsKeys = sortedKeys(flat)
		return
	}

	hostVars, err := store.GetHostVars(m.inventoryPath, m.varsTargetName)
	if err != nil || hostVars == nil {
		hostVars = make(map[string]interface{})
	}
	hostFlat := make(map[string]interface{})
	store.FlattenMap(hostVars, "", hostFlat)
	stripInternalVars(hostFlat)

	inv, err := store.LoadInventory(m.inventoryPath)
	if err == nil && inv != nil {
		if _, ok := inv.Hosts[m.varsTargetName]; ok {
			effective := store.GetEffectiveHostVars(inv, m.varsTargetName)
			effFlat := make(map[string]interface{})
			store.FlattenMap(effective, "", effFlat)
			stripInternalVars(effFlat)

			inherited := make(map[string]interface{})
			for k, v := range effFlat {
				if _, isHostVar := hostFlat[k]; !isHostVar {
					inherited[k] = v
				}
			}

			if len(inherited) > 0 {
				source := make(map[string]string)
				byGroup := make(map[string]map[string]interface{})
				hostGroups := store.FindHostGroups(inv, m.varsTargetName)
				for i := len(hostGroups) - 1; i >= 0; i-- {
					gName := hostGroups[i]
					g, ok := inv.Groups[gName]
					if !ok || g.Vars == nil {
						continue
					}
					gFlat := make(map[string]interface{})
					store.FlattenMap(g.Vars, "", gFlat)
					stripInternalVars(gFlat)
					for k := range gFlat {
						if _, ok := inherited[k]; ok {
							if _, claimed := source[k]; !claimed {
								source[k] = gName
							}
						}
					}
				}
				for k, v := range inherited {
					if gName, ok := source[k]; ok {
						if byGroup[gName] == nil {
							byGroup[gName] = make(map[string]interface{})
						}
						byGroup[gName][k] = v
					}
				}
				m.inheritedByGroup = byGroup
				m.inheritedKeysByGroup = make(map[string][]string)
				m.inheritedGroups = make([]string, 0, len(byGroup))
				for gName := range byGroup {
					m.inheritedGroups = append(m.inheritedGroups, gName)
					m.inheritedKeysByGroup[gName] = sortedKeys(byGroup[gName])
				}
				sort.Strings(m.inheritedGroups)
			}
		}
	}

	m.varsNested = hostVars
	stripInternalVars(m.varsNested)
	m.varsValues = hostFlat
	m.varsKeys = sortedKeys(hostFlat)
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (m *MainTUIModel) buildVarsForm(key, val string) *FormModel {
	title := "ADD VARIABLE TO " + strings.ToUpper(m.varsTargetName)
	if m.varsFormType == "edit" {
		title = "EDIT VARIABLE FOR " + strings.ToUpper(m.varsTargetName)
	}

	form := NewForm(title)

	if m.varsFormType == "add" {
		form.AddTextBox("key", "Variable Key", "", "e.g. ansible_user")
	}

	form.AddTextBox("value", "Variable Value", "", "e.g. admin")
	if m.varsFormType == "edit" {
		form.SetValue("value", val)
	}

	form.AddButton("Save", func(f *FormModel) tea.Cmd {
		var targetKey string
		if m.varsFormType == "add" {
			targetKey = f.GetString("key")
		} else {
			targetKey = key
		}

		targetVal := f.GetString("value")
		if strings.TrimSpace(targetKey) == "" {
			return nil
		}

		var err error
		if m.varsTargetIsGroup {
			err = store.SetGroupVar(m.inventoryPath, m.varsTargetName, targetKey, targetVal)
		} else {
			err = store.SetHostVar(m.inventoryPath, m.varsTargetName, targetKey, targetVal)
		}

		if err == nil {
			f.Submitted = true
		}
		return nil
	})

	form.AddButton("Cancel", func(f *FormModel) tea.Cmd {
		f.Quitting = true
		return nil
	})

	return form
}
