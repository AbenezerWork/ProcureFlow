# ProcureFlow

ProcureFlow is a multi-tenant procurement workflow platform. The backend is a Go service with a clean-architecture scaffold, and the frontend lives in a separate Vite/React app under `web/`.

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
├── internal/interfaces
└── web
```

## Running locally

```bash
export AUTH_JWT_SECRET=replace-with-a-long-random-secret
export DB_USER=procureflow
export DB_PASSWORD=procureflow
export DB_NAME=procureflow
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

## Frontend

The frontend is a root-level app in `web/` so it can evolve alongside the Go backend without changing the backend package layout.

```bash
cd web
npm install
npm run dev
```

Default frontend settings:

- Vite dev server: `http://localhost:5173`
- API base URL: `VITE_API_BASE_URL=http://localhost:8080`

The frontend API wrapper sends bearer tokens returned by `/api/v1/auth/login` or `/api/v1/auth/register`, and organization-scoped calls include the selected organization ID in `X-Tenant-ID`.

Generate TypeScript API types from the backend OpenAPI document:

```bash
cd web
npm run generate:api
```

The generated output is written to `web/src/shared/api/generated/schema.ts`.

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
- quotations: create/list/get/update, item pricing, submit, reject, compare submitted RFQ quotations
- awards: create and look up the RFQ award decision
- activity logs: query entity timelines for organization-scoped workflow activity

Known Phase 1 gaps:

- Handler, repository, and real-database integration test coverage should be broadened for production confidence

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

The application refuses to start unless `AUTH_JWT_SECRET`, `DB_USER`, `DB_PASSWORD`, and `DB_NAME` are explicitly set.

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

### Demo seed data

For local validation, the migration command can populate a deterministic demo organization after applying migrations:

```bash
go run ./cmd/migrate up -seed-demo
```

With Docker Compose:

```bash
docker compose run --rm migrate up -seed-demo
```

The seed is idempotent and only manages records with the fixed demo IDs and `demo.*@procureflow.test` accounts. It creates:

- organization: `ProcureFlow Demo Organization`
- tenant ID: `10000000-0000-4000-8000-000000000001`
- password for all demo users: `DemoPass123!`
- active role users: `demo.owner@procureflow.test`, `demo.admin@procureflow.test`, `demo.requester@procureflow.test`, `demo.approver@procureflow.test`, `demo.procurement@procureflow.test`, `demo.viewer@procureflow.test`
- membership-status users: `demo.invited@procureflow.test`, `demo.suspended@procureflow.test`, `demo.removed@procureflow.test`

The demo organization includes active and archived vendors, procurement requests across every request status, RFQs across every RFQ status, quotations across every quotation status, one award, and activity-log entries for timeline validation.

## Integration tests

Real PostgreSQL-backed integration tests live under `internal/integration` and are guarded by the `integration` build tag. They are skipped unless `PROCUREFLOW_TEST_DATABASE_URL` is set.

Example:

```bash
docker run --rm -d --name procureflow-integration-test \
  -e POSTGRES_DB=procureflow_test \
  -e POSTGRES_USER=procureflow \
  -e POSTGRES_PASSWORD=procureflow \
  -p 55432:5432 \
  postgres:18-alpine

PROCUREFLOW_TEST_DATABASE_URL='postgres://procureflow:procureflow@localhost:55432/procureflow_test?sslmode=disable' \
  go test -tags=integration ./internal/integration

docker stop procureflow-integration-test
```

The integration harness applies embedded migrations, truncates domain tables between runs, and exercises the full procurement request to award flow through real repositories and services.

## API docs

The service now serves its OpenAPI document and Swagger UI directly:

```bash
curl http://localhost:8080/openapi.yaml
```

Then open `http://localhost:8080/swagger` in a browser.

Swagger UI uses the served OpenAPI document at `/openapi.yaml`. The UI assets are loaded from the `swagger-ui-dist` CDN in the browser, so internet access is required when viewing `/swagger`.

## Docker workflow

`docker compose up --build` now runs a one-shot `migrate` service before the API starts.
