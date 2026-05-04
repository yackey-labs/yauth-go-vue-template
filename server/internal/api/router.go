package api

import (
	"net/http"

	yauth "github.com/yackey-labs/yauth-go"

	appmw "github.com/yackey-labs/yauth-go-vue-template/server/internal/api/middleware"
)

// NewRouter mounts every HTTP route this server exposes.
//
// Layout:
//
//	/api/auth/*    yauth-go's own routes (register, login, session, admin)
//	/api/<route>   App routes — typed via Huma, wrapped by RequireAuth
//	/openapi.json  Application OpenAPI 3.1 spec (drives the generated TS client)
//	/docs          Stoplight Elements UI for /openapi.json
//
// Adding a new app route: declare it in handlers/, then add a
// `mux.Handle(...)` line below for whichever security wrapping it needs.
func NewRouter(ya *yauth.YAuth) http.Handler {
	mux := http.NewServeMux()

	// yauth-managed routes.
	mux.Handle("/api/auth/", http.StripPrefix("/api/auth", ya.Router()))

	// Application routes via Huma. Calling api.New populates humaMux
	// with the actual handlers AND emits OpenAPI operations for them.
	humaMux := http.NewServeMux()
	_ = New(humaMux)

	requireAuth := ya.Middleware().RequireAuth(humaMux)

	// Protected app endpoints — every new RequireAuth route gets a line here.
	mux.Handle("/api/me", requireAuth)

	// OpenAPI spec + interactive docs. Always public.
	mux.Handle("/openapi.json", humaMux)
	mux.Handle("/docs", humaMux)

	return appmw.RequestLogger(mux)
}
