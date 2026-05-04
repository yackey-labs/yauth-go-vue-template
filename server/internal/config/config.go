// Package config loads runtime configuration. Almost everything lives in
// yauth.yaml and is consumed by yauth-go via yauthcfg; this layer wraps
// it so the rest of the app has one Load() call to remember.
//
// Add app-specific fields to Config (and to Load()) when you have settings
// that aren't part of the yauth schema (e.g. third-party API keys).
package config

import (
	"github.com/yackey-labs/yauth-go/yauthcfg"
)

type Config struct {
	// YAuth is the parsed yauth.yaml document, with `${VAR}` placeholders
	// resolved from the environment (yauthcfg does this in Load).
	YAuth *yauthcfg.Config
}

// Load reads yauth.yaml from path and resolves YAUTH_* env overrides.
// Production deployments should keep secrets out of the file and inject
// them via env (`${DB_PASSWORD}` placeholders, etc.).
func Load(path string) (Config, error) {
	yc, err := yauthcfg.Load(path)
	if err != nil {
		return Config{}, err
	}
	return Config{YAuth: yc}, nil
}
