package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lazysql/lazysql/pkg/db"
	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	conn *sql.DB
	path string
}

func New() *SQLite {
	return &SQLite{}
}

func (s *SQLite) Connect(dsn string) error {
	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return err
	}
	s.conn = conn
	s.path = dsn
	return nil
}

func (s *SQLite) Disconnect() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func (s *SQLite) Ping() error {
	if s.conn == nil {
		return fmt.Errorf("not connected")
	}
	return s.conn.Ping()
}

func (s *SQLite) ListDatabases() ([]string, error) {
	return []string{s.path}, nil
}

func (s *SQLite) ListTables(_ string) ([]string, error) {
	rows, err := s.conn.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
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

func (s *SQLite) DescribeTable(table string) ([]db.Column, error) {
	escaped := strings.ReplaceAll(table, `"`, `""`)
	rows, err := s.conn.Query(fmt.Sprintf(`PRAGMA table_info("%s")`, escaped))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []db.Column
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt *string
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			return nil, err
		}
		columns = append(columns, db.Column{
			Name:       name,
			Type:       colType,
			Nullable:   notNull == 0,
			PrimaryKey: pk > 0,
			Default:    dflt,
		})
	}
	return columns, rows.Err()
}

func (s *SQLite) Query(query string) (*db.Result, error) {
	start := time.Now()
	rows, err := s.conn.Query(query)
	elapsed := time.Since(start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.ScanRows(rows, elapsed)
}

func (s *SQLite) Execute(query string) (*db.ExecResult, error) {
	start := time.Now()
	result, err := s.conn.Exec(query)
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

func (s *SQLite) DriverName() db.DriverType {
	return db.DriverSQLite
}

func (s *SQLite) CurrentDatabase() string {
	return s.path
}
