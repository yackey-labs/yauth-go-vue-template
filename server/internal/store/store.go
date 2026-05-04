// Package store owns the database connection. yauth-go's gormrepo handles
// all auth-related persistence; add your own repositories alongside as
// the app grows (e.g. store.NewBilling(db), store.NewPosts(db)).
package store

import (
	"context"
	"fmt"

	"github.com/yackey-labs/yauth-go/repo/gormrepo"
	"github.com/yackey-labs/yauth-go/yauthcfg"
	"gorm.io/gorm"
)

// Open connects to the database described by cfg. Currently only Postgres
// is exercised by the template, but yauth-go's gormrepo accepts SQLite
// and MySQL too — flip cfg.Driver in yauth.yaml.
func Open(cfg yauthcfg.DatabaseConfig) (*gorm.DB, error) {
	switch cfg.Driver {
	case "postgres":
		return gormrepo.OpenPostgresSchema(cfg.DSN, cfg.Schema)
	case "sqlite":
		return gormrepo.OpenSQLite(cfg.DSN)
	case "mysql":
		return gormrepo.OpenMySQL(cfg.DSN)
	default:
		return nil, fmt.Errorf("store: unsupported database driver %q", cfg.Driver)
	}
}

// Migrate runs yauth-go's idempotent schema migration. When you add your
// own GORM models, append `db.AutoMigrate(&YourModel{})` calls here so
// `server migrate` migrates the whole schema in one shot.
func Migrate(ctx context.Context, db *gorm.DB) error {
	return gormrepo.Migrate(ctx, db)
}
