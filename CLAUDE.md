# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...                            # Build all packages
go test ./...                             # Run all tests
go test ./rejecteddrug/ -run TestName     # Run a single test
go run . serve                            # Run the API server (listens on :8080)
make protoc                               # Regenerate kafkapb/*.pb.go from kafkapb/*.proto
```

There is no lint configuration; use `gofmt`/`go vet`.

Commit messages follow Conventional Commits (`feat:`, `fix:`, `chore:`, with optional scopes like `feat(drug): ...`). Pushes to `main` trigger automatic semver tagging, changelog generation, a Docker image publish to ghcr.io, and deployment (see `.github/workflows/`).

## What this service is

A proxy in front of [Vmedis](https://vmedis.com), a pharmacy management web app. It scrapes/fetches data from Vmedis, dumps it into its own database, and serves it through a cleaner HTTP API plus scheduled jobs (Kafka consumers, email reports). All times are hardcoded to `Asia/Jakarta` with the `id_ID` locale (set in `main.go`); domain language is Indonesian pharmacy terminology (e.g. "stock opname" = inventory count).

## Architecture

Single binary with Cobra subcommands (`cmd/`): `serve` runs the HTTP API; `drugs`, `sales`, `procurements`, `stock-opnames`, `shifts` have `dump` subcommands that fetch from Vmedis and write to the DB; `tokens refresh` keeps Vmedis session tokens alive; `drugs run-updated-drugs-consumer` runs the Kafka consumer; `reports send-to-iqvia` emails monthly reports.

- **Configuration** (`cmd/root.go`): Viper, layered flags → `VMEDIS_*` env vars → YAML file from `./config/config.yaml` or `~/.vmedis-proxy-api/config.yaml` (see `config/config.yaml.example`). Config keys use underscores (`postgres_dsn`), flags use dashes (`--postgres-dsn`).
- **Dependency wiring** (`cmd/dependencies.go`): every shared dependency (DB, Redis, Vmedis client, services, handlers) is a lazily-initialized singleton behind an `atomic.Pointer` with a `getX()` accessor. New services/handlers get wired here.
- **Domain packages** (`drug`, `sale`, `procurement`, `stockopname`, `shift`, `rejecteddrug`, `auth`) all follow the same file layout: `models.go` (API DTOs), `db.go` (GORM queries wrapped in a `Database` type), `service.go`, `handler.go` (Gin handlers, named `ApiHandler`), `commands.go` (entry points called by `cmd/`), optionally `cache.go` (Redis) and `consumer.go`/`producer.go` (Kafka).
- **Vmedis clients** (`vmedis/`): `v1` scrapes the Vmedis web UI — HTML parsed with goquery into structs via `vmedis` struct tags interpreted by `schema.go` and the `*_parser.go` files; authenticated by session-ID tokens stored in the DB and kept alive by `vmedis/v1/token` (provider/refresher/service); all requests go through a shared rate limiter. `v2` calls the Vmedis API gateway with encrypted payloads (`crypt.go`).
- **HTTP server** (`proxy/`): routes are declared in `proxy/api.go` under `/api/v1` (raw data) and `/api/v2` (backend-driven UI: the server returns display-ready components — tables, forms, option lists — built from the `cui` (common UI) package types, and the user's role changes the output, e.g. `drug/handler_v2.go` builds role-gated sections). New v2 endpoints should respond with `cui` types rather than raw data, so the frontend can render them generically. Responses are cached in Redis via gin-cache with a zstd-compressing store (`proxy/cache.go`).
- **Auth** (`auth/`): the `X-Email` request header identifies the user (no password); the Gin middleware resolves it to a role (`admin`, `staff`, `reseller`, `guest`) from the `users` table, defaulting to guest. Role checks happen in handlers, not in routing.
- **Database** (`database/`): GORM with Postgres (`postgres_dsn` set) or SQLite fallback. Schema is managed by `AutoMigrate` on startup — add new models to `database/models/` and register them in `database/db.go`.
- **Kafka** (`kafkapb/`, `drug/producer.go`, `drug/consumer.go`): drug updates are published as protobuf messages (topics `drug_vmedis_id.updated` and `drug_vmedis_code.updated`, declared in `drug/producer.go`); the consumer re-fetches full drug details from Vmedis and upserts them.

The HTTP API is documented in `docs/openapi.yaml` — keep it in sync when changing endpoints.
