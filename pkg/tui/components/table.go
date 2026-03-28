package components

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lazysql/lazysql/pkg/tui/styles"
)

type TableModel struct {
	table  table.Model
	Width  int
	Height int
	empty  bool
}

func NewTable() TableModel {
	t := table.New(
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.ColorBorder).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#F8F8F2")).
		Background(styles.ColorPrimary).
		Bold(false)
	t.SetStyles(s)

	return TableModel{table: t, empty: true}
}

func (t TableModel) Init() tea.Cmd {
	return nil
}

func (t TableModel) Update(msg tea.Msg) (TableModel, tea.Cmd) {
	var cmd tea.Cmd
	t.table, cmd = t.table.Update(msg)
	return t, cmd
}

func (t TableModel) View() string {
	if t.empty {
		return styles.MutedText.Render("  No results yet. Run a query to see data here.")
	}
	return t.table.View()
}

func (t *TableModel) SetData(columns []string, rows [][]string) {
	if len(columns) == 0 {
		t.empty = true
		return
	}
	t.empty = false

	colWidth := t.Width / len(columns)
	if colWidth < 10 {
		colWidth = 10
	}

	cols := make([]table.Column, len(columns))
	for i, c := range columns {
		cols[i] = table.Column{Title: c, Width: colWidth}
	}

	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		tableRows[i] = table.Row(r)
	}

	t.table.SetRows(nil)
	t.table.SetColumns(cols)
	t.table.SetRows(tableRows)
}

func (t *TableModel) SetSize(w, h int) {
	t.Width = w
	t.Height = h
	t.table.SetWidth(w)
	t.table.SetHeight(h)
}

func (t *TableModel) Focus() {
	t.table.Focus()
}

func (t *TableModel) Blur() {
	t.table.Blur()
}
