package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lazysql/lazysql/pkg/config"
	"github.com/lazysql/lazysql/pkg/tui"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	model := tui.New(cfg)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	_, err = p.Run()
	return err
}
