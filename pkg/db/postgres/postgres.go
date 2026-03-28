package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lazysql/lazysql/pkg/db"
	_ "github.com/lib/pq"
)

type Postgres struct {
	conn     *sql.DB
	database string
}

func New() *Postgres {
	return &Postgres{}
}

func (p *Postgres) Connect(dsn string) error {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return err
	}
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	p.conn = conn

	// Extract current database name
	var dbName string
	if err := conn.QueryRow("SELECT current_database()").Scan(&dbName); err == nil {
		p.database = dbName
	}
	return nil
}

func (p *Postgres) Disconnect() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func (p *Postgres) Ping() error {
	if p.conn == nil {
		return fmt.Errorf("not connected")
	}
	return p.conn.Ping()
}

func (p *Postgres) ListDatabases() ([]string, error) {
	rows, err := p.conn.Query("SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname")
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

func (p *Postgres) ListTables(_ string) ([]string, error) {
	rows, err := p.conn.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name")
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

func (p *Postgres) DescribeTable(table string) ([]db.Column, error) {
	query := `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable,
			CASE WHEN tc.constraint_type = 'PRIMARY KEY' THEN true ELSE false END AS is_pk,
			c.column_default
		FROM information_schema.columns c
		LEFT JOIN information_schema.key_column_usage kcu
			ON c.table_name = kcu.table_name AND c.column_name = kcu.column_name
		LEFT JOIN information_schema.table_constraints tc
			ON kcu.constraint_name = tc.constraint_name AND tc.constraint_type = 'PRIMARY KEY'
		WHERE c.table_name = $1 AND c.table_schema = 'public'
		ORDER BY c.ordinal_position`

	rows, err := p.conn.Query(query, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []db.Column
	for rows.Next() {
		var name, colType, nullable string
		var isPK bool
		var dflt *string
		if err := rows.Scan(&name, &colType, &nullable, &isPK, &dflt); err != nil {
			return nil, err
		}
		columns = append(columns, db.Column{
			Name:       name,
			Type:       colType,
			Nullable:   nullable == "YES",
			PrimaryKey: isPK,
			Default:    dflt,
		})
	}
	return columns, rows.Err()
}

func (p *Postgres) Query(query string) (*db.Result, error) {
	start := time.Now()
	rows, err := p.conn.Query(query)
	elapsed := time.Since(start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.ScanRows(rows, elapsed)
}

func (p *Postgres) Execute(query string) (*db.ExecResult, error) {
	start := time.Now()
	result, err := p.conn.Exec(query)
	elapsed := time.Since(start)
	if err != nil {
		return nil, err
	}

	affected, _ := result.RowsAffected()
	return &db.ExecResult{
		RowsAffected: affected,
		LastInsertID: 0, // PostgreSQL doesn't support LastInsertId via database/sql
		Elapsed:      elapsed,
	}, nil
}

func (p *Postgres) DriverName() db.DriverType {
	return db.DriverPostgres
}

func (p *Postgres) CurrentDatabase() string {
	return p.database
}
