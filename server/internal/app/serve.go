package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yackey-labs/yauth-go-vue-template/server/internal/api"
	"github.com/yackey-labs/yauth-go-vue-template/server/internal/store"
)

// Serve loads config, optionally runs migrations, then starts the HTTP
// server with graceful SIGINT/SIGTERM shutdown.
//
// Migration policy: yauth.yaml's `database.auto_migrate: true` flips on
// startup migrations (convenient for local dev). In prod, leave it
// false and run `server migrate` (or `task migrate`) as a separate step
// before rolling out replicas.
func Serve(ctx context.Context, cfgPath string) error {
	a, err := New(cfgPath)
	if err != nil {
		return err
	}

	if a.Cfg.YAuth.Database.AutoMigrate {
		slog.Info("auto_migrate=true, migrating before serve")
		if err := store.Migrate(ctx, a.DB); err != nil {
			return err
		}
	}

	addr := a.Cfg.YAuth.Server.Addr
	if addr == "" {
		addr = ":3000"
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(a.YAuth),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errs := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
		close(errs)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errs:
		return err
	case sig := <-stop:
		slog.Info("draining", "signal", sig)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("http shutdown", "err", err)
	}
	if err := a.YAuth.Shutdown(shutdownCtx); err != nil {
		slog.Error("yauth shutdown", "err", err)
	}
	return nil
}
