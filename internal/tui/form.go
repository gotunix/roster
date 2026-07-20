// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The Roster Authors
// =============================================================================================== //
//                                                                                                 //
//            /$$$$$$                                                                              //
//           /$$__  $$                                                                             //
//          | $$  \__/  /$$$$$$  /$$   /$$  /$$$$$$   /$$$$$$$  /$$$$$$                            //
//          |  $$$$$$  /$$__  $$| $$  | $$ /$$__  $$ /$$_____/ /$$__  $$                           //
//           \____  $$| $$  \ $$| $$  | $$| $$  \__/| $$      | $$$$$$$$                           //
//           /$$  \ $$| $$  | $$| $$  | $$| $$      | $$      | $$_____/                           //
//          |  $$$$$$/|  $$$$$$/|  $$$$$$/| $$      |  $$$$$$$|  $$$$$$$                           //
//           \______/  \______/  \______/ |__/       \_______/ \_______/                           //
//                                                                                                 //
//                                             /$$    /$$                    /$$   /$$             //
//                                            | $$   | $$                   | $$  | $$             //
//                                            | $$   | $$ /$$$$$$  /$$   /$$| $$ /$$$$$$           //
//                                            |  $$ / $$/|____  $$| $$  | $$| $$|_  $$_/           //
//                                             \  $$ $$/  /$$$$$$$| $$  | $$| $$  | $$             //
//                                              \  $$$/  /$$__  $$| $$  | $$| $$  | $$ /$$         //
//                                               \  $/  |  $$$$$$$|  $$$$$$/| $$  |  $$$$/         //
//                                                \_/    \_______/ \______/ |__/   \___/           //
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

package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FieldType identifies the type of widget rendered in the form.
type FieldType int

const (
	FieldText FieldType = iota
	FieldTextArea
	FieldSelector
	FieldBoolean
	FieldButton
	FieldMultiSelector
	FieldSearchableSelector
	FieldSearchableMultiSelector
)

// FormField defines a single input or action control in a generic form.
type FormField struct {
	Type  FieldType
	Name  string
	Label string
	Help  string

	// Input models
	TextInput textinput.Model
	TextArea  textarea.Model

	// State for selectors & checkbox
	Options         []string
	Selected        int
	SelectedIndices map[int]bool

	// Filter state for searchable selectors
	FilterText    string
	FilterInput   textinput.Model
	FilteredOpts  []string
	FilterIndices []int

	// Action for buttons
	ButtonBgColor lipgloss.Color
	ButtonFgColor lipgloss.Color
	Action        func(form *FormModel) tea.Cmd
}

// FormModel manages a generic, reusable form layout.
type FormModel struct {
	Title          string
	Fields         []*FormField
	FocusIndex     int
	Width          int
	Height         int
	terminalWidth  int
	terminalHeight int
	WidthPct       float64
	HeightPct      float64
	Quitting       bool
	Submitted      bool
	Styles         Styles
}

// NewForm creates an initialized FormModel with default global styles.
func NewForm(title string) *FormModel {
	return &FormModel{
		Title:     title,
		WidthPct:  0.95,
		HeightPct: 0.95,
		Styles:    GlobalStyles,
	}
}

// WithStyles sets a custom Styles configuration for the form, helping dynamic styling.
func (f *FormModel) WithStyles(s Styles) *FormModel {
	f.Styles = s
	return f
}

// AddTextBox adds a single-line text input field to the form.
func (f *FormModel) AddTextBox(name, label, placeholder, help string) {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = "  "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(GlobalTheme.Accent)
	ti.TextStyle = lipgloss.NewStyle().Foreground(GlobalTheme.Text)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(GlobalTheme.Primary)
	f.Fields = append(f.Fields, &FormField{
		Type:      FieldText,
		Name:      name,
		Label:     label,
		Help:      help,
		TextInput: ti,
	})
}

// AddTextArea adds a multi-line text input field to the form.
func (f *FormModel) AddTextArea(name, label, placeholder, help string) {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.ShowLineNumbers = false
	ta.SetHeight(3)
	ta.Prompt = "  "
	ta.FocusedStyle.Base = lipgloss.NewStyle().Foreground(GlobalTheme.Text)
	ta.Cursor.Style = lipgloss.NewStyle().Foreground(GlobalTheme.Primary)
	ta.BlurredStyle.Base = lipgloss.NewStyle().Foreground(GlobalTheme.Subtext)
	f.Fields = append(f.Fields, &FormField{
		Type:     FieldTextArea,
		Name:     name,
		Label:    label,
		Help:     help,
		TextArea: ta,
	})
}

// AddSelector adds a single-choice options list to the form.
func (f *FormModel) AddSelector(name, label string, options []string, help string) {
	f.Fields = append(f.Fields, &FormField{
		Type:    FieldSelector,
		Name:    name,
		Label:   label,
		Help:    help,
		Options: options,
	})
}

// AddMultiSelector adds a multi-choice options list to the form.
func (f *FormModel) AddMultiSelector(name, label string, options []string, help string) {
	f.Fields = append(f.Fields, &FormField{
		Type:            FieldMultiSelector,
		Name:            name,
		Label:           label,
		Help:            help,
		Options:         options,
		Selected:        0,
		SelectedIndices: make(map[int]bool),
	})
}

// AddSearchableSelector adds a single-select dropdown with real-time filtering.
func (f *FormModel) AddSearchableSelector(name, label string, options []string, help string) {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.Prompt = "/ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(GlobalTheme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(GlobalTheme.Text)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(GlobalTheme.Primary)

	filteredOpts := make([]string, len(options))
	copy(filteredOpts, options)
	filterIndices := make([]int, len(options))
	for i := range options {
		filterIndices[i] = i
	}

	f.Fields = append(f.Fields, &FormField{
		Type:          FieldSearchableSelector,
		Name:          name,
		Label:         label,
		Help:          help,
		Options:       options,
		FilteredOpts:  filteredOpts,
		FilterIndices: filterIndices,
		FilterInput:   ti,
	})
}

// AddSearchableMultiSelector adds a multi-select with real-time filtering.
func (f *FormModel) AddSearchableMultiSelector(name, label string, options []string, help string) {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.Prompt = "/ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(GlobalTheme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(GlobalTheme.Text)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(GlobalTheme.Primary)

	filteredOpts := make([]string, len(options))
	copy(filteredOpts, options)
	filterIndices := make([]int, len(options))
	for i := range options {
		filterIndices[i] = i
	}

	f.Fields = append(f.Fields, &FormField{
		Type:            FieldSearchableMultiSelector,
		Name:            name,
		Label:           label,
		Help:            help,
		Options:         options,
		FilteredOpts:    filteredOpts,
		FilterIndices:   filterIndices,
		SelectedIndices: make(map[int]bool),
		FilterInput:     ti,
	})
}

// AddBoolean adds a Yes/No option field to the form.
func (f *FormModel) AddBoolean(name, label, help string) {
	f.Fields = append(f.Fields, &FormField{
		Type:    FieldBoolean,
		Name:    name,
		Label:   label,
		Help:    help,
		Options: []string{"True", "False"},
	})
}

// AddButton adds an action button to the form.
func (f *FormModel) AddButton(label string, action func(form *FormModel) tea.Cmd) {
	f.Fields = append(f.Fields, &FormField{
		Type:          FieldButton,
		Name:          label,
		Label:         label,
		ButtonBgColor: GlobalTheme.Overlay,
		ButtonFgColor: GlobalTheme.Text,
		Action:        action,
	})
}

// GetString returns the value of an input field by its name.
func (f *FormModel) GetString(name string) string {
	for _, field := range f.Fields {
		if field.Name == name {
			switch field.Type {
			case FieldText:
				return field.TextInput.Value()
			case FieldTextArea:
				return field.TextArea.Value()
			case FieldSelector, FieldBoolean:
				if len(field.Options) > 0 && field.Selected >= 0 && field.Selected < len(field.Options) {
					return field.Options[field.Selected]
				}
			case FieldSearchableSelector, FieldSearchableMultiSelector:
				if len(field.FilteredOpts) > 0 && field.Selected >= 0 && field.Selected < len(field.FilteredOpts) && field.Selected < len(field.FilterIndices) {
					origIdx := field.FilterIndices[field.Selected]
					if origIdx >= 0 && origIdx < len(field.Options) {
						return field.Options[origIdx]
					}
				}
			}
		}
	}
	return ""
}

// SetValue sets the value of an input field by name.
func (f *FormModel) SetValue(name, value string) {
	for _, field := range f.Fields {
		if field.Name == name {
			switch field.Type {
			case FieldText:
				field.TextInput.SetValue(value)
			case FieldTextArea:
				field.TextArea.SetValue(value)
			}
			return
		}
	}
}

// GetMultiSelect returns the list of selected options for a multi-selector field.
func (f *FormModel) GetMultiSelect(name string) []string {
	for _, field := range f.Fields {
		if field.Name == name {
			if field.Type == FieldMultiSelector || field.Type == FieldSearchableMultiSelector {
				var selected []string
				for idx, val := range field.SelectedIndices {
					if val && idx >= 0 && idx < len(field.Options) {
						selected = append(selected, field.Options[idx])
					}
				}
				sort.Strings(selected)
				return selected
			}
		}
	}
	return nil
}

// GetBool returns the boolean value for a boolean field.
func (f *FormModel) GetBool(name string) bool {
	for _, field := range f.Fields {
		if field.Name == name {
			if field.Type == FieldBoolean {
				if field.Selected == 0 {
					return true // "True" option
				}
				return false // "False" option
			}
		}
	}
	return false
}

func (f *FormModel) updateFilter(field *FormField) {
	if field.FilterText == "" {
		field.FilteredOpts = make([]string, len(field.Options))
		copy(field.FilteredOpts, field.Options)
		field.FilterIndices = make([]int, len(field.Options))
		for i := range field.Options {
			field.FilterIndices[i] = i
		}
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
	if len(field.FilteredOpts) > 0 {
		field.Selected = 0
	}
}

func (f *FormModel) updateFocus() {
	for i, field := range f.Fields {
		if field.Type == FieldText {
			if i == f.FocusIndex {
				field.TextInput.Focus()
			} else {
				field.TextInput.Blur()
			}
		} else if field.Type == FieldTextArea {
			if i == f.FocusIndex {
				field.TextArea.Focus()
			} else {
				field.TextArea.Blur()
			}
		} else if field.Type == FieldSearchableSelector || field.Type == FieldSearchableMultiSelector {
			if i == f.FocusIndex {
				field.FilterInput.Focus()
			} else {
				field.FilterInput.Blur()
			}
		}
	}
}

// Init initializes the form focus stack.
func (f *FormModel) Init() tea.Cmd {
	f.updateFocus()
	return nil
}

// Update processes key events and size messages.
func (f *FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.terminalWidth = msg.Width
		f.terminalHeight = msg.Height
		f.Width = int(float64(msg.Width) * f.WidthPct)
		if f.Width < 60 {
			f.Width = 60
		}
		f.Height = int(float64(msg.Height) * f.HeightPct)

		for _, field := range f.Fields {
			if field.Type == FieldText {
				field.TextInput.Width = f.Width - 10
			}
			if field.Type == FieldTextArea {
				field.TextArea.SetWidth(f.Width - 10)
			}
			if field.Type == FieldSearchableSelector || field.Type == FieldSearchableMultiSelector {
				field.FilterInput.Width = f.Width - 14
			}
		}
		return f, nil

	case tea.KeyMsg:
		isSearchableField := func(field *FormField) bool {
			return field.Type == FieldSearchableSelector || field.Type == FieldSearchableMultiSelector
		}
		inFilter := func(field *FormField) bool {
			return isSearchableField(field) && field.FilterInput.Focused()
		}

		switch msg.String() {
		case "ctrl+c":
			f.Quitting = true
			return f, tea.Quit
		case "esc":
			if len(f.Fields) > 0 {
				field := f.Fields[f.FocusIndex]
				if inFilter(field) && field.FilterText != "" {
					field.FilterText = ""
					field.FilterInput.SetValue("")
					f.updateFilter(field)
					return f, nil
				}
				if isSearchableField(field) && !inFilter(field) {
					field.FilterInput.Focus()
					return f, nil
				}
			}
			f.Quitting = true
			return f, tea.Quit
		case "tab":
			if len(f.Fields) > 0 {
				f.FocusIndex = (f.FocusIndex + 1) % len(f.Fields)
				f.updateFocus()
			}
			return f, nil
		case "shift+tab":
			if len(f.Fields) > 0 {
				f.FocusIndex = (f.FocusIndex - 1 + len(f.Fields)) % len(f.Fields)
				f.updateFocus()
			}
			return f, nil
		case "enter":
			if len(f.Fields) > 0 {
				field := f.Fields[f.FocusIndex]
				if inFilter(field) {
					field.FilterInput.Blur()
					return f, nil
				}
				if field.Type == FieldButton {
					return f, field.Action(f)
				}
				if field.Type == FieldMultiSelector {
					field.SelectedIndices[field.Selected] = !field.SelectedIndices[field.Selected]
					return f, nil
				}
				if field.Type == FieldSearchableSelector && !inFilter(field) {
					f.FocusIndex = (f.FocusIndex + 1) % len(f.Fields)
					f.updateFocus()
					return f, nil
				}
				if field.Type == FieldSearchableMultiSelector {
					origIdx := field.FilterIndices[field.Selected]
					field.SelectedIndices[origIdx] = !field.SelectedIndices[origIdx]
					return f, nil
				}
			}
		case " ":
			if len(f.Fields) > 0 {
				field := f.Fields[f.FocusIndex]
				if !inFilter(field) && field.Type == FieldMultiSelector {
					field.SelectedIndices[field.Selected] = !field.SelectedIndices[field.Selected]
					return f, nil
				}
				if !inFilter(field) && field.Type == FieldSearchableMultiSelector {
					origIdx := field.FilterIndices[field.Selected]
					field.SelectedIndices[origIdx] = !field.SelectedIndices[origIdx]
					return f, nil
				}
			}
		case "up", "k":
			if len(f.Fields) > 0 {
				field := f.Fields[f.FocusIndex]
				if isSearchableField(field) && field.FilterInput.Focused() {
					// Let the key fall through to the text input handler
				} else if (field.Type == FieldMultiSelector || (field.Type == FieldSelector && len(field.Options) > 4)) && field.Selected > 0 {
					field.Selected--
					return f, nil
				} else if isSearchableField(field) && field.Selected > 0 {
					field.Selected--
					return f, nil
				}
			}
		case "down", "j":
			if len(f.Fields) > 0 {
				field := f.Fields[f.FocusIndex]
				if isSearchableField(field) && field.FilterInput.Focused() {
					// Let the key fall through to the text input handler
				} else if (field.Type == FieldMultiSelector || (field.Type == FieldSelector && len(field.Options) > 4)) && field.Selected < len(field.Options)-1 {
					field.Selected++
					return f, nil
				} else if isSearchableField(field) && field.Selected < len(field.FilteredOpts)-1 {
					field.Selected++
					return f, nil
				}
			}
		case "left", "h":
			if len(f.Fields) > 0 {
				field := f.Fields[f.FocusIndex]
				if (field.Type == FieldSelector || field.Type == FieldBoolean) && len(field.Options) <= 4 && field.Selected > 0 {
					field.Selected--
					return f, nil
				}
			}
		case "right", "l":
			if len(f.Fields) > 0 {
				field := f.Fields[f.FocusIndex]
				if (field.Type == FieldSelector || field.Type == FieldBoolean) && len(field.Options) <= 4 && field.Selected < len(field.Options)-1 {
					field.Selected++
					return f, nil
				}
			}
		}
	}

	if len(f.Fields) > 0 {
		field := f.Fields[f.FocusIndex]
		if field.Type == FieldText {
			var cmd tea.Cmd
			field.TextInput, cmd = field.TextInput.Update(msg)
			return f, cmd
		} else if field.Type == FieldTextArea {
			var cmd tea.Cmd
			field.TextArea, cmd = field.TextArea.Update(msg)
			return f, cmd
		} else if field.Type == FieldSearchableSelector || field.Type == FieldSearchableMultiSelector {
			if field.FilterInput.Focused() {
				var cmd tea.Cmd
				oldVal := field.FilterText
				field.FilterInput, cmd = field.FilterInput.Update(msg)
				field.FilterText = field.FilterInput.Value()
				if field.FilterText != oldVal {
					f.updateFilter(field)
				}
				return f, cmd
			}
		}
	}

	return f, nil
}

// View outputs the complete visual layout.
func (f *FormModel) View() string {
	if f.Quitting || f.Submitted {
		return ""
	}

	var sections []string
	headerFull := f.Styles.HeaderStyle.Copy().Width(f.Width).Align(lipgloss.Center).Render(" " + strings.ToUpper(f.Title) + " ")
	sections = append(sections, headerFull)

	contentWidth := f.Width - 4

	var buttonViews []string

	for i, field := range f.Fields {
		focused := i == f.FocusIndex

		if field.Type == FieldButton {
			var view string
			btnBg := field.ButtonBgColor
			btnFg := field.ButtonFgColor
			if focused {
				// Focused action button
				view = lipgloss.NewStyle().
					Foreground(GlobalTheme.Base).
					Background(GlobalTheme.Success).
					Padding(0, 2).
					Bold(true).
					Render("▶ " + field.Label + " ◀")
			} else {
				view = lipgloss.NewStyle().
					Foreground(btnFg).
					Background(btnBg).
					Padding(0, 3).
					Render(field.Label)
			}
			buttonViews = append(buttonViews, view)
			continue
		}

		var lbl string
		if focused {
			lbl = f.Styles.FocusedLabelStyle.Render("▶ " + field.Label)
		} else {
			lbl = f.Styles.LabelStyle.Render("  " + field.Label)
		}

		var view string
		switch field.Type {
		case FieldText:
			view = field.TextInput.View()
		case FieldTextArea:
			view = field.TextArea.View()
		case FieldSearchableSelector, FieldSearchableMultiSelector:
			statusStr := ""

			filterView := field.FilterInput.View()
			if len(field.Options) > 0 {
				matchCount := len(field.FilteredOpts)
				totalCount := len(field.Options)
				countStr := fmt.Sprintf(" (%d/%d)", matchCount, totalCount)
				filterView += lipgloss.NewStyle().Foreground(GlobalTheme.Overlay).Render(countStr)
			}
			statusStr = filterView + "\n"

			if field.FilterInput.Focused() {
				statusStr += "  " + lipgloss.NewStyle().Foreground(GlobalTheme.Overlay).Render("Enter to browse results") + "\n"
			}

			maxVisible := 8
			if field.Selected < 0 {
				field.Selected = 0
			}
			if len(field.FilteredOpts) > 0 && field.Selected >= len(field.FilteredOpts) {
				field.Selected = len(field.FilteredOpts) - 1
			}

			if len(field.FilteredOpts) == 0 {
				statusStr += "  " + lipgloss.NewStyle().Foreground(GlobalTheme.Overlay).Italic(true).Render("No matches") + "\n"
			} else {
				start := 0
				end := len(field.FilteredOpts)
				if len(field.FilteredOpts) > maxVisible {
					start = field.Selected - (maxVisible / 2)
					if start < 0 {
						start = 0
					}
					end = start + maxVisible
					if end > len(field.FilteredOpts) {
						end = len(field.FilteredOpts)
						start = end - maxVisible
					}
				}

				var lines []string
				if start > 0 {
					lines = append(lines, "  "+lipgloss.NewStyle().Foreground(GlobalTheme.Overlay).Render("▲ ... more above ..."))
				}

				isMulti := field.Type == FieldSearchableMultiSelector
				optionsAreFocused := focused && !field.FilterInput.Focused()

				for idx := start; idx < end; idx++ {
					opt := field.FilteredOpts[idx]
					origIdx := field.FilterIndices[idx]
					isCursor := idx == field.Selected

					var line string
					if isMulti {
						isChecked := field.SelectedIndices[origIdx]
						box := "[ ]"
						if isChecked {
							box = "[✔]"
						}
						itemText := fmt.Sprintf("%s %s", box, opt)
						if optionsAreFocused && isCursor {
							line = lipgloss.NewStyle().
								Foreground(GlobalTheme.Base).
								Background(GlobalTheme.Primary).
								Bold(true).
								Render("  " + itemText + "  ")
						} else {
							line = "  " + lipgloss.NewStyle().Foreground(GlobalTheme.Text).Render(itemText)
						}
					} else {
						if optionsAreFocused && isCursor {
							line = lipgloss.NewStyle().
								Foreground(GlobalTheme.Base).
								Background(GlobalTheme.Primary).
								Bold(true).
								Render("   ▶ " + opt + "   ")
						} else if isCursor {
							line = "   ● " + lipgloss.NewStyle().Foreground(GlobalTheme.Primary).Bold(true).Render(opt)
						} else {
							line = "     " + lipgloss.NewStyle().Foreground(GlobalTheme.Text).Render(opt)
						}
					}
					lines = append(lines, line)
				}

				if end < len(field.FilteredOpts) {
					lines = append(lines, "  "+lipgloss.NewStyle().Foreground(GlobalTheme.Overlay).Render("▼ ... more below ..."))
				}

				statusStr += strings.Join(lines, "\n")
			}
			view = statusStr
		case FieldSelector, FieldBoolean:
			statusStr := ""
			if len(field.Options) > 4 {
				maxVisible := 10

				if field.Selected < 0 {
					field.Selected = 0
				}
				if len(field.Options) > 0 && field.Selected >= len(field.Options) {
					field.Selected = len(field.Options) - 1
				}

				start := 0
				end := len(field.Options)
				if len(field.Options) > maxVisible {
					start = field.Selected - (maxVisible / 2)
					if start < 0 {
						start = 0
					}
					end = start + maxVisible
					if end > len(field.Options) {
						end = len(field.Options)
						start = end - maxVisible
					}
				}

				var lines []string
				if start > 0 {
					lines = append(lines, "  "+lipgloss.NewStyle().Foreground(GlobalTheme.Overlay).Render("▲ ... more above ..."))
				}

				for idx := start; idx < end; idx++ {
					opt := field.Options[idx]
					isCursor := idx == field.Selected

					var line string
					if isCursor {
						if focused {
							line = lipgloss.NewStyle().
								Foreground(GlobalTheme.Base).
								Background(GlobalTheme.Primary).
								Bold(true).
								Render("   ▶ " + opt + "   ")
						} else {
							line = "   ● " + lipgloss.NewStyle().Foreground(GlobalTheme.Primary).Bold(true).Render(opt)
						}
					} else {
						line = "     " + lipgloss.NewStyle().Foreground(GlobalTheme.Text).Render(opt)
					}
					lines = append(lines, line)
				}

				if end < len(field.Options) {
					lines = append(lines, "  "+lipgloss.NewStyle().Foreground(GlobalTheme.Overlay).Render("▼ ... more below ..."))
				}

				statusStr = strings.Join(lines, "\n")
			} else {
				for j, opt := range field.Options {
					prefix := "○"
					if j == field.Selected {
						prefix = "⊙"
					}
					if focused && j == field.Selected {
						statusStr += lipgloss.NewStyle().
							Foreground(GlobalTheme.Base).
							Background(GlobalTheme.Success).
							Bold(true).
							Render(fmt.Sprintf(" %s %s ", prefix, opt))
					} else {
						statusStr += fmt.Sprintf(" %s %s ", prefix, opt)
					}
				}
			}
			view = statusStr
		case FieldMultiSelector:
			statusStr := ""
			maxVisible := 10

			if field.Selected < 0 {
				field.Selected = 0
			}
			if len(field.Options) > 0 && field.Selected >= len(field.Options) {
				field.Selected = len(field.Options) - 1
			}

			start := 0
			end := len(field.Options)
			if len(field.Options) > maxVisible {
				start = field.Selected - (maxVisible / 2)
				if start < 0 {
					start = 0
				}
				end = start + maxVisible
				if end > len(field.Options) {
					end = len(field.Options)
					start = end - maxVisible
				}
			}

			var lines []string
			if start > 0 {
				lines = append(lines, "  "+lipgloss.NewStyle().Foreground(GlobalTheme.Overlay).Render("▲ ... more above ..."))
			}

			for idx := start; idx < end; idx++ {
				opt := field.Options[idx]
				isChecked := field.SelectedIndices[idx]
				box := "[ ]"
				if isChecked {
					box = "[✔]"
				}

				itemText := fmt.Sprintf("%s %s", box, opt)
				isCursor := idx == field.Selected

				var line string
				if focused && isCursor {
					line = lipgloss.NewStyle().
						Foreground(GlobalTheme.Base).
						Background(GlobalTheme.Primary).
						Bold(true).
						Render("  " + itemText + "  ")
				} else {
					line = "  " + lipgloss.NewStyle().Foreground(GlobalTheme.Text).Render(itemText)
				}
				lines = append(lines, line)
			}

			if end < len(field.Options) {
				lines = append(lines, "  "+lipgloss.NewStyle().Foreground(GlobalTheme.Overlay).Render("▼ ... more below ..."))
			}

			statusStr = strings.Join(lines, "\n")
			view = statusStr
		}

		if lbl != "" {
			sections = append(sections, lbl)
		}

		sections = append(sections, view)

		if field.Help != "" {
			sections = append(sections, f.Styles.HelperStyle.Render("    "+field.Help))
		}

		sections = append(sections, "") // Spacing between fields
	}

	if len(buttonViews) > 0 {
		buttonsStr := strings.Join(buttonViews, "   ")
		sections = append(sections, lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Background(GlobalTheme.Base).Render(buttonsStr))
	}

	sections = append(sections, "") // Spacing before footer

	helpText := " tab/shift+tab: move • enter/space: select • esc: quit "
	footer := f.Styles.HeaderStyle.Copy().Width(f.Width).Align(lipgloss.Center).Render(helpText)
	sections = append(sections, footer)

	formContent := strings.Join(sections, "\n")

	win := lipgloss.NewStyle().Background(GlobalTheme.Base).Width(f.Width).Render(formContent)
	return lipgloss.Place(f.terminalWidth, f.terminalHeight, lipgloss.Center, lipgloss.Top, win, lipgloss.WithWhitespaceBackground(GlobalTheme.Base))
}
