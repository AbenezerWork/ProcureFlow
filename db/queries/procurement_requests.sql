-- name: CreateProcurementRequest :one
INSERT INTO procurement_requests (
    organization_id,
    requester_user_id,
    title,
    description,
    justification,
    currency_code,
    estimated_total_amount
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetProcurementRequestByID :one
SELECT *
FROM procurement_requests
WHERE organization_id = $1
  AND id = $2;

-- name: ListProcurementRequests :many
SELECT *
FROM procurement_requests
WHERE organization_id = $1
  AND (
      sqlc.narg(status)::procurement_request_status IS NULL
      OR status = sqlc.narg(status)::procurement_request_status
  )
ORDER BY created_at DESC;

-- name: ListApprovalInboxRequests :many
SELECT *
FROM procurement_requests
WHERE organization_id = $1
  AND status = 'submitted'
ORDER BY submitted_at ASC NULLS LAST;

-- name: UpdateDraftProcurementRequest :one
UPDATE procurement_requests
SET
    title = $3,
    description = $4,
    justification = $5,
    currency_code = $6,
    estimated_total_amount = $7,
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status = 'draft'
RETURNING *;

-- name: SubmitProcurementRequest :one
UPDATE procurement_requests
SET
    status = 'submitted',
    submitted_at = now(),
    submitted_by_user_id = $3,
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status = 'draft'
RETURNING *;

-- name: ApproveProcurementRequest :one
UPDATE procurement_requests
SET
    status = 'approved',
    approved_at = now(),
    approved_by_user_id = $3,
    decision_comment = $4,
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status = 'submitted'
RETURNING *;

-- name: RejectProcurementRequest :one
UPDATE procurement_requests
SET
    status = 'rejected',
    rejected_at = now(),
    rejected_by_user_id = $3,
    decision_comment = $4,
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status = 'submitted'
RETURNING *;

-- name: CancelProcurementRequest :one
UPDATE procurement_requests
SET
    status = 'cancelled',
    cancelled_at = now(),
    cancelled_by_user_id = $3,
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status IN ('draft', 'submitted')
RETURNING *;

-- name: CreateProcurementRequestItem :one
INSERT INTO procurement_request_items (
    organization_id,
    procurement_request_id,
    line_number,
    item_name,
    description,
    quantity,
    unit,
    estimated_unit_price,
    needed_by_date
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: ListProcurementRequestItems :many
SELECT *
FROM procurement_request_items
WHERE organization_id = $1
  AND procurement_request_id = $2
ORDER BY line_number;

-- name: UpdateProcurementRequestItem :one
UPDATE procurement_request_items
SET
    item_name = $4,
    description = $5,
    quantity = $6,
    unit = $7,
    estimated_unit_price = $8,
    needed_by_date = $9,
    updated_at = now()
WHERE organization_id = $1
  AND procurement_request_id = $2
  AND id = $3
RETURNING *;

-- name: DeleteProcurementRequestItem :execrows
DELETE FROM procurement_request_items
WHERE organization_id = $1
  AND procurement_request_id = $2
  AND id = $3;
