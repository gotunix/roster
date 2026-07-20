package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"gotunix.net/roster/internal/store"
)

func (m MainTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ExternalEditorFinishedMsg:
		m.loadVars()
		return m, nil

	case SyncFinishedMsg:
		m.state = StateSyncResult
		m.syncError = msg.err
		return m, nil

	case ExportFinishedMsg:
		m.state = StateExportResult
		m.exportError = msg.err
		m.exportNumHosts = msg.numHosts
		m.exportOutFile = msg.outFile
		return m, nil

	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height

		vModel, cmd := m.versionModel.Update(msg)
		m.versionModel = vModel.(VersionWindowModel)
		cmds = append(cmds, cmd)

		if m.varsForm != nil {
			fModel, cmd := m.varsForm.Update(msg)
			m.varsForm = fModel.(*FormModel)
			cmds = append(cmds, cmd)
		}

		if m.tuiForm != nil {
			fModel, cmd := m.tuiForm.Update(msg)
			m.tuiForm = fModel.(*FormModel)
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		if m.state == StateMenu {
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit
			case "up", "k":
				if m.menuCursor > 0 {
					m.menuCursor--
				}
			case "down", "j":
				if m.menuCursor < len(m.menuItems)-1 {
					m.menuCursor++
				}
			case "enter":
				switch m.menuCursor {
				case 0:
					m.state = StateDashboard
				case 1:
					m.state = StateHosts
					m.treeCursor = 0
				case 2:
					m.state = StateManageHostsMenu
					m.subMenuCursor = 0
					m.subMenuItems = []string{"1. Add Host", "2. Remove Host", "3. Move Host", "4. Clone Host", "5. Back"}
				case 3:
					m.state = StateManageGroupsMenu
					m.subMenuCursor = 0
					m.subMenuItems = []string{"1. Add Group", "2. Remove Group", "3. Assign Host to Group", "4. Assign Group to Group", "5. Clone Group", "6. Back"}
				case 4:
					m.state = StateSyncMenu
					m.tuiForm = m.buildSyncForm()
					m.tuiForm.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
					return m, m.tuiForm.Init()
				case 5:
					m.state = StateExportMenu
					m.tuiForm = m.buildExportForm()
					m.tuiForm.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
					return m, m.tuiForm.Init()
				case 6:
					m.state = StateVersion
					vModel, cmd := m.versionModel.Update(tea.WindowSizeMsg{
						Width:  m.terminalWidth,
						Height: m.terminalHeight,
					})
					m.versionModel = vModel.(VersionWindowModel)
					return m, cmd
				case 7:
					return m, tea.Quit
				}
			}
			return m, nil
		} else if m.state == StateDashboard {
			switch msg.String() {
			case "esc", "q":
				m.state = StateMenu
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}
		} else if m.state == StateHosts {
			if m.treeFilterActive {
				switch msg.String() {
				case "esc":
					m.treeFilterActive = false
					m.treeFilterText = ""
					m.treeFilterInput.SetValue("")
					m.treeFilterInput.Blur()
					m.treeFilteredRows = nil
					return m, nil
				case "enter":
					m.treeFilterActive = false
					m.treeFilterInput.Blur()
					return m, nil
				case "ctrl+c":
					return m, tea.Quit
				default:
					var cmd tea.Cmd
					m.treeFilterInput, cmd = m.treeFilterInput.Update(msg)
					m.treeFilterText = m.treeFilterInput.Value()
					m.applyTreeFilter()
					return m, cmd
				}
			}

			switch msg.String() {
			case "esc", "q":
				m.state = StateMenu
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			case "/":
				m.treeFilterActive = true
				m.treeFilterInput.Focus()
				m.treeFilterInput.SetValue("")
				m.treeFilterText = ""
				m.treeFilteredRows = nil
				return m, nil
			case "up", "k":
				if m.treeCursor > 0 {
					m.treeCursor--
				}
				return m, nil
			case "down", "j":
				visibleRows := m.getVisibleRows()
				if m.treeCursor < len(visibleRows)-1 {
					m.treeCursor++
				}
				return m, nil
			case "pgup", "pageup":
				pageSize := m.terminalHeight - 6
				if pageSize < 1 {
					pageSize = 10
				}
				m.treeCursor -= pageSize
				if m.treeCursor < 0 {
					m.treeCursor = 0
				}
				return m, nil
			case "pgdn", "pagedown":
				pageSize := m.terminalHeight - 6
				if pageSize < 1 {
					pageSize = 10
				}
				visibleRows := m.getVisibleRows()
				m.treeCursor += pageSize
				if m.treeCursor >= len(visibleRows) {
					m.treeCursor = len(visibleRows) - 1
				}
				if m.treeCursor < 0 {
					m.treeCursor = 0
				}
				return m, nil
			case "v", "enter":
				visibleRows := m.getVisibleRows()
				if len(visibleRows) > 0 && m.treeCursor < len(visibleRows) {
					row := visibleRows[m.treeCursor]
					if !row.IsGroup && msg.String() == "enter" {
						m.state = StateVars
						m.varsTargetName = row.Name
						m.varsTargetIsGroup = false
						m.varsCursor = 0
						m.loadVars()
						return m, nil
					} else if msg.String() == "v" {
						m.state = StateVars
						m.varsTargetName = row.Name
						m.varsTargetIsGroup = row.IsGroup
						m.varsCursor = 0
						m.loadVars()
						return m, nil
					} else if row.IsGroup && msg.String() == "enter" {
						m.groupExpanded[row.GroupName] = !m.groupExpanded[row.GroupName]
						m.rowsDirty = true
						if m.treeFilterText != "" {
							m.applyTreeFilter()
						}
						return m, nil
					}
				}
				return m, nil
			}
		} else if m.state == StateVars {
			switch msg.String() {
			case "esc", "q":
				if m.treeFilterText != "" {
					m.treeFilterText = ""
					m.treeFilterInput.SetValue("")
					m.treeFilteredRows = nil
					return m, nil
				}
				m.state = StateMenu
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			case "up", "k":
				if m.varsCursor <= 0 {
					return m, nil
				}
				items := m.buildVarsItems()
				m.varsCursor--
				for m.varsCursor > 0 && items[m.varsCursor].cursorSlots == 0 {
					m.varsCursor--
				}
				return m, nil
			case "down", "j":
				items := m.buildVarsItems()
				maxItem := len(items) - 1
				if m.varsCursor >= maxItem {
					return m, nil
				}
				m.varsCursor++
				for m.varsCursor < maxItem && items[m.varsCursor].cursorSlots == 0 {
					m.varsCursor++
				}
				return m, nil
			case "enter":
				var path string
				if m.varsTargetIsGroup {
					path = store.GetGroupVarsPath(m.inventoryPath, m.varsTargetName)
				} else {
					path = store.GetHostVarsPath(m.inventoryPath, m.varsTargetName)
				}
				_ = os.MkdirAll(filepath.Dir(path), 0755)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					_ = os.WriteFile(path, []byte("{}\n"), 0644)
				}
				editor := os.Getenv("EDITOR")
				if editor == "" {
					editor = "vi"
				}
				c := exec.Command(editor, path)
				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					return ExternalEditorFinishedMsg{err: err}
				})
			case "d", "delete", "backspace":
				if m.varsCursor > 0 {
					items := m.buildVarsItems()
					if m.varsCursor < len(items) && items[m.varsCursor].key != "" {
						key := items[m.varsCursor].key
						var err error
						if m.varsTargetIsGroup {
							err = store.DeleteGroupVar(m.inventoryPath, m.varsTargetName, key)
						} else {
							err = store.DeleteHostVar(m.inventoryPath, m.varsTargetName, key)
						}
						if err == nil {
							m.loadVars()
							items2 := m.buildVarsItems()
							if m.varsCursor >= len(items2) {
								m.varsCursor = len(items2) - 1
							}
						}
					}
				}
				return m, nil
			case "e":
				var path string
				if m.varsTargetIsGroup {
					path = store.GetGroupVarsPath(m.inventoryPath, m.varsTargetName)
				} else {
					path = store.GetHostVarsPath(m.inventoryPath, m.varsTargetName)
				}
				_ = os.MkdirAll(filepath.Dir(path), 0755)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					_ = os.WriteFile(path, []byte("{}"), 0644)
				}
				editor := os.Getenv("EDITOR")
				if editor == "" {
					editor = "vi"
				}
				c := exec.Command(editor, path)
				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					return ExternalEditorFinishedMsg{err: err}
				})
			}
		} else if m.state == StateVarsForm {
			var cmd tea.Cmd
			mModel, cmd := m.varsForm.Update(msg)
			m.varsForm = mModel.(*FormModel)

			if m.varsForm.Quitting {
				m.state = StateVars
				return m, nil
			}
			if m.varsForm.Submitted {
				m.state = StateVars
				m.loadVars()
				return m, nil
			}
			return m, cmd
		} else if m.state == StateManageHostsMenu {
			switch msg.String() {
			case "esc", "q":
				m.state = StateMenu
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			case "up", "k":
				if m.subMenuCursor > 0 {
					m.subMenuCursor--
				}
				return m, nil
			case "down", "j":
				if m.subMenuCursor < len(m.subMenuItems)-1 {
					m.subMenuCursor++
				}
				return m, nil
			case "enter":
				switch m.subMenuCursor {
				case 0:
					m.state = StateTUIForm
					m.tuiFormAction = "add_host"
					m.tuiForm = m.buildTUIForm("add_host")
					m.tuiForm.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
					return m, m.tuiForm.Init()
				case 1:
					m.state = StateTUIForm
					m.tuiFormAction = "remove_host"
					m.tuiForm = m.buildTUIForm("remove_host")
					m.tuiForm.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
					return m, m.tuiForm.Init()
				case 2:
					m.state = StateTUIForm
					m.tuiFormAction = "move_host"
					m.tuiForm = m.buildTUIForm("move_host")
					m.tuiForm.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
					return m, m.tuiForm.Init()
				case 3:
					m.state = StateTUIForm
					m.tuiFormAction = "clone_host"
					m.tuiForm = m.buildTUIForm("clone_host")
					m.tuiForm.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
					return m, m.tuiForm.Init()
				case 4:
					m.state = StateMenu
					return m, nil
				}
			}
			return m, nil
		} else if m.state == StateManageGroupsMenu {
			switch msg.String() {
			case "esc", "q":
				m.state = StateMenu
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			case "up", "k":
				if m.subMenuCursor > 0 {
					m.subMenuCursor--
				}
				return m, nil
			case "down", "j":
				if m.subMenuCursor < len(m.subMenuItems)-1 {
					m.subMenuCursor++
				}
				return m, nil
			case "enter":
				switch m.subMenuCursor {
				case 0:
					m.state = StateTUIForm
					m.tuiFormAction = "add_group"
					m.tuiForm = m.buildTUIForm("add_group")
					m.tuiForm.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
					return m, m.tuiForm.Init()
				case 1:
					m.state = StateTUIForm
					m.tuiFormAction = "remove_group"
					m.tuiForm = m.buildTUIForm("remove_group")
					m.tuiForm.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
					return m, m.tuiForm.Init()
				case 2:
					m.state = StateTUIForm
					m.tuiFormAction = "assign_host"
					m.tuiForm = m.buildTUIForm("assign_host")
					m.tuiForm.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
					return m, m.tuiForm.Init()
				case 3:
					m.state = StateTUIForm
					m.tuiFormAction = "assign_group"
					m.tuiForm = m.buildTUIForm("assign_group")
					m.tuiForm.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
					return m, m.tuiForm.Init()
				case 4:
					m.state = StateTUIForm
					m.tuiFormAction = "clone_group"
					m.tuiForm = m.buildTUIForm("clone_group")
					m.tuiForm.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
					return m, m.tuiForm.Init()
				case 5:
					m.state = StateMenu
					return m, nil
				}
			}
			return m, nil
		} else if m.state == StateTUIForm {
			var cmd tea.Cmd
			mModel, cmd := m.tuiForm.Update(msg)
			m.tuiForm = mModel.(*FormModel)

			if m.tuiForm.Quitting {
				if m.tuiFormAction == "add_host" || m.tuiFormAction == "remove_host" || m.tuiFormAction == "move_host" || m.tuiFormAction == "clone_host" {
					m.state = StateManageHostsMenu
				} else {
					m.state = StateManageGroupsMenu
				}
				return m, nil
			}
			if m.tuiForm.Submitted {
				m.cachedInventory = nil
				if m.tuiFormAction == "add_host" || m.tuiFormAction == "remove_host" || m.tuiFormAction == "move_host" || m.tuiFormAction == "clone_host" {
					m.state = StateManageHostsMenu
				} else {
					m.state = StateManageGroupsMenu
				}
				return m, nil
			}
			return m, cmd
		} else if m.state == StateSyncMenu {
			var cmd tea.Cmd
			mModel, cmd := m.tuiForm.Update(msg)
			m.tuiForm = mModel.(*FormModel)

			if m.tuiForm.Quitting {
				m.state = StateMenu
				return m, nil
			}
			if m.tuiForm.Submitted {
				url := m.tuiForm.GetString("url")
				token := m.tuiForm.GetString("token")
				filter := m.tuiForm.GetString("filter")
				syncConfigContexts := m.tuiForm.GetBool("sync_config_contexts")
				syncHosts := m.tuiForm.GetBool("sync_hosts")
				syncInterfaces := m.tuiForm.GetBool("sync_interfaces")
				syncIPs := m.tuiForm.GetBool("sync_ips")
				syncVMDisks := m.tuiForm.GetBool("sync_vm_disks")
				m.state = StateSyncing
				m.syncError = nil
				m.syncLogBuffer = &SafeBuffer{}
				return m, tea.Batch(
					m.syncSpinner.Init(),
					m.runNetboxSync(url, token, filter, syncConfigContexts, syncHosts, syncInterfaces, syncIPs, syncVMDisks),
				)
			}
			return m, cmd
		} else if m.state == StateSyncing {
			var cmd tea.Cmd
			m.syncSpinner, cmd = m.syncSpinner.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		} else if m.state == StateSyncResult || m.state == StateExportResult {
			switch msg.String() {
			case "esc", "q":
				m.state = StateMenu
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}
		} else if m.state == StateExportMenu {
			var cmd tea.Cmd
			mModel, cmd := m.tuiForm.Update(msg)
			m.tuiForm = mModel.(*FormModel)

			if m.tuiForm.Quitting {
				m.state = StateMenu
				return m, nil
			}
			if m.tuiForm.Submitted {
				output := m.tuiForm.GetString("output")
				emailAddr := m.tuiForm.GetString("email")
				varsStr := m.tuiForm.GetString("vars")
				excludeStr := m.tuiForm.GetString("exclude")
				groupFilter := strings.TrimSpace(m.tuiForm.GetString("group"))

				var vars []string
				for _, v := range strings.Split(varsStr, ",") {
					if strings.TrimSpace(v) != "" {
						vars = append(vars, strings.TrimSpace(v))
					}
				}

				var exclude []string
				for _, e := range strings.Split(excludeStr, ",") {
					if strings.TrimSpace(e) != "" {
						exclude = append(exclude, strings.TrimSpace(e))
					}
				}

				m.state = StateExportResult
				m.exportError = nil
				return m, m.runExport(output, emailAddr, vars, exclude, groupFilter)
			}
			return m, cmd
		} else if m.state == StateVersion {
			switch msg.String() {
			case "esc", "q":
				m.state = StateMenu
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}

			vModel, cmd := m.versionModel.Update(msg)
			m.versionModel = vModel.(VersionWindowModel)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}
	}

	if m.state == StateVersion {
		vModel, cmd := m.versionModel.Update(msg)
		m.versionModel = vModel.(VersionWindowModel)
		cmds = append(cmds, cmd)
	}

	if m.state == StateSyncing {
		var cmd tea.Cmd
		m.syncSpinner, cmd = m.syncSpinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
