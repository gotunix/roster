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

package store

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Try to lock with a 10-second timeout
	locked, err := globalFlock.TryLockContext(nil, 100*time.Millisecond)
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
		globalFlock.Unlock()
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
	groups := findHostGroups(inv, hostname)

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

// findHostGroups returns all groups (including ancestors) a host belongs to
func findHostGroups(inv *models.Inventory, hostname string) []string {
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
		bytes, _ := yaml.Marshal(data)
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
	return filepath.Join(baseDir, "host_vars", hostname+".yml")
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
	return filepath.Join(baseDir, "group_vars", groupname+".yml")
}

// GetHostVars reads variables for a specific host from host_vars/<hostname>.yml
func GetHostVars(baseDir, hostname string) (map[string]interface{}, error) {
	paths := []string{
		filepath.Join(baseDir, "host_vars", hostname+".yaml"),
		filepath.Join(baseDir, "host_vars", hostname+".yml"),
		filepath.Join(baseDir, "host_vars", hostname), // Ansible also supports dir with main.yml, but we'll stick to files for now
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			data, _ := os.ReadFile(p)
			var vars map[string]interface{}
			yaml.Unmarshal(data, &vars)
			return vars, nil
		}
	}
	return make(map[string]interface{}), nil
}

// GetGroupVars reads variables for a specific group from group_vars/<groupname>.yml
func GetGroupVars(baseDir, groupname string) (map[string]interface{}, error) {
	paths := []string{
		filepath.Join(baseDir, "group_vars", groupname+".yaml"),
		filepath.Join(baseDir, "group_vars", groupname+".yml"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			data, _ := os.ReadFile(p)
			var vars map[string]interface{}
			yaml.Unmarshal(data, &vars)
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
	yaml.Unmarshal(data, &raw)
	if raw == nil {
		raw = make(map[string]interface{})
	}

	all, ok := raw["all"].(map[string]interface{})
	if !ok {
		all = make(map[string]interface{})
		raw["all"] = all
	}

	hosts, ok := all["hosts"].(map[string]interface{})
	if !ok {
		hosts = make(map[string]interface{})
		all["hosts"] = hosts
	}

	hosts[hostname] = make(map[string]interface{})

	bytes, _ := yaml.Marshal(raw)
	return os.WriteFile(mainPath, bytes, 0644)
}

// SaveGroup writes a group's data to a specific YAML file
func SaveGroup(baseDir, groupName string, group *models.Group) error {
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

	bytes, _ := yaml.Marshal(data)
	return os.WriteFile(path, bytes, 0644)
}

// AssignHostToGroup adds a host to a specific group file
func AssignHostToGroup(baseDir, hostname, groupName string) error {
	inv, _ := LoadInventory(baseDir)
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
	inv, _ := LoadInventory(baseDir)
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

// SetHostVar sets a variable for a host in host_vars/<hostname>.yml
func SetHostVar(baseDir, hostname, key, value string) error {
	inv, err := LoadInventory(baseDir)
	if err != nil {
		return err
	}
	if _, ok := inv.Hosts[hostname]; !ok {
		return fmt.Errorf("host %q not found in inventory", hostname)
	}

	path := filepath.Join(baseDir, "host_vars", hostname+".yml")
	vars, _ := GetHostVars(baseDir, hostname)
	if vars == nil {
		vars = make(map[string]interface{})
	}
	vars[key] = value

	bytes, _ := yaml.Marshal(vars)
	return os.WriteFile(path, bytes, 0644)
}

// SetGroupVar sets a variable for a group in group_vars/<groupname>.yml
func SetGroupVar(baseDir, groupname, key, value string) error {
	inv, err := LoadInventory(baseDir)
	if err != nil {
		return err
	}
	if _, ok := inv.Groups[groupname]; !ok {
		return fmt.Errorf("group %q not found in inventory", groupname)
	}

	path := filepath.Join(baseDir, "group_vars", groupname+".yml")
	vars, _ := GetGroupVars(baseDir, groupname)
	if vars == nil {
		vars = make(map[string]interface{})
	}
	vars[key] = value

	bytes, _ := yaml.Marshal(vars)
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
	data, _ := os.ReadFile(mainPath)
	var raw map[string]interface{}
	yaml.Unmarshal(data, &raw)
	if all, ok := raw["all"].(map[string]interface{}); ok {
		if hosts, ok := all["hosts"].(map[string]interface{}); ok {
			delete(hosts, hostname)
		}
	}
	bytes, _ := yaml.Marshal(raw)
	os.WriteFile(mainPath, bytes, 0644)

	// 2. Remove from all other group files
	files, _ := filepath.Glob(filepath.Join(baseDir, "*.yaml"))
	ymlFiles, _ := filepath.Glob(filepath.Join(baseDir, "*.yml"))
	files = append(files, ymlFiles...)

	for _, f := range files {
		if filepath.Base(f) == "main.yaml" {
			continue
		}
		data, _ := os.ReadFile(f)
		var gRaw map[string]interface{}
		yaml.Unmarshal(data, &gRaw)

		changed := false
		for _, content := range gRaw {
			if cMap, ok := content.(map[string]interface{}); ok {
				if hosts, ok := cMap["hosts"].(map[string]interface{}); ok {
					if _, ok := hosts[hostname]; ok {
						delete(hosts, hostname)
						changed = true
					}
				}
			}
		}
		if changed {
			bytes, _ := yaml.Marshal(gRaw)
			os.WriteFile(f, bytes, 0644)
		}
	}

	// 3. Remove host_vars file
	os.Remove(filepath.Join(baseDir, "host_vars", hostname+".yml"))
	os.Remove(filepath.Join(baseDir, "host_vars", hostname+".yaml"))

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

	path := filepath.Join(baseDir, groupName+".yml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = filepath.Join(baseDir, groupName+".yaml")
	}
	os.Remove(path)
	os.Remove(filepath.Join(baseDir, "group_vars", groupName+".yml"))
	os.Remove(filepath.Join(baseDir, "group_vars", groupName+".yaml"))

	// Also remove from children of other groups (TODO)
	return nil
}

// SaveHostVars overwrites a host's variables in host_vars/<hostname>.yml
func SaveHostVars(baseDir, hostname string, vars map[string]interface{}) error {
	path := filepath.Join(baseDir, "host_vars", hostname+".yml")
	bytes, _ := yaml.Marshal(vars)
	return os.WriteFile(path, bytes, 0644)
}

// SaveGroupVars overwrites a group's variables in group_vars/<groupname>.yml
func SaveGroupVars(baseDir, groupname string, vars map[string]interface{}) error {
	path := filepath.Join(baseDir, "group_vars", groupname+".yml")
	bytes, _ := yaml.Marshal(vars)
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
	hv, _ := GetHostVars(sourceDir, hostname)
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
		sourceGv, _ := GetGroupVars(sourceDir, gName)

		if _, ok := destInv.Groups[gName]; !ok {
			// Auto-create group in destination
			newG := &models.Group{Name: gName}
			if len(sourceGv) > 0 {
				SaveGroupVars(destDir, gName, sourceGv)
				newG.Vars = sourceGv
			}
			SaveGroup(destDir, gName, newG)
			destInv.Groups[gName] = newG
		} else {
			// Merge group_vars if group already exists
			destGv, _ := GetGroupVars(destDir, gName)
			if len(sourceGv) > 0 {
				// Dest wins on conflict, but new keys from source are added
				merged := make(map[string]interface{})
				for k, v := range sourceGv {
					merged[k] = v
				}
				for k, v := range destGv {
					merged[k] = v
				}
				SaveGroupVars(destDir, gName, merged)
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

	if eType == "host" {
		vars, err = GetHostVars(sourceDir, name)
		if err == nil && len(vars) > 0 {
			err = SaveHostVars(destDir, name, vars)
		}
	} else if eType == "group" {
		vars, err = GetGroupVars(sourceDir, name)
		if err == nil && len(vars) > 0 {
			err = SaveGroupVars(destDir, name, vars)
		}
	} else {
		return fmt.Errorf("invalid entity type %q", eType)
	}

	return err
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
	gv, _ := GetGroupVars(baseDir, sourceName)
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
	ymlFiles, _ := filepath.Glob(filepath.Join(baseDir, "*.yml"))
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

			if cMap, ok := content.(map[string]interface{}); ok {
				// Handle hosts in group
				if hosts, ok := cMap["hosts"].(map[string]interface{}); ok {
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
				if children, ok := cMap["children"].(map[string]interface{}); ok {
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
				if v, ok := cMap["vars"].(map[string]interface{}); ok {
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
		gv, _ := GetGroupVars(baseDir, gName)
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
		hv, _ := GetHostVars(baseDir, hName)
		if len(hv) > 0 {
			h.Vars = hv
		}
	}

	return inv, nil
}
