-- name: CreateQuotation :one
INSERT INTO quotations (
    organization_id,
    rfq_id,
    rfq_vendor_id,
    currency_code,
    lead_time_days,
    payment_terms,
    notes,
    created_by_user_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetQuotationByID :one
SELECT *
FROM quotations
WHERE organization_id = $1
  AND id = $2;

-- name: ListQuotationsByRFQ :many
SELECT q.*, v.name AS vendor_name
FROM quotations AS q
JOIN rfq_vendors AS rv
    ON rv.rfq_id = q.rfq_id
   AND rv.id = q.rfq_vendor_id
JOIN vendors AS v
    ON v.organization_id = rv.organization_id
   AND v.id = rv.vendor_id
WHERE q.organization_id = $1
  AND q.rfq_id = $2
ORDER BY q.created_at DESC;

-- name: UpdateDraftQuotation :one
UPDATE quotations
SET
    currency_code = $3,
    lead_time_days = $4,
    payment_terms = $5,
    notes = $6,
    updated_by_user_id = $7,
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status = 'draft'
RETURNING *;

-- name: SubmitQuotation :one
UPDATE quotations
SET
    status = 'submitted',
    submitted_at = now(),
    submitted_by_user_id = $3,
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status = 'draft'
RETURNING *;

-- name: RejectQuotation :one
UPDATE quotations
SET
    status = 'rejected',
    rejected_at = now(),
    rejected_by_user_id = $3,
    rejection_reason = $4,
    updated_at = now()
WHERE organization_id = $1
  AND id = $2
  AND status IN ('draft', 'submitted')
RETURNING *;

-- name: CreateQuotationItem :one
INSERT INTO quotation_items (
    organization_id,
    quotation_id,
    rfq_id,
    rfq_item_id,
    line_number,
    item_name,
    description,
    quantity,
    unit,
    unit_price,
    delivery_days,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: ListQuotationItems :many
SELECT *
FROM quotation_items
WHERE organization_id = $1
  AND quotation_id = $2
ORDER BY line_number;

-- name: UpdateQuotationItem :one
UPDATE quotation_items
SET
    item_name = $4,
    description = $5,
    quantity = $6,
    unit = $7,
    unit_price = $8,
    delivery_days = $9,
    notes = $10,
    updated_at = now()
WHERE organization_id = $1
  AND quotation_id = $2
  AND id = $3
RETURNING *;

-- name: DeleteQuotationItem :execrows
DELETE FROM quotation_items
WHERE organization_id = $1
  AND quotation_id = $2
  AND id = $3;

-- name: CompareRFQQuotations :many
SELECT
    q.id AS quotation_id,
    q.rfq_id,
    q.status,
    q.currency_code,
    q.lead_time_days,
    v.id AS vendor_id,
    v.name AS vendor_name,
    COALESCE(SUM(qi.quantity * qi.unit_price), 0)::NUMERIC(18, 2) AS total_amount,
    MIN(qi.delivery_days) AS fastest_item_delivery_days
FROM quotations AS q
JOIN rfq_vendors AS rv
    ON rv.rfq_id = q.rfq_id
   AND rv.id = q.rfq_vendor_id
JOIN vendors AS v
    ON v.organization_id = rv.organization_id
   AND v.id = rv.vendor_id
LEFT JOIN quotation_items AS qi
    ON qi.organization_id = q.organization_id
   AND qi.quotation_id = q.id
WHERE q.organization_id = $1
  AND q.rfq_id = $2
GROUP BY q.id, q.rfq_id, q.status, q.currency_code, q.lead_time_days, v.id, v.name
ORDER BY total_amount ASC, q.created_at ASC;

-- name: CreateRFQAward :one
INSERT INTO rfq_awards (
    organization_id,
    rfq_id,
    quotation_id,
    awarded_by_user_id,
    reason
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetRFQAwardByRFQ :one
SELECT *
FROM rfq_awards
WHERE organization_id = $1
  AND rfq_id = $2;
