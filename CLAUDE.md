# CLAUDE.md — yauth-go-vue-template

## What This Is

A working starter for a Go + Vue 3 web app authenticated by
[yauth-go](https://github.com/yackey-labs/yauth-go). Email/password +
session cookies, GORM-backed Postgres, and a Vite-based SPA that uses
the published `@yackey-labs/yauth-ui-vue` components.

Two halves under one repo:

| Path                | Stack                                                                          |
| ------------------- | ------------------------------------------------------------------------------ |
| `server/`           | Go module — `yauth-go` + `gormrepo` (Postgres) + `email-password` + `status` + `admin` plugins. Subcommands: `serve` (default), `migrate`. |
| `web/`              | Vue 3 + TS + vue-router + Vite 8 + Tailwind v4 (`@tailwindcss/vite`). pnpm + vp ([vite-plus](https://viteplus.dev)). Consumes `@yackey-labs/yauth-{client,ui-vue}@^0.12.2`. |
| `docker-compose.yml`| Postgres 17-alpine on `:5432` with healthcheck.                                |
| `Taskfile.yml`      | Single source of truth for setup/dev/build/lint/test/CI commands.              |
| `task`              | Tiny shell wrapper that runs Task via `go tool task` (declared in `server/go.mod`). |
| `.github/workflows/ci.yml` | Server (Go) + Web (Vue) jobs, both driven by `./task` targets.          |

## Key Commands

Always use `./task <name>` (or `task <name>` if Task is installed
globally). Never use `go run` / `pnpm dev` directly in docs or
suggestions — the Taskfile is authoritative.

```bash
./task                    # list every task
./task setup              # one-time bootstrap (.env, deps, modules)
./task dev                # Postgres + backend + frontend together
./task migrate            # explicit schema migration (idempotent)
./task ci                 # lint + typecheck + build + test (matches CI)
```

`./task dev` brings up the only URL anyone should hit in a browser:
**`http://localhost:5173`**. Vite serves the SPA there and proxies
`/api` to the Go backend on `:3000`. Same-origin in dev → no CORS,
session cookies are first-party.

## Local Architecture

```
                 ┌──────────────────────┐
 Browser ──────► │ Vite dev (:5173)     │
                 │ • serves SPA         │
                 │ • proxy /api → :3000 │
                 └──────────┬───────────┘
                            │ /api/*
                            ▼
                 ┌──────────────────────┐    DSN     ┌────────────────┐
                 │ Go server (:3000)    │ ─────────► │ Postgres :5432 │
                 │ • /api/auth/* yauth  │            │ (docker)       │
                 │ • /api/me protected  │            └────────────────┘
                 │ • /openapi.json /docs│
                 └──────────────────────┘
```

## Migrations

Schema migrations are GORM `AutoMigrate` (idempotent). Two ways to run them:

- **Auto on startup** (default) — convenient for `./task dev`. Set
  `AUTO_MIGRATE=false` to disable.
- **Explicit** — `./task migrate` (which runs `server migrate`). Use
  this in CI/CD before rolling out new replicas so two booting
  instances don't race the migration.

CI exercises both paths: it calls `./task migrate` as its own step,
then boots the server with `AUTO_MIGRATE=false` for the smoke test to
prove the explicit path works.

## yauth-go API Notes

Things that bit me writing this template — keep in mind:

- `middleware.AuthUserFromContext(ctx)` returns `(*domain.AuthUser, bool)`,
  not a single value. Always destructure both.
- `AuthUser` embeds `User` and `Session`. Email/role/etc. live at
  `user.User.Email`, **not** `user.Email`.
- `user.Method` (string) is the credential class — `cookie`, `bearer`,
  `apikey` — not `user.AuthMethod`.
- Mount with `http.StripPrefix("/api/auth", ya.Router())` if you want
  routes under `/api/auth/...`. Without `StripPrefix`, the router
  expects them at the root.
- The published `@yackey-labs/yauth-ui-vue@0.12.0` is broken (lazy
  `import()` race + envelope-not-unwrapped + workspace:* leaks). Use
  `^0.12.2`.

## Vue / yauth-ui-vue Notes

- Components use **callback props**, not Vue emits:
  `<LoginForm :on-success="handler" />` — `@success="..."` silently
  does nothing.
- `useSession()` returns `{ user, loading, isAuthenticated, isLoading,
  isEmailVerified, userRole, userEmail, displayName, refetch, logout }`.
  No `error` field.
- The yauth-ui-vue components ship Tailwind class names referencing
  semantic tokens (`text-destructive`, `bg-input`, `ring-ring`, ...).
  `web/src/style.css` maps those to concrete colors via Tailwind v4's
  `@theme` block — keep that in sync if the components add new tokens.
- `YAuthPlugin` is installed with `{ baseUrl: '/api/auth' }`. Don't
  switch to a pre-built `client` unless you have a reason; the
  `baseUrl` path is correct in 0.12.2+.

## Config

All runtime config goes through env vars (read in `server/main.go`):

| Env             | Default                                                                  | What                                              |
| --------------- | ------------------------------------------------------------------------ | ------------------------------------------------- |
| `DATABASE_URL`  | `postgres://yauth:yauth@127.0.0.1:5432/yauth_app?sslmode=disable`        | Postgres DSN.                                     |
| `PORT`          | `3000`                                                                   | HTTP listener.                                    |
| `CORS_ORIGINS`  | (unset)                                                                  | Comma-separated. Empty = CORS off (Vite proxy mode). |
| `DISABLE_HIBP`  | (unset)                                                                  | Set to `true` to skip the HIBP password check (dev). |
| `AUTO_MIGRATE`  | `true`                                                                   | Set to `false` to disable startup auto-migrate.   |

`AutoAdminFirstUser: true` is hardcoded in `serve()` — first registered
user becomes admin. Remove that line for prod if you want manual admin
provisioning.

## Conventions

- **Go** — `go vet` clean, `gofmt -l .` must be empty. `./task lint-server`
  enforces both.
- **Vue / TS** — `vp lint` (oxlint, 93 rules) and `vue-tsc --noEmit`.
  `./task lint` + `./task typecheck`. Default `tsconfig.app.json` keeps
  `erasableSyntaxOnly: true` on (works with yauth-ui-vue 0.12.2+).
- **JS package manager** — pnpm only. Never use npm/yarn/bun for
  scripts. `packageManager` is pinned in `web/package.json` so corepack
  resolves the right pnpm.
- **Build tool** — `vp` (vite-plus) for the web side. `pnpm dev` /
  `pnpm build` invoke `vp dev` / `vp build`.
- **Tailwind** — v4, via `@tailwindcss/vite`. The `@theme` block in
  `src/style.css` is the design-tokens contract; don't replace it with
  `tailwind.config.js`.

## CI

`.github/workflows/ci.yml` runs the same `./task` targets you use locally:

- `server` job: `./task lint-server`, `build-server`, `test-server`,
  `migrate`, then a curl smoke (register → login → `/api/me`) against a
  Postgres 17 service container.
- `web` job: `./task lint-web`, `typecheck`, `build-web`. pnpm store
  cached across runs.

Both jobs `go mod download` against `server/go.sum` so the Task tool
itself is fetched and cached.

## Don't

- Don't add `go run` / `pnpm` invocations directly to docs or scripts —
  add a Taskfile target and reference that.
- Don't reintroduce `npm publish` for the TS packages (they're in the
  yauth repo, not here, but if you ever vendor them: pnpm publish
  rewrites `workspace:*`; npm doesn't).
- Don't bypass the Vite proxy in dev — point fetches at relative `/api/...`
  so cookies are first-party. If you absolutely need to call the
  backend directly, set `CORS_ORIGINS` and use `credentials: 'include'`.
