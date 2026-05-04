// Package middleware contains HTTP middleware specific to this app.
// yauth-go's middleware (RequireAuth/RequireAdmin/etc.) lives in its
// own package and is imported as needed.
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// RequestLogger logs every request with method, path, status code, and
// duration via slog. It's the outermost wrap — if you add tracing or
// request IDs later, do that here.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}
