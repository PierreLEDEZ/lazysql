package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lazysql/lazysql/pkg/tui/styles"
)

type ConfirmResultMsg struct {
	Confirmed bool
	Payload   tea.Msg
}

type ConfirmModel struct {
	message string
	onYes   tea.Msg
}

func NewConfirm(message string, onYes tea.Msg) ConfirmModel {
	return ConfirmModel{
		message: message,
		onYes:   onYes,
	}
}

func (c ConfirmModel) Init() tea.Cmd {
	return nil
}

func (c ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			payload := c.onYes
			return c, func() tea.Msg {
				return ConfirmResultMsg{Confirmed: true, Payload: payload}
			}
		case "n", "N", "esc":
			return c, func() tea.Msg {
				return ConfirmResultMsg{Confirmed: false}
			}
		}
	}
	return c, nil
}

func (c ConfirmModel) View() string {
	msg := styles.WarningText.Render(c.message)
	hint := fmt.Sprintf("\n\n%s / %s",
		styles.StatusKey.Render("[Y]"),
		styles.MutedText.Render("es   "),
	) + fmt.Sprintf("%s / %s",
		styles.StatusKey.Render("[N]"),
		styles.MutedText.Render("o"),
	)
	return msg + hint
}
