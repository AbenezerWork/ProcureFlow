# Next Session TODOs

These are the unfinished database-tooling tasks intentionally left out of this commit so the repository stays buildable.

## Tooling

- install `sqlc` locally or provide a pinned project-local binary path
- choose the migration execution approach for the repo:
  - `golang-migrate` CLI
  - embedded Go migration command
  - containerized migration step in Docker Compose

## sqlc generation

- run `sqlc generate` using `sqlc.yaml`
- review generated types for enum and nullable field ergonomics
- decide whether to keep the generated package directly under `internal/infrastructure/database/sqlc`
- add a thin store/transaction wrapper after generated code exists

## Database runtime

- add a PostgreSQL connection package using `pgx/v5`
- add a migration command or script that applies `db/migrations`
- decide whether the API should auto-run migrations on startup in local/dev only

## First implementation slice

- start with `identity`, `organizations`, and `memberships`
- define organization context resolution from authenticated user plus active membership
- implement auth token strategy before adding protected CRUD modules
