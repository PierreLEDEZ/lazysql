package components

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type EditorModel struct {
	textarea textarea.Model
	Width    int
	Height   int
}

func NewEditor() EditorModel {
	ta := textarea.New()
	ta.Placeholder = "Type SQL here..."
	ta.CharLimit = 0
	ta.ShowLineNumbers = false
	return EditorModel{textarea: ta}
}

func (e EditorModel) Init() tea.Cmd {
	return textarea.Blink
}

func (e EditorModel) Update(msg tea.Msg) (EditorModel, tea.Cmd) {
	var cmd tea.Cmd
	e.textarea, cmd = e.textarea.Update(msg)
	return e, cmd
}

func (e EditorModel) View() string {
	return e.textarea.View()
}

func (e *EditorModel) Value() string {
	return e.textarea.Value()
}

func (e *EditorModel) SetValue(s string) {
	e.textarea.SetValue(s)
}

func (e *EditorModel) Focus() tea.Cmd {
	return e.textarea.Focus()
}

func (e *EditorModel) Blur() {
	e.textarea.Blur()
}

func (e *EditorModel) Focused() bool {
	return e.textarea.Focused()
}

func (e *EditorModel) SetSize(w, h int) {
	e.Width = w
	e.Height = h
	e.textarea.SetWidth(w)
	e.textarea.SetHeight(h)
}
