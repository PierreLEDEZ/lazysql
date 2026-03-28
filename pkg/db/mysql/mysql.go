package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lazysql/lazysql/pkg/db"
	_ "github.com/go-sql-driver/mysql"
)

type MySQL struct {
	conn     *sql.DB
	database string
}

func New() *MySQL {
	return &MySQL{}
}

func (m *MySQL) Connect(dsn string) error {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return err
	}
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	m.conn = conn

	// Extract current database name
	var dbName string
	if err := conn.QueryRow("SELECT DATABASE()").Scan(&dbName); err == nil {
		m.database = dbName
	}
	return nil
}

func (m *MySQL) Disconnect() error {
	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}

func (m *MySQL) Ping() error {
	if m.conn == nil {
		return fmt.Errorf("not connected")
	}
	return m.conn.Ping()
}

func (m *MySQL) ListDatabases() ([]string, error) {
	rows, err := m.conn.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		databases = append(databases, name)
	}
	return databases, rows.Err()
}

func (m *MySQL) ListTables(_ string) ([]string, error) {
	rows, err := m.conn.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

func (m *MySQL) DescribeTable(table string) ([]db.Column, error) {
	query := `SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_DEFAULT
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`

	rows, err := m.conn.Query(query, m.database, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []db.Column
	for rows.Next() {
		var name, colType, nullable, key string
		var dflt *string
		if err := rows.Scan(&name, &colType, &nullable, &key, &dflt); err != nil {
			return nil, err
		}
		columns = append(columns, db.Column{
			Name:       name,
			Type:       colType,
			Nullable:   nullable == "YES",
			PrimaryKey: key == "PRI",
			Default:    dflt,
		})
	}
	return columns, rows.Err()
}

func (m *MySQL) Query(query string) (*db.Result, error) {
	start := time.Now()
	rows, err := m.conn.Query(query)
	elapsed := time.Since(start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.ScanRows(rows, elapsed)
}

func (m *MySQL) Execute(query string) (*db.ExecResult, error) {
	start := time.Now()
	result, err := m.conn.Exec(query)
	elapsed := time.Since(start)
	if err != nil {
		return nil, err
	}

	affected, _ := result.RowsAffected()
	lastID, _ := result.LastInsertId()
	return &db.ExecResult{
		RowsAffected: affected,
		LastInsertID: lastID,
		Elapsed:      elapsed,
	}, nil
}

func (m *MySQL) DriverName() db.DriverType {
	return db.DriverMySQL
}

func (m *MySQL) CurrentDatabase() string {
	return m.database
}
