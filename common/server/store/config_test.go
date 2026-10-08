package store

import (
	"testing"

	"github.com/Crows-Storm/Axis/common/config"
)

func TestDBConfigFromMaria(t *testing.T) {
	got := dbConfigFrom(config.DBConfig{
		DBType:     "mariadb",
		DBHost:     "db",
		DBPort:     3306,
		DBLoginID:  "app",
		DBPassword: "secret",
		DBSchema:   "axis",
		DBSslMode:  "disable",
	})
	if got.Type != DBTypeMaria || got.Host != "db" || got.Port != 3306 || got.User != "app" || got.DBName != "axis" || got.SSLMode != "disable" {
		t.Fatalf("mariadb config = %+v", got)
	}
	if got.Password != "secret" {
		t.Fatal("password was not copied")
	}
}

func TestDBConfigFromDefaultsToSQLite(t *testing.T) {
	got := dbConfigFrom(config.DBConfig{DBType: "sqlite", DBPath: "axis.db"})
	if got.Type != DBTypeSQLite || got.Path != "axis.db" {
		t.Fatalf("sqlite config = %+v", got)
	}
}
