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

API documentation:

- OpenAPI spec: `http://localhost:8080/openapi.yaml`
- Swagger UI: `http://localhost:8080/swagger`

Implemented API areas:

- auth: register, login, current user
- organizations: create, list mine, get details, update
- organization memberships: list, create, update
- organization ownership transfer: dedicated owner-to-member transfer flow
- vendors: create, list, get, update, archive
- procurement requests: draft create/list/get/update, item CRUD, submit, approve, reject, cancel
- rfqs: create from approved requests, snapshot items, attach/remove vendors, publish, close, evaluate, cancel

## Running with Docker

The repository includes a multi-stage `Dockerfile` for the API and migration binaries, plus a `compose.yaml` stack that starts PostgreSQL, runs migrations, and then starts the API.

```bash
docker compose up --build
```

Set the required runtime secrets first, for example:

```bash
export AUTH_JWT_SECRET=replace-with-a-long-random-secret
export DB_USER=procureflow
export DB_PASSWORD=replace-with-a-strong-password
export DB_NAME=procureflow
docker compose up --build
```

Services:

- API: `http://localhost:8080`
- PostgreSQL: `localhost:5432`

The application reads database settings from environment variables:

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_SSLMODE`

JWT signing configuration:

- `AUTH_JWT_SECRET`

In `APP_ENV=production`, the application refuses to start if `AUTH_JWT_SECRET`, `DB_USER`, `DB_PASSWORD`, or `DB_NAME` are still using the built-in development defaults.

## Phase 1 design baseline

The initial product-to-backend mapping is documented in `docs/phase1-foundation.md`.
The current API usage guide lives in `docs/api-guide.md`.
The current implementation-pass deliverables checklist lives in `docs/implementation-pass-checklist.md`.

Current organization roles:

- `owner`
- `admin`
- `requester`
- `approver`
- `procurement_officer`
- `viewer`

Current membership statuses:

- `invited`
- `active`
- `suspended`
- `removed`

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

## API docs

The service now serves its OpenAPI document and Swagger UI directly:

```bash
curl http://localhost:8080/openapi.yaml
```

Then open `http://localhost:8080/swagger` in a browser.

Swagger UI uses the served OpenAPI document at `/openapi.yaml`. The UI assets are loaded from the `swagger-ui-dist` CDN in the browser, so internet access is required when viewing `/swagger`.

## Docker workflow

`docker compose up --build` now runs a one-shot `migrate` service before the API starts.
