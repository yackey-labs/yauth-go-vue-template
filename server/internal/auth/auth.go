// Package auth wires yauth-go from the loaded yauth.yaml config. We
// build the YAuth instance manually rather than calling yauth.NewFromConfig
// so we can attach extra plugins (status, admin) on top of email-password
// — NewFromConfig only wires email-password and telemetry today.
package auth

import (
	yauth "github.com/yackey-labs/yauth-go"
	"github.com/yackey-labs/yauth-go/auth/passwordpolicy"
	"github.com/yackey-labs/yauth-go/plugins/admin"
	"github.com/yackey-labs/yauth-go/plugins/emailpassword"
	"github.com/yackey-labs/yauth-go/plugins/status"
	"github.com/yackey-labs/yauth-go/repo/gormrepo"
	"github.com/yackey-labs/yauth-go/yauthcfg"
	"gorm.io/gorm"

	"github.com/yackey-labs/yauth-go-vue-template/server/internal/config"
)

// New builds a configured *yauth.YAuth from the application config and an
// open GORM connection. Add WithPlugin(...) calls below to opt into more
// auth methods (bearer JWT, API keys, MFA, ...).
func New(db *gorm.DB, cfg config.Config) (*yauth.YAuth, error) {
	yc := cfg.YAuth
	ycfg := yauthRuntime(yc)

	epCfg := emailPassword(yc.Plugins.EmailPassword)

	return yauth.New(gormrepo.New(db), ycfg).
		WithPlugin(emailpassword.New(epCfg)).
		WithPlugin(status.New()).
		WithPlugin(admin.New()).
		Build()
}

// yauthRuntime translates the parsed yauth.yaml into yauth's runtime
// config struct. Only the knobs we expect callers to flip are mapped;
// the rest stay at NewDefaultConfig values.
func yauthRuntime(yc *yauthcfg.Config) yauth.YAuthConfig {
	out := yauth.NewDefaultConfig()

	if yc.Session.TTL > 0 {
		out.SessionTTL = yc.Session.TTL
	}
	if yc.Session.CookieName != "" {
		out.CookieName = yc.Session.CookieName
	}
	if yc.Session.CookiePath != "" {
		out.CookiePath = yc.Session.CookiePath
	}
	out.CookieDomain = yc.Session.CookieDomain
	out.CookieSecure = yc.Session.CookieSecure
	if yc.Session.CookieSameSite != "" {
		out.CookieSameSite = yc.Session.CookieSameSite
	}

	out.BaseURL = yc.Server.BaseURL
	if yc.Server.AllowSignups != nil {
		out.AllowSignups = *yc.Server.AllowSignups
	}
	out.AutoAdminFirstUser = yc.Server.AutoAdminFirstUser
	out.CORS = yauth.CORSConfig{
		AllowedOrigins:   yc.Server.CORS.AllowedOrigins,
		AllowedMethods:   yc.Server.CORS.AllowedMethods,
		AllowedHeaders:   yc.Server.CORS.AllowedHeaders,
		AllowCredentials: yc.Server.CORS.AllowCredentials,
		MaxAge:           yc.Server.CORS.MaxAge,
	}

	return out
}

func emailPassword(ep yauthcfg.EmailPasswordPluginConfig) emailpassword.Config {
	cfg := emailpassword.Config{
		MinPasswordLength:        ep.MinPasswordLength,
		RequireEmailVerification: ep.RequireEmailVerification,
		PasswordPolicy: passwordpolicy.Policy{
			MinLength:      ep.PasswordPolicy.MinLength,
			MaxLength:      ep.PasswordPolicy.MaxLength,
			RequireUpper:   ep.PasswordPolicy.RequireUpper,
			RequireLower:   ep.PasswordPolicy.RequireLower,
			RequireDigit:   ep.PasswordPolicy.RequireDigit,
			RequireSpecial: ep.PasswordPolicy.RequireSpecial,
			DisallowCommon: ep.PasswordPolicy.DisallowCommon,
			HistoryCount:   ep.PasswordPolicy.HistoryCount,
		},
	}
	if ep.HIBPCheck != nil {
		cfg.HIBPCheck = *ep.HIBPCheck
		cfg.HIBPCheckSet = true
	}
	return cfg
}
