-- name: CreateRFQ :one
INSERT INTO rfqs (
    organization_id,
    procurement_request_id,
    reference_number,
    title,
    description,
    created_by_user_id
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetRFQByID :one
SELECT *
FROM rfqs
WHERE organization_id = $1
  AND id = $2;

-- name: ListRFQs :many
SELECT *
FROM rfqs
WHERE organization_id = $1
  AND (
      sqlc.narg(status)::rfq_status IS NULL
      OR status = sqlc.narg(status)::rfq_status
  )
ORDER BY created_at DESC;

-- name: UpdateDraftRFQ :one
UPDATE rfqs
SET
    reference_number = $3,
    title = $4,
    description = $5,
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status = 'draft'
RETURNING *;

-- name: PublishRFQ :one
UPDATE rfqs
SET
    status = 'published',
    published_at = now(),
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status = 'draft'
RETURNING *;

-- name: CloseRFQ :one
UPDATE rfqs
SET
    status = 'closed',
    closed_at = now(),
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status = 'published'
RETURNING *;

-- name: EvaluateRFQ :one
UPDATE rfqs
SET
    status = 'evaluated',
    evaluated_at = now(),
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status = 'closed'
RETURNING *;

-- name: MarkRFQAwarded :one
UPDATE rfqs
SET
    status = 'awarded',
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status = 'evaluated'
RETURNING *;

-- name: CancelRFQ :one
UPDATE rfqs
SET
    status = 'cancelled',
    cancelled_at = now(),
    cancelled_by_user_id = $3,
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status IN ('draft', 'published', 'closed', 'evaluated')
RETURNING *;

-- name: CreateRFQItem :one
INSERT INTO rfq_items (
    organization_id,
    rfq_id,
    source_request_item_id,
    line_number,
    item_name,
    description,
    quantity,
    unit,
    target_date
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: ListRFQItems :many
SELECT *
FROM rfq_items
WHERE organization_id = $1
  AND rfq_id = $2
ORDER BY line_number;

-- name: AttachVendorToRFQ :one
INSERT INTO rfq_vendors (
    organization_id,
    rfq_id,
    vendor_id,
    attached_by_user_id
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: RemoveVendorFromRFQ :execrows
DELETE FROM rfq_vendors
WHERE organization_id = $1
  AND rfq_id = $2
  AND vendor_id = $3;

-- name: ListRFQVendors :many
SELECT rv.*, v.name AS vendor_name, v.status AS vendor_status
FROM rfq_vendors AS rv
JOIN vendors AS v
    ON v.organization_id = rv.organization_id
   AND v.id = rv.vendor_id
WHERE rv.organization_id = $1
  AND rv.rfq_id = $2
ORDER BY v.name;
