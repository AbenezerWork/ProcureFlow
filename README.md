# ProcureFlow Backend

ProcureFlow is a Go backend for a multi-tenant procurement workflow. This repository starts with a clean-architecture scaffold so core business rules can evolve independently from transport, persistence, and framework details.

## Initial architecture

The project is organized around four boundaries:

- `internal/domain`: enterprise entities and core business concepts
- `internal/application`: use cases and application services
- `internal/interfaces`: delivery adapters such as HTTP handlers and middleware
- `internal/infrastructure`: technical details such as configuration and server runtime

## Project layout

```text
.
├── cmd/api
├── db/migrations
├── docs
├── internal/application
├── internal/bootstrap
├── internal/domain
├── internal/infrastructure
└── internal/interfaces
```

## Running locally

```bash
go run ./cmd/migrate up
go run ./cmd/api
```

Default server settings:

- HTTP address: `:8080`
- Tenant header: `X-Tenant-ID`
- Environment: `development`

Health check:

```bash
curl http://localhost:8080/healthz
```

## Running with Docker

The repository includes a multi-stage `Dockerfile` for the API and migration binaries, plus a `compose.yaml` stack that starts PostgreSQL, runs migrations, and then starts the API.

```bash
docker compose up --build
```

Services:

- API: `http://localhost:8080`
- PostgreSQL: `localhost:5432`

Default PostgreSQL credentials:

- Database: `procureflow`
- User: `procureflow`
- Password: `procureflow`

The application now also reads database settings from environment variables so the repository is ready for repository and migration wiring:

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_SSLMODE`

## Phase 1 design baseline

The initial product-to-backend mapping is documented in `docs/phase1-foundation.md`.

Database schema changes should be added as raw SQL migrations under `db/migrations`.

## Database tooling

The repository uses:

- embedded Go migrations via `cmd/migrate`
- `sqlc` generated query code under `internal/infrastructure/database/sqlc`
- `pgx/v5` for the PostgreSQL runtime connection and transactions

Useful commands:

```bash
sqlc generate
go run ./cmd/migrate up
go run ./cmd/migrate version
```

The API does not auto-run migrations on startup. Apply them explicitly in local development and CI.

## Docker workflow

`docker compose up --build` now runs a one-shot `migrate` service before the API starts.
