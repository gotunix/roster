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

package interactive

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
	"gotunix.net/roster/internal/store"
	"gotunix.net/roster/internal/ui"
)

func getCustomKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	// Set Enter to NewLine
	km.Text.NewLine = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "new line"),
	)
	// Set Ctrl+Enter to Submit
	km.Text.Submit = key.NewBinding(
		key.WithKeys("ctrl+enter", "ctrl+j", "ctrl+s", "ctrl+d"),
		key.WithHelp("ctrl+enter/ctrl+d", "submit"),
	)
	return km
}

func EditHostInteractive(baseDir, hostname string) error {
	inv, err := store.LoadInventory(baseDir)
	if err != nil {
		return err
	}
	if _, ok := inv.Hosts[hostname]; !ok {
		return fmt.Errorf("host %q not found in inventory", hostname)
	}

	vars, err := store.GetHostVars(baseDir, hostname)
	if err != nil {
		return err
	}

	// Convert vars to YAML string for editing
	bytes, err := yaml.Marshal(vars)
	if err != nil {
		return err
	}
	varsStr := string(bytes)
	if varsStr == "{}\n" {
		varsStr = ""
	}

	theme := huh.ThemeCharm()
	theme.Focused.Base = theme.Focused.Base.BorderForeground(ui.Magenta)
	theme.Group.Base = theme.Group.Base.Border(lipgloss.NormalBorder()).BorderForeground(ui.Gray)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Host Variables (YAML)").
				Value(&varsStr).
				EditorExtension(".yaml").
				Lines(15),
		).Title("EDIT HOST: " + hostname),
	).WithTheme(theme).WithKeyMap(getCustomKeyMap())

	if err := form.Run(); err != nil {
		return err
	}

	// Parse back into map
	newVars := make(map[string]interface{})
	if strings.TrimSpace(varsStr) != "" {
		if err := yaml.Unmarshal([]byte(varsStr), &newVars); err != nil {
			return fmt.Errorf("failed to parse YAML: %v", err)
		}
	}

	// Save
	return store.SaveHostVars(baseDir, hostname, newVars)
}

func EditGroupInteractive(baseDir, groupname string) error {
	inv, err := store.LoadInventory(baseDir)
	if err != nil {
		return err
	}
	if _, ok := inv.Groups[groupname]; !ok {
		return fmt.Errorf("group %q not found in inventory", groupname)
	}

	vars, err := store.GetGroupVars(baseDir, groupname)
	if err != nil {
		return err
	}

	// Convert vars to YAML string for editing
	bytes, err := yaml.Marshal(vars)
	if err != nil {
		return err
	}
	varsStr := string(bytes)
	if varsStr == "{}\n" {
		varsStr = ""
	}

	theme := huh.ThemeCharm()
	theme.Focused.Base = theme.Focused.Base.BorderForeground(ui.Magenta)
	theme.Group.Base = theme.Group.Base.Border(lipgloss.NormalBorder()).BorderForeground(ui.Gray)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Group Variables (YAML)").
				Value(&varsStr).
				EditorExtension(".yaml").
				Lines(15),
		).Title("EDIT GROUP: " + groupname),
	).WithTheme(theme).WithKeyMap(getCustomKeyMap())

	if err := form.Run(); err != nil {
		return err
	}

	// Parse back into map
	newVars := make(map[string]interface{})
	if strings.TrimSpace(varsStr) != "" {
		if err := yaml.Unmarshal([]byte(varsStr), &newVars); err != nil {
			return fmt.Errorf("failed to parse YAML: %v", err)
		}
	}

	// Save
	return store.SaveGroupVars(baseDir, groupname, newVars)
}
