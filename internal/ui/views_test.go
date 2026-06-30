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
	"strings"
	"testing"

	"gotunix.net/roster/internal/models"
)

func TestRenderHostList(t *testing.T) {
	// 1. Empty inventory
	inv := &models.Inventory{Hosts: make(map[string]*models.Host)}
	out := RenderHostList(inv, "", false)
	if !strings.Contains(out, "No hosts found") {
		t.Errorf("expected empty message, got %q", out)
	}

	// 2. Populated inventory
	inv.Hosts["srv1"] = &models.Host{Name: "srv1"}
	inv.Hosts["srv2"] = &models.Host{Name: "srv2"}
	out = RenderHostList(inv, "", false)

	if !strings.Contains(out, "HOSTS") {
		t.Error("missing HOSTS header")
	}
	if !strings.Contains(out, "srv1") || !strings.Contains(out, "srv2") {
		t.Error("missing hostnames in output")
	}
}

func TestRenderGroupList(t *testing.T) {
	inv := &models.Inventory{Groups: make(map[string]*models.Group)}

	// Populated
	inv.Groups["web"] = &models.Group{Name: "web"}
	out := RenderGroupList(inv)

	if !strings.Contains(out, "GROUPS") {
		t.Error("missing GROUPS header")
	}
	if !strings.Contains(out, "web") {
		t.Error("missing groupname in output")
	}
}

func TestRenderDashboard(t *testing.T) {
	inv := &models.Inventory{
		Hosts: map[string]*models.Host{
			"srv1": {Name: "srv1", Vars: map[string]interface{}{"description": "Primary Web Server"}},
			"srv2": {Name: "srv2"},
		},
		Groups: map[string]*models.Group{
			"web": {
				Name:  "web",
				Hosts: []string{"srv1", "srv2"},
			},
		},
	}

	out := RenderDashboard(inv)

	if !strings.Contains(out, "ANSIBLE INVENTORY") {
		t.Error("missing dashboard header")
	}
	if !strings.Contains(out, "Primary Web Server") {
		t.Error("missing host description in dashboard")
	}
	if !strings.Contains(out, "srv1") || !strings.Contains(out, "srv2") {
		t.Error("missing hostnames in dashboard")
	}
}
