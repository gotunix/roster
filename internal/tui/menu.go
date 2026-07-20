package tui

import (
	"bytes"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"gotunix.net/roster/internal/models"
)

type SafeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *SafeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *SafeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

type MenuState int

const (
	StateMenu MenuState = iota
	StateDashboard
	StateHosts
	StateVars
	StateVarsForm
	StateManageHostsMenu
	StateManageGroupsMenu
	StateTUIForm
	StateSyncMenu
	StateSyncing
	StateSyncResult
	StateExportMenu
	StateExportResult
	StateVersion
)

type HostTreeRow struct {
	Name      string
	IsGroup   bool
	GroupName string
	IsLast    bool
}

type SyncFinishedMsg struct {
	err error
}

type ExportFinishedMsg struct {
	err      error
	numHosts int
	outFile  string
}

type ExternalEditorFinishedMsg struct {
	err error
}

type MainTUIModel struct {
	state             MenuState
	menuCursor        int
	menuItems         []string
	versionModel      VersionWindowModel
	groupExpanded     map[string]bool
	treeCursor        int
	terminalWidth     int
	terminalHeight    int
	inventoryPath     string
	varsTargetName    string
	varsTargetIsGroup bool
	varsCursor        int
	varsKeys          []string
	varsValues        map[string]interface{}
	varsNested        map[string]interface{}
	inheritedGroups      []string
	inheritedByGroup     map[string]map[string]interface{}
	inheritedKeysByGroup map[string][]string
	varsForm          *FormModel
	varsFormType      string
	varsFormKey       string
	subMenuCursor     int
	subMenuItems      []string
	tuiForm           *FormModel
	tuiFormAction     string
	syncSpinner       SpinnerModel
	syncError         error
	syncLogBuffer     *SafeBuffer
	exportSpinner     SpinnerModel
	exportError       error
	exportNumHosts    int
	exportOutFile     string

	cachedInventory   *models.Inventory
	cachedRows        []HostTreeRow
	rowsDirty         bool
	inventoryModTime  string

	treeFilterInput   textinput.Model
	treeFilterText    string
	treeFilterActive  bool
	treeFilteredRows  []HostTreeRow
}

func NewMainTUIModel(appName, version, commit, date string, inventoryPath string) MainTUIModel {
	m := MainTUIModel{
		state:         StateMenu,
		menuItems: []string{
			"1. View Dashboard",
			"2. View Hosts & Groups (Tree)",
			"3. Manage Hosts",
			"4. Manage Groups",
			"5. Sync NetBox",
			"6. Export Inventory",
			"7. View Version Details",
			"8. Exit",
		},
		versionModel:      NewVersionWindow(appName, version, commit, date),
		groupExpanded:     make(map[string]bool),
		inventoryPath:     inventoryPath,
		syncSpinner:       NewSpinner("Syncing from NetBox..."),
		exportSpinner:     NewSpinner("Exporting to CSV..."),
		rowsDirty:         true,
		inventoryModTime:  "",
		treeFilterInput:   textinput.New(),
	}
	m.treeFilterInput.Placeholder = "Filter groups/hosts..."
	m.treeFilterInput.Prompt = "/ "
	m.treeFilterInput.PromptStyle = lipgloss.NewStyle().Foreground(GlobalTheme.Primary)
	m.treeFilterInput.TextStyle = lipgloss.NewStyle().Foreground(GlobalTheme.Text)
	m.treeFilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(GlobalTheme.Primary)
	return m
}

func (m MainTUIModel) Init() tea.Cmd {
	return nil
}
