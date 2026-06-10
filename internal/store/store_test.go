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
	InitInventory(tmpDir)

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
	InitInventory(tmpDir)

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
	InitInventory(tmpDir)

	hostname := "srv1"
	groupName := "web"
	AddHostToMain(tmpDir, hostname)
	SaveGroup(tmpDir, groupName, &models.Group{Name: groupName})

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
	InitInventory(sourceDir)
	InitInventory(destDir)

	hostname := "migrating-host"
	AddHostToMain(sourceDir, hostname)
	SetHostVar(sourceDir, hostname, "key", "val")

	// Create a group in source and assign host
	SaveGroup(sourceDir, "web", &models.Group{Name: "web", Hosts: []string{hostname}})
	SetGroupVar(sourceDir, "web", "port", "80")

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
