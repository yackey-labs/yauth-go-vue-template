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
├── Makefile                   # Convenience targets
├── .env.example               # Copy to .env
└── .github/workflows/ci.yml   # Lint + build + test for both halves
```

## Quick start

```bash
# 1. Bootstrap (copies .env, fetches modules, installs JS deps)
make setup

# 2. Bring up Postgres + backend + frontend in one shell
make dev
```

Open <http://localhost:5173>. Register, log in, watch the dashboard
populate from `GET /api/auth/session`. The dashboard also calls a
demo `GET /api/me` route to show how to protect your own handlers
with `ya.Middleware().RequireAuth(...)`.

To stop, hit Ctrl-C (`make dev` cleans up both processes), and
`make db-down` if you want to stop Postgres too.

### What's running

| Process              | Where                | What                                                                     |
| -------------------- | -------------------- | ------------------------------------------------------------------------ |
| Postgres             | `localhost:5432`     | `docker compose up -d postgres` — DSN in `.env` `DATABASE_URL`          |
| Go backend           | `localhost:3000`     | `make server` — yauth at `/api/auth/*`, OpenAPI docs at `/docs`         |
| Vite dev server      | `localhost:5173`     | `make web` — proxies `/api` to the backend                              |

The Vite proxy means the frontend and backend share an origin in dev,
so cross-origin/cookie issues don't bite. For a deployed setup that
serves the SPA from a different origin, see "CORS for split origins"
below.

## Project layout

### `server/` — Go backend

- [`main.go`](server/main.go) — env-driven config, builds the yauth
  router, mounts `/api/auth/*`, `/api/me` (protected demo), and the
  OpenAPI UI at `/docs`. Graceful shutdown on SIGTERM/SIGINT.
- Plugins: `email-password`, `status`, `admin`. Bearer-JWT and API-key
  plugins are commented out — uncomment to opt in.
- Repo: `gormrepo.OpenPostgres(DATABASE_URL)` + `gormrepo.Migrate`.
  Schema migrations run on startup (idempotent via GORM AutoMigrate).
- The first user registered is auto-promoted to `admin`
  (`AutoAdminFirstUser: true` in config) so you can hit
  `/api/auth/admin/users` without manual setup.

### `web/` — Vue 3 SPA

- [`vite.config.ts`](web/vite.config.ts) — proxies `/api` to
  `http://localhost:3000`, includes `@tailwindcss/vite` so the yauth-ui
  components render styled out of the box.
- [`src/main.ts`](web/src/main.ts) — installs `YAuthPlugin` with
  `{ baseUrl: '/api/auth' }` and the router.
- [`src/views/`](web/src/views/) — `LoginView`, `RegisterView`,
  `DashboardView`. The dashboard demonstrates `useSession()` (user +
  loading + logout) and a manual `fetch('/api/me', { credentials: 'include' })`
  call.

### `Makefile`

| Target          | What                                              |
| --------------- | ------------------------------------------------- |
| `make setup`    | First-time bootstrap                              |
| `make db-up`    | Start Postgres                                    |
| `make db-down`  | Stop Postgres                                     |
| `make db-reset` | Drop the Postgres volume + restart                |
| `make dev`      | Postgres + backend + frontend in one shell        |
| `make server`   | Backend only                                      |
| `make web`      | Frontend only                                     |
| `make build`    | Compile server + build web bundle                 |
| `make lint`     | Both `go vet` + `gofmt -l` and `vp lint`          |
| `make test`     | `go test ./...` (web has no tests by default)     |
| `make clean`    | Remove build outputs                              |

## CORS for split origins

The Vite proxy means CORS doesn't apply in dev. If you deploy the SPA
to a different origin than the backend, configure `YAuthConfig.CORS`
in `server/main.go`:

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

## Deployment notes

- Set `DATABASE_URL`, `PORT`, and (if needed) `CORS_ORIGINS`.
- The cookie defaults are `Secure: false` for dev. In production set
  `cfg.Cookie.Secure = true` (or use `yauthcfg` to load the cookie
  block from YAML).
- For Postgres, point at a managed service. The local `docker-compose`
  service is for development only.

## CI

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs on every
push and PR:

- **Server** — `go vet`, `gofmt -l` (must be empty), `go build`, and
  `go test ./...` against a Postgres 17 service container.
- **Web** — `pnpm install --frozen-lockfile`, `pnpm typecheck`,
  `pnpm lint`, `pnpm build`.

## License

MIT
