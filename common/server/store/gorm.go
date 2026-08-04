package store

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var gormDB *gorm.DB

type DBType string

const (
	DBTypeSQLite DBType = "sqlite"
	DBTypeMaria  DBType = "mariadb"
	// other db, maybe sqlite
)

// DB returns the GORM database connection
func DB() *gorm.DB {
	return gormDB
}

func InitGormWithConfig(cfg DBConfig) (*gorm.DB, error) {
	switch cfg.Type {
	case DBTypeSQLite:
		return InitSQLite(cfg.Path)
	case DBTypeMaria:
		return InitGormMaria(
			cfg.Host,
			cfg.Port,
			cfg.User,
			cfg.Password,
			cfg.DBName,
			cfg.SSLMode,
		)

	default:
		return nil, fmt.Errorf("unsupported DB_TYPE: %s (use 'sqlite' or 'postgres')", cfg.Type)
	}
}

func InitGormMaria(host string, port int, user string, password string, name string, mode string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		user, password, host, port, name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		//Logger: logger.Default.LogMode(logger.Silent),
		// Use UTC for all auto-generated timestamps (autoCreateTime, autoUpdateTime)
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open MariaDB database: %w", err)
	}

	// Set connection pool for MariaDB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	gormDB = db
	return db, nil
}

// InitSQLite initializes GORM with SQLite
func InitSQLite(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		// Use UTC for all auto-generated timestamps (autoCreateTime, autoUpdateTime)
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Set connection pool for SQLite
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// Enable foreign keys for SQLite
	db.Exec("PRAGMA foreign_keys = ON")
	db.Exec("PRAGMA journal_mode = DELETE")
	db.Exec("PRAGMA synchronous = FULL")
	db.Exec("PRAGMA busy_timeout = 5000")

	gormDB = db
	return db, nil
}
