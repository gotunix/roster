# TUI Form Search/Filter Plan

## Problem
- `FieldSelector` and `FieldMultiSelector` show all options in a scrollable list
- Navigation with `j/k` or `↑/↓` is slow for 50+ items (hosts, groups)
- Need real-time filtering as user types

## Solution: Searchable Selector Field Types

### New Field Types
Add two new field types to `internal/tui/form.go`:

```go
const (
    // ... existing
    FieldSearchableSelector    FieldType = iota // single-select with filter
    FieldSearchableMultiSelector               // multi-select with filter
)
```

### FormField Struct Changes
Add filter-related fields to `FormField` (line 64-83):

```go
type FormField struct {
    // ... existing fields
    FilterText    string        // current filter text
    FilterInput   textinput.Model
    FilteredOpts  []string      // computed filtered options
    FilterIndices []int         // maps filtered index -> original index
}
```

### New Methods on FormModel

```go
// AddSearchableSelector adds a single-select dropdown with real-time filtering
func (f *FormModel) AddSearchableSelector(name, label string, options []string, help string)

// AddSearchableMultiSelector adds a multi-select with real-time filtering
func (f *FormModel) AddSearchableMultiSelector(name, label string, options []string, help string)
```

### Behavior

**Focus Flow:**
1. User tabs to selector field → focus goes to filter input first
2. User types → `FilteredOpts` updates in real-time
3. User presses `Enter` → focus moves to filtered option list
4. User presses `Esc` in filter input → clears filter, back to form fields

**Keybindings (when filter input focused):**
| Key | Action |
|-----|--------|
| Type | Update filter, recompute `FilteredOpts` |
| `Enter` | Move focus to filtered option list |
| `Esc` | Clear filter, return to form field navigation |
| `Ctrl+U` | Clear filter text |

**Keybindings (when option list focused):**
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate filtered options |
| `Enter` | Select (single) / Submit form |
| `Space` | Toggle (multi-select) |
| `Esc` | Back to filter input |

### View Rendering
- Show filter input at top of selector view
- Display "Showing X of Y" indicator
- Render filtered options with existing scroll logic
- Preserve existing styling

### Implementation Steps

1. **Add constants** for new field types
2. **Extend FormField** with filter fields
3. **Add constructor methods** `AddSearchableSelector` / `AddSearchableMultiSelector`
4. **Update `Update()`** to handle filter input focus and filtering logic
5. **Update `View()`** to render filter input + filtered list
6. **Update `GetString()` / `GetMultiSelect()`** to map filtered index → original option
7. **Add filter helper** function to recompute `FilteredOpts` and `FilterIndices`

### Filter Logic
```go
func (f *FormModel) updateFilter(field *FormField) {
    if field.FilterText == "" {
        field.FilteredOpts = field.Options
        field.FilterIndices = nil // 1:1 mapping
    } else {
        field.FilteredOpts = nil
        field.FilterIndices = nil
        lowerFilter := strings.ToLower(field.FilterText)
        for i, opt := range field.Options {
            if strings.Contains(strings.ToLower(opt), lowerFilter) {
                field.FilteredOpts = append(field.FilteredOpts, opt)
                field.FilterIndices = append(field.FilterIndices, i)
            }
        }
    }
    // Reset selection to first filtered item
    if len(field.FilteredOpts) > 0 {
        field.Selected = 0
    }
}
```

### Usage in menu.go (buildTUIForm)
Replace:
```go
form.AddMultiSelector("hostname", "Select Hosts", hostOptions, "...")
```
With:
```go
form.AddSearchableMultiSelector("hostname", "Select Hosts", hostOptions, "Type to filter...")
```

### Backward Compatibility
- Existing `FieldSelector` / `FieldMultiSelector` unchanged
- Old `AddSelector` / `AddMultiSelector` methods still work
- New methods are additive

### Testing Checklist
- [ ] Filter works with 500+ options
- [ ] Selection returns correct original value
- [ ] Multi-select toggles work with filtered list
- [ ] Esc clears filter and returns to form
- [ ] Tab navigation works: form → filter → options → next field
- [ ] Empty filter shows all options
- [ ] No matches shows "No matches" message

---

# TUI Host/Group Tree View Filter Plan

## Problem
- "View Hosts & Groups (Tree)" screen (StateHosts) shows all groups/hosts in a tree
- With 100+ hosts, scrolling is slow and finding specific hosts is difficult
- Need real-time filter to narrow down visible items

## Solution: Tree View Filter Mode

### New State Fields (MainTUIModel)
```go
type MainTUIModel struct {
    // ... existing
    treeFilterText   string        // current filter text
    treeFilterInput  textinput.Model
    treeFilterActive bool          // whether filter mode is active
    treeFilteredRows []HostTreeRow  // filtered visible rows
}
```

### Keybindings (in StateHosts)
| Key | Action |
|-----|--------|
| `/` | Activate filter mode, focus filter input |
| `Esc` | Clear filter, exit filter mode |
| `Enter` (in filter) | Apply filter, return to tree navigation |
| Type (in filter) | Update filter in real-time |
| `↑/↓` / `j/k` | Navigate filtered tree (when not in filter input) |
| `PgUp/PgDn` | Page through filtered results |

### Filter Logic
- Filter matches against: group name, host name, host description
- Case-insensitive substring match
- When filter active:
  - Show only matching groups AND their matching hosts
  - Always show parent groups of matching hosts (for context)
  - Or: flat list of matching hosts with group context

### View Rendering (StateHosts)
- When `treeFilterActive`: show filter input at top of tree view
- Display "Filter: xyz (X of Y matches)" indicator
- Render `treeFilteredRows` instead of all visible rows
- Preserve expand/collapse state for groups

### Implementation Steps
- [x] Add `treeFilterInput` to `NewMainTUIModel()` initialization
- [x] Add filter key handling in `Update()` for `StateHosts`
- [x] Add `applyTreeFilter()` method that builds `treeFilteredRows`
- [x] Update `View()` for `StateHosts` to show filter input + filtered rows
- [x] Ensure cursor navigation works within filtered results

### Filter Matching Rules
```go
func matchesFilter(row HostTreeRow, filter string) bool {
    if filter == "" {
        return true
    }
    lowerFilter := strings.ToLower(filter)
    // Match group name
    if row.IsGroup && strings.Contains(strings.ToLower(row.Name), lowerFilter) {
        return true
    }
    // Match host name
    if !row.IsGroup && strings.Contains(strings.ToLower(row.Name), lowerFilter) {
        return true
    }
    // Match host description (would need to look up host vars)
    return false
}
```

### Tree Filtering Strategy
**Option A: Flat filtered list** - Show only matching rows, lose hierarchy
**Option B: Context-preserving** - Show matching groups + their matching children + parent groups
**Option C: Expand matches** - Auto-expand groups containing matches

Recommend **Option B** for usability - preserves context while filtering.

### Integration with Existing Features
- Works with existing expand/collapse (`Enter` on group)
- Works with variable view (`v` on host/group)
- PgUp/PgDn navigate filtered results
- Filter persists until cleared with `Esc`

---

# Universal List Filter Component (Reusable)

## Goal
Create a single `ListFilter` component that can be embedded in ANY TUI list view:
- Host/Group Tree (StateHosts)
- Group Hierarchy (StateDashboard "groups" view)
- Host List view (StateHosts "hosts" view)
- Variable list views (StateVars)
- Export form selectors
- Any future list views

## Design: ListFilter Component

### Struct (internal/tui/listfilter.go)
```go
type ListFilter struct {
    Input        textinput.Model
    Active       bool
    FilterText   string
    MatchCount   int
    TotalCount   int
    OnFilter     func(string)      // callback to apply filter
    OnClear      func()            // callback when cleared
    Placeholder  string
    HelpText     string
}

func NewListFilter(placeholder, helpText string) *ListFilter {
    ti := textinput.New()
    ti.Placeholder = placeholder
    ti.Prompt = "/ "
    ti.PromptStyle = lipgloss.NewStyle().Foreground(GlobalTheme.Primary)
    ti.TextStyle = lipgloss.NewStyle().Foreground(GlobalTheme.Text)
    ti.Cursor.Style = lipgloss.NewStyle().Foreground(GlobalTheme.Primary)
    
    return &ListFilter{
        Input:       ti,
        Placeholder: placeholder,
        HelpText:    helpText,
    }
}

func (lf *ListFilter) Activate() {
    lf.Active = true
    lf.Input.Focus()
    lf.Input.SetValue("")
    lf.FilterText = ""
}

func (lf *ListFilter) Deactivate() {
    lf.Active = false
    lf.Input.Blur()
    lf.Input.SetValue("")
    lf.FilterText = ""
    if lf.OnClear != nil {
        lf.OnClear()
    }
}

func (lf *ListFilter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if !lf.Active {
        return lf, nil
    }
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "esc":
            lf.Deactivate()
            return lf, nil
        case "enter":
            lf.Active = false
            lf.Input.Blur()
            return lf, nil
        }
    }
    
    var cmd tea.Cmd
    lf.Input, cmd = lf.Input.Update(msg)
    lf.FilterText = lf.Input.Value()
    
    if lf.OnFilter != nil {
        lf.OnFilter(lf.FilterText)
    }
    
    return lf, cmd
}

func (lf *ListFilter) View() string {
    if !lf.Active {
        return ""
    }
    
    help := ""
    if lf.HelpText != "" {
        help = lf.HelpText
    }
    
    countStr := ""
    if lf.TotalCount > 0 {
        countStr = fmt.Sprintf(" (%d/%d)", lf.MatchCount, lf.TotalCount)
    }
    
    return lipgloss.JoinVertical(lipgloss.Left,
        lf.Input.View()+countStr,
        lipgloss.NewStyle().Foreground(GlobalTheme.Overlay).Render(help),
    )
}
```

### Usage in Any List View

```go
// In MainTUIModel for StateHosts:
func (m *MainTUIModel) initHostsFilter() {
    m.treeFilter = tui.NewListFilter(
        "Filter groups/hosts...",
        "Type to filter • Enter: apply • Esc: clear",
    )
    m.treeFilter.OnFilter = func(filter string) {
        m.applyTreeFilter(filter)
    }
    m.treeFilter.OnClear = func() {
        m.treeFilteredRows = nil // show all
    }
}
```

### Filter Matching Strategies (configurable per view)

```go
type FilterStrategy int

const (
    FilterSubstring FilterStrategy = iota  // contains (default)
    FilterPrefix                            // starts with
    FilterFuzzy                             // fuzzy match
    FilterRegex                             // regex
)

type FilterableItem interface {
    FilterText() string        // primary text to match
    FilterSecondary() string   // optional secondary (e.g., description)
    FilterTags() []string      // optional tags (e.g., group names for hosts)
}

func (lf *ListFilter) SetStrategy(s FilterStrategy) { ... }
```

### Views to Upgrade

| View | State | Items | Filter Fields |
|------|-------|-------|---------------|
| Host/Group Tree | StateHosts | HostTreeRow | name, description, groups |
| Group Hierarchy | StateDashboard (groups) | Group | name, children |
| Host List (compact) | StateHosts | Host | name, description, groups |
| Variable List | StateVars | key:value | key, value |
| Export Selectors | StateExportMenu | hosts/groups | name |

### Keybindings (Standardized)

| Key | Context | Action |
|-----|---------|--------|
| `/` | List focused | Activate filter |
| `Esc` | Filter active | Clear & deactivate |
| `Enter` | Filter active | Apply & return to list |
| `↑/↓`, `j/k` | List focused | Navigate (filtered) |
| `PgUp/PgDn` | List focused | Page (filtered) |

### Implementation Plan

1. **Create `internal/tui/listfilter.go`** with `ListFilter` component
2. **Add to `MainTUIModel`**: one `ListFilter` per list view needing it
3. **Implement `applyXxxFilter(filter string)`** methods for each view
4. **Wire keybindings** in each state's Update handler
5. **Render filter bar** in each view's View method
6. **Test with 500+ items** for performance

### Benefits
- **Single implementation** - consistent behavior everywhere
- **Reusable** - drop into any list view
- **Configurable matching** - substring/prefix/fuzzy/regex per view
- **Keyboard-driven** - standard `/` to search, `Esc` to clear
- **Performance** - filters in-memory, no re-render of full list