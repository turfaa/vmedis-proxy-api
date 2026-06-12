# Vmedis Proxy API

A proxy service in front of [Vmedis](https://vmedis.com), a pharmacy management web application. It periodically pulls pharmacy data (drugs, sales, procurements, stock opnames, shifts) out of Vmedis, stores it in its own database, and exposes it through a clean, cacheable HTTP API — along with background jobs such as Kafka consumers and scheduled email reports.

## Features

- **HTTP API** (`/api/v1` and `/api/v2`) for sales, drugs, procurements (including procurement recommendations and invoice calculators), stock opnames, shifts, and rejected drugs. The full API is documented in [`docs/openapi.yaml`](docs/openapi.yaml).
- **Data dumpers** that scrape or fetch data from Vmedis and persist it to Postgres/SQLite.
- **Vmedis session management** — session tokens are stored in the database and kept alive by a refresher job.
- **Kafka pipeline** — drug updates are published as protobuf messages and a consumer re-fetches full drug details from Vmedis.
- **Backend-driven UI** — `/api/v2` endpoints return display-ready UI components (tables, forms, option lists) built with the [`cui`](cui) (common UI) package, so frontends can render them generically without domain logic.
- **Role-based responses** — users are identified by an `X-Email` header and mapped to `admin`, `staff`, `reseller`, or `guest` roles; `/api/v2` endpoints tailor their output to the caller's role.
- **Reports** — e.g. monthly sales/procurement reports emailed to IQVIA as Excel attachments.

All times use the `Asia/Jakarta` timezone with the `id_ID` locale.

## Getting started

### Prerequisites

- Go 1.24+
- Redis (response caching and token/user caches)
- Optional: PostgreSQL (falls back to SQLite when `postgres_dsn` is not set)
- Optional: Kafka (only needed for the drug update pipeline)

### Configuration

Configuration is read with [Viper](https://github.com/spf13/viper) from, in order of precedence:

1. Command-line flags (e.g. `--postgres-dsn`)
2. Environment variables prefixed with `VMEDIS_` (e.g. `VMEDIS_POSTGRES_DSN`)
3. A YAML config file at `./config/config.yaml` or `~/.vmedis-proxy-api/config.yaml` (or passed via `--config`)

Copy [`config/config.yaml.example`](config/config.yaml.example) to `config/config.yaml` and fill in your values:

```yaml
base_url: "https://xxx.vmedis.com"   # your Vmedis instance
sqlite_path: "data/db.sqlite"        # or set postgres_dsn instead
redis_address: "localhost:6379"
kafka_brokers:
  - "localhost:9092"
```

### Run

```bash
# Run the API server on :8080
go run . serve

# One-time dumpers
go run . drugs dump
go run . sales dump
go run . procurements dump
go run . stock-opnames dump
go run . shifts dump

# Keep Vmedis session tokens fresh
go run . tokens refresh

# Run the updated-drugs Kafka consumer
go run . drugs run-updated-drugs-consumer

# Send last month's report to IQVIA
go run . reports send-to-iqvia
```

### Docker

```bash
docker build -t vmedis-proxy-api .
docker run -p 8080:8080 vmedis-proxy-api
```

Tagged releases are automatically published to GitHub Container Registry (`ghcr.io/turfaa/vmedis-proxy-api`).

## API overview

Authentication is a simple `X-Email` header; requests without it are treated as the `guest` user. Endpoints that accept a time range use `date`, or `from` + `until`/`to` query parameters (`YYYY-MM-DD`), defaulting to today.

| Area | Examples |
|------|----------|
| Sales | `GET /api/v1/sales`, `GET /api/v1/sales/statistics`, `POST /api/v1/sales/dump` |
| Drugs | `GET /api/v1/drugs`, `GET /api/v1/drugs/to-stock-opname`, `GET /api/v2/drugs` |
| Procurements | `GET /api/v1/procurements/recommendations`, `GET /api/v1/procurements/invoice-calculators` |
| Stock opnames | `GET /api/v1/stock-opnames`, `GET /api/v1/stock-opnames/summaries` |
| Shifts | `GET /api/v2/shifts` |
| Rejected drugs | `GET /api/v2/rejected-drugs` |
| Vmedis tokens | `GET /api/v2/vmedis/tokens`, `POST /api/v2/vmedis/tokens` |
| Auth | `POST /api/v1/auth/login` |

See [`docs/openapi.yaml`](docs/openapi.yaml) for the complete, authoritative specification.

## Development

```bash
go build ./...    # build
go test ./...     # test
make protoc       # regenerate Kafka protobuf code from kafkapb/*.proto
```

Commits follow [Conventional Commits](https://www.conventionalcommits.org/); pushes to `main` automatically create semver tags, update [`CHANGELOG.md`](CHANGELOG.md), publish the Docker image, and trigger deployment via GitHub Actions.
