package dbconnect

import (
	"fmt"

	"github.com/lazysql/lazysql/pkg/config"
	"github.com/lazysql/lazysql/pkg/db"
	"github.com/lazysql/lazysql/pkg/db/mysql"
	"github.com/lazysql/lazysql/pkg/db/postgres"
	"github.com/lazysql/lazysql/pkg/db/sqlite"
	"github.com/lazysql/lazysql/pkg/tunnel"
)

func NewDriver(t db.DriverType) (db.Driver, error) {
	switch t {
	case db.DriverPostgres:
		return postgres.New(), nil
	case db.DriverMySQL:
		return mysql.New(), nil
	case db.DriverSQLite:
		return sqlite.New(), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %s", t)
	}
}

func Connect(conn config.Connection) (db.Driver, *tunnel.Tunnel, error) {
	var tun *tunnel.Tunnel

	host := conn.Host
	port := conn.Port

	if conn.SSH != nil {
		tun = tunnel.New(*conn.SSH, conn.Host, conn.Port)
		localPort, err := tun.Start()
		if err != nil {
			return nil, nil, fmt.Errorf("SSH tunnel: %w", err)
		}
		host = "127.0.0.1"
		port = localPort
	}

	driver, err := NewDriver(db.DriverType(conn.Driver))
	if err != nil {
		if tun != nil {
			tun.Stop()
		}
		return nil, nil, err
	}

	dsn := buildDSN(conn.Driver, host, port, conn.User, conn.Password, conn.Database, conn.Path)
	if err := driver.Connect(dsn); err != nil {
		if tun != nil {
			tun.Stop()
		}
		return nil, nil, fmt.Errorf("connect: %w", err)
	}

	return driver, tun, nil
}

func buildDSN(driver, host string, port int, user, password, database, path string) string {
	switch driver {
	case "postgres":
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, database)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, database)
	case "sqlite":
		return path
	default:
		return ""
	}
}
