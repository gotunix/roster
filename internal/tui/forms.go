package tui

import (
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"gotunix.net/roster/internal/models"
	"gotunix.net/roster/internal/store"
)

func (m *MainTUIModel) buildTUIForm(formType string) *FormModel {
	inv, err := store.LoadInventory(m.inventoryPath)
	var hostOptions []string
	var groupOptions []string

	if err == nil && inv != nil {
		for h := range inv.Hosts {
			hostOptions = append(hostOptions, h)
		}
		sort.Strings(hostOptions)

		for g := range inv.Groups {
			groupOptions = append(groupOptions, g)
		}
		sort.Strings(groupOptions)
	}

	if len(hostOptions) == 0 {
		hostOptions = []string{"(no hosts)"}
	}
	if len(groupOptions) == 0 {
		groupOptions = []string{"(no groups)"}
	}

	var form *FormModel
	switch formType {
	case "add_host":
		form = NewForm("ADD NEW HOST")
		form.AddTextBox("hostname", "Host Name", "", "e.g. webserver-03")
	case "remove_host":
		form = NewForm("REMOVE HOSTS")
		form.AddSearchableMultiSelector("hostname", "Select Hosts to Remove", hostOptions, "Type to filter, Space to select/deselect")
	case "move_host":
		form = NewForm("MOVE HOSTS TO ANOTHER INVENTORY")
		form.AddSearchableMultiSelector("hostname", "Select Hosts", hostOptions, "Type to filter, Space to select/deselect")
		form.AddTextBox("dest_dir", "Destination Directory", "", "Path to dest inventory dir")
	case "clone_host":
		form = NewForm("CLONE HOST")
		form.AddSearchableSelector("hostname", "Source Host", hostOptions, "Type to filter")
		form.AddTextBox("dest_name", "New Host Name", "", "e.g. webserver-03-backup")
	case "add_group":
		form = NewForm("ADD NEW GROUP")
		form.AddTextBox("groupname", "Group Name", "", "e.g. database_servers")
	case "remove_group":
		form = NewForm("REMOVE GROUPS")
		form.AddSearchableMultiSelector("groupname", "Select Groups to Remove", groupOptions, "Type to filter, Space to select/deselect")
	case "assign_host":
		form = NewForm("ASSIGN HOSTS TO GROUPS")
		form.AddSearchableMultiSelector("hostname", "Select Hosts", hostOptions, "Type to filter, Space to select/deselect")
		form.AddSearchableMultiSelector("groupname", "Select Groups", groupOptions, "Type to filter, Space to select/deselect")
	case "assign_group":
		form = NewForm("NEST GROUPS INSIDE PARENT GROUP")
		form.AddSearchableMultiSelector("child", "Select Child Groups", groupOptions, "Type to filter, Space to select/deselect")
		form.AddSearchableSelector("parent", "Select Parent Group", groupOptions, "Type to filter")
	case "clone_group":
		form = NewForm("CLONE GROUP")
		form.AddSearchableSelector("groupname", "Source Group", groupOptions, "Type to filter")
		form.AddTextBox("dest_name", "New Group Name", "", "e.g. database_servers_backup")
	}

	if form == nil {
		return nil
	}

	form.AddButton("Save", func(f *FormModel) tea.Cmd {
		var err error
		switch formType {
		case "add_host":
			name := f.GetString("hostname")
			if strings.TrimSpace(name) != "" {
				err = store.AddHostToMain(m.inventoryPath, name)
			}
		case "remove_host":
			names := f.GetMultiSelect("hostname")
			for _, name := range names {
				if name != "" && name != "(no hosts)" {
					if e := store.RemoveHost(m.inventoryPath, name); e != nil {
						err = e
					}
				}
			}
		case "move_host":
			hNames := f.GetMultiSelect("hostname")
			destDir := f.GetString("dest_dir")
			if len(hNames) > 0 && strings.TrimSpace(destDir) != "" {
				for _, hName := range hNames {
					if hName != "" && hName != "(no hosts)" {
						if e := store.MoveHost(m.inventoryPath, destDir, hName); e != nil {
							err = e
						}
					}
				}
			}
		case "add_group":
			name := strings.ReplaceAll(strings.TrimSpace(f.GetString("groupname")), "-", "_")
			if name != "" {
				err = store.SaveGroup(m.inventoryPath, name, &models.Group{Name: name})
			}
		case "remove_group":
			names := f.GetMultiSelect("groupname")
			for _, name := range names {
				if name != "" && name != "(no groups)" {
					if e := store.RemoveGroup(m.inventoryPath, name); e != nil {
						err = e
					}
				}
			}
		case "assign_host":
			hNames := f.GetMultiSelect("hostname")
			gNames := f.GetMultiSelect("groupname")
			for _, gName := range gNames {
				if gName != "" && gName != "(no groups)" {
					for _, hName := range hNames {
						if hName != "" && hName != "(no hosts)" {
							if e := store.AssignHostToGroup(m.inventoryPath, hName, gName); e != nil {
								err = e
							}
						}
					}
				}
			}
		case "assign_group":
			children := f.GetMultiSelect("child")
			parent := f.GetString("parent")
			if parent != "" && parent != "(no groups)" {
				for _, child := range children {
					if child != "" && child != "(no groups)" {
						if e := store.AssignGroupToGroup(m.inventoryPath, child, parent); e != nil {
							err = e
						}
					}
				}
			}
		case "clone_host":
			src := f.GetString("hostname")
			dest := f.GetString("dest_name")
			if src != "" && src != "(no hosts)" && strings.TrimSpace(dest) != "" {
				if e := store.CopyHost(m.inventoryPath, src, strings.TrimSpace(dest)); e != nil {
					err = e
				}
			}
		case "clone_group":
			src := f.GetString("groupname")
			dest := strings.ReplaceAll(strings.TrimSpace(f.GetString("dest_name")), "-", "_")
			if src != "" && src != "(no groups)" && dest != "" {
				if e := store.CopyGroup(m.inventoryPath, src, dest); e != nil {
					err = e
				}
			}
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

func (m *MainTUIModel) buildSyncForm() *FormModel {
	form := NewForm("SYNC NETBOX INVENTORY")

	defaultURL := "https://netbox.example.com"
	defaultToken := os.Getenv("NETBOX_TOKEN")
	defaultFilter := ""

	confURL, confToken, confFilter := store.LoadRosterConf()
	if confURL != "" {
		defaultURL = confURL
	}
	if confToken != "" {
		defaultToken = confToken
	}
	if confFilter != "" {
		defaultFilter = confFilter
	}

	form.AddTextBox("url", "NetBox API URL", "https://netbox.example.com", "e.g. https://netbox.example.com")
	form.AddTextBox("token", "API Token (Sensitive)", "", "NETBOX_TOKEN")
	form.AddTextBox("filter", "Filter Query Parameters (Optional)", "", "e.g. status=active")
	form.AddBoolean("sync_config_contexts", "Sync Config Contexts", "")
	form.AddBoolean("sync_hosts", "Sync Hosts", "")
	form.AddBoolean("sync_interfaces", "Sync Interfaces", "")
	form.AddBoolean("sync_ips", "Sync IP Addresses", "")
	form.AddBoolean("sync_vm_disks", "Sync VM Disks", "")

	form.SetValue("url", defaultURL)
	form.SetValue("token", defaultToken)
	form.SetValue("filter", defaultFilter)
	form.SetValue("sync_config_contexts", "True")
	form.SetValue("sync_hosts", "True")
	form.SetValue("sync_interfaces", "True")
	form.SetValue("sync_ips", "True")
	form.SetValue("sync_vm_disks", "True")

	form.AddButton("Sync", func(f *FormModel) tea.Cmd {
		f.Submitted = true
		return nil
	})

	form.AddButton("Cancel", func(f *FormModel) tea.Cmd {
		f.Quitting = true
		return nil
	})

	return form
}

func (m *MainTUIModel) buildExportForm() *FormModel {
	form := NewForm("EXPORT TO CSV")

	form.AddTextBox("output", "Output CSV File Path", "export.csv", "e.g. inventory.csv")
	form.AddTextBox("vars", "Variables as Columns (Comma separated)", "", "e.g. ansible_host,ansible_port")
	form.AddTextBox("group", "Filter by Group (Optional)", "", "e.g. webservers")
	form.AddTextBox("exclude", "Exclude Groups (Comma separated)", "", "e.g. testing,staging")
	form.AddTextBox("email", "Email Destination (Optional)", "", "e.g. admin@example.com")

	form.SetValue("output", "export.csv")

	form.AddButton("Export", func(f *FormModel) tea.Cmd {
		f.Submitted = true
		return nil
	})

	form.AddButton("Cancel", func(f *FormModel) tea.Cmd {
		f.Quitting = true
		return nil
	})

	return form
}
