package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lazysql/lazysql/pkg/config"
	"github.com/lazysql/lazysql/pkg/db"
	"github.com/lazysql/lazysql/pkg/tui/components"
	"github.com/lazysql/lazysql/pkg/tui/styles"
	"github.com/lazysql/lazysql/pkg/tui/views"
	"github.com/lazysql/lazysql/pkg/tunnel"
)

type Panel int

const (
	PanelConnections Panel = iota
	PanelTables
	PanelStructure
	PanelQuery
	PanelResults
	panelCount = 5
)

type Model struct {
	activePanel Panel
	driver      db.Driver
	tunnel      *tunnel.Tunnel
	connName    string

	// Sub-models
	connections views.ConnectionsModel
	tables      views.TablesModel
	structure   views.StructureModel
	query       views.QueryModel
	results     views.ResultsModel

	// Modal
	activeModal *components.ModalModel

	// State
	config         *config.Config
	keys           KeyMap
	queryHistory   []string
	historyIdx     int
	connIndex      int // index of active connection in config, -1 if none
	spinner        spinner.Model
	loading        bool
	width          int
	height         int
	status         string
	lastLeftPanel  Panel
	lastRightPanel Panel
	pendingSaveSQL string
}

func New(cfg *config.Config) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorPrimary)

	conns := views.NewConnections(cfg.Connections)
	conns.SetActive(true)

	return Model{
		activePanel:    PanelConnections,
		connections:    conns,
		tables:         views.NewTables(),
		structure:      views.NewStructure(),
		query:          views.NewQuery(),
		results:        views.NewResults(),
		config:         cfg,
		keys:           DefaultKeyMap(),
		historyIdx:     -1,
		connIndex:      -1,
		spinner:        s,
		lastLeftPanel:  PanelConnections,
		lastRightPanel: PanelQuery,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Spinner tick
	switch msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Window resize — always handle
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()
		if m.activeModal != nil {
			m.activeModal.SetScreenSize(m.width, m.height)
		}
		return m, tea.Batch(cmds...)
	}

	// Modal mode: route everything to modal
	if m.activeModal != nil {
		return m.updateModal(msg, cmds)
	}

	// Normal mode
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg, cmds)

	case views.ConnectMsg:
		return m.handleConnectMsg(msg, cmds)

	case views.TablesLoadedMsg:
		if msg.Err != nil {
			m.status = fmt.Sprintf("Error: %s", msg.Err)
		}
		var cmd tea.Cmd
		m.tables, cmd = m.tables.Update(msg)
		cmds = append(cmds, cmd)
		m.loading = false
		return m, tea.Batch(cmds...)

	case views.TableSelectedMsg:
		m.query.SetValue(fmt.Sprintf("SELECT * FROM %s LIMIT 100;", msg.Table))
		cmd := m.switchPanel(PanelQuery)
		cmds = append(cmds, cmd)
		// Auto-load structure
		if m.driver != nil {
			driver := m.driver
			table := msg.Table
			cmds = append(cmds, func() tea.Msg {
				cols, err := driver.DescribeTable(table)
				return views.StructureUpdatedMsg{Columns: cols, Err: err}
			})
		}
		return m, tea.Batch(cmds...)

	case views.DescribeTableResultMsg:
		if msg.Err != nil {
			m.status = fmt.Sprintf("Error: %s", msg.Err)
		} else {
			// Also update the structure panel inline
			m.structure.Update(views.StructureUpdatedMsg{Columns: msg.Columns})
		}
		return m, tea.Batch(cmds...)

	case views.StructureUpdatedMsg:
		var cmd tea.Cmd
		m.structure, cmd = m.structure.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case views.QueryResultMsg:
		if msg.Err == nil && msg.Result != nil {
			m.status = fmt.Sprintf("%d rows in %s", msg.Result.RowCount, msg.Result.Elapsed.Round(1e6))
		} else if msg.Err != nil {
			m.status = fmt.Sprintf("Error: %s", msg.Err)
		}
		m.loading = false
		var cmd tea.Cmd
		m.results, cmd = m.results.Update(msg)
		cmds = append(cmds, cmd)
		switchCmd := m.switchPanel(PanelResults)
		cmds = append(cmds, switchCmd)
		return m, tea.Batch(cmds...)

	case views.ExecResultMsg:
		m.loading = false
		var cmd tea.Cmd
		m.results, cmd = m.results.Update(msg)
		cmds = append(cmds, cmd)
		if msg.Err == nil && msg.Result != nil {
			m.status = fmt.Sprintf("%d rows affected in %s", msg.Result.RowsAffected, msg.Result.Elapsed.Round(1e6))
		} else if msg.Err != nil {
			m.status = fmt.Sprintf("Error: %s", msg.Err)
		}
		return m, tea.Batch(cmds...)

	case views.ExecuteConfirmedMsg:
		return m.executeSQL(msg.SQL, cmds)

	case views.LoadSavedQueryMsg:
		m.activeModal = nil
		m.query.SetValue(msg.SQL)
		cmd := m.switchPanel(PanelQuery)
		m.status = "Query loaded"
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case views.DeleteSavedQueryMsg:
		m.activeModal = nil
		if m.connIndex >= 0 && m.connIndex < len(m.config.Connections) {
			conn := &m.config.Connections[m.connIndex]
			if msg.Index >= 0 && msg.Index < len(conn.SavedQueries) {
				name := conn.SavedQueries[msg.Index].Name
				conn.SavedQueries = append(conn.SavedQueries[:msg.Index], conn.SavedQueries[msg.Index+1:]...)
				if err := config.Save(m.config); err != nil {
					m.status = fmt.Sprintf("Error: %s", err)
				} else {
					m.status = fmt.Sprintf("Query '%s' deleted", name)
				}
			}
		}
		return m, tea.Batch(cmds...)

	case views.SaveQueryMsg:
		m.activeModal = nil
		if m.connIndex >= 0 && m.connIndex < len(m.config.Connections) {
			conn := &m.config.Connections[m.connIndex]
			conn.SavedQueries = append(conn.SavedQueries, config.SavedQuery{
				Name: msg.Name,
				SQL:  msg.SQL,
			})
			if err := config.Save(m.config); err != nil {
				m.status = fmt.Sprintf("Error saving: %s", err)
			} else {
				m.status = fmt.Sprintf("Query '%s' saved", msg.Name)
			}
		}
		return m, tea.Batch(cmds...)

	case views.ErrorMsg:
		m.status = fmt.Sprintf("Error: %s", msg.Err)
		return m, tea.Batch(cmds...)

	case components.ConfirmResultMsg:
		return m.handleConfirmResult(msg, cmds)

	case components.FormSubmitMsg:
		return m.handleFormSubmit(msg, cmds)

	case views.ConnectionSavedMsg:
		if msg.Err != nil {
			m.status = fmt.Sprintf("Error saving: %s", msg.Err)
		} else {
			m.status = "Connection saved"
			m.connections = views.NewConnections(m.config.Connections)
			m.updateSizes()
		}
		return m, tea.Batch(cmds...)
	}

	// Delegate to active panel
	switch m.activePanel {
	case PanelConnections:
		var cmd tea.Cmd
		m.connections, cmd = m.connections.Update(msg)
		cmds = append(cmds, cmd)
	case PanelTables:
		var cmd tea.Cmd
		m.tables, cmd = m.tables.Update(msg)
		cmds = append(cmds, cmd)
	case PanelStructure:
		var cmd tea.Cmd
		m.structure, cmd = m.structure.Update(msg)
		cmds = append(cmds, cmd)
	case PanelQuery:
		var cmd tea.Cmd
		m.query, cmd = m.query.Update(msg)
		cmds = append(cmds, cmd)
	case PanelResults:
		var cmd tea.Cmd
		m.results, cmd = m.results.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleKeyMsg processes key events in normal (non-modal) mode.
func (m Model) handleKeyMsg(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	queryFocused := m.activePanel == PanelQuery && m.query.Focused()

	switch msg.String() {
	case "ctrl+c":
		return m, m.cleanup()

	case "q":
		if !queryFocused {
			return m, m.cleanup()
		}

	case "?":
		if !queryFocused {
			m.openHelp()
			return m, nil
		}

	case "tab":
		cmd := m.switchPanel((m.activePanel + 1) % panelCount)
		return m, cmd

	case "shift+tab":
		cmd := m.switchPanel((m.activePanel + panelCount - 1) % panelCount)
		return m, cmd

	case "1", "2", "3", "4", "5":
		if !queryFocused {
			panel := Panel(msg.String()[0] - '1')
			cmd := m.switchPanel(panel)
			return m, cmd
		}

	// Lazygit-style spatial navigation
	case "h", "left":
		if !queryFocused {
			cmd := m.jumpToColumn(columnLeft)
			return m, cmd
		}

	case "l", "right":
		if !queryFocused {
			cmd := m.jumpToColumn(columnRight)
			return m, cmd
		}

	case "[":
		if !queryFocused {
			cmd := m.cyclePanelInColumn(-1)
			return m, cmd
		}

	case "]":
		if !queryFocused {
			cmd := m.cyclePanelInColumn(1)
			return m, cmd
		}

	case "ctrl+d":
		if m.driver != nil {
			m.disconnect()
			return m, nil
		}

	// Connection panel actions
	case "a":
		if m.activePanel == PanelConnections && !queryFocused {
			m.openConnectionForm("add", nil, -1)
			return m, nil
		}

	case "e":
		if m.activePanel == PanelConnections && !queryFocused {
			idx := m.connections.CursorIndex()
			if idx >= 0 && idx < len(m.config.Connections) {
				conn := m.config.Connections[idx]
				m.openConnectionForm("edit", &conn, idx)
			}
			return m, nil
		}

	case "d", "delete":
		if m.activePanel == PanelConnections && !queryFocused {
			idx := m.connections.CursorIndex()
			if idx >= 0 && idx < len(m.config.Connections) {
				name := m.config.Connections[idx].Name
				m.openConfirm(
					fmt.Sprintf("Delete connection '%s'?", name),
					views.DeleteConnectionMsg{Index: idx},
				)
			}
			return m, nil
		}

	// Tables panel actions
	case "s":
		if m.activePanel == PanelTables && !queryFocused && m.driver != nil {
			table := m.tables.SelectedTable()
			if table != "" {
				driver := m.driver
				return m, func() tea.Msg {
					cols, err := driver.DescribeTable(table)
					return views.DescribeTableResultMsg{Table: table, Columns: cols, Err: err}
				}
			}
			return m, nil
		}

	// Query history
	case "ctrl+p":
		if m.activePanel == PanelQuery && len(m.queryHistory) > 0 {
			if m.historyIdx < len(m.queryHistory)-1 {
				m.historyIdx++
				m.query.SetValue(m.queryHistory[len(m.queryHistory)-1-m.historyIdx])
			}
			return m, nil
		}

	case "ctrl+n":
		if m.activePanel == PanelQuery {
			if m.historyIdx > 0 {
				m.historyIdx--
				m.query.SetValue(m.queryHistory[len(m.queryHistory)-1-m.historyIdx])
			} else if m.historyIdx == 0 {
				m.historyIdx = -1
				m.query.SetValue("")
			}
			return m, nil
		}

	case "ctrl+s":
		if m.activePanel == PanelQuery && m.connIndex >= 0 {
			sql := strings.TrimSpace(m.query.Value())
			if sql != "" {
				m.openSaveQueryPrompt(sql)
			}
			return m, nil
		}

	case "S":
		if !queryFocused && m.connIndex >= 0 {
			m.openSavedQueries()
			return m, nil
		}

	case "ctrl+e", "f5":
		if m.activePanel == PanelQuery {
			sql := strings.TrimSpace(m.query.Value())
			if sql != "" && m.driver != nil {
				// Check for destructive query
				if isDestructive(sql) {
					m.openConfirm(
						fmt.Sprintf("Execute destructive query?\n\n%s", truncate(sql, 100)),
						views.ExecuteConfirmedMsg{SQL: sql},
					)
					return m, nil
				}
				return m.executeSQL(sql, cmds)
			}
			return m, nil
		}

	}

	// Delegate to active panel for other keys
	switch m.activePanel {
	case PanelConnections:
		var cmd tea.Cmd
		m.connections, cmd = m.connections.Update(msg)
		cmds = append(cmds, cmd)
	case PanelTables:
		var cmd tea.Cmd
		m.tables, cmd = m.tables.Update(msg)
		cmds = append(cmds, cmd)
	case PanelStructure:
		var cmd tea.Cmd
		m.structure, cmd = m.structure.Update(msg)
		cmds = append(cmds, cmd)
	case PanelQuery:
		var cmd tea.Cmd
		m.query, cmd = m.query.Update(msg)
		cmds = append(cmds, cmd)
	case PanelResults:
		var cmd tea.Cmd
		m.results, cmd = m.results.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleConnectMsg(msg views.ConnectMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	// Forward to connections model to clear loading state
	var cmd tea.Cmd
	m.connections, cmd = m.connections.Update(msg)
	cmds = append(cmds, cmd)

	if msg.Err != nil {
		m.status = fmt.Sprintf("Connection error: %s", msg.Err)
		m.loading = false
		return m, tea.Batch(cmds...)
	}
	if m.driver != nil {
		m.driver.Disconnect()
	}
	if m.tunnel != nil {
		m.tunnel.Stop()
	}
	m.driver = msg.Driver
	m.tunnel = msg.Tunnel
	m.query.SetDriver(m.driver)
	m.connName = m.driver.CurrentDatabase()
	m.connIndex = m.connections.CursorIndex()
	m.status = fmt.Sprintf("Connected to %s (%s)", m.connName, m.driver.DriverName())

	driver := m.driver
	cmds = append(cmds, func() tea.Msg {
		tables, err := driver.ListTables(driver.CurrentDatabase())
		return views.TablesLoadedMsg{Tables: tables, Err: err}
	})
	return m, tea.Batch(cmds...)
}

func (m Model) handleConfirmResult(msg components.ConfirmResultMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	m.activeModal = nil
	if !msg.Confirmed {
		return m, tea.Batch(cmds...)
	}

	switch payload := msg.Payload.(type) {
	case views.DeleteConnectionMsg:
		if err := m.config.RemoveConnection(payload.Index); err != nil {
			m.status = fmt.Sprintf("Error: %s", err)
		} else {
			if err := config.Save(m.config); err != nil {
				m.status = fmt.Sprintf("Error saving: %s", err)
			} else {
				m.status = "Connection deleted"
				m.connections = views.NewConnections(m.config.Connections)
				m.updateSizes()
			}
		}
	case views.ExecuteConfirmedMsg:
		return m.executeSQL(payload.SQL, cmds)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleFormSubmit(msg components.FormSubmitMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	m.activeModal = nil
	vals := msg.Values

	// Save query form — only has "name" key, no "driver"
	if _, hasDriver := vals["driver"]; !hasDriver && m.pendingSaveSQL != "" {
		name := strings.TrimSpace(vals["name"])
		if name == "" {
			m.status = "Query name cannot be empty"
			m.pendingSaveSQL = ""
			return m, tea.Batch(cmds...)
		}
		sql := m.pendingSaveSQL
		m.pendingSaveSQL = ""
		return m.Update(views.SaveQueryMsg{Name: name, SQL: sql})
	}

	port := 0
	if vals["port"] != "" {
		fmt.Sscanf(vals["port"], "%d", &port)
	}

	conn := config.Connection{
		Name:     vals["name"],
		Driver:   vals["driver"],
		Host:     vals["host"],
		Port:     port,
		User:     vals["user"],
		Password: vals["password"],
		Database: vals["database"],
		Path:     vals["path"],
	}

	if err := conn.Validate(); err != nil {
		m.status = fmt.Sprintf("Validation error: %s", err)
		return m, tea.Batch(cmds...)
	}

	m.config.AddConnection(conn)
	if err := config.Save(m.config); err != nil {
		m.status = fmt.Sprintf("Error saving: %s", err)
	} else {
		m.status = fmt.Sprintf("Connection '%s' added", conn.Name)
		m.connections = views.NewConnections(m.config.Connections)
		m.updateSizes()
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) executeSQL(sql string, cmds []tea.Cmd) (Model, tea.Cmd) {
	// Add to history
	if len(m.queryHistory) == 0 || m.queryHistory[len(m.queryHistory)-1] != sql {
		m.queryHistory = append(m.queryHistory, sql)
	}
	m.historyIdx = -1
	m.loading = true

	driver := m.driver
	if isQuery(sql) {
		cmds = append(cmds, func() tea.Msg {
			result, err := driver.Query(sql)
			return views.QueryResultMsg{Result: result, Err: err}
		})
	} else {
		cmds = append(cmds, func() tea.Msg {
			result, err := driver.Execute(sql)
			return views.ExecResultMsg{Result: result, Err: err}
		})
	}
	return *m, tea.Batch(cmds...)
}

// updateModal handles all messages when a modal is active.
func (m Model) updateModal(msg tea.Msg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case components.DismissModalMsg:
		m.activeModal = nil
		return m, tea.Batch(cmds...)
	case components.ConfirmResultMsg:
		return m.handleConfirmResult(msg.(components.ConfirmResultMsg), cmds)
	case components.FormSubmitMsg:
		return m.handleFormSubmit(msg.(components.FormSubmitMsg), cmds)
	case views.LoadSavedQueryMsg:
		return m.Update(msg)
	case views.DeleteSavedQueryMsg:
		return m.Update(msg)
	case views.SaveQueryMsg:
		return m.Update(msg)
	}

	if m.activeModal != nil {
		var cmd tea.Cmd
		*m.activeModal, cmd = m.activeModal.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

// ─── View ────────────────────────────────────────────────────

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Modal overlay takes full screen
	if m.activeModal != nil {
		return m.activeModal.View()
	}

	leftW, rightW, connH, tablesH, structH, queryH, resultsH := m.panelSizes()
	panelsH := m.height - 1 // everything except status bar

	// Left column: stacked panels
	connView := styles.PanelStyle(m.activePanel == PanelConnections, leftW, connH).
		Render(m.connections.View())
	tablesView := styles.PanelStyle(m.activePanel == PanelTables, leftW, tablesH).
		Render(m.tables.View())
	structView := styles.PanelStyle(m.activePanel == PanelStructure, leftW, structH).
		Render(m.structure.View())

	leftCol := lipgloss.JoinVertical(lipgloss.Left, connView, tablesView, structView)
	leftCol = clampHeight(leftCol, panelsH)

	// Right column: query + results
	queryView := styles.PanelStyle(m.activePanel == PanelQuery, rightW, queryH).
		Render(m.query.View())
	resultsView := styles.PanelStyle(m.activePanel == PanelResults, rightW, resultsH).
		Render(m.results.View())

	rightCol := lipgloss.JoinVertical(lipgloss.Left, queryView, resultsView)
	rightCol = clampHeight(rightCol, panelsH)

	// Main layout
	main := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)

	// Status bar
	statusBar := m.renderStatusBar()

	screen := main + "\n" + statusBar
	screen = clampHeight(screen, m.height)

	return screen
}

func (m Model) renderStatusBar() string {
	// Left: connection info
	left := ""
	if m.driver != nil {
		left = fmt.Sprintf(" %s | %s", m.driver.DriverName(), m.connName)
	}
	if m.loading {
		left += " " + m.spinner.View()
	}

	// Center: status message
	center := ""
	if m.status != "" {
		center = " │ " + m.status
	}

	// Right: context-sensitive keybindings
	right := m.contextHints()

	info := left + center
	contentW := m.width - 1 // account for StatusBar's PaddingLeft(1)
	padding := contentW - lipgloss.Width(info) - lipgloss.Width(right)
	if padding < 1 {
		padding = 1
	}

	bar := info + strings.Repeat(" ", padding) + right

	return styles.StatusBar.Width(contentW).Render(bar)
}

func (m Model) contextHints() string {
	queryFocused := m.activePanel == PanelQuery && m.query.Focused()

	hint := func(key, desc string) string {
		return styles.StatusKey.Render(key) + styles.StatusDesc.Render(" "+desc+"  ")
	}

	var h strings.Builder
	h.WriteString(hint("?", "help"))
	h.WriteString(hint("h/l", "cols"))
	h.WriteString(hint("[/]", "panels"))

	switch m.activePanel {
	case PanelConnections:
		h.WriteString(hint("a", "add"))
		h.WriteString(hint("e", "edit"))
		h.WriteString(hint("d", "del"))
		h.WriteString(hint("enter", "connect"))
		if m.driver != nil {
			h.WriteString(hint("ctrl+d", "disconnect"))
		}
	case PanelTables:
		h.WriteString(hint("enter", "select"))
		h.WriteString(hint("s", "structure"))
	case PanelStructure:
		// minimal hints
	case PanelQuery:
		if queryFocused {
			h.WriteString(hint("esc", "normal"))
			h.WriteString(hint("ctrl+e", "execute"))
			h.WriteString(hint("ctrl+s", "save"))
		} else {
			h.WriteString(hint("i", "insert"))
			h.WriteString(hint("ctrl+e", "execute"))
			h.WriteString(hint("ctrl+s", "save"))
			h.WriteString(hint("S", "list"))
			h.WriteString(hint("ctrl+p/n", "history"))
		}
	case PanelResults:
		h.WriteString(hint("↑↓", "rows"))
		h.WriteString(hint("←→", "cols"))
	}

	h.WriteString(hint("q", "quit"))
	return h.String()
}

// ─── Panel management ────────────────────────────────────────

type column int

const (
	columnLeft column = iota
	columnRight
)

var leftPanels = []Panel{PanelConnections, PanelTables, PanelStructure}
var rightPanels = []Panel{PanelQuery, PanelResults}

func panelColumn(p Panel) column {
	switch p {
	case PanelConnections, PanelTables, PanelStructure:
		return columnLeft
	default:
		return columnRight
	}
}

// jumpToColumn switches to the other column, remembering the last active panel per column.
func (m *Model) jumpToColumn(col column) tea.Cmd {
	if panelColumn(m.activePanel) == col {
		return nil // already there
	}
	// Jump to first panel in target column
	if col == columnLeft {
		return m.switchPanel(m.lastLeftPanel)
	}
	return m.switchPanel(m.lastRightPanel)
}

// cyclePanelInColumn moves to the next/prev panel within the current column.
func (m *Model) cyclePanelInColumn(dir int) tea.Cmd {
	panels := leftPanels
	if panelColumn(m.activePanel) == columnRight {
		panels = rightPanels
	}
	idx := 0
	for i, p := range panels {
		if p == m.activePanel {
			idx = i
			break
		}
	}
	idx = (idx + dir + len(panels)) % len(panels)
	return m.switchPanel(panels[idx])
}

func (m *Model) switchPanel(p Panel) tea.Cmd {
	if m.activePanel == PanelQuery {
		m.query.Blur()
	}
	if m.activePanel == PanelResults {
		m.results.Blur()
	}

	// Clear active flags
	m.connections.SetActive(false)
	m.tables.SetActive(false)
	m.structure.SetActive(false)
	m.query.SetActive(false)
	m.results.SetActive(false)

	m.activePanel = p

	// Set active flag on new panel
	switch p {
	case PanelConnections:
		m.connections.SetActive(true)
	case PanelTables:
		m.tables.SetActive(true)
	case PanelStructure:
		m.structure.SetActive(true)
	case PanelQuery:
		m.query.SetActive(true)
	case PanelResults:
		m.results.SetActive(true)
	}

	// Remember last visited panel per column
	if panelColumn(p) == columnLeft {
		m.lastLeftPanel = p
	} else {
		m.lastRightPanel = p
	}

	if p == PanelQuery {
		return m.query.Focus()
	}
	if p == PanelResults {
		m.results.Focus()
	}
	return nil
}

func (m *Model) updateSizes() {
	leftW, rightW, connH, tablesH, structH, queryH, resultsH := m.panelSizes()
	m.connections.SetSize(leftW, connH)
	m.tables.SetSize(leftW, tablesH)
	m.structure.SetSize(leftW, structH)
	m.query.SetSize(rightW, queryH)
	m.results.SetSize(rightW, resultsH)
}

func (m Model) panelSizes() (leftW, rightW, connH, tablesH, structH, queryH, resultsH int) {
	// Height budget (lipgloss Height() = content height, border adds 2 extra rows):
	//   Left:  connH + tablesH + structH + 6 (borders 3×2) = m.height - 1
	//   Right: queryH + resultsH + 4 (borders 2×2)          = m.height - 1
	//   StatusBar: 1
	//   Total: m.height
	//
	// So: left content = m.height - 7, right content = m.height - 5
	availH := m.height - 7
	if availH < 9 {
		availH = 9
	}

	// Left: 3 equal panels
	connH = availH / 3
	tablesH = availH / 3
	structH = availH - connH - tablesH

	// Width: left 30%, right 70%
	// Left: 1 panel column = 2 border cols
	// Right: 1 panel column = 2 border cols
	// Separator between left and right = 0 (borders touch)
	availW := m.width - 4 // 2 borders left + 2 borders right
	if availW < 40 {
		availW = 40
	}
	leftW = availW * 30 / 100
	if leftW < 16 {
		leftW = 16
	}
	rightW = availW - leftW

	// Right: query 35%, results 65%
	rightAvailH := availH
	// Right has only 2 panels = 4 border rows, but left has 3 = 6
	// Extra 2 rows for right to use
	rightAvailH = availH + 2
	queryH = rightAvailH * 35 / 100
	if queryH < 4 {
		queryH = 4
	}
	resultsH = rightAvailH - queryH

	return
}

// ─── Actions ─────────────────────────────────────────────────

func (m *Model) disconnect() {
	if m.driver != nil {
		m.driver.Disconnect()
		m.driver = nil
	}
	if m.tunnel != nil {
		m.tunnel.Stop()
		m.tunnel = nil
	}
	m.connName = ""
	m.connIndex = -1
	m.connections.ClearConnected()
	m.tables.Clear()
	m.structure.Clear()
	m.query.SetDriver(nil)
	m.status = "Disconnected"
}

func (m *Model) openHelp() {
	help := views.NewHelp(m.keys)
	modal := components.NewModal("Help", help, 0.7, 0.8)
	modal.SetScreenSize(m.width, m.height)
	m.activeModal = &modal
}

func (m *Model) openConfirm(message string, onYes tea.Msg) {
	confirm := components.NewConfirm(message, onYes)
	modal := components.NewModal("Confirm", confirm, 0.5, 0.3)
	modal.SetScreenSize(m.width, m.height)
	m.activeModal = &modal
}

func (m *Model) openConnectionForm(mode string, conn *config.Connection, index int) {
	fields := connectionFormFields(conn)
	form := components.NewForm(fields)
	title := "Add Connection"
	if mode == "edit" {
		title = "Edit Connection"
	}
	modal := components.NewModal(title, form, 0.6, 0.7)
	modal.SetScreenSize(m.width, m.height)
	m.activeModal = &modal
}

func (m *Model) openSavedQueries() {
	var queries []config.SavedQuery
	if m.connIndex >= 0 && m.connIndex < len(m.config.Connections) {
		queries = m.config.Connections[m.connIndex].SavedQueries
	}
	sq := views.NewSavedQueries(queries)
	connName := m.connName
	modal := components.NewModal(fmt.Sprintf("Saved Queries — %s", connName), sq, 0.7, 0.6)
	modal.SetScreenSize(m.width, m.height)
	m.activeModal = &modal
}

func (m *Model) openSaveQueryPrompt(sql string) {
	fields := []components.FormField{
		{Label: "Name", Key: "name", Placeholder: "e.g. list-active-users"},
	}
	form := components.NewForm(fields)
	// Store SQL in status temporarily — will be used by handleSaveQueryForm
	m.pendingSaveSQL = sql
	modal := components.NewModal("Save Query", form, 0.5, 0.3)
	modal.SetScreenSize(m.width, m.height)
	m.activeModal = &modal
}

func (m Model) cleanup() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			if m.driver != nil {
				m.driver.Disconnect()
			}
			if m.tunnel != nil {
				m.tunnel.Stop()
			}
			return nil
		},
		tea.Quit,
	)
}

// ─── Helpers ─────────────────────────────────────────────────

func isQuery(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	for _, prefix := range []string{"SELECT", "SHOW", "PRAGMA", "DESCRIBE", "EXPLAIN", "WITH"} {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

func isDestructive(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	for _, prefix := range []string{"DELETE", "DROP", "TRUNCATE", "ALTER"} {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

func clampHeight(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func connectionFormFields(conn *config.Connection) []components.FormField {
	fields := []components.FormField{
		{Label: "Name", Key: "name", Placeholder: "my-database"},
		{Label: "Driver", Key: "driver", Placeholder: "postgres | mysql | sqlite"},
		{Label: "Host", Key: "host", Placeholder: "localhost"},
		{Label: "Port", Key: "port", Placeholder: "5432"},
		{Label: "User", Key: "user", Placeholder: "admin"},
		{Label: "Password", Key: "password", Placeholder: ""},
		{Label: "Database", Key: "database", Placeholder: "mydb"},
		{Label: "Path", Key: "path", Placeholder: "/path/to/db.sqlite"},
	}
	if conn != nil {
		for i := range fields {
			switch fields[i].Key {
			case "name":
				fields[i].Value = conn.Name
			case "driver":
				fields[i].Value = conn.Driver
			case "host":
				fields[i].Value = conn.Host
			case "port":
				if conn.Port > 0 {
					fields[i].Value = fmt.Sprintf("%d", conn.Port)
				}
			case "user":
				fields[i].Value = conn.User
			case "password":
				fields[i].Value = conn.Password
			case "database":
				fields[i].Value = conn.Database
			case "path":
				fields[i].Value = conn.Path
			}
		}
	}
	return fields
}
