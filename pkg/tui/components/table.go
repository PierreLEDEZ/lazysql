package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lazysql/lazysql/pkg/tui/styles"
)

const (
	minColWidth = 12 // minimum characters per column
	colGap      = 3  // padding/gap bubbles/table adds per column
)

type TableModel struct {
	table     table.Model
	Width     int
	Height    int
	empty     bool
	colOffset int
	allCols   []string
	allRows   [][]string
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "H":
			if t.colOffset > 0 {
				t.colOffset--
				t.refreshVisibleColumns()
			}
			return t, nil
		case "right", "L":
			maxOffset := len(t.allCols) - t.fitColCount()
			if maxOffset < 0 {
				maxOffset = 0
			}
			if t.colOffset < maxOffset {
				t.colOffset++
				t.refreshVisibleColumns()
			}
			return t, nil
		}
	}

	var cmd tea.Cmd
	t.table, cmd = t.table.Update(msg)
	return t, cmd
}

func (t TableModel) View() string {
	if t.empty {
		return styles.MutedText.Render("  No results yet. Run a query to see data here.")
	}

	output := t.table.View()

	totalCols := len(t.allCols)
	visCols := t.fitColCount()
	if totalCols > visCols {
		endCol := t.colOffset + visCols
		if endCol > totalCols {
			endCol = totalCols
		}
		hint := styles.MutedText.Render(
			fmt.Sprintf(" cols %d-%d of %d (←→ scroll)",
				t.colOffset+1, endCol, totalCols))
		output += "\n" + hint
	}

	// Hard-clip to prevent any overflow beyond panel border
	return lipgloss.NewStyle().MaxWidth(t.Width).Render(output)
}

func (t *TableModel) SetData(columns []string, rows [][]string) {
	if len(columns) == 0 {
		t.empty = true
		return
	}
	t.empty = false
	t.allCols = columns
	t.allRows = rows
	t.colOffset = 0
	t.refreshVisibleColumns()
}

func (t *TableModel) SetSize(w, h int) {
	t.Width = w
	t.Height = h
	t.table.SetWidth(w)
	t.table.SetHeight(h - 1) // reserve 1 line for scroll hint
	if !t.empty {
		t.refreshVisibleColumns()
	}
}

func (t *TableModel) Focus() {
	t.table.Focus()
}

func (t *TableModel) Blur() {
	t.table.Blur()
}

// fitColCount returns how many columns fit within Width,
// accounting for the gap that bubbles/table adds per column.
func (t *TableModel) fitColCount() int {
	if len(t.allCols) == 0 || t.Width <= 0 {
		return 0
	}
	n := t.Width / (minColWidth + colGap)
	if n < 1 {
		n = 1
	}
	if n > len(t.allCols) {
		n = len(t.allCols)
	}
	return n
}

func (t *TableModel) refreshVisibleColumns() {
	visCols := t.fitColCount()
	if visCols == 0 {
		return
	}

	end := t.colOffset + visCols
	if end > len(t.allCols) {
		end = len(t.allCols)
	}
	actual := end - t.colOffset

	// Distribute available width evenly across visible columns
	colW := (t.Width / actual) - colGap
	if colW < minColWidth {
		colW = minColWidth
	}

	cols := make([]table.Column, actual)
	for i := 0; i < actual; i++ {
		cols[i] = table.Column{
			Title: t.allCols[t.colOffset+i],
			Width: colW,
		}
	}

	tableRows := make([]table.Row, len(t.allRows))
	for i, r := range t.allRows {
		row := make(table.Row, actual)
		for j := 0; j < actual; j++ {
			srcIdx := t.colOffset + j
			if srcIdx < len(r) {
				row[j] = r[srcIdx]
			}
		}
		tableRows[i] = row
	}

	t.table.SetColumns(cols)
	t.table.SetRows(tableRows)
}
