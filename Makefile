SHELL := /bin/bash
.DEFAULT_GOAL := help

ENV_FILE := .env

# Source .env so env-var-driven targets see DATABASE_URL etc.
ifneq (,$(wildcard $(ENV_FILE)))
include $(ENV_FILE)
export
endif

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: setup
setup: ## First-time setup — copy .env, install web deps, fetch Go modules
	@test -f .env || cp .env.example .env
	cd web && pnpm install --frozen-lockfile=false
	cd server && go mod download

.PHONY: db-up
db-up: ## Start the Postgres container in the background
	docker compose up -d postgres
	@echo "Waiting for Postgres to be healthy..."
	@until docker compose ps postgres --format json | grep -q '"Health":"healthy"'; do sleep 1; done
	@echo "Postgres is ready at $${DATABASE_URL}"

.PHONY: db-down
db-down: ## Stop the Postgres container
	docker compose down

.PHONY: db-reset
db-reset: ## Drop and recreate the Postgres volume
	docker compose down -v
	$(MAKE) db-up

.PHONY: server
server: ## Run the Go backend (foreground)
	cd server && go run .

.PHONY: web
web: ## Run the Vue dev server (foreground, proxies /api → backend)
	cd web && pnpm dev

.PHONY: dev
dev: db-up ## Bring up Postgres and start backend + frontend in parallel (Ctrl-C stops both)
	@trap 'kill 0' INT TERM; \
	(cd server && go run .) & \
	(cd web && pnpm dev) & \
	wait

.PHONY: build
build: build-server build-web ## Build server binary + web bundle

.PHONY: build-server
build-server: ## Compile the server binary into server/server
	cd server && go build -o server .

.PHONY: build-web
build-web: ## Build the production web bundle
	cd web && pnpm build

.PHONY: lint
lint: lint-server lint-web ## Lint everything

.PHONY: lint-server
lint-server: ## go vet + gofmt check
	cd server && go vet ./... && test -z "$$(gofmt -l .)"

.PHONY: lint-web
lint-web: ## vp lint
	cd web && pnpm lint

.PHONY: test
test: test-server ## Run all tests

.PHONY: test-server
test-server: ## Go tests
	cd server && go test ./...

.PHONY: clean
clean: ## Remove build artifacts
	rm -f server/server
	rm -rf web/dist web/node_modules/.vite
