package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lazysql/lazysql/pkg/tui/styles"
)

type FormSubmitMsg struct {
	Values map[string]string
}

type FormField struct {
	Label       string
	Key         string
	Placeholder string
	Value       string
	Hidden      bool
	EchoMode    textinput.EchoMode
}

type FormModel struct {
	fields  []formEntry
	focused int
	driver  string // current driver selection for dynamic fields
}

type formEntry struct {
	label string
	key   string
	input textinput.Model
	hide  func(driver string) bool
}

func NewForm(fields []FormField) FormModel {
	entries := make([]formEntry, len(fields))
	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.Placeholder
		ti.CharLimit = 256
		if f.Value != "" {
			ti.SetValue(f.Value)
		}
		if f.EchoMode != 0 {
			ti.EchoMode = f.EchoMode
		}
		if i == 0 {
			ti.Focus()
		}

		var hideFn func(string) bool
		if f.Hidden {
			hideFn = func(_ string) bool { return true }
		}

		entries[i] = formEntry{
			label: f.Label,
			key:   f.Key,
			input: ti,
			hide:  hideFn,
		}
	}

	m := FormModel{fields: entries}
	// Set dynamic hide rules for connection form
	for i := range m.fields {
		switch m.fields[i].key {
		case "host", "port", "user", "password", "database":
			m.fields[i].hide = func(driver string) bool { return driver == "sqlite" }
		case "path":
			m.fields[i].hide = func(driver string) bool { return driver != "sqlite" }
		}
	}

	// Detect initial driver
	for _, e := range entries {
		if e.key == "driver" {
			m.driver = e.input.Value()
			break
		}
	}
	if m.driver == "" {
		m.driver = "postgres"
	}

	return m
}

func (f FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (f FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			return f.focusNext(), textinput.Blink
		case "shift+tab", "up":
			return f.focusPrev(), textinput.Blink
		case "enter":
			// If on last visible field, submit
			visible := f.visibleFields()
			if len(visible) > 0 && f.focused == visible[len(visible)-1] {
				return f, f.submit()
			}
			return f.focusNext(), textinput.Blink
		case "ctrl+s":
			return f, f.submit()
		}
	}

	// Update focused field
	if f.focused < len(f.fields) {
		var cmd tea.Cmd
		f.fields[f.focused].input, cmd = f.fields[f.focused].input.Update(msg)
		// Track driver changes
		if f.fields[f.focused].key == "driver" {
			f.driver = f.fields[f.focused].input.Value()
		}
		return f, cmd
	}
	return f, nil
}

func (f FormModel) View() string {
	var b strings.Builder
	for i, entry := range f.fields {
		if entry.hide != nil && entry.hide(f.driver) {
			continue
		}
		label := styles.FormLabel.Render(entry.label + ":")
		field := entry.input.View()
		if i == f.focused {
			label = styles.FormLabel.Foreground(styles.ColorPrimary).Render(entry.label + ":")
		}
		b.WriteString(fmt.Sprintf("%s %s\n", label, field))
	}
	b.WriteString("\n")
	b.WriteString(styles.MutedText.Render("[tab] next field  [ctrl+s] save  [esc] cancel"))

	// Driver hint
	driverHint := ""
	switch f.driver {
	case "sqlite":
		driverHint = "  (sqlite: only path needed)"
	case "postgres":
		driverHint = "  (postgres: host/port/user/password/database)"
	case "mysql":
		driverHint = "  (mysql: host/port/user/password/database)"
	}
	if driverHint != "" {
		b.WriteString("\n")
		b.WriteString(styles.MutedText.Render(driverHint))
	}

	return b.String()
}

func (f FormModel) Values() map[string]string {
	vals := make(map[string]string)
	for _, entry := range f.fields {
		vals[entry.key] = entry.input.Value()
	}
	return vals
}

func (f FormModel) visibleFields() []int {
	var visible []int
	for i, entry := range f.fields {
		if entry.hide == nil || !entry.hide(f.driver) {
			visible = append(visible, i)
		}
	}
	return visible
}

func (f FormModel) focusNext() FormModel {
	visible := f.visibleFields()
	f.fields[f.focused].input.Blur()
	currentPos := -1
	for i, idx := range visible {
		if idx == f.focused {
			currentPos = i
			break
		}
	}
	if currentPos < len(visible)-1 {
		f.focused = visible[currentPos+1]
	} else if len(visible) > 0 {
		f.focused = visible[0]
	}
	f.fields[f.focused].input.Focus()
	return f
}

func (f FormModel) focusPrev() FormModel {
	visible := f.visibleFields()
	f.fields[f.focused].input.Blur()
	currentPos := -1
	for i, idx := range visible {
		if idx == f.focused {
			currentPos = i
			break
		}
	}
	if currentPos > 0 {
		f.focused = visible[currentPos-1]
	} else if len(visible) > 0 {
		f.focused = visible[len(visible)-1]
	}
	f.fields[f.focused].input.Focus()
	return f
}

func (f FormModel) submit() tea.Cmd {
	vals := f.Values()
	return func() tea.Msg {
		return FormSubmitMsg{Values: vals}
	}
}

