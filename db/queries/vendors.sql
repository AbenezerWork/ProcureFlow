-- name: CreateVendor :one
INSERT INTO vendors (
    organization_id,
    name,
    legal_name,
    contact_name,
    email,
    phone,
    tax_identifier,
    address_line1,
    address_line2,
    city,
    state_region,
    postal_code,
    country,
    notes,
    created_by_user_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13, $14, $15
)
RETURNING *;

-- name: GetVendorByID :one
SELECT *
FROM vendors
WHERE organization_id = $1
  AND id = $2;

-- name: ListVendors :many
SELECT *
FROM vendors
WHERE organization_id = $1
  AND (
      sqlc.narg(status)::vendor_status IS NULL
      OR status = sqlc.narg(status)::vendor_status
  )
ORDER BY created_at DESC;

-- name: SearchVendors :many
SELECT *
FROM vendors
WHERE organization_id = $1
  AND (
      lower(name) LIKE '%' || lower(sqlc.arg(search)) || '%'
      OR lower(COALESCE(legal_name, '')) LIKE '%' || lower(sqlc.arg(search)) || '%'
      OR lower(COALESCE(email, '')) LIKE '%' || lower(sqlc.arg(search)) || '%'
  )
ORDER BY name;

-- name: UpdateVendor :one
UPDATE vendors
SET
    name = $3,
    legal_name = $4,
    contact_name = $5,
    email = $6,
    phone = $7,
    tax_identifier = $8,
    address_line1 = $9,
    address_line2 = $10,
    city = $11,
    state_region = $12,
    postal_code = $13,
    country = $14,
    notes = $15,
    updated_by_user_id = $16,
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
RETURNING *;

-- name: ArchiveVendor :one
UPDATE vendors
SET
    status = 'archived',
    archived_at = now(),
    updated_by_user_id = $3,
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
RETURNING *;
