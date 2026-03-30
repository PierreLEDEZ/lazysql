package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lazysql/lazysql/pkg/config"
	"github.com/lazysql/lazysql/pkg/tui/styles"
)

type SavedQueriesModel struct {
	queries []config.SavedQuery
	cursor  int
	offset  int
	width   int
	height  int
}

func NewSavedQueries(queries []config.SavedQuery) SavedQueriesModel {
	return SavedQueriesModel{queries: queries}
}

func (s SavedQueriesModel) Init() tea.Cmd { return nil }

func (s SavedQueriesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
			if s.cursor < len(s.queries)-1 {
				s.cursor++
				viewH := s.viewHeight()
				if s.cursor >= s.offset+viewH {
					s.offset = s.cursor - viewH + 1
				}
			}
		case "enter":
			if s.cursor < len(s.queries) {
				sql := s.queries[s.cursor].SQL
				return s, func() tea.Msg {
					return LoadSavedQueryMsg{SQL: sql}
				}
			}
		case "d", "delete":
			if s.cursor < len(s.queries) {
				idx := s.cursor
				return s, func() tea.Msg {
					return DeleteSavedQueryMsg{Index: idx}
				}
			}
		}
	}
	return s, nil
}

func (s SavedQueriesModel) View() string {
	if len(s.queries) == 0 {
		return styles.MutedText.Render("  No saved queries.\n  Use ctrl+s to save the current query.")
	}

	viewH := s.viewHeight()
	end := s.offset + viewH
	if end > len(s.queries) {
		end = len(s.queries)
	}

	var b strings.Builder
	for i := s.offset; i < end; i++ {
		q := s.queries[i]
		prefix := "  "
		if i == s.cursor {
			prefix = "> "
		}

		name := styles.StatusKey.Render(q.Name)
		snippet := truncateSQL(q.SQL, 60)
		line := fmt.Sprintf("%s%s  %s", prefix, name, styles.MutedText.Render(snippet))

		if i == s.cursor {
			b.WriteString(styles.SuccessText.Render(prefix) + name + "  " + styles.MutedText.Render(snippet))
		} else {
			b.WriteString(line)
		}
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n")
	b.WriteString(styles.MutedText.Render("  [enter] load  [d] delete  [esc] close"))

	return b.String()
}

func (s SavedQueriesModel) viewHeight() int {
	if s.height <= 3 {
		return len(s.queries)
	}
	return s.height - 3 // hints
}

func truncateSQL(sql string, max int) string {
	// Collapse to single line
	s := strings.ReplaceAll(sql, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
