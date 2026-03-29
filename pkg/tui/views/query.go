package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lazysql/lazysql/pkg/db"
	"github.com/lazysql/lazysql/pkg/tui/components"
	"github.com/lazysql/lazysql/pkg/tui/styles"
)

type VimMode int

const (
	VimNormal VimMode = iota
	VimInsert
)

type QueryModel struct {
	editor  components.EditorModel
	driver  db.Driver
	mode    VimMode
	active  bool
	Width   int
	Height  int
}

func NewQuery() QueryModel {
	return QueryModel{
		editor: components.NewEditor(),
		mode:   VimNormal,
	}
}

func (q QueryModel) Init() tea.Cmd {
	return q.editor.Init()
}

func (q QueryModel) Update(msg tea.Msg) (QueryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch q.mode {
		case VimNormal:
			switch msg.String() {
			case "i":
				q.mode = VimInsert
				return q, q.editor.Focus()
			case "a":
				q.mode = VimInsert
				return q, q.editor.Focus()
			}
			// In normal mode, don't forward keys to editor
			return q, nil
		case VimInsert:
			if msg.String() == "esc" {
				q.mode = VimNormal
				q.editor.Blur()
				return q, nil
			}
		}
	}

	// Forward to editor only in insert mode
	if q.mode == VimInsert {
		var cmd tea.Cmd
		q.editor, cmd = q.editor.Update(msg)
		return q, cmd
	}
	return q, nil
}

func (q QueryModel) View() string {
	modeStr := styles.VimNormal.Render(" NORMAL ")
	if q.mode == VimInsert {
		modeStr = styles.VimInsert.Render(" INSERT ")
	}
	title := styles.PanelTitle(4, "Query", q.active) + " " + modeStr

	if q.driver == nil {
		return fmt.Sprintf("%s\n%s", title,
			styles.MutedText.Render("  Connect to a database to write queries."))
	}

	hint := ""
	if q.mode == VimNormal {
		hint = styles.MutedText.Render("  [i] insert  [ctrl+e] execute")
	}

	content := q.editor.View()
	if hint != "" {
		return fmt.Sprintf("%s\n%s\n%s", title, content, hint)
	}
	return fmt.Sprintf("%s\n%s", title, content)
}

func (q *QueryModel) SetDriver(d db.Driver) {
	q.driver = d
}

func (q *QueryModel) SetValue(s string) {
	q.editor.SetValue(s)
}

func (q *QueryModel) Value() string {
	return q.editor.Value()
}

func (q *QueryModel) Focus() tea.Cmd {
	// Don't auto-enter insert mode — stay in normal mode
	return nil
}

func (q *QueryModel) Blur() {
	q.mode = VimNormal
	q.editor.Blur()
}

func (q *QueryModel) Focused() bool {
	return q.mode == VimInsert
}

func (q *QueryModel) Mode() VimMode {
	return q.mode
}

func (q *QueryModel) SetSize(w, h int) {
	q.Width = w
	q.Height = h
	editorH := h - 1 // title
	if q.mode == VimNormal {
		editorH-- // hint line
	}
	if editorH < 1 {
		editorH = 1
	}
	q.editor.SetSize(w, editorH)
}

func (q *QueryModel) SetActive(v bool) { q.active = v }
