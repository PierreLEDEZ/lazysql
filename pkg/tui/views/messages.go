package views

import (
	"github.com/lazysql/lazysql/pkg/config"
	"github.com/lazysql/lazysql/pkg/db"
	"github.com/lazysql/lazysql/pkg/tunnel"

	tea "github.com/charmbracelet/bubbletea"
)

// Connection lifecycle
type ConnectMsg struct {
	Driver db.Driver
	Tunnel *tunnel.Tunnel
	Err    error
}

type DisconnectMsg struct{}

// Table loading
type TablesLoadedMsg struct {
	Tables []string
	Err    error
}

type TableSelectedMsg struct {
	Table string
}

// Structure (describe table)
type StructureUpdatedMsg struct {
	Columns []db.Column
	Err     error
}

type DescribeTableResultMsg struct {
	Table   string
	Columns []db.Column
	Err     error
}

// Query execution
type QueryResultMsg struct {
	Result *db.Result
	Err    error
}

type ExecResultMsg struct {
	Result *db.ExecResult
	Err    error
}

type ExecuteConfirmedMsg struct {
	SQL string
}

// Modal control
type DismissModalMsg struct{}

type RequestConfirmMsg struct {
	Message   string
	OnConfirm tea.Msg
}

type ConfirmResultMsg struct {
	Confirmed bool
	Payload   tea.Msg
}

// Connection management (CRUD)
type OpenConnectionFormMsg struct {
	Mode       string
	Connection *config.Connection
	Index      int
}

type DeleteConnectionMsg struct {
	Index int
}

type ConnectionSavedMsg struct {
	Err error
}

// Saved queries
type SaveQueryMsg struct {
	Name string
	SQL  string
}

type DeleteSavedQueryMsg struct {
	Index int
}

type LoadSavedQueryMsg struct {
	SQL string
}

// Errors
type ErrorMsg struct {
	Err error
}
