# Implementation Pass Checklist

This checklist tracks the current practical implementation pass for the ProcureFlow backend as it exists in the repository.

Completed means the slice is implemented end to end for the current backend stack:

- domain and application rules exist
- persistence is wired into the concrete repository layer
- HTTP routes and handlers are exposed where applicable
- the behavior has at least basic automated or documented manual verification

Database schema, raw SQL, or generated `sqlc` code alone do not count as a completed deliverable.

## Current Assessment

The repository is functionally strong for Phase 1. Auth, organizations, memberships, tenant enforcement, vendors, procurement requests, approvals, RFQs, quotations, quotation comparison, awards, activity logs, migrations, SQL generation, `pgx`, transaction handling, OpenAPI, and Docker startup are present.

Strictly against the original Phase 1 target, the remaining open work is mostly verification hardening rather than missing workflow implementation.

Current intentional design differences from the earlier Phase 1 wording:

- Award creation is restricted to active `owner`, `admin`, and `procurement_officer` memberships; `approver` participates in procurement request approvals, not award creation.
- Membership statuses include `removed` in addition to `invited`, `active`, and `suspended`.
- RFQ and quotation item snapshots are explicit backend behavior, which is stronger than the original generic workflow wording.

Current rough status:

- Functional Phase 1 completion: about 95%
- Verification hardening completion: about 75-80%

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

## Delivered End-To-End Slices

### Identity

- [x] User registration
- [x] User login
- [x] Current authenticated user endpoint
- [x] Password hashing
- [x] Bearer token issuance and validation
- [x] Identity persistence in PostgreSQL

### Organizations And Memberships

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

### Tenant Context

- [x] Enforced tenant context on organization-scoped protected routes using `X-Tenant-ID`

### Vendors

- [x] Vendor creation
- [x] Vendor listing
- [x] Vendor lookup
- [x] Vendor update
- [x] Vendor archive
- [x] Organization-scoped vendor authorization rules

### Procurement Requests And Approvals

- [x] Draft procurement request creation
- [x] Procurement request listing and lookup
- [x] Draft procurement request update
- [x] Procurement request item create, list, update, and delete
- [x] Procurement request submit flow
- [x] Approval inbox for active `owner`, `admin`, and `approver` memberships
- [x] Procurement request approve flow with decision comment support
- [x] Procurement request reject flow with decision comment support
- [x] Procurement request cancel flow
- [x] Workflow-transition enforcement for procurement request lifecycle states

### RFQs

- [x] RFQ creation from approved procurement requests
- [x] RFQ item snapshotting from procurement request items
- [x] RFQ listing and lookup
- [x] Draft RFQ update
- [x] RFQ vendor attach, list, and remove
- [x] RFQ publish flow
- [x] RFQ close flow
- [x] RFQ evaluate flow
- [x] RFQ cancel flow
- [x] Workflow-transition enforcement for RFQ lifecycle states

## Verification And Documentation Present In Repo

- [x] README updated to reflect the current runnable stack
- [x] API usage guide updated for the implemented auth, organization, vendor, procurement, approval, RFQ, quotation, award, and activity-log flows
- [x] Application-service tests for identity, organization, vendor, procurement, RFQ, quotation, award, and activity-log slices
- [x] Router coverage for the mounted API surface
- [x] Focused HTTP handler coverage for the organization slice
- [x] Manual smoke-test procedure documented for a real PostgreSQL-backed API instance

## Delivered Since Earlier Partial Groundwork

- [x] Quotation tables, SQL, and generated `sqlc` code exist, and the quotation domain, application service, repository wiring, HTTP surface, and focused test coverage are now delivered
- [x] Award tables, SQL, and generated `sqlc` code exist, and the award domain, application service, repository wiring, HTTP surface, and focused test coverage are now delivered
- [x] Activity log tables, SQL, generated `sqlc` code, transactional workflow writes, and an entity-scoped read endpoint are now wired as a deliverable timeline feature

## Remaining Deliverables For A Fuller Phase 1 Pass

### Quotation Comparison

- [x] Add RFQ quotation comparison domain/application read model
- [x] Implement repository read queries for comparing submitted quotations by RFQ and line item
- [x] Expose an RFQ comparison endpoint at `/api/v1/organizations/{organizationID}/rfqs/{rfqID}/comparison`
- [x] Add OpenAPI and API guide coverage for the comparison response
- [x] Add automated tests for comparison authorization, totals, and line-item comparison behavior

### Quotations

- [x] Add quotation domain models and application-service rules
- [x] Implement repository wiring for quotation create, list, get, update, submit, reject, and item pricing flows
- [x] Expose quotation HTTP handlers, routes, request validation, and OpenAPI documentation
- [x] Define quotation authorization rules across buyer and vendor-facing operations
- [x] Add automated tests for quotation lifecycle transitions and access control
- [x] Add manual API guide coverage for quotation flows

### Awards

- [x] Add award domain models and application-service rules
- [x] Implement repository wiring for award decision creation and award lookup
- [x] Expose award HTTP handlers, routes, request validation, and OpenAPI documentation
- [x] Enforce RFQ-to-award rules so only eligible quotations can be awarded
- [x] Add automated tests for award authorization and lifecycle behavior
- [x] Add manual API guide coverage for award flows

### Activity Logs

- [x] Write activity logs transactionally for major workflow actions
- [x] Define a stable event taxonomy for procurement requests, RFQs, quotations, awards, organizations, and memberships
- [x] Expose activity log query/read models needed for timeline rendering
- [x] Add automated verification that activity log writes occur on the intended workflows

### Workflow Completeness

- [x] Add workflow-transition enforcement for quotation lifecycle states
- [x] Add workflow-transition enforcement for award lifecycle states
- [ ] Verify cross-slice transitions from approved procurement request to RFQ to quotation to award

### Verification Depth

- [ ] Add HTTP handler tests for vendor, procurement, approval, and RFQ endpoints
- [ ] Add repository-level tests for the implemented PostgreSQL-backed data access paths
- [ ] Add broader integration coverage that exercises the full implemented flow against a real database
