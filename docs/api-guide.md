# API Guide

This guide summarizes the currently implemented API surface and the authorization rules around organization, vendor, procurement request, approval, and RFQ management.

## Base URLs

- API: `http://localhost:8080`
- OpenAPI: `http://localhost:8080/openapi.yaml`
- Swagger UI: `http://localhost:8080/swagger`

## Authentication

Authenticate with bearer tokens returned by:

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

Send the token on protected routes:

```http
Authorization: Bearer <access-token>
```

## Implemented routes

### Public

- `GET /healthz`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

### Authenticated

- `GET /api/v1/auth/me`
- `GET /api/v1/organizations/`
- `POST /api/v1/organizations/`
- `GET /api/v1/organizations/{organizationID}`
- `PATCH /api/v1/organizations/{organizationID}`
- `GET /api/v1/organizations/{organizationID}/memberships`
- `POST /api/v1/organizations/{organizationID}/memberships`
- `PATCH /api/v1/organizations/{organizationID}/memberships/{userID}`
- `POST /api/v1/organizations/{organizationID}/ownership-transfer`
- `GET /api/v1/organizations/{organizationID}/vendors/`
- `POST /api/v1/organizations/{organizationID}/vendors/`
- `GET /api/v1/organizations/{organizationID}/vendors/{vendorID}`
- `PATCH /api/v1/organizations/{organizationID}/vendors/{vendorID}`
- `POST /api/v1/organizations/{organizationID}/vendors/{vendorID}/archive`
- `GET /api/v1/organizations/{organizationID}/procurement-requests/`
- `POST /api/v1/organizations/{organizationID}/procurement-requests/`
- `GET /api/v1/organizations/{organizationID}/approvals/procurement-requests/`
- `GET /api/v1/organizations/{organizationID}/procurement-requests/{requestID}`
- `PATCH /api/v1/organizations/{organizationID}/procurement-requests/{requestID}`
- `POST /api/v1/organizations/{organizationID}/procurement-requests/{requestID}/submit`
- `POST /api/v1/organizations/{organizationID}/procurement-requests/{requestID}/approve`
- `POST /api/v1/organizations/{organizationID}/procurement-requests/{requestID}/reject`
- `POST /api/v1/organizations/{organizationID}/procurement-requests/{requestID}/cancel`
- `GET /api/v1/organizations/{organizationID}/procurement-requests/{requestID}/items`
- `POST /api/v1/organizations/{organizationID}/procurement-requests/{requestID}/items`
- `PATCH /api/v1/organizations/{organizationID}/procurement-requests/{requestID}/items/{itemID}`
- `DELETE /api/v1/organizations/{organizationID}/procurement-requests/{requestID}/items/{itemID}`
- `GET /api/v1/organizations/{organizationID}/rfqs/`
- `POST /api/v1/organizations/{organizationID}/rfqs/`
- `GET /api/v1/organizations/{organizationID}/rfqs/{rfqID}`
- `PATCH /api/v1/organizations/{organizationID}/rfqs/{rfqID}`
- `POST /api/v1/organizations/{organizationID}/rfqs/{rfqID}/publish`
- `POST /api/v1/organizations/{organizationID}/rfqs/{rfqID}/close`
- `POST /api/v1/organizations/{organizationID}/rfqs/{rfqID}/evaluate`
- `POST /api/v1/organizations/{organizationID}/rfqs/{rfqID}/cancel`
- `GET /api/v1/organizations/{organizationID}/rfqs/{rfqID}/items`
- `GET /api/v1/organizations/{organizationID}/rfqs/{rfqID}/vendors`
- `POST /api/v1/organizations/{organizationID}/rfqs/{rfqID}/vendors`
- `DELETE /api/v1/organizations/{organizationID}/rfqs/{rfqID}/vendors/{vendorID}`

## Organization roles

- `owner`
- `admin`
- `requester`
- `approver`
- `procurement_officer`
- `viewer`

## Membership statuses

- `invited`
- `active`
- `suspended`
- `removed`

## Current authorization rules

- Any organization access requires an active membership in that organization.
- Organization-scoped routes require `X-Tenant-ID` to match the target organization ID.
- Only `owner` and `admin` can update organization details.
- Only `owner` and `admin` can list or manage memberships.
- Only an `owner` can create another `owner` membership.
- Generic membership updates cannot modify existing `owner` memberships.
- Users cannot modify their own membership through the generic membership update route.
- Ownership transfer requires the caller to be the current active `owner`.
- Ownership transfer requires the target user to already have an active membership in the organization.
- Any active organization member can list and get vendors.
- Only `owner`, `admin`, and `procurement_officer` can create, update, or archive vendors.
- Any active organization member can list and get procurement requests and request items.
- `owner`, `admin`, `procurement_officer`, and `requester` can create procurement request drafts.
- Draft updates, item writes, submit, and cancel are allowed for manager roles or the original requester on their own request.
- Only active `owner`, `admin`, and `approver` memberships can access the procurement approval inbox.
- Only active `owner`, `admin`, and `approver` memberships can approve or reject submitted procurement requests.
- Any active organization member can list and get RFQs, RFQ items, and RFQ vendors.
- Only active `owner`, `admin`, and `procurement_officer` memberships can create and manage RFQs.
- RFQs can only be created from approved procurement requests.
- RFQ creation snapshots the current procurement request items into immutable RFQ items.
- Vendor attachment and removal are only allowed while the RFQ is in `draft`.
- Publishing an RFQ requires at least one attached vendor.

## Manual test flow

Register a user:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"owner@example.com","password":"secret123","full_name":"Owner User"}'
```

Login and capture a token:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"owner@example.com","password":"secret123"}'
```

Create an organization:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/ \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Acme Corporation","slug":"acme"}'
```

List organizations:

```bash
curl http://localhost:8080/api/v1/organizations/ \
  -H "Authorization: Bearer $TOKEN"
```

Load one organization:

```bash
curl http://localhost:8080/api/v1/organizations/$ORG_ID \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN"
```

Invite an existing user by email:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/memberships \
  -H 'Content-Type: application/json' \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email":"approver@example.com","role":"approver","status":"invited"}'
```

Activate or change a member role:

```bash
curl -X PATCH http://localhost:8080/api/v1/organizations/$ORG_ID/memberships/$USER_ID \
  -H 'Content-Type: application/json' \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"role":"approver","status":"active"}'
```

Transfer ownership to another active member:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/ownership-transfer \
  -H 'Content-Type: application/json' \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"target_user_id":"'$USER_ID'","current_owner_new_role":"admin"}'
```

Create a vendor:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/vendors/ \
  -H 'Content-Type: application/json' \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Blue Nile Supplies","email":"sales@bluenile.example","country":"ET"}'
```

List vendors:

```bash
curl http://localhost:8080/api/v1/organizations/$ORG_ID/vendors/ \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN"
```

Update a vendor:

```bash
curl -X PATCH http://localhost:8080/api/v1/organizations/$ORG_ID/vendors/$VENDOR_ID \
  -H 'Content-Type: application/json' \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"contact_name":"Abebe Buyer","phone":"+251900000000"}'
```

Archive a vendor:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/vendors/$VENDOR_ID/archive \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN"
```

Create a procurement request draft:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/procurement-requests/ \
  -H 'Content-Type: application/json' \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Office Chairs","justification":"Expand seating","currency_code":"ETB","estimated_total_amount":"10000.00"}'
```

Add an item to a draft:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/procurement-requests/$REQUEST_ID/items \
  -H 'Content-Type: application/json' \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"item_name":"Ergonomic Chair","quantity":"10","unit":"pcs","estimated_unit_price":"1000.00","needed_by_date":"2026-04-15"}'
```

Submit a draft request:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/procurement-requests/$REQUEST_ID/submit \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN"
```

List the approval inbox as an approver:

```bash
curl http://localhost:8080/api/v1/organizations/$ORG_ID/approvals/procurement-requests/ \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $APPROVER_TOKEN"
```

Approve a submitted request:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/procurement-requests/$REQUEST_ID/approve \
  -H 'Content-Type: application/json' \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $APPROVER_TOKEN" \
  -d '{"decision_comment":"Approved for vendor sourcing"}'
```

Reject a submitted request:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/procurement-requests/$REQUEST_ID/reject \
  -H 'Content-Type: application/json' \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $APPROVER_TOKEN" \
  -d '{"decision_comment":"Insufficient business justification"}'
```

Create an RFQ from an approved procurement request:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/rfqs/ \
  -H 'Content-Type: application/json' \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"procurement_request_id":"'$REQUEST_ID'","reference_number":"RFQ-2026-001"}'
```

Attach a vendor to a draft RFQ:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/rfqs/$RFQ_ID/vendors \
  -H 'Content-Type: application/json' \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"vendor_id":"'$VENDOR_ID'"}'
```

Publish an RFQ:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/rfqs/$RFQ_ID/publish \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN"
```

Close and evaluate an RFQ:

```bash
curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/rfqs/$RFQ_ID/close \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN"

curl -X POST http://localhost:8080/api/v1/organizations/$ORG_ID/rfqs/$RFQ_ID/evaluate \
  -H "X-Tenant-ID: $ORG_ID" \
  -H "Authorization: Bearer $TOKEN"
```
