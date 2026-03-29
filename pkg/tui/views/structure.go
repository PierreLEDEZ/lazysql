package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lazysql/lazysql/pkg/db"
	"github.com/lazysql/lazysql/pkg/tui/styles"
)

type StructureModel struct {
	columns []db.Column
	cursor  int
	offset  int
	active  bool
	Width   int
	Height  int
}

func NewStructure() StructureModel {
	return StructureModel{}
}

func (s StructureModel) Init() tea.Cmd {
	return nil
}

func (s StructureModel) Update(msg tea.Msg) (StructureModel, tea.Cmd) {
	switch msg := msg.(type) {
	case StructureUpdatedMsg:
		if msg.Err == nil {
			s.columns = msg.Columns
			s.cursor = 0
			s.offset = 0
		}
		return s, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
				if s.cursor < s.offset {
					s.offset = s.cursor
				}
			}
		case "down", "j":
			if s.cursor < len(s.columns)-1 {
				s.cursor++
				viewH := s.viewableHeight()
				if s.cursor >= s.offset+viewH {
					s.offset = s.cursor - viewH + 1
				}
			}
		}
	}
	return s, nil
}

func (s StructureModel) View() string {
	title := styles.PanelTitle(3, fmt.Sprintf("Structure (%d)", len(s.columns)), s.active)
	if len(s.columns) == 0 {
		return title + "\n" + styles.MutedText.Render("  Select a table to view structure.")
	}

	viewH := s.viewableHeight()
	end := s.offset + viewH
	if end > len(s.columns) {
		end = len(s.columns)
	}

	var b strings.Builder
	b.WriteString(title)
	for i := s.offset; i < end; i++ {
		col := s.columns[i]
		prefix := "  "
		if i == s.cursor {
			prefix = "> "
		}

		line := formatColumn(col)

		b.WriteString("\n")
		if i == s.cursor {
			b.WriteString(lipgloss.NewStyle().Foreground(styles.ColorYellow).Bold(true).Render(prefix + line))
		} else {
			b.WriteString(prefix + line)
		}
	}
	return b.String()
}

func (s *StructureModel) SetSize(w, h int) {
	s.Width = w
	s.Height = h - 1 // title
}

func (s *StructureModel) SetActive(v bool) { s.active = v }

func (s *StructureModel) Clear() {
	s.columns = nil
	s.cursor = 0
	s.offset = 0
}

func (s StructureModel) viewableHeight() int {
	if s.Height <= 0 {
		return len(s.columns)
	}
	return s.Height
}

func formatColumn(col db.Column) string {
	var parts []string
	parts = append(parts, col.Name)
	parts = append(parts, col.Type)

	if col.PrimaryKey {
		parts = append(parts, "PK")
	}
	if col.Nullable {
		parts = append(parts, "NULL")
	}
	if col.Default != nil {
		parts = append(parts, fmt.Sprintf("=%s", *col.Default))
	}

	return strings.Join(parts, " ")
}
