# API Guide

This guide summarizes the currently implemented API surface and the authorization rules around organization and vendor management.

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
