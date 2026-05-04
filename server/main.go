package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	yauth "github.com/yackey-labs/yauth-go"
	"github.com/yackey-labs/yauth-go/middleware"
	"github.com/yackey-labs/yauth-go/openapi"
	"github.com/yackey-labs/yauth-go/plugins/admin"
	"github.com/yackey-labs/yauth-go/plugins/emailpassword"
	"github.com/yackey-labs/yauth-go/plugins/status"
	"github.com/yackey-labs/yauth-go/repo/gormrepo"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dsn := envOr("DATABASE_URL", "postgres://yauth:yauth@127.0.0.1:5432/yauth_app?sslmode=disable")
	port := envOr("PORT", "3000")

	db, err := gormrepo.OpenPostgres(dsn)
	if err != nil {
		return err
	}
	if err := gormrepo.Migrate(context.Background(), db); err != nil {
		return err
	}

	cfg := yauth.NewDefaultConfig()
	cfg.AutoAdminFirstUser = true
	if origins := splitNonEmpty(os.Getenv("CORS_ORIGINS"), ","); len(origins) > 0 {
		cfg.CORS = yauth.CORSConfig{
			AllowedOrigins:   origins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type"},
			AllowCredentials: true,
		}
	}

	epCfg := emailpassword.Config{}
	if os.Getenv("DISABLE_HIBP") == "true" {
		epCfg.HIBPCheck = false
		epCfg.HIBPCheckSet = true
	}

	ya, err := yauth.New(gormrepo.New(db), cfg).
		WithPlugin(emailpassword.New(epCfg)).
		WithPlugin(status.New()).
		WithPlugin(admin.New()).
		Build()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/api/auth/", http.StripPrefix("/api/auth", ya.Router()))
	mux.Handle("/api/me", ya.Middleware().RequireAuth(http.HandlerFunc(meHandler)))
	mux.Handle("/", openapi.YAuth(ya))

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errs := make(chan error, 1)
	go func() {
		log.Printf("yauth-go-vue-template backend listening on :%s", port)
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
		log.Printf("received %s, draining...", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
	if err := ya.Shutdown(ctx); err != nil {
		log.Printf("yauth shutdown: %v", err)
	}
	return nil
}

// meHandler returns the resolved AuthUser as JSON. It demonstrates how to
// protect a custom route with `ya.Middleware().RequireAuth(...)` — anything
// that reaches this handler is already an authenticated user, accessible via
// middleware.AuthUserFromContext.
func meHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":             user.User.ID,
		"email":          user.User.Email,
		"role":           user.User.Role,
		"email_verified": user.User.EmailVerified,
		"auth_method":    user.Method,
	})
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitNonEmpty(s, sep string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, sep)
	out := parts[:0]
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
