package api

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	"github.com/yackey-labs/yauth-go-vue-template/server/internal/api/handlers"
)

// humaConfig defines the OpenAPI document the rest of the world sees.
// Update Title/Version/Description here when shaping your public API.
func humaConfig() huma.Config {
	cfg := huma.DefaultConfig("yauth-go-vue-template API", "0.1.0")
	cfg.Info.Description = "Application HTTP surface for the " +
		"yauth-go + Vue starter. yauth's own auth/session/admin endpoints " +
		"live under /api/auth/* and are documented separately via the " +
		"@yackey-labs/yauth-client npm package."
	cfg.Info.License = &huma.License{Name: "MIT", Identifier: "MIT"}
	// Document session-cookie security so generated clients know how
	// authenticated routes expect to be called.
	if cfg.Components == nil {
		cfg.Components = &huma.Components{}
	}
	if cfg.Components.SecuritySchemes == nil {
		cfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{}
	}
	cfg.Components.SecuritySchemes["sessionCookie"] = &huma.SecurityScheme{
		Type: "apiKey",
		In:   "cookie",
		Name: "yauth_session",
	}
	return cfg
}

// New builds a Huma API on the given mux. All handlers are registered
// onto mux at this point; the returned huma.API is kept around for the
// `gen-spec` subcommand which serializes the OpenAPI document.
func New(mux *http.ServeMux) huma.API {
	api := humago.New(mux, humaConfig())
	handlers.Register(api)
	return api
}

// Spec builds the API on a throwaway mux and returns the OpenAPI doc.
// `server gen-spec` calls this — no HTTP listener is started.
func Spec() *huma.OpenAPI {
	return New(http.NewServeMux()).OpenAPI()
}
