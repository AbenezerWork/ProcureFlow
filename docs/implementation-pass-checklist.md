# Implementation Pass Checklist

This checklist tracks one practical end-to-end implementation pass for the current ProcureFlow backend scope.

Completed means the slice exists in code, is wired through the running API where applicable, and has at least basic automated or manual verification.

## Foundation

- [x] Clean architecture project layout under `internal/domain`, `internal/application`, `internal/interfaces`, and `internal/infrastructure`
- [x] API entrypoint under `cmd/api`
- [x] Embedded migration command under `cmd/migrate`
- [x] Initial PostgreSQL schema migration covering Phase 1 domain tables
- [x] PostgreSQL runtime connection via `pgx/v5`
- [x] `sqlc` query generation wired into the database layer
- [x] Transaction wrapper for application use cases
- [x] Dockerfile for API and migration binaries
- [x] Docker Compose workflow for Postgres, migrations, and API startup

## API Platform

- [x] Environment-based application configuration
- [x] Health check endpoint
- [x] Shared HTTP router and middleware wiring
- [x] Authentication middleware for protected routes
- [x] OpenAPI document served by the application
- [x] Swagger UI served by the application

## Identity

- [x] User registration
- [x] User login
- [x] Current authenticated user endpoint
- [x] Password hashing
- [x] Bearer token issuance and validation
- [x] Identity persistence in PostgreSQL

## Organizations And Memberships

- [x] Organization creation with transactional owner-membership creation
- [x] List organizations for the authenticated user
- [x] Get organization details for an active member
- [x] Update organization details
- [x] Organization membership listing
- [x] Organization membership creation by user ID or email lookup
- [x] Organization membership role update
- [x] Organization membership status update
- [x] Dedicated ownership transfer flow
- [x] Role-based organization management rules for `owner` and `admin`
- [x] Membership-status enforcement for protected organization access
- [x] Protection against generic updates to existing owner memberships
- [x] Protection against self-modifying membership updates

## Documentation And Verification

- [x] README updated to reflect the current runnable stack
- [x] API usage guide for the implemented auth and organization flows
- [x] Router and application tests for the implemented organization slice
- [x] Manual smoke test against a real PostgreSQL-backed API instance

## Remaining For A Broader Phase 1 Pass

- [x] Enforced tenant context on organization-scoped protected routes using `X-Tenant-ID`
- [x] Vendor module: CRUD, archive, and organization-scoped queries
- [x] Procurement module: draft, items, submit, cancel
- [ ] Approval module: approver inbox, approve, reject, comment
- [ ] RFQ module: create, publish, close, evaluate, cancel
- [ ] Quotation module: draft, submit, reject, item pricing
- [ ] Award module: award decision and award lookup
- [ ] Activity log module for major domain actions
- [ ] Workflow-transition enforcement for procurement, RFQ, quotation, and award slices
- [ ] Broader integration coverage beyond the current auth and organization flows
