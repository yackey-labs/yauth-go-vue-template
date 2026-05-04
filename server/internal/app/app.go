// Package app is the composition root: it assembles config, the
// database, yauth-go, and the HTTP layer, then exposes Serve / Migrate /
// GenSpec entry points called from main.
package app

import (
	"context"
	"os"

	yauth "github.com/yackey-labs/yauth-go"
	"gorm.io/gorm"

	"github.com/yackey-labs/yauth-go-vue-template/server/internal/auth"
	"github.com/yackey-labs/yauth-go-vue-template/server/internal/config"
	"github.com/yackey-labs/yauth-go-vue-template/server/internal/store"
	apptel "github.com/yackey-labs/yauth-go-vue-template/server/internal/telemetry"
)

// DefaultConfigPath is the location of yauth.yaml relative to wherever
// you invoke the binary from. CI and `task` always run from the repo
// root, so the default is correct there. Override with `-c <path>` if
// you want to ship the binary elsewhere.
const DefaultConfigPath = "yauth.yaml"

// noopShutdown satisfies the `func(context.Context) error` shape when
// telemetry is disabled. It's installed as the App's TelShutdown so
// callers can always call it.
func noopShutdown(context.Context) error { return nil }

// App owns the long-lived dependencies that handlers and services need.
// Tests can construct an App with their own DB / yauth and run the same
// production code paths.
type App struct {
	Cfg   config.Config
	DB    *gorm.DB
	YAuth *yauth.YAuth

	// TelShutdown drains the OpenTelemetry exporter on graceful exit.
	// Always non-nil — a no-op when telemetry was disabled in
	// yauth.yaml. Serve calls this from its shutdown sequence.
	TelShutdown func(context.Context) error
}

// New wires up the application from yauth.yaml. Order of operations:
//
//  1. Load and validate config.
//  2. Initialize OpenTelemetry (if telemetry.enabled). This MUST run
//     before yauth.New so yauth's plugin handlers, our otelhttp router
//     wrap, and any future tracer.Start() calls all see the right
//     global TracerProvider + propagator.
//  3. Open the database.
//  4. Build the yauth instance.
func New(ctx context.Context, cfgPath string) (*App, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}

	telShutdown := noopShutdown
	if cfg.YAuth.Telemetry.Enabled {
		// telemetry.OTLPEndpoint blanks fall back to the SDK's
		// OTEL_EXPORTER_OTLP_ENDPOINT env var, so the same yauth.yaml
		// ships dev/prod and the env var swaps endpoints.
		endpoint := cfg.YAuth.Telemetry.OTLPEndpoint
		if endpoint == "" {
			endpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		}
		shutdown, err := apptel.Init(ctx, apptel.Config{
			Enabled:     true,
			Endpoint:    endpoint,
			ServiceName: cfg.YAuth.Telemetry.ServiceName,
		})
		if err != nil {
			return nil, err
		}
		telShutdown = shutdown
	}

	db, err := store.Open(cfg.YAuth.Database)
	if err != nil {
		return nil, err
	}

	ya, err := auth.New(db, cfg)
	if err != nil {
		return nil, err
	}

	return &App{Cfg: cfg, DB: db, YAuth: ya, TelShutdown: telShutdown}, nil
}
