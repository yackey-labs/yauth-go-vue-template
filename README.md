# yauth-go-vue-template

Production-shaped starter for a Go + Vue app authenticated by
[yauth-go](https://github.com/yackey-labs/yauth-go), with email/password
login, session cookies, GORM-backed Postgres persistence, and a typed
Vue 3 frontend driven by the published `@yackey-labs/yauth-ui-vue`
components.

```
yauth-go-vue-template/
├── server/                    # Go backend (yauth-go + GORM Postgres)
├── web/                       # Vue 3 SPA (Vite + vp + Tailwind v4)
├── docker-compose.yml         # Local Postgres
├── Taskfile.yml               # Convenience targets (taskfile.dev)
├── task                       # `./task <name>` — wraps `go tool task`
├── .env.example               # Copy to .env
└── .github/workflows/ci.yml   # Lint + build + test for both halves
```

## Quick start

```bash
# 1. One-time bootstrap (copies .env, installs JS deps, fetches Go modules)
./task setup

# 2. Bring up Postgres + backend + frontend in one shell
./task dev
```

Open <http://localhost:5173>. **That's the only URL you need in your
browser** — Vite serves the SPA there and proxies `/api` to the Go
backend at `http://localhost:3000`, so the SPA and the API share an
origin in dev (no CORS, session cookies Just Work).

Register, log in, watch the dashboard populate from
`GET /api/auth/session`. The dashboard also calls a demo `GET /api/me`
route to show how to protect your own handlers with
`ya.Middleware().RequireAuth(...)`.

To stop, hit Ctrl-C (`./task dev` cleans up both processes), and
`./task db-down` if you want to stop Postgres too.

## Task runner

`./task` is a tiny wrapper that runs Task ([taskfile.dev](https://taskfile.dev))
via the `tool` directive in [`server/go.mod`](server/go.mod), so you don't
need a global `task` binary on PATH — just Go ≥ 1.24. If you do have
Task installed (`brew install go-task`), you can use `task <name>`
directly instead.

| Target               | What                                                      |
| -------------------- | --------------------------------------------------------- |
| `./task` (no args)   | List all tasks with descriptions                          |
| `./task setup`       | First-time bootstrap                                      |
| `./task dev`         | Postgres + backend + frontend in one shell                |
| `./task server`      | Backend only (foreground)                                 |
| `./task web`         | Frontend only (foreground)                                |
| `./task migrate`     | Run schema migrations and exit (idempotent)               |
| `./task db-up`       | Start the Postgres container                              |
| `./task db-down`     | Stop the Postgres container                               |
| `./task db-reset`    | Drop the Postgres volume and start it again               |
| `./task build`       | Compile the server binary + build the web bundle          |
| `./task lint`        | `go vet`, `gofmt -l` (must be empty), and `vp lint`       |
| `./task typecheck`   | `vue-tsc --noEmit` over the web app                       |
| `./task test`        | `go test ./...`                                           |
| `./task ci`          | Same checks CI runs locally — lint + typecheck + build + test |
| `./task clean`       | Remove build outputs and Vite caches                      |

### What's running while `./task dev` is up

| Process              | Port        | What                                                            |
| -------------------- | ----------- | --------------------------------------------------------------- |
| Postgres             | `:5432`     | `docker compose up -d postgres` — DSN in `.env` `DATABASE_URL`  |
| Go backend           | `:3000`     | `yauth-go` at `/api/auth/*`, OpenAPI UI at `/docs`              |
| Vite dev server      | `:5173`     | Serves the SPA, proxies `/api` → `:3000` (the only URL you hit) |

## Project layout

### `server/` — Go backend

- [`main.go`](server/main.go) — env-driven config, builds the yauth
  router, mounts `/api/auth/*`, `/api/me` (protected demo), and the
  OpenAPI UI at `/docs`. Graceful shutdown on SIGTERM/SIGINT.
- Subcommands: `serve` (default — runs the HTTP server) and `migrate`
  (runs schema migrations and exits).
- Plugins: `email-password`, `status`, `admin`. Bearer-JWT and API-key
  plugins are commented out — uncomment to opt in.
- Repo: `gormrepo.OpenPostgres(DATABASE_URL)`.
- Two ways to run schema migrations:
  - **Auto, on startup** — default. Convenient for `./task dev`. Set
    `AUTO_MIGRATE=false` to disable.
  - **Explicit** — `./task migrate` (or `./server migrate`).
    Idempotent. Use this in CI/CD before rolling out a new replica set
    so two booting replicas don't race the migration.
- The first user registered is auto-promoted to `admin`
  (`AutoAdminFirstUser: true` in config), so you can hit
  `/api/auth/admin/users` without manual setup.

### `web/` — Vue 3 SPA

- [`vite.config.ts`](web/vite.config.ts) — proxies `/api` to
  `http://localhost:3000`, includes `@tailwindcss/vite` so the yauth-ui
  components render styled out of the box.
- [`src/main.ts`](web/src/main.ts) — installs `YAuthPlugin` with
  `{ baseUrl: '/api/auth' }` and the router.
- [`src/views/`](web/src/views/) — `LoginView`, `RegisterView`,
  `DashboardView`. The dashboard demonstrates `useSession()` (user +
  loading + logout) and a manual
  `fetch('/api/me', { credentials: 'include' })` call against the
  protected backend route.

## CORS for split origins

The Vite proxy means CORS doesn't apply in dev — the browser sees one
origin (`localhost:5173`) and Vite forwards `/api` server-side, so
cookies are first-party. If you deploy the SPA to a different origin
than the backend, configure `YAuthConfig.CORS` in
[`server/main.go`](server/main.go):

```go
cfg := yauth.NewDefaultConfig()
cfg.CORS = yauth.CORSConfig{
    AllowedOrigins:   []string{"https://app.example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Content-Type"},
    AllowCredentials: true, // required — session cookies are credentials
}
```

`AllowCredentials: true` is essential — without it the browser refuses
to include cookies on cross-origin requests, which breaks session auth.
The starter already wires `CORS_ORIGINS=…,…` from `.env` into this
block, so the production path is "just set the env var."

## Deployment notes

- Set `DATABASE_URL`, `PORT`, and (if needed) `CORS_ORIGINS`.
- Set `AUTO_MIGRATE=false` and run `./server migrate` (or
  `./task migrate`) as a separate step in your deploy pipeline before
  rolling out new replicas.
- The cookie defaults are `Secure: false` for dev. In production set
  `cfg.Cookie.Secure = true` (or use `yauthcfg` to load the cookie
  block from YAML).
- For Postgres, point at a managed service. The local `docker-compose`
  service is for development only.

## CI

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs on every
push and PR. It uses the same `./task` targets as your local shell:

- **Server (Go)** — `./task lint-server` + `./task build-server` +
  `./task test-server` + `./task migrate` + a curl smoke test
  (register → login → `/api/me`) against a Postgres 17 service
  container.
- **Web (Vue)** — `./task lint-web` + `./task typecheck` +
  `./task build-web`, with the pnpm store cached across runs.

## License

MIT
