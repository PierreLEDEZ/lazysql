package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lazysql/lazysql/pkg/tui/components"
	"github.com/lazysql/lazysql/pkg/tui/styles"
)

type TablesModel struct {
	list   components.ListModel
	tables []string
	active bool
	Width  int
	Height int
}

func NewTables() TablesModel {
	return TablesModel{
		list: components.NewList("Tables", nil),
	}
}

func (t TablesModel) Init() tea.Cmd {
	return nil
}

func (t TablesModel) Update(msg tea.Msg) (TablesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case TablesLoadedMsg:
		if msg.Err == nil {
			t.tables = msg.Tables
			t.list.SetItems(msg.Tables)
		}
		return t, nil
	case components.ItemSelectedMsg:
		if msg.Index >= 0 && msg.Index < len(t.tables) {
			tableName := t.tables[msg.Index]
			return t, func() tea.Msg {
				return TableSelectedMsg{Table: tableName}
			}
		}
	}

	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return t, cmd
}

func (t TablesModel) View() string {
	title := styles.PanelTitle(2, fmt.Sprintf("Tables (%d)", len(t.tables)), t.active)
	content := t.list.View()
	if len(t.tables) == 0 {
		content = styles.MutedText.Render("  Connect to a database first.")
	}
	return fmt.Sprintf("%s\n%s", title, content)
}

func (t *TablesModel) SetSize(w, h int) {
	t.Width = w
	t.Height = h
	t.list.SetSize(w, h-1)
}

func (t *TablesModel) Clear() {
	t.tables = nil
	t.list.SetItems(nil)
}

func (t *TablesModel) SetActive(v bool) { t.active = v }

func (t TablesModel) SelectedTable() string {
	if t.list.Cursor >= 0 && t.list.Cursor < len(t.tables) {
		return t.tables[t.list.Cursor]
	}
	return ""
}
