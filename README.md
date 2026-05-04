# yauth-go-vue-template

Production-shaped starter for a Go + Vue app authenticated by
[yauth-go](https://github.com/yackey-labs/yauth-go), with email/password
session cookies, GORM-backed Postgres persistence, and a Vue 3 SPA whose
custom routes are typed end-to-end via Huma → OpenAPI → orval.

```
yauth-go-vue-template/
├── server/                    # Go backend (yauth-go + GORM Postgres + Huma)
│   ├── main.go                # subcommand dispatcher (serve | migrate | gen-spec)
│   └── internal/
│       ├── app/               # composition root: New / Serve / Migrate / GenSpec
│       ├── config/            # wraps yauthcfg.Load("yauth.yaml")
│       ├── store/             # DB open + migrations (place for app repos)
│       ├── auth/              # builds *yauth.YAuth from config
│       └── api/               # router + Huma + middleware + handlers/
├── web/                       # Vue 3 SPA (Vite + vp + Tailwind v4)
│   ├── src/api/fetcher.ts     # custom fetch wrapper (credentials: 'include')
│   └── src/generated/api.ts   # orval-generated typed client (committed)
├── yauth.yaml                 # single source of truth for yauth knobs
├── docker-compose.yml         # Local Postgres
├── Taskfile.yml               # taskfile.dev — every entry point
├── task                       # `./task <name>` — wraps `go tool task`
├── yauth                      # `./yauth <subcommand>` — wraps the yauth-go CLI
├── .env.example               # Copy to .env (DATABASE_URL etc.)
└── .github/workflows/ci.yml   # Server (Go) + Web (Vue) jobs, both via ./task
```

## Quick start

```bash
# 1. One-time bootstrap (copies .env, installs JS deps, fetches Go modules)
./task setup

# 2. Bring up Postgres + backend + frontend in one shell
./task dev
```

Open <http://localhost:5173>. **That's the only URL you need in your
browser** — Vite serves the SPA and proxies `/api` to the Go backend at
`http://localhost:3000`. SPA and API share an origin in dev, so session
cookies are first-party (no CORS).

The dashboard demonstrates two ways to talk to the backend:
- **`useSession()`** from `@yackey-labs/yauth-ui-vue` — yauth's own
  reactive session state.
- **`getMe()`** from `web/src/generated/api.ts` — a typed call to your
  application's own protected `/api/me` endpoint, generated from the
  Huma-emitted OpenAPI spec.

## Tooling

`./task` is the single entry point. Both the Task runner and the
yauth-go operator CLI are declared as Go tools in
[`server/go.mod`](server/go.mod), so contributors don't need global
binaries — just Go ≥ 1.24.

| Wrapper          | Backed by                  | Use it for                                            |
| ---------------- | -------------------------- | ----------------------------------------------------- |
| `./task <name>`  | `go tool task`             | dev / build / lint / test / migrate / gen-spec / gen  |
| `./yauth <cmd>`  | `go tool yauth`            | `migrate`, `check`, `status`, `dump-schema`, `gen-secrets`, ... |

### Common task targets

| Target                 | What                                                          |
| ---------------------- | ------------------------------------------------------------- |
| `./task` (no args)     | List every task with descriptions                             |
| `./task setup`         | First-time bootstrap                                          |
| `./task dev`           | Postgres + backend + frontend in one shell                    |
| `./task migrate`       | `server migrate` — idempotent schema migration                |
| `./task schema-check`  | `yauth check` — verify live DB matches enabled plugins        |
| `./task yauth-status`  | `yauth status` — load + validate yauth.yaml                   |
| `./task gen-spec`      | Emit the OpenAPI spec to `web/openapi.json`                   |
| `./task gen`           | gen-spec + run orval to refresh the typed TS client           |
| `./task build`         | Compile server binary + build web bundle                      |
| `./task lint`          | `go vet`, `gofmt`, and `vp lint`                              |
| `./task typecheck`     | `vue-tsc --noEmit` over the web app                           |
| `./task test`          | `go test ./...`                                               |
| `./task ci`            | Same checks CI runs locally — lint + typecheck + build + test |

## How config works

[`yauth.yaml`](yauth.yaml) is the single source of truth for every knob
yauth-go cares about — DB driver/DSN, session/cookie settings, CORS,
plugin config, mailer, telemetry. It's loaded by
[`yauthcfg.Load`](https://pkg.go.dev/github.com/yackey-labs/yauth-go/yauthcfg#Load),
the same function the standalone `yauth` CLI uses, so the running
server, `./task migrate`, and `./yauth check` all read the same config.

Secret values use `env:NAME` placeholders — yauthcfg substitutes the env
var at load time. Example: `database.dsn: env:DATABASE_URL` reads from
your `.env` (copied by `./task setup`) or your platform's secret store
in production. There is no fallback — if the env var isn't set,
yauthcfg fails loudly.

## Migrations

Schema migrations are GORM `AutoMigrate` (idempotent). Two ways to run:

- **Auto on startup** — set `database.auto_migrate: true` in
  `yauth.yaml`. Convenient for `./task dev`.
- **Explicit** — set `auto_migrate: false`, then run `./task migrate`
  (or `./yauth migrate -c yauth.yaml`) before rolling out replicas.
  This is what you want in prod — concurrent AutoMigrate calls across
  booting replicas race.

CI exercises both paths: it runs `./task migrate` as its own step, then
boots the server (which sees the schema already in place) for the
smoke test.

## End-to-end typed API client

Your app's custom routes are typed all the way to the browser:

1. **Server** — declare a route with [Huma](https://huma.rocks) in
   [`internal/api/handlers/`](server/internal/api/handlers). Input/output
   structs become OpenAPI schemas; the operation metadata becomes a path.
   See [`me.go`](server/internal/api/handlers/me.go) for the pattern.
2. **Spec** — `./task gen-spec` writes `web/openapi.json`. No HTTP
   listener needed; the spec is built from the handler declarations.
3. **Client** — `./task gen` runs gen-spec + [orval](https://orval.dev),
   which writes `web/src/generated/api.ts`. Every operation surfaces as
   an exported function with typed input + output.
4. **Browser** — import from `./generated/api` and call. The custom
   fetcher in [`src/api/fetcher.ts`](web/src/api/fetcher.ts) adds
   `credentials: 'include'` so the yauth session cookie travels along.

CI runs `./task gen` and fails if anything diffs — server-side handler
changes must ship with a regenerated client.

### Adding a new route

1. Drop a new file under `server/internal/api/handlers/` defining the
   input/output structs and a `register*(api huma.API)` function.
2. Call it from `Register()` in
   [`handlers/handlers.go`](server/internal/api/handlers/handlers.go).
3. If the route is protected, add a `mux.Handle("/api/<path>", requireAuth)`
   line to [`router.go`](server/internal/api/router.go).
4. Run `./task gen`, commit the regenerated `web/openapi.json` and
   `web/src/generated/api.ts`, and use the typed function in your Vue
   components.

## CORS for split origins

The Vite proxy means CORS doesn't apply in dev. If you deploy the SPA
on a different origin than the backend, set
`server.cors.allowed_origins` in `yauth.yaml`:

```yaml
server:
  cors:
    allowed_origins: ["https://app.example.com"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["Content-Type"]
    allow_credentials: true   # required — session cookies are credentials
```

`allow_credentials: true` is essential — without it the browser refuses
to include cookies on cross-origin requests, breaking session auth.

## Deployment notes

- Configure `DATABASE_URL` (referenced by `yauth.yaml` via
  `env:DATABASE_URL`).
- Set `session.cookie_secure: true` in `yauth.yaml`.
- Set `database.auto_migrate: false` and run `./task migrate` (or
  `./yauth migrate -c yauth.yaml`) as a separate step in your deploy
  pipeline before rolling out new replicas.
- Populate `server.cors.allowed_origins` if the SPA isn't served from
  the same origin as the backend.

## CI

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs the same
`./task` targets you use locally:

- **Server (Go)** — `lint-server`, `build-server`, `test-server`,
  `./yauth status`, `./task migrate`, plus a curl smoke (register →
  login → `/api/me`) against a Postgres 17 service container.
- **Web (Vue)** — `lint-web`, `typecheck`, `build-web`, plus a stale
  check that fails if `./task gen` would change the committed
  `web/openapi.json` or `web/src/generated`.

## License

MIT
