package store

import (
	"database/sql"
	"fmt"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"gorm.io/gorm"
)

type Store struct {
	gdb    *gorm.DB  // GORM database connection
	db     *sql.DB   // Legacy sql.DB for backward compatibility
	driver *DBDriver // Database driver for abstraction (legacy)
}

func NewWithConfig(cfg DBConfig) (*Store, error) {
	gdb, err := InitGormWithConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Get underlying sql.DB for legacy compatibility
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	s := &Store{gdb: gdb, db: sqlDB}

	// Here can do some initialization of table or data, but there's not such demand now

	dbTypeStr := "SQLite"
	if cfg.Type == DBTypeMaria {
		dbTypeStr = "Mariadb"
	}
	logger.Infof("✅ Database initialized (GORM, %s)", dbTypeStr)
	return s, nil
}

func (s *Store) DB() *gorm.DB {
	return s.gdb
}

// Close closes database connection
func (s *Store) Close() error {
	if s.driver != nil {
		return s.driver.Close()
	}
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Store) Driver() *DBDriver {
	return s.driver
}

func (s *Store) Transaction(fn func(tx *gorm.DB) error) error {
	return s.gdb.Transaction(fn)
}
