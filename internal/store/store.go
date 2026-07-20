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

package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"gopkg.in/yaml.v3"
	"gotunix.net/roster/internal/models"
)

var globalFlock *flock.Flock

// LockInventory creates a lock file to prevent concurrent access
func LockInventory(baseDir string) error {
	lockPath := filepath.Join(baseDir, ".roster.lock")
	globalFlock = flock.New(lockPath)

	// Try to lock with a 100ms timeout
	locked, err := globalFlock.TryLockContext(context.Background(), 100*time.Millisecond)
	if err != nil {
		return fmt.Errorf("failed to acquire inventory lock: %v", err)
	}
	if !locked {
		return fmt.Errorf("inventory is currently locked by another process")
	}
	return nil
}

// UnlockInventory releases the inventory lock
func UnlockInventory() {
	if globalFlock != nil {
		_ = globalFlock.Unlock()
	}
}

// GetEffectiveHostVars calculates the full variable map for a host, respecting inheritance
func GetEffectiveHostVars(inv *models.Inventory, hostname string) map[string]interface{} {
	effective := make(map[string]interface{})

	// 1. Resolve 'all' group
	if all, ok := inv.Groups["all"]; ok {
		for k, v := range all.Vars {
			effective[k] = v
		}
	}

	// 2. Find all groups and their ancestors
	groups := FindHostGroups(inv, hostname)

	// To handle precedence correctly, we should merge from "least specific" to "most specific".
	// Ansible's actual precedence is complex, but a simple depth-based or discovery-order
	// merge is a good start. We'll use the order found by findHostGroups.
	for _, gName := range groups {
		if gName == "all" {
			continue
		}
		if g, ok := inv.Groups[gName]; ok {
			for k, v := range g.Vars {
				effective[k] = v
			}
		}
	}

	// 3. Host-specific vars (highest precedence)
	if h, ok := inv.Hosts[hostname]; ok {
		for k, v := range h.Vars {
			effective[k] = v
		}
	}

	return effective
}

// FindHostGroups returns all groups (including ancestors) a host belongs to
func FindHostGroups(inv *models.Inventory, hostname string) []string {
	var direct []string
	for gName, g := range inv.Groups {
		for _, h := range g.Hosts {
			if h == hostname {
				direct = append(direct, gName)
				break
			}
		}
	}

	// Find ancestors
	allGroups := make(map[string]bool)
	for _, g := range direct {
		allGroups[g] = true
		for _, ancestor := range findAncestors(inv, g) {
			allGroups[ancestor] = true
		}
	}

	var result []string
	for g := range allGroups {
		result = append(result, g)
	}
	return topoSortGroups(inv, result)
}

func topoSortGroups(inv *models.Inventory, groups []string) []string {
	// Sort alphabetically first for deterministic baseline
	sort.Strings(groups)

	inSet := make(map[string]bool)
	for _, g := range groups {
		inSet[g] = true
	}

	var result []string
	visited := make(map[string]bool)
	temp := make(map[string]bool)

	var visit func(string)
	visit = func(node string) {
		if visited[node] {
			return
		}
		if temp[node] {
			return
		}
		temp[node] = true

		var parents []string
		for pName, g := range inv.Groups {
			if !inSet[pName] {
				continue
			}
			for _, child := range g.Children {
				if child == node {
					parents = append(parents, pName)
					break
				}
			}
		}
		sort.Strings(parents)

		for _, p := range parents {
			visit(p)
		}

		temp[node] = false
		visited[node] = true
		result = append(result, node)
	}

	for _, g := range groups {
		visit(g)
	}

	return result
}

func findAncestors(inv *models.Inventory, childName string) []string {
	var ancestors []string
	for gName, g := range inv.Groups {
		for _, child := range g.Children {
			if child == childName {
				ancestors = append(ancestors, gName)
				ancestors = append(ancestors, findAncestors(inv, gName)...)
				break
			}
		}
	}
	return ancestors
}

// ResolveNestedVar looks up a value in a map using dot notation (e.g. "networking.ip")
func ResolveNestedVar(vars map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = vars

	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current, ok = m[part]
		if !ok {
			return nil
		}
	}
	return current
}

// InitInventory scaffolds a standard Ansible inventory directory structure
func InitInventory(baseDir string) error {
	dirs := []string{
		filepath.Join(baseDir, "host_vars"),
		filepath.Join(baseDir, "group_vars"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", d, err)
		}
	}

	mainPath := filepath.Join(baseDir, "main.yaml")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		// Create empty main.yaml with 'all' group
		data := map[string]interface{}{
			"all": map[string]interface{}{
				"hosts": make(map[string]interface{}),
			},
		}
		bytes, err := yaml.Marshal(data)
		if err != nil {
			return err
		}
		return os.WriteFile(mainPath, bytes, 0644)
	}

	return nil
}

// GetHostVarsPath returns the existing path to a host's vars file, or a default one
func GetHostVarsPath(baseDir, hostname string) string {
	paths := []string{
		filepath.Join(baseDir, "host_vars", hostname+".yaml"),
		filepath.Join(baseDir, "host_vars", hostname+".yml"),
		filepath.Join(baseDir, "host_vars", hostname),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join(baseDir, "host_vars", hostname+".yaml")
}

// GetGroupVarsPath returns the existing path to a group's vars file, or a default one
func GetGroupVarsPath(baseDir, groupname string) string {
	paths := []string{
		filepath.Join(baseDir, "group_vars", groupname+".yaml"),
		filepath.Join(baseDir, "group_vars", groupname+".yml"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join(baseDir, "group_vars", groupname+".yaml")
}

// GetHostVars reads variables for a specific host from host_vars/<hostname>.yaml
func GetHostVars(baseDir, hostname string) (map[string]interface{}, error) {
	paths := []string{
		filepath.Join(baseDir, "host_vars", hostname+".yaml"),
		filepath.Join(baseDir, "host_vars", hostname+".yml"),
		filepath.Join(baseDir, "host_vars", hostname), // Ansible also supports dir with main.yaml, but we'll stick to files for now
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			data, err := os.ReadFile(p)
			if err != nil {
				return nil, err
			}
			var vars map[string]interface{}
			if err := yaml.Unmarshal(data, &vars); err != nil {
				return nil, err
			}
			return vars, nil
		}
	}
	return make(map[string]interface{}), nil
}

// GetGroupVars reads variables for a specific group from group_vars/<groupname>.yaml
func GetGroupVars(baseDir, groupname string) (map[string]interface{}, error) {
	paths := []string{
		filepath.Join(baseDir, "group_vars", groupname+".yaml"),
		filepath.Join(baseDir, "group_vars", groupname+".yml"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			data, err := os.ReadFile(p)
			if err != nil {
				return nil, err
			}
			var vars map[string]interface{}
			if err := yaml.Unmarshal(data, &vars); err != nil {
				return nil, err
			}
			return vars, nil
		}
	}
	return make(map[string]interface{}), nil
}

// AddHostToMain adds a host to the 'all' group in main.yaml
func AddHostToMain(baseDir, hostname string) error {
	mainPath := filepath.Join(baseDir, "main.yaml")
	data, err := os.ReadFile(mainPath)
	if err != nil {
		return err
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw == nil {
		raw = make(map[string]interface{})
	}

	all, ok := raw["all"].(map[string]interface{})
	if !ok || all == nil {
		all = make(map[string]interface{})
		raw["all"] = all
	}

	hosts, ok := all["hosts"].(map[string]interface{})
	if !ok || hosts == nil {
		hosts = make(map[string]interface{})
		all["hosts"] = hosts
	}

	hosts[hostname] = make(map[string]interface{})

	bytes, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}
	return os.WriteFile(mainPath, bytes, 0644)
}

// SaveGroup writes a group's data to a specific YAML file
func SaveGroup(baseDir, groupName string, group *models.Group) error {
	if group == nil {
		return fmt.Errorf("group cannot be nil")
	}
	path := filepath.Join(baseDir, groupName+".yaml")

	// Construct Ansible-compatible group map
	hostsMap := make(map[string]interface{})
	for _, h := range group.Hosts {
		hostsMap[h] = make(map[string]interface{})
	}

	childrenMap := make(map[string]interface{})
	for _, c := range group.Children {
		childrenMap[c] = make(map[string]interface{})
	}

	gData := map[string]interface{}{
		"hosts": hostsMap,
	}
	if len(childrenMap) > 0 {
		gData["children"] = childrenMap
	}
	if len(group.Vars) > 0 {
		gData["vars"] = group.Vars
	}

	data := map[string]interface{}{
		groupName: gData,
	}

	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

// AssignHostToGroup adds a host to a specific group file
func AssignHostToGroup(baseDir, hostname, groupName string) error {
	inv, err := LoadInventory(baseDir)
	if err != nil {
		return err
	}
	group, ok := inv.Groups[groupName]
	if !ok {
		return fmt.Errorf("group %q not found in inventory", groupName)
	}

	// Check if already assigned
	for _, h := range group.Hosts {
		if h == hostname {
			return nil
		}
	}

	group.Hosts = append(group.Hosts, hostname)
	return SaveGroup(baseDir, groupName, group)
}

// AssignGroupToGroup adds a child group to a parent group
func AssignGroupToGroup(baseDir, childName, parentName string) error {
	inv, err := LoadInventory(baseDir)
	if err != nil {
		return err
	}
	parent, ok := inv.Groups[parentName]
	if !ok {
		return fmt.Errorf("parent group %q not found in inventory", parentName)
	}

	if _, ok := inv.Groups[childName]; !ok {
		return fmt.Errorf("child group %q not found in inventory", childName)
	}

	if childName == parentName {
		return fmt.Errorf("cannot nest a group within itself")
	}

	// Check if already assigned
	for _, c := range parent.Children {
		if c == childName {
			return nil
		}
	}

	parent.Children = append(parent.Children, childName)
	return SaveGroup(baseDir, parentName, parent)
}

// SetHostVar sets a variable for a host in host_vars/<hostname>.yaml
func SetHostVar(baseDir, hostname, key, value string) error {
	inv, err := LoadInventory(baseDir)
	if err != nil {
		return err
	}
	if _, ok := inv.Hosts[hostname]; !ok {
		return fmt.Errorf("host %q not found in inventory", hostname)
	}

	path := GetHostVarsPath(baseDir, hostname)
	vars, err := GetHostVars(baseDir, hostname)
	if err != nil {
		return err
	}
	if vars == nil {
		vars = make(map[string]interface{})
	}
	SetNestedVar(vars, key, value)

	bytes, err := yaml.Marshal(vars)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

// SetGroupVar sets a variable for a group in group_vars/<groupname>.yaml
func SetGroupVar(baseDir, groupname, key, value string) error {
	inv, err := LoadInventory(baseDir)
	if err != nil {
		return err
	}
	if _, ok := inv.Groups[groupname]; !ok {
		return fmt.Errorf("group %q not found in inventory", groupname)
	}

	path := GetGroupVarsPath(baseDir, groupname)
	vars, err := GetGroupVars(baseDir, groupname)
	if err != nil {
		return err
	}
	if vars == nil {
		vars = make(map[string]interface{})
	}
	SetNestedVar(vars, key, value)

	bytes, err := yaml.Marshal(vars)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

// DeleteHostVar removes a variable for a host in host_vars/<hostname>.yaml
func DeleteHostVar(baseDir, hostname, key string) error {
	vars, err := GetHostVars(baseDir, hostname)
	if err != nil {
		return err
	}
	if vars == nil {
		return nil
	}
	DeleteNestedVar(vars, key)

	path := GetHostVarsPath(baseDir, hostname)
	bytes, err := yaml.Marshal(vars)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

// DeleteGroupVar removes a variable for a group in group_vars/<groupname>.yaml
func DeleteGroupVar(baseDir, groupname, key string) error {
	vars, err := GetGroupVars(baseDir, groupname)
	if err != nil {
		return err
	}
	if vars == nil {
		return nil
	}
	DeleteNestedVar(vars, key)

	path := GetGroupVarsPath(baseDir, groupname)
	bytes, err := yaml.Marshal(vars)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

// RemoveHost removes a host from all inventory files
func RemoveHost(baseDir, hostname string) error {
	inv, err := LoadInventory(baseDir)
	if err != nil {
		return err
	}
	if _, ok := inv.Hosts[hostname]; !ok {
		return fmt.Errorf("host %q not found in inventory", hostname)
	}

	// 1. Remove from main.yaml
	mainPath := filepath.Join(baseDir, "main.yaml")
	data, err := os.ReadFile(mainPath)
	if err != nil {
		return err
	}
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}
	if all, ok := raw["all"].(map[string]interface{}); ok && all != nil {
		if hosts, ok := all["hosts"].(map[string]interface{}); ok && hosts != nil {
			delete(hosts, hostname)
		}
	}
	bytes, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}
	if err := os.WriteFile(mainPath, bytes, 0644); err != nil {
		return err
	}

	// 2. Remove from all other group files
	files, err := filepath.Glob(filepath.Join(baseDir, "*.yaml"))
	if err != nil {
		return err
	}
	ymlFiles, err := filepath.Glob(filepath.Join(baseDir, "*.yml"))
	if err != nil {
		return err
	}
	files = append(files, ymlFiles...)

	for _, f := range files {
		if filepath.Base(f) == "main.yaml" {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var gRaw map[string]interface{}
		if err := yaml.Unmarshal(data, &gRaw); err != nil {
			continue
		}

		changed := false
		for _, content := range gRaw {
			if cMap, ok := content.(map[string]interface{}); ok && cMap != nil {
				if hosts, ok := cMap["hosts"].(map[string]interface{}); ok && hosts != nil {
					if _, ok := hosts[hostname]; ok {
						delete(hosts, hostname)
						changed = true
					}
				}
			}
		}
		if changed {
			bytes, err := yaml.Marshal(gRaw)
			if err != nil {
				return err
			}
			if err := os.WriteFile(f, bytes, 0644); err != nil {
				return err
			}
		}
	}

	// 3. Remove host_vars file
	_ = os.Remove(filepath.Join(baseDir, "host_vars", hostname+".yml"))
	_ = os.Remove(filepath.Join(baseDir, "host_vars", hostname+".yaml"))

	return nil
}

// RemoveGroup deletes a group file and removes references
func RemoveGroup(baseDir, groupName string) error {
	inv, err := LoadInventory(baseDir)
	if err != nil {
		return err
	}
	if _, ok := inv.Groups[groupName]; !ok {
		return fmt.Errorf("group %q not found in inventory", groupName)
	}

	path := filepath.Join(baseDir, groupName+".yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = filepath.Join(baseDir, groupName+".yml")
	}
	_ = os.Remove(path)
	_ = os.Remove(filepath.Join(baseDir, "group_vars", groupName+".yml"))
	_ = os.Remove(filepath.Join(baseDir, "group_vars", groupName+".yaml"))

	// Also remove from children of other groups
	for otherName, g := range inv.Groups {
		if otherName == groupName {
			continue
		}

		found := false
		var newChildren []string
		for _, child := range g.Children {
			if child == groupName {
				found = true
			} else {
				newChildren = append(newChildren, child)
			}
		}

		if found {
			g.Children = newChildren
			if err := SaveGroup(baseDir, otherName, g); err != nil {
				return fmt.Errorf("failed to update parent group %s: %v", otherName, err)
			}
		}
	}

	return nil
}

// MergeHostVars merges a map of variables into a host's existing host_vars file (recursive)
func MergeHostVars(baseDir, hostname string, newVars map[string]interface{}) error {
	existing, err := GetHostVars(baseDir, hostname)
	if err != nil {
		return err
	}
	if existing == nil {
		existing = make(map[string]interface{})
	}

	DeepMerge(existing, newVars)
	return SaveHostVars(baseDir, hostname, existing)
}

// SaveHostVars overwrites a host's variables in host_vars/<hostname>.yaml
func SaveHostVars(baseDir, hostname string, vars map[string]interface{}) error {
	path := GetHostVarsPath(baseDir, hostname)
	bytes, err := yaml.Marshal(vars)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

// MergeGroupVars merges a map of variables into a group's existing group_vars file (recursive)
func MergeGroupVars(baseDir, groupname string, newVars map[string]interface{}) error {
	existing, err := GetGroupVars(baseDir, groupname)
	if err != nil {
		return err
	}
	if existing == nil {
		existing = make(map[string]interface{})
	}

	DeepMerge(existing, newVars)
	return SaveGroupVars(baseDir, groupname, existing)
}

// DeepMerge recursively merges src into dst, returns true if any changes were made
func DeepMerge(dst, src map[string]interface{}) bool {
	changed := false
	for k, v := range src {
		if srcMap, ok := v.(map[string]interface{}); ok {
			if dstMap, ok := dst[k].(map[string]interface{}); ok {
				if DeepMerge(dstMap, srcMap) {
					changed = true
				}
				continue
			}
		}
		// Only write if the value has actually changed to avoid spurious UpdatedHosts detections
		if existing, ok := dst[k]; ok && valuesEqual(existing, v) {
			continue
		}
		dst[k] = v
		changed = true
	}
	return changed
}

// valuesEqual compares two values for semantic equality, handling type differences
func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Use reflect.DeepEqual for most cases
	if reflect.DeepEqual(a, b) {
		return true
	}

	// Handle numeric type mismatches (JSON float64 vs YAML int)
	if isNumber(a) && isNumber(b) {
		return toFloat64(a) == toFloat64(b)
	}

	// Handle common type mismatches (e.g., map[string]interface{} vs map[string]string)
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	if aVal.Kind() == bVal.Kind() && aVal.Kind() == reflect.Map {
		if aVal.Type().Key() != bVal.Type().Key() || aVal.Type().Elem() != bVal.Type().Elem() {
			// Different map types - compare by converting to same type
			return mapsEqual(a, b)
		}
	}

	if aVal.Kind() == bVal.Kind() && aVal.Kind() == reflect.Slice {
		if aVal.Type().Elem() != bVal.Type().Elem() {
			return slicesEqual(a, b)
		}
	}

	return false
}

func isNumber(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	}
	return false
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	}
	return 0
}

func mapsEqual(a, b interface{}) bool {
	aMap := reflect.ValueOf(a)
	bMap := reflect.ValueOf(b)
	
	if aMap.Len() != bMap.Len() {
		return false
	}
	
	for _, key := range aMap.MapKeys() {
		aVal := aMap.MapIndex(key)
		bVal := bMap.MapIndex(key)
		if !bVal.IsValid() {
			return false
		}
		if !valuesEqual(aVal.Interface(), bVal.Interface()) {
			return false
		}
	}
	return true
}

func slicesEqual(a, b interface{}) bool {
	aSlice := reflect.ValueOf(a)
	bSlice := reflect.ValueOf(b)
	
	if aSlice.Len() != bSlice.Len() {
		return false
	}
	
	for i := 0; i < aSlice.Len(); i++ {
		if !valuesEqual(aSlice.Index(i).Interface(), bSlice.Index(i).Interface()) {
			return false
		}
	}
	return true
}

// SaveGroupVars overwrites a group's variables in group_vars/<groupname>.yaml
func SaveGroupVars(baseDir, groupname string, vars map[string]interface{}) error {
	path := GetGroupVarsPath(baseDir, groupname)
	bytes, err := yaml.Marshal(vars)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

// MoveHost migrates a host and its context from one inventory to another
func MoveHost(sourceDir, destDir, hostname string) error {
	// 1. Resolve source and ensure host exists
	sourceInv, err := LoadInventory(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to load source inventory: %v", err)
	}
	if _, ok := sourceInv.Hosts[hostname]; !ok {
		return fmt.Errorf("host %q not found in source inventory", hostname)
	}

	// 2. Ensure dest inventory is valid
	if _, err := os.Stat(filepath.Join(destDir, "main.yaml")); os.IsNotExist(err) {
		return fmt.Errorf("destination directory %q is not a valid inventory (missing main.yaml)", destDir)
	}
	destInv, err := LoadInventory(destDir)
	if err != nil {
		return fmt.Errorf("failed to load destination inventory: %v", err)
	}

	// 3. Migrate host_vars
	hv, err := GetHostVars(sourceDir, hostname)
	if err != nil {
		return fmt.Errorf("failed to get source host_vars: %v", err)
	}
	if len(hv) > 0 {
		if err := SaveHostVars(destDir, hostname, hv); err != nil {
			return fmt.Errorf("failed to migrate host_vars: %v", err)
		}
	}

	// 4. Find all groups host belongs to in source
	var sourceGroups []string
	for gName, g := range sourceInv.Groups {
		if gName == "all" {
			continue
		}
		for _, h := range g.Hosts {
			if h == hostname {
				sourceGroups = append(sourceGroups, gName)
				break
			}
		}
	}

	// 5. Migrate membership and merge groups/group_vars
	for _, gName := range sourceGroups {
		sourceGv, err := GetGroupVars(sourceDir, gName)
		if err != nil {
			return err
		}

		if _, ok := destInv.Groups[gName]; !ok {
			// Auto-create group in destination
			newG := &models.Group{Name: gName}
			if len(sourceGv) > 0 {
				if err := SaveGroupVars(destDir, gName, sourceGv); err != nil {
					return err
				}
				newG.Vars = sourceGv
			}
			if err := SaveGroup(destDir, gName, newG); err != nil {
				return err
			}
			destInv.Groups[gName] = newG
		} else {
			// Merge group_vars if group already exists
			destGv, err := GetGroupVars(destDir, gName)
			if err != nil {
				return err
			}
			if len(sourceGv) > 0 {
				// Dest wins on conflict, but new keys from source are added
				merged := make(map[string]interface{})
				for k, v := range sourceGv {
					merged[k] = v
				}
				for k, v := range destGv {
					merged[k] = v
				}
				if err := SaveGroupVars(destDir, gName, merged); err != nil {
					return err
				}
			}
		}
		// Assign to group in dest
		if err := AssignHostToGroup(destDir, hostname, gName); err != nil {
			return fmt.Errorf("failed to assign host to group %s in dest: %v", gName, err)
		}
	}

	// 6. Add to main.yaml in destination
	if err := AddHostToMain(destDir, hostname); err != nil {
		return fmt.Errorf("failed to add host to destination main.yaml: %v", err)
	}

	// 7. Cleanup source
	return RemoveHost(sourceDir, hostname)
}

// SyncVars copies variables for a host or group from source inventory to destination
func SyncVars(sourceDir, destDir, eType, name string) error {
	var vars map[string]interface{}
	var err error

	switch eType {
	case "host":
		vars, err = GetHostVars(sourceDir, name)
		if err == nil && len(vars) > 0 {
			err = SaveHostVars(destDir, name, vars)
		}
	case "group":
		vars, err = GetGroupVars(sourceDir, name)
		if err == nil && len(vars) > 0 {
			err = SaveGroupVars(destDir, name, vars)
		}
	default:
		return fmt.Errorf("invalid entity type %q", eType)
	}

	return err
}

// CopyHost clones an existing host and its vars/group memberships to a new name
func CopyHost(baseDir, sourceName, destName string) error {
	inv, err := LoadInventory(baseDir)
	if err != nil {
		return err
	}

	if _, ok := inv.Hosts[sourceName]; !ok {
		return fmt.Errorf("source host %q not found", sourceName)
	}
	if _, ok := inv.Hosts[destName]; ok {
		return fmt.Errorf("destination host %q already exists", destName)
	}

	// 1. Clone host_vars
	hv, err := GetHostVars(baseDir, sourceName)
	if err != nil {
		return fmt.Errorf("failed to get source host_vars: %v", err)
	}
	if len(hv) > 0 {
		if err := SaveHostVars(baseDir, destName, hv); err != nil {
			return fmt.Errorf("failed to clone host_vars: %v", err)
		}
	}

	// 2. Find all groups source belongs to (excluding "all")
	var sourceGroups []string
	for gName, g := range inv.Groups {
		if gName == "all" {
			continue
		}
		for _, h := range g.Hosts {
			if h == sourceName {
				sourceGroups = append(sourceGroups, gName)
				break
			}
		}
	}

	// 3. Add dest host to all those groups
	for _, gName := range sourceGroups {
		if err := AssignHostToGroup(baseDir, destName, gName); err != nil {
			return fmt.Errorf("failed to assign host to group %s: %v", gName, err)
		}
	}

	// 4. Add to main.yaml
	return AddHostToMain(baseDir, destName)
}

// CopyGroup clones an existing group to a new name, including structure and group_vars
func CopyGroup(baseDir, sourceName, destName string) error {
	inv, err := LoadInventory(baseDir)
	if err != nil {
		return err
	}

	source, ok := inv.Groups[sourceName]
	if !ok {
		return fmt.Errorf("source group %q not found", sourceName)
	}

	if _, ok := inv.Groups[destName]; ok {
		return fmt.Errorf("destination group %q already exists", destName)
	}

	// 1. Create new group structure
	destGroup := &models.Group{
		Name:     destName,
		Hosts:    append([]string{}, source.Hosts...),
		Children: append([]string{}, source.Children...),
		Vars:     make(map[string]interface{}),
	}
	for k, v := range source.Vars {
		destGroup.Vars[k] = v
	}

	if err := SaveGroup(baseDir, destName, destGroup); err != nil {
		return fmt.Errorf("failed to save cloned group: %v", err)
	}

	// 2. Clone group_vars file if it exists (LoadInventory already merged inline vars, but we want the file too)
	gv, err := GetGroupVars(baseDir, sourceName)
	if err != nil {
		return fmt.Errorf("failed to get source group_vars: %v", err)
	}
	if len(gv) > 0 {
		if err := SaveGroupVars(baseDir, destName, gv); err != nil {
			return fmt.Errorf("failed to clone group_vars: %v", err)
		}
	}

	return nil
}

// LoadInventory reads all YAML files in the directory and merges them
func LoadInventory(baseDir string) (*models.Inventory, error) {
	inv := &models.Inventory{
		Hosts:  make(map[string]*models.Host),
		Groups: make(map[string]*models.Group),
	}

	files, err := filepath.Glob(filepath.Join(baseDir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	ymlFiles, err := filepath.Glob(filepath.Join(baseDir, "*.yml"))
	if err != nil {
		return nil, err
	}
	files = append(files, ymlFiles...)

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		var raw map[string]interface{}
		if err := yaml.Unmarshal(data, &raw); err != nil {
			continue // Skip invalid YAML
		}

		for groupName, content := range raw {
			group, ok := inv.Groups[groupName]
			if !ok {
				group = &models.Group{Name: groupName}
				inv.Groups[groupName] = group
			}

			if cMap, ok := content.(map[string]interface{}); ok && cMap != nil {
				// Handle hosts in group
				if hosts, ok := cMap["hosts"].(map[string]interface{}); ok && hosts != nil {
					for hName := range hosts {
						found := false
						for _, existing := range group.Hosts {
							if existing == hName {
								found = true
								break
							}
						}
						if !found {
							group.Hosts = append(group.Hosts, hName)
						}

						if _, ok := inv.Hosts[hName]; !ok {
							inv.Hosts[hName] = &models.Host{Name: hName}
						}
					}
				}
				// Handle children groups
				if children, ok := cMap["children"].(map[string]interface{}); ok && children != nil {
					for childName := range children {
						found := false
						for _, existing := range group.Children {
							if existing == childName {
								found = true
								break
							}
						}
						if !found {
							group.Children = append(group.Children, childName)
						}
					}
				}
				// Handle inline vars (though Ansible often uses group_vars files)
				if v, ok := cMap["vars"].(map[string]interface{}); ok && v != nil {
					if group.Vars == nil {
						group.Vars = make(map[string]interface{})
					}
					for k, val := range v {
						group.Vars[k] = val
					}
				}
			}
		}
	}

	// Load external vars
	for gName, g := range inv.Groups {
		gv, err := GetGroupVars(baseDir, gName)
		if err != nil {
			return nil, err
		}
		if len(gv) > 0 {
			if g.Vars == nil {
				g.Vars = make(map[string]interface{})
			}
			for k, v := range gv {
				g.Vars[k] = v
			}
		}
	}

	for hName, h := range inv.Hosts {
		hv, err := GetHostVars(baseDir, hName)
		if err != nil {
			return nil, err
		}
		if len(hv) > 0 {
			h.Vars = hv
		}
	}

	return inv, nil
}

// SetNestedVar sets a value in a nested map using dot-notation (e.g. "a.b.c")
func SetNestedVar(m map[string]interface{}, keyPath string, value interface{}) {
	parts := strings.Split(keyPath, ".")
	curr := m
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		next, ok := curr[part]
		if !ok {
			nextMap := make(map[string]interface{})
			curr[part] = nextMap
			curr = nextMap
		} else if nextMap, ok := next.(map[string]interface{}); ok {
			curr = nextMap
		} else {
			// Overwrite non-map value with a map
			nextMap := make(map[string]interface{})
			curr[part] = nextMap
			curr = nextMap
		}
	}
	curr[parts[len(parts)-1]] = value
}

// DeleteNestedVar deletes a key in a nested map using dot-notation, pruning empty parent maps
func DeleteNestedVar(m map[string]interface{}, keyPath string) bool {
	parts := strings.Split(keyPath, ".")
	if len(parts) == 0 {
		return false
	}

	var remove func(map[string]interface{}, []string) bool
	remove = func(curr map[string]interface{}, path []string) bool {
		if len(path) == 1 {
			delete(curr, path[0])
			return len(curr) == 0
		}

		part := path[0]
		next, ok := curr[part]
		if !ok {
			return false
		}

		nextMap, ok := next.(map[string]interface{})
		if !ok {
			return false
		}

		isEmpty := remove(nextMap, path[1:])
		if isEmpty {
			delete(curr, part)
		}
		return len(curr) == 0
	}

	remove(m, parts)
	return true
}

// LoadRosterConf loads NetBox sync settings from roster.conf, .roster.conf, or ~/.roster.conf
func LoadRosterConf() (url, token, filter string) {
	return loadConfValues("netbox_url", "url"), loadConfValues("netbox_token", "token"), loadConfValues("netbox_filter", "filter")
}

// LoadSMTPConfig loads SMTP settings from roster.conf, .roster.conf, or ~/.roster.conf
func LoadSMTPConfig() (host, port, from, user, pass string) {
	return loadConfValues("smtp_host"), loadConfValues("smtp_port"), loadConfValues("smtp_from"), loadConfValues("smtp_user"), loadConfValues("smtp_pass")
}

// loadConfValues is a helper that returns values for the given keys from roster conf files
func loadConfValues(keys ...string) string {
	paths := []string{"roster.conf", ".roster.conf"}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".roster.conf"))
	}

	var data []byte
	var err error
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if len(data) == 0 {
		return ""
	}

	vals := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		var cleanLine strings.Builder
		inQuote := false
		var quoteChar rune
		for _, r := range line {
			if r == '"' || r == '\'' {
				if !inQuote {
					inQuote = true
					quoteChar = r
				} else if r == quoteChar {
					inQuote = false
				}
			}
			if (r == '#' || r == ';') && !inQuote {
				break
			}
			cleanLine.WriteRune(r)
		}
		line = strings.TrimSpace(cleanLine.String())

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.ToLower(strings.TrimSpace(parts[0]))
		v := strings.TrimSpace(parts[1])
		v = strings.Trim(v, `"'`)

		for _, key := range keys {
			if k == key {
				vals[key] = v
			}
		}
	}

	for _, key := range keys {
		if v, ok := vals[key]; ok {
			return v
		}
	}
	return ""
}

// FlattenMap flattens a nested map into a single level map with dot-notation keys
func FlattenMap(m map[string]interface{}, prefix string, result map[string]interface{}) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		if childMap, ok := v.(map[string]interface{}); ok {
			FlattenMap(childMap, key, result)
		} else {
			result[key] = v
		}
	}
}
