package app

import (
	"context"
	"log/slog"

	yauth "github.com/yackey-labs/yauth-go"

	"github.com/yackey-labs/yauth-go-vue-template/server/internal/config"
)

// Migrate runs the yauth-go schema migration (idempotent) against the
// database described in yauth.yaml and returns. It uses the same code
// path as the standalone `yauth migrate` CLI.
//
// Use this before rolling out a new replica set in prod, paired with
// `database.auto_migrate: false` in yauth.yaml so booting replicas
// don't race the migration.
func Migrate(ctx context.Context, cfgPath string) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	slog.Info("running migrations", "driver", cfg.YAuth.Database.Driver)
	if err := yauth.Migrate(ctx, cfg.YAuth); err != nil {
		return err
	}
	slog.Info("migrations complete")
	return nil
}
