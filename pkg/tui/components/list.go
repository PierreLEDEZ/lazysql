package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lazysql/lazysql/pkg/tui/styles"
)

type ListModel struct {
	Items    []string
	Cursor   int
	Selected int
	Title    string
	Width    int
	Height   int
	offset   int
}

type ItemSelectedMsg struct {
	Index int
	Item  string
}

func NewList(title string, items []string) ListModel {
	return ListModel{
		Items:    items,
		Cursor:   0,
		Selected: -1,
		Title:    title,
	}
}

func (l ListModel) Init() tea.Cmd {
	return nil
}

func (l ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if l.Cursor > 0 {
				l.Cursor--
				if l.Cursor < l.offset {
					l.offset = l.Cursor
				}
			}
		case "down", "j":
			if l.Cursor < len(l.Items)-1 {
				l.Cursor++
				visibleHeight := l.viewableHeight()
				if l.Cursor >= l.offset+visibleHeight {
					l.offset = l.Cursor - visibleHeight + 1
				}
			}
		case "enter":
			if len(l.Items) > 0 {
				l.Selected = l.Cursor
				return l, func() tea.Msg {
					return ItemSelectedMsg{Index: l.Cursor, Item: l.Items[l.Cursor]}
				}
			}
		}
	}
	return l, nil
}

func (l ListModel) View() string {
	if len(l.Items) == 0 {
		return styles.MutedText.Render("  (empty)")
	}

	visibleHeight := l.viewableHeight()
	if visibleHeight <= 0 {
		visibleHeight = len(l.Items)
	}

	var b strings.Builder
	end := l.offset + visibleHeight
	if end > len(l.Items) {
		end = len(l.Items)
	}

	for i := l.offset; i < end; i++ {
		cursor := "  "
		style := lipgloss.NewStyle()

		if i == l.Cursor {
			cursor = "> "
			style = style.Foreground(styles.ColorPrimary).Bold(true)
		}
		if i == l.Selected {
			style = style.Foreground(styles.ColorSuccess)
		}

		line := fmt.Sprintf("%s%s", cursor, l.Items[i])
		b.WriteString(style.Render(line))
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (l ListModel) viewableHeight() int {
	if l.Height <= 0 {
		return len(l.Items)
	}
	return l.Height
}

func (l *ListModel) SetItems(items []string) {
	l.Items = items
	l.Cursor = 0
	l.Selected = -1
	l.offset = 0
}

func (l *ListModel) SetSize(w, h int) {
	l.Width = w
	l.Height = h
}
