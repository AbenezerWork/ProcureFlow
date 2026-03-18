# Phase 1 Foundation

This document translates the ProcureFlow Phase 1 product requirements into backend module boundaries, workflow rules, and the initial PostgreSQL schema.

## Module boundaries

The modular monolith should be split into these business modules:

- `identity`: registration, login, current user, password hashing, token issuance
- `organizations`: organization creation, organization lookup, organization context resolution
- `memberships`: organization membership lifecycle, role assignment, membership status enforcement
- `vendors`: vendor CRUD, archive, organization-scoped vendor search
- `procurement`: procurement request drafts, items, submission, cancellation
- `approvals`: approver inbox, approve, reject, decision comment
- `rfqs`: RFQ creation from approved requests, vendor attachment, publish, close, evaluate, cancel
- `quotations`: quotation entry, item pricing, submit, reject
- `awards`: RFQ award decision and award lookup
- `activitylog`: timeline-ready audit records for major domain actions

## Tenant model

Tenant isolation is organization-based.

- A user may belong to multiple organizations through `organization_memberships`
- All organization-owned records carry `organization_id`
- Application services must require an organization context for protected operations
- List and detail queries must always filter by `organization_id`
- The schema also uses tenant-aware foreign keys for org-owned children to reduce accidental cross-organization links

## Role model

Membership roles are organization-scoped:

- `owner`
- `admin`
- `requester`
- `approver`
- `procurement_officer`
- `viewer`

Membership status is explicit:

- `invited`
- `active`
- `suspended`
- `removed`

Authorization should be enforced in the application layer using both role and membership status.

## Workflow states

Procurement request status:

- `draft`
- `submitted`
- `approved`
- `rejected`
- `cancelled`

RFQ status:

- `draft`
- `published`
- `closed`
- `evaluated`
- `awarded`
- `cancelled`

Quotation status:

- `draft`
- `submitted`
- `rejected`

These are intentionally modeled as backend-owned states. HTTP handlers should not bypass transition rules.

## Transaction boundaries

The following operations should be treated as transactional application use cases:

- organization creation plus owner membership creation
- procurement request submission plus activity log write
- procurement request approval or rejection plus activity log write
- RFQ creation from an approved request plus RFQ item snapshot creation plus activity log write
- quotation submission or rejection plus activity log write
- RFQ award creation plus RFQ status update plus activity log write

## Schema notes

The initial schema lives in `db/migrations/000001_init_schema.up.sql`.

Highlights:

- `users` stores authenticated identities
- `organizations` and `organization_memberships` model tenancy and org-scoped roles
- `vendors` is organization-owned and archiveable
- `procurement_requests` and `procurement_request_items` capture internal demand
- `rfqs`, `rfq_items`, and `rfq_vendors` manage sourcing
- `quotations` and `quotation_items` store vendor offers in comparable line-item form
- `rfq_awards` records the winning quotation
- `activity_logs` supports timeline rendering by `entity_type` and `entity_id`

The schema favors explicit foreign keys and duplicate `organization_id` columns on org-owned tables because the product requirements prioritize tenant isolation over normalization purity.
