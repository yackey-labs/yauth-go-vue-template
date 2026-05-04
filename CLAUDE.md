# CLAUDE.md — yauth-go-vue-template

## What This Is

A working starter for a Go + Vue 3 web app authenticated by
[yauth-go](https://github.com/yackey-labs/yauth-go). Email/password +
session cookies, GORM-backed Postgres, a Vite-based SPA, and a typed
TS client generated from the server's own OpenAPI spec (Huma → orval).

## Layout

```
yauth-go-vue-template/
├── server/                          # Go module
│   ├── main.go                      # subcommand dispatcher
│   └── internal/
│       ├── app/                     # composition root
│       │   ├── app.go               # App: Cfg + DB + YAuth, wired via New()
│       │   ├── serve.go             # HTTP server with graceful shutdown
│       │   ├── migrate.go           # explicit migration entry
│       │   └── genspec.go           # writes OpenAPI JSON for the typed client
│       ├── config/                  # wraps yauthcfg.Load("yauth.yaml")
│       ├── store/                   # DB Open() + Migrate() (yauth + your repos)
│       ├── auth/                    # builds *yauth.YAuth from cfg
│       └── api/
│           ├── api.go               # Huma factory + Spec()
│           ├── router.go            # mounts yauth routes + Huma routes + middleware
│           ├── middleware/          # custom net/http middleware (logging, …)
│           └── handlers/            # typed app handlers, one file each
├── web/
│   ├── src/api/fetcher.ts           # custom fetch (credentials: 'include' + ApiError)
│   ├── src/generated/api.ts         # orval output — committed for hermetic builds
│   ├── orval.config.ts              # generation config
│   └── openapi.json                 # spec snapshot — committed; CI checks freshness
├── yauth.yaml                       # single source of truth for yauth knobs
├── docker-compose.yml               # Local Postgres 17
├── Taskfile.yml                     # taskfile.dev — every entry point
├── task                             # ./task — wraps `go tool task`
└── yauth                            # ./yauth — wraps `go tool yauth` (operator CLI)
```

`internal/` is a Go convention: anything inside is private to this
module. Add new packages there; no external repo can import them.

## Key Commands

Always go through `./task <name>` and `./yauth <subcommand>`. Both
wrap `go tool ...` so contributors don't need global binaries — only
Go ≥ 1.24.

```bash
./task                    # list every task
./task setup              # one-time bootstrap (.env, deps, modules)
./task dev                # Postgres + backend + frontend together
./task migrate            # explicit schema migration (idempotent)
./task gen                # gen-spec + orval → fresh typed TS client
./task ci                 # lint + typecheck + build + test (matches CI)

./yauth status -c yauth.yaml      # validate config + plugins
./yauth check  -c yauth.yaml      # preflight DB schema vs enabled plugins
./yauth dump-schema -c yauth.yaml # CREATE TABLE statements for the live schema
```

Browser URL during dev is **http://localhost:5173** (Vite proxies `/api` →
`:3000`). Don't tell users to hit `:3000` directly.

## Layered architecture

```
┌──────────────────────────────────────────────────────────┐
│                       main.go                            │
│   subcommand dispatcher: serve | migrate | gen-spec      │
└─────────────────────────┬────────────────────────────────┘
                          ▼
┌──────────────────────────────────────────────────────────┐
│              internal/app  (composition root)            │
│  • New(cfgPath) → loads config, opens DB, builds yauth   │
│  • Serve  / Migrate  / GenSpec                           │
└──────┬─────────────────┬───────────────────┬─────────────┘
       │                 │                   │
       ▼                 ▼                   ▼
   internal/config   internal/store      internal/auth
   (yauthcfg.Load)   (gormrepo)          (*yauth.YAuth)
                          │
                          ▼
                   internal/api
                   ├── api.go        Huma factory + Spec()
                   ├── router.go     mounts every route + middleware
                   ├── middleware/   request logging
                   └── handlers/     typed Huma operations
```

**Adding a new app route:**
1. New file in `internal/api/handlers/` with input/output structs and a
   `register*(api huma.API)` function. Reference
   [`me.go`](server/internal/api/handlers/me.go).
2. Add the call to `Register()` in
   [`handlers/handlers.go`](server/internal/api/handlers/handlers.go).
3. If protected, mount under `requireAuth` in
   [`router.go`](server/internal/api/router.go).
4. `./task gen` and commit `web/openapi.json` +
   `web/src/generated/api.ts`.

**Adding a new repository / store:**
- Drop it in `internal/store/` next to `store.go`. Keep the existing
  `Open` / `Migrate` shape. App handlers and services depend on
  store-level types, not on `*gorm.DB` directly, so future swap-outs
  are local.

**Adding a service (business logic):**
- Create `internal/service/<thing>.go`. A service typically holds a
  `*store.Foo` and exposes high-level methods. Handlers depend on
  services, not stores.

## Config: yauth.yaml + yauthcfg

[`yauth.yaml`](yauth.yaml) at the repo root holds every yauth-go knob.
[`internal/config.Load`](server/internal/config/config.go) wraps
`yauthcfg.Load`, which:

1. Reads the YAML
2. Substitutes `env:NAME` placeholders from the process environment
3. Validates the resulting struct

**Don't add a parallel env-loading layer.** If you need a new config
field, prefer adding it to yauth.yaml's schema (or a sibling YAML
section parsed alongside) rather than inventing fresh `os.Getenv`
plumbing.

## Migrations

- `database.auto_migrate: true` in yauth.yaml → server runs
  `gormrepo.Migrate` on startup. Local-only convenience.
- `auto_migrate: false` (production) → run `./task migrate` (or
  `./yauth migrate -c yauth.yaml`) as a separate deploy step.
- CI exercises both: it migrates explicitly via `./task migrate`, then
  boots the server expecting the schema is already there.

## OpenAPI / typed client

The contract:
- **Source of truth** — Huma operation declarations in
  `internal/api/handlers/`. Tags on Go structs (`json`, `format`,
  `enum`, `doc`) flow into the spec.
- **Spec** — `./task gen-spec` writes `web/openapi.json`. No HTTP
  listener needed (the spec is built from declarations alone).
- **Client** — `./task gen` runs gen-spec + orval → typed
  `web/src/generated/api.ts`.
- **Both committed** — CI fails if `./task gen` would change either.

`web/src/api/fetcher.ts` is the orval mutator — add custom headers,
retry logic, telemetry there. The two-arg `apiFetch(url, init)`
signature matches what orval emits for `client: 'fetch'`.

## Observability (OpenTelemetry)

Telemetry is on by default. [`internal/telemetry`](server/internal/telemetry)
initializes the global `TracerProvider` + W3C TraceContext propagator
once in `app.New`. The HTTP layer wraps the mux in `otelhttp.NewHandler`
so every request — yauth plugin handlers and our custom routes —
emits a server span and propagates context. yauth-go's plugins call
`otel.Tracer(...)` on the global provider, so they automatically pick
up our setup; we deliberately do NOT call yauth's `WithTelemetry` /
`WithTelemetryShutdown` on the builder (would only add a duplicate
yauth-only middleware).

**Why a custom telemetry pkg instead of `yauth.telemetry.Init`?** Two
real problems with that helper today:

1. It merges `resource.Default()` (semconv 1.40+ in current SDKs) with
   an explicit semconv 1.26.0 attribute — fails fast with
   `conflicting Schema URL`.
2. It exports via `otlptracegrpc`, but the
   `otel-{local,}.yackey.cloud` collectors only accept OTLP/HTTP at
   `/v1/traces`.

Our local `internal/telemetry`:
- Builds a schemaless resource (one attribute: `service.name`).
- Uses `otlptracehttp` with `WithEndpointURL(cfg.Endpoint)` — the
  exporter appends `/v1/traces` automatically.
- Falls back to the SDK's default endpoint behavior when both
  `yauth.yaml.telemetry.otlp_endpoint` and
  `OTEL_EXPORTER_OTLP_ENDPOINT` are blank.

**Endpoint convention** (matches ghostline's Rust setup):
- dev → `https://otel-local.yackey.cloud`
- prod → `https://otel.yackey.cloud`

Set via `.env` in dev, secret manager in prod. CORS already allows
`traceparent`/`tracestate` headers in `yauth.yaml`.

### Frontend OTel

[`web/src/otel.ts`](web/src/otel.ts) installs `WebTracerProvider` +
`FetchInstrumentation`. Imported FIRST in `main.ts` so every typed
client call is auto-spanned and carries `traceparent`.

**Don't post browser spans cross-origin.** The SPA always exports to
the same-origin path `/v1/traces`. In dev, the Vite proxy forwards it
to the configured collector (see `VITE_OTEL_EXPORTER_OTLP_ENDPOINT`
+ the `proxy` block in `vite.config.ts`). In prod, your SPA host's
reverse proxy must do the same forwarding for `/v1/traces`. This
skips the CORS dance with the collector entirely.

Disable browser-side OTel by leaving `VITE_OTEL_EXPORTER_OTLP_ENDPOINT`
unset — `web/src/otel.ts` becomes a no-op.

## yauth-go API gotchas

Things that bit me writing this template — keep in mind:

- `middleware.AuthUserFromContext(ctx)` returns
  `(*domain.AuthUser, bool)`, not a single value.
- `AuthUser` embeds `User` and `Session`. Email/role/etc. live at
  `user.User.Email`, **not** `user.Email`.
- `user.Method` (string) is the credential class — `cookie`, `bearer`,
  `apikey` — not `user.AuthMethod`.
- `yauthcfg` env-var substitution uses `env:NAME` syntax (e.g.
  `dsn: env:DATABASE_URL`), NOT `${NAME}`.
- `yauth.NewFromConfig` only wires `email_password` + `telemetry`
  today. Admin/status/etc. are TODO upstream — that's why this template
  builds yauth manually in [`internal/auth/auth.go`](server/internal/auth/auth.go)
  with explicit `WithPlugin(...)` calls.

## Vue / yauth-ui-vue

- Components use **callback props**, not Vue emits:
  `<LoginForm :on-success="handler" />`. `@success="..."` silently
  does nothing.
- `useSession()` returns `{ user, loading, isAuthenticated, isLoading,
  isEmailVerified, userRole, userEmail, displayName, refetch, logout }`.
  No `error` field.
- yauth-ui-vue components ship Tailwind class names referencing
  semantic tokens (`text-destructive`, `bg-input`, `ring-ring`, ...).
  `web/src/style.css` maps them via Tailwind v4's `@theme` block —
  keep that in sync if upstream adds new tokens.
- `YAuthPlugin` is installed with `{ baseUrl: '/api/auth' }`. Don't
  switch to a pre-built `client` unless you have a reason; the
  `baseUrl` path is correct in 0.12.2+.

## Conventions

- **Go** — `go vet` clean, `gofmt -l .` must be empty. `./task lint-server`
  enforces both.
- **Vue / TS** — `vp lint` (oxlint, 93 rules) and `vue-tsc --noEmit`.
  `./task lint` + `./task typecheck`. Default `tsconfig.app.json` keeps
  `erasableSyntaxOnly: true` on (works with yauth-ui-vue 0.12.2+).
- **JS package manager** — pnpm only. Never use npm/yarn/bun for
  scripts. `packageManager` is pinned in `web/package.json`.
- **Build tool** — `vp` (vite-plus) for the web side. `pnpm dev` /
  `pnpm build` invoke `vp dev` / `vp build`.
- **Tailwind** — v4 via `@tailwindcss/vite`. The `@theme` block in
  `src/style.css` is the design-tokens contract; don't replace it
  with `tailwind.config.js`.
- **Logging** — `slog` everywhere on the Go side. Don't introduce
  `log.Println` in new code; use `slog.Info` / `slog.Error` with
  structured fields.

## CI

`.github/workflows/ci.yml` runs the same `./task` targets you use locally:

- `server` job: `lint-server`, `build-server`, `test-server`,
  `./yauth status`, `./task migrate`, then curl smoke (register →
  login → `/api/me`) against a Postgres 17 service container.
- `web` job: `lint-web`, `typecheck`, `build-web`, plus a stale check
  that fails if `./task gen` would change committed artifacts.

Both jobs `go mod download` so the Task and yauth tools are fetched
and cached.

## Don't

- Don't add `go run` / `pnpm` invocations directly to docs or scripts —
  add a Taskfile target and reference that.
- Don't drift the typed client. Whenever you change a handler's input
  or output, run `./task gen` and commit the diff.
- Don't bypass the Vite proxy in dev — point fetches at relative
  `/api/...` so cookies stay first-party.
- Don't add a parallel env-loading layer alongside yauthcfg. Extend
  the yaml schema or load adjacent YAML.
- Don't rebuild the dispatcher in main.go; add subcommands by extending
  the switch and adding a `func` to `internal/app/`.
