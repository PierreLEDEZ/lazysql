package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lazysql/lazysql/pkg/tui/styles"
)

type DismissModalMsg struct{}

type ModalModel struct {
	Inner     tea.Model
	title     string
	widthPct  float64
	heightPct float64
	screenW   int
	screenH   int
}

func NewModal(title string, inner tea.Model, widthPct, heightPct float64) ModalModel {
	return ModalModel{
		Inner:     inner,
		title:     title,
		widthPct:  widthPct,
		heightPct: heightPct,
	}
}

func (m ModalModel) Init() tea.Cmd {
	return m.Inner.Init()
}

func (m ModalModel) Update(msg tea.Msg) (ModalModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			return m, func() tea.Msg { return DismissModalMsg{} }
		}
	case tea.WindowSizeMsg:
		m.screenW = msg.Width
		m.screenH = msg.Height
	}

	var cmd tea.Cmd
	m.Inner, cmd = m.Inner.Update(msg)
	return m, cmd
}

func (m ModalModel) View() string {
	w := int(float64(m.screenW) * m.widthPct)
	h := int(float64(m.screenH) * m.heightPct)
	if w < 30 {
		w = 30
	}
	if h < 10 {
		h = 10
	}
	// Account for border + padding (4 cols, 4 rows)
	contentW := w - 8
	contentH := h - 6

	title := styles.ModalTitle.Render(m.title)
	hint := styles.MutedText.Render("[esc] close")
	header := fmt.Sprintf("%s  %s", title, hint)

	content := m.Inner.View()

	box := styles.ModalBorder.
		Width(contentW).
		Height(contentH).
		Render(fmt.Sprintf("%s\n%s", header, content))

	return lipgloss.Place(m.screenW, m.screenH,
		lipgloss.Center, lipgloss.Center,
		box)
}

func (m *ModalModel) SetScreenSize(w, h int) {
	m.screenW = w
	m.screenH = h
}
