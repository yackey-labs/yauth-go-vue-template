// Package app is the composition root: it assembles config, the
// database, yauth-go, and the HTTP layer, then exposes Serve / Migrate /
// GenSpec entry points called from main.
package app

import (
	yauth "github.com/yackey-labs/yauth-go"
	"gorm.io/gorm"

	"github.com/yackey-labs/yauth-go-vue-template/server/internal/auth"
	"github.com/yackey-labs/yauth-go-vue-template/server/internal/config"
	"github.com/yackey-labs/yauth-go-vue-template/server/internal/store"
)

// DefaultConfigPath is the location of yauth.yaml relative to wherever
// you invoke the binary from. CI and `task` always run from the repo
// root, so the default is correct there. Override with `-c <path>` if
// you want to ship the binary elsewhere.
const DefaultConfigPath = "yauth.yaml"

// App owns the long-lived dependencies that handlers and services need.
// Tests can construct an App with their own DB / yauth and run the same
// production code paths.
type App struct {
	Cfg   config.Config
	DB    *gorm.DB
	YAuth *yauth.YAuth
}

// New loads config and wires up the database + yauth instance. It does
// NOT run migrations — Serve and Migrate decide when to do that.
func New(cfgPath string) (*App, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}
	db, err := store.Open(cfg.YAuth.Database)
	if err != nil {
		return nil, err
	}
	ya, err := auth.New(db, cfg)
	if err != nil {
		return nil, err
	}
	return &App{Cfg: cfg, DB: db, YAuth: ya}, nil
}
