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
	"os"
	"path/filepath"
	"testing"

	"gotunix.net/roster/internal/models"
)

func TestInitInventory(t *testing.T) {
	tmpDir := t.TempDir()

	err := InitInventory(tmpDir)
	if err != nil {
		t.Fatalf("InitInventory failed: %v", err)
	}

	// Check directories
	dirs := []string{"host_vars", "group_vars"}
	for _, d := range dirs {
		if _, err := os.Stat(filepath.Join(tmpDir, d)); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", d)
		}
	}

	// Check main.yaml
	if _, err := os.Stat(filepath.Join(tmpDir, "main.yaml")); os.IsNotExist(err) {
		t.Error("expected main.yaml to exist")
	}
}

func TestHostManagement(t *testing.T) {
	tmpDir := t.TempDir()
	if err := InitInventory(tmpDir); err != nil {
		t.Fatalf("InitInventory failed: %v", err)
	}

	hostname := "test-host"

	// 1. Add host
	err := AddHostToMain(tmpDir, hostname)
	if err != nil {
		t.Fatalf("AddHostToMain failed: %v", err)
	}

	// 2. Load and verify
	inv, err := LoadInventory(tmpDir)
	if err != nil {
		t.Fatalf("LoadInventory failed: %v", err)
	}

	if _, ok := inv.Hosts[hostname]; !ok {
		t.Errorf("expected host %q to be in inventory", hostname)
	}

	// 3. Remove host
	err = RemoveHost(tmpDir, hostname)
	if err != nil {
		t.Fatalf("RemoveHost failed: %v", err)
	}

	// 4. Verify removed
	inv, _ = LoadInventory(tmpDir)
	if _, ok := inv.Hosts[hostname]; ok {
		t.Errorf("expected host %q to be removed from inventory", hostname)
	}
}

func TestGroupManagement(t *testing.T) {
	tmpDir := t.TempDir()
	if err := InitInventory(tmpDir); err != nil {
		t.Fatalf("InitInventory failed: %v", err)
	}

	groupName := "webservers"
	hostName := "srv1"

	// 1. Create group
	group := &models.Group{Name: groupName, Hosts: []string{hostName}}
	err := SaveGroup(tmpDir, groupName, group)
	if err != nil {
		t.Fatalf("SaveGroup failed: %v", err)
	}

	// 2. Verify existence and membership
	inv, _ := LoadInventory(tmpDir)
	g, ok := inv.Groups[groupName]
	if !ok {
		t.Fatalf("expected group %q to exist", groupName)
	}

	found := false
	for _, h := range g.Hosts {
		if h == hostName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected host %q to be in group %q", hostName, groupName)
	}

	// 3. Assign to group
	err = AssignHostToGroup(tmpDir, "srv2", groupName)
	if err != nil {
		t.Fatalf("AssignHostToGroup failed: %v", err)
	}

	inv, _ = LoadInventory(tmpDir)
	if len(inv.Groups[groupName].Hosts) != 2 {
		t.Errorf("expected 2 hosts in group, got %d", len(inv.Groups[groupName].Hosts))
	}
}

func TestVariables(t *testing.T) {
	tmpDir := t.TempDir()
	if err := InitInventory(tmpDir); err != nil {
		t.Fatalf("InitInventory failed: %v", err)
	}

	hostname := "srv1"
	groupName := "web"
	if err := AddHostToMain(tmpDir, hostname); err != nil {
		t.Fatalf("AddHostToMain failed: %v", err)
	}
	if err := SaveGroup(tmpDir, groupName, &models.Group{Name: groupName}); err != nil {
		t.Fatalf("SaveGroup failed: %v", err)
	}

	// 1. Host Vars
	err := SetHostVar(tmpDir, hostname, "ansible_user", "root")
	if err != nil {
		t.Fatalf("SetHostVar failed: %v", err)
	}

	hVars, _ := GetHostVars(tmpDir, hostname)
	if hVars["ansible_user"] != "root" {
		t.Errorf("expected ansible_user to be root, got %v", hVars["ansible_user"])
	}

	// 2. Group Vars
	err = SetGroupVar(tmpDir, groupName, "http_port", "80")
	if err != nil {
		t.Fatalf("SetGroupVar failed: %v", err)
	}

	gVars, _ := GetGroupVars(tmpDir, groupName)
	if gVars["http_port"] != "80" {
		t.Errorf("expected http_port to be 80, got %v", gVars["http_port"])
	}
}

func TestMoveHost(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()
	if err := InitInventory(sourceDir); err != nil {
		t.Fatalf("InitInventory failed: %v", err)
	}
	if err := InitInventory(destDir); err != nil {
		t.Fatalf("InitInventory failed: %v", err)
	}

	hostname := "migrating-host"
	if err := AddHostToMain(sourceDir, hostname); err != nil {
		t.Fatalf("AddHostToMain failed: %v", err)
	}
	if err := SetHostVar(sourceDir, hostname, "key", "val"); err != nil {
		t.Fatalf("SetHostVar failed: %v", err)
	}

	// Create a group in source and assign host
	if err := SaveGroup(sourceDir, "web", &models.Group{Name: "web", Hosts: []string{hostname}}); err != nil {
		t.Fatalf("SaveGroup failed: %v", err)
	}
	if err := SetGroupVar(sourceDir, "web", "port", "80"); err != nil {
		t.Fatalf("SetGroupVar failed: %v", err)
	}

	// MOVE
	err := MoveHost(sourceDir, destDir, hostname)
	if err != nil {
		t.Fatalf("MoveHost failed: %v", err)
	}

	// Verify in destination
	destInv, _ := LoadInventory(destDir)
	if _, ok := destInv.Hosts[hostname]; !ok {
		t.Error("host not found in destination")
	}

	dv, _ := GetHostVars(destDir, hostname)
	if dv["key"] != "val" {
		t.Error("host_vars did not migrate")
	}

	if _, ok := destInv.Groups["web"]; !ok {
		t.Error("group was not auto-created in destination")
	}

	// Verify removed from source
	sourceInv, _ := LoadInventory(sourceDir)
	if _, ok := sourceInv.Hosts[hostname]; ok {
		t.Error("host still exists in source")
	}
}

func TestGetEffectiveHostVarsPrecedence(t *testing.T) {
	// Setup inventory
	inv := &models.Inventory{
		Hosts: map[string]*models.Host{
			"srv1": {
				Name: "srv1",
				Vars: map[string]interface{}{
					"var_host":  "host_val",
					"var_clash": "host_wins",
				},
			},
		},
		Groups: map[string]*models.Group{
			"all": {
				Name: "all",
				Vars: map[string]interface{}{
					"var_all":   "all_val",
					"var_clash": "all_loses",
				},
			},
			"parent_group": {
				Name: "parent_group",
				Vars: map[string]interface{}{
					"var_parent": "parent_val",
					"var_clash":  "parent_loses",
				},
				Children: []string{"child_group"},
			},
			"child_group": {
				Name:  "child_group",
				Hosts: []string{"srv1"},
				Vars: map[string]interface{}{
					"var_child": "child_val",
					"var_clash": "child_wins_over_parent",
				},
			},
			"sibling_group_a": {
				Name:  "sibling_group_a",
				Hosts: []string{"srv1"},
				Vars: map[string]interface{}{
					"var_clash_sibling": "sibling_a_val",
				},
			},
			"sibling_group_b": {
				Name:  "sibling_group_b",
				Hosts: []string{"srv1"},
				Vars: map[string]interface{}{
					"var_clash_sibling": "sibling_b_val",
				},
			},
		},
	}

	// Verify inheritance order
	// Precedence order (least to most specific): all -> parent_group -> child_group -> sibling_group_a/b -> host
	// Since sibling_group_a and sibling_group_b are both direct groups of srv1 and have no ancestor relationship,
	// they should be sorted alphabetically: sibling_group_a comes before sibling_group_b.
	// So sibling_group_b should override sibling_group_a.
	// Therefore, var_clash_sibling should resolve to "sibling_b_val" deterministically.
	// Also child_group is a child of parent_group, so child_group wins over parent_group.
	// Host wins over all of them.

	for i := 0; i < 50; i++ {
		effective := GetEffectiveHostVars(inv, "srv1")

		if effective["var_all"] != "all_val" {
			t.Errorf("expected var_all = all_val, got %v", effective["var_all"])
		}
		if effective["var_parent"] != "parent_val" {
			t.Errorf("expected var_parent = parent_val, got %v", effective["var_parent"])
		}
		if effective["var_child"] != "child_val" {
			t.Errorf("expected var_child = child_val, got %v", effective["var_child"])
		}
		if effective["var_host"] != "host_val" {
			t.Errorf("expected var_host = host_val, got %v", effective["var_host"])
		}

		// Clash on host vs all others
		if effective["var_clash"] != "host_wins" {
			t.Errorf("expected var_clash = host_wins, got %v", effective["var_clash"])
		}

		// Sibling group alphabetical precedence (sibling_group_b overrides sibling_group_a)
		if effective["var_clash_sibling"] != "sibling_b_val" {
			t.Errorf("expected var_clash_sibling = sibling_b_val, got %v", effective["var_clash_sibling"])
		}
	}
}

func TestRemoveGroup(t *testing.T) {
	tmpDir := t.TempDir()
	if err := InitInventory(tmpDir); err != nil {
		t.Fatalf("InitInventory failed: %v", err)
	}

	parentName := "parent"
	childName := "child"

	// 1. Create parent and child groups
	parentGroup := &models.Group{Name: parentName, Children: []string{childName}}
	if err := SaveGroup(tmpDir, parentName, parentGroup); err != nil {
		t.Fatalf("failed to save parent group: %v", err)
	}

	childGroup := &models.Group{Name: childName}
	if err := SaveGroup(tmpDir, childName, childGroup); err != nil {
		t.Fatalf("failed to save child group: %v", err)
	}

	// 2. Verify nesting
	inv, err := LoadInventory(tmpDir)
	if err != nil {
		t.Fatalf("failed to load inventory: %v", err)
	}
	if _, ok := inv.Groups[childName]; !ok {
		t.Fatal("expected child group to exist")
	}
	if len(inv.Groups[parentName].Children) != 1 || inv.Groups[parentName].Children[0] != childName {
		t.Fatal("expected child to be nested in parent")
	}

	// 3. Remove child group
	if err := RemoveGroup(tmpDir, childName); err != nil {
		t.Fatalf("RemoveGroup failed: %v", err)
	}

	// 4. Verify references cleaned up
	inv2, err := LoadInventory(tmpDir)
	if err != nil {
		t.Fatalf("failed to load inventory after removal: %v", err)
	}

	if _, ok := inv2.Groups[childName]; ok {
		t.Error("expected child group to be removed from inventory")
	}

	// Parent group should no longer reference childName in Children
	parentChildren := inv2.Groups[parentName].Children
	for _, c := range parentChildren {
		if c == childName {
			t.Error("expected child name to be removed from parent group children list")
		}
	}
}
