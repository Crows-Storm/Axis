package store

import "database/sql"

type DBConfig struct {
	Type     DBType // sqlite or mariadb
	Path     string // SQLite file path (for sqlite)
	Host     string // mariadb host (for postgres)
	Port     int    // mariadb port (for postgres)
	User     string // mariadb user (for postgres)
	Password string // mariadb password (for postgres)
	DBName   string // mariadb database name (for postgres)
	SSLMode  string // mariadb SSL mode (for postgres)
}

// DBDriver database driver abstraction
type DBDriver struct {
	Type DBType
	db   *sql.DB
}

func (d *DBDriver) DB() *sql.DB {
	return d.db
}

// Close closes database connection
func (d *DBDriver) Close() error {
	return d.db.Close()
}
