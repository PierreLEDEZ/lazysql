package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lazysql/lazysql/pkg/config"
	"github.com/lazysql/lazysql/pkg/dbconnect"
	"github.com/lazysql/lazysql/pkg/tui/components"
	"github.com/lazysql/lazysql/pkg/tui/styles"
)

type ConnectionsModel struct {
	list        components.ListModel
	connections []config.Connection
	loading     bool
	connected   int // index of the connected connection, -1 if none
	active      bool
	spinner     spinner.Model
	Width       int
	Height      int
}

func NewConnections(connections []config.Connection) ConnectionsModel {
	names := make([]string, len(connections))
	for i, c := range connections {
		names[i] = fmt.Sprintf("%s (%s)", c.Name, c.Driver)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorPrimary)

	return ConnectionsModel{
		list:        components.NewList("Connections", names),
		connections: connections,
		connected:   -1,
		spinner:     s,
	}
}

func (c ConnectionsModel) Init() tea.Cmd {
	return nil
}

func (c ConnectionsModel) Update(msg tea.Msg) (ConnectionsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if c.loading {
			var cmd tea.Cmd
			c.spinner, cmd = c.spinner.Update(msg)
			return c, cmd
		}
	case components.ItemSelectedMsg:
		if msg.Index >= 0 && msg.Index < len(c.connections) {
			c.loading = true
			conn := c.connections[msg.Index]
			return c, tea.Batch(
				c.spinner.Tick,
				func() tea.Msg {
					driver, tun, err := dbconnect.Connect(conn)
					return ConnectMsg{Driver: driver, Tunnel: tun, Err: err}
				},
			)
		}
	case ConnectMsg:
		c.loading = false
		if msg.Err == nil {
			c.connected = c.list.Cursor
			c.list.Selected = c.list.Cursor
		}
	}

	var cmd tea.Cmd
	c.list, cmd = c.list.Update(msg)
	return c, cmd
}

func (c ConnectionsModel) View() string {
	title := styles.PanelTitle(1, fmt.Sprintf("Connections (%d)", len(c.connections)), c.active)
	var content string
	switch {
	case len(c.connections) == 0:
		content = styles.MutedText.Render("  No connections.\n  Press 'a' to add one.")
	default:
		content = c.list.View()
		if c.loading {
			content += "\n" + fmt.Sprintf("  %s Connecting...", c.spinner.View())
		}
	}
	return fmt.Sprintf("%s\n%s", title, content)
}

func (c *ConnectionsModel) SetSize(w, h int) {
	c.Width = w
	c.Height = h
	listH := h - 1 // title
	if c.loading {
		listH-- // spinner line
	}
	c.list.SetSize(w, listH)
}

func (c ConnectionsModel) CursorIndex() int {
	return c.list.Cursor
}

func (c *ConnectionsModel) ClearConnected() {
	c.connected = -1
	c.list.Selected = -1
}

func (c *ConnectionsModel) SetActive(v bool) {
	c.active = v
}
