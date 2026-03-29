package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lazysql/lazysql/pkg/tui/styles"
)

type HelpKeyMap interface {
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

type HelpModel struct {
	content string
}

func NewHelp(keys interface{}) HelpModel {
	return HelpModel{content: buildHelpContent(keys)}
}

func (h HelpModel) Init() tea.Cmd { return nil }

func (h HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return h, nil
}

func (h HelpModel) View() string {
	return h.content
}

func buildHelpContent(keys interface{}) string {
	var b strings.Builder

	section := func(title string) {
		b.WriteString(styles.HelpSection.Render(title))
		b.WriteString("\n")
	}

	binding := func(key, desc string) {
		b.WriteString(fmt.Sprintf("  %s%s\n",
			styles.HelpKey.Render(key),
			styles.HelpDesc.Render(desc)))
	}

	section("Global")
	binding("tab / shift+tab", "Cycle through all panels")
	binding("h / ← / l / →", "Jump between left and right columns")
	binding("[ / ]", "Cycle panels within current column")
	binding("1-5", "Jump to panel directly")
	binding("?", "Toggle this help screen")
	binding("ctrl+d", "Disconnect from database")
	binding("q / ctrl+c", "Quit lazysql")

	section("Connections Panel")
	binding("j/k or ↑/↓", "Navigate connections")
	binding("enter", "Connect to selected database")
	binding("a", "Add new connection")
	binding("e", "Edit selected connection")
	binding("d / delete", "Delete selected connection")

	section("Tables Panel")
	binding("j/k or ↑/↓", "Navigate tables")
	binding("enter", "Generate SELECT query for table")
	binding("s", "View table structure in detail")

	section("Structure Panel")
	binding("j/k or ↑/↓", "Scroll through columns")
	binding("", "Auto-updates when a table is selected")

	section("Query Panel (vim-style)")
	binding("i / a", "Enter INSERT mode (start typing)")
	binding("esc", "Return to NORMAL mode")
	binding("ctrl+e / F5", "Execute current query")
	binding("ctrl+p", "Previous query from history")
	binding("ctrl+n", "Next query from history")
	binding("", "Destructive queries require confirmation")

	section("Results Panel")
	binding("j/k or ↑/↓", "Scroll through results")
	binding("← →", "Scroll columns horizontally")

	return b.String()
}
