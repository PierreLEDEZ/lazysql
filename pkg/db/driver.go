package db

import "time"

type DriverType string

const (
	DriverPostgres DriverType = "postgres"
	DriverMySQL    DriverType = "mysql"
	DriverSQLite   DriverType = "sqlite"
)

type Driver interface {
	Connect(dsn string) error
	Disconnect() error
	Ping() error

	ListDatabases() ([]string, error)
	ListTables(database string) ([]string, error)
	DescribeTable(table string) ([]Column, error)

	Query(sql string) (*Result, error)
	Execute(sql string) (*ExecResult, error)

	DriverName() DriverType
	CurrentDatabase() string
}

type Column struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
	Default    *string
}

type Result struct {
	Columns  []string
	Rows     [][]interface{}
	Elapsed  time.Duration
	RowCount int
}

type ExecResult struct {
	RowsAffected int64
	LastInsertID int64
	Elapsed      time.Duration
}
