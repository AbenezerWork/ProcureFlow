package demo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DemoPassword = "DemoPass123!"

const demoPasswordHash = "$2a$10$AhDHBp7eeQevsxhZKbii9ejW7/qgAk55zTmPdPw398Ed4Sg9zjHPS"

type Result struct {
	OrganizationID string
	UserCount      int
}

func Seed(ctx context.Context, pool *pgxpool.Pool) (Result, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("begin demo seed transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	if _, err := tx.Exec(ctx, seedSQL); err != nil {
		return Result{}, fmt.Errorf("seed demo data: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("commit demo seed transaction: %w", err)
	}

	committed = true
	return Result{
		OrganizationID: "10000000-0000-4000-8000-000000000001",
		UserCount:      10,
	}, nil
}

const seedSQL = `
WITH demo_users(id, email, full_name, is_active) AS (
    VALUES
        ('20000000-0000-4000-8000-000000000001'::uuid, 'demo.owner@procureflow.test', 'Olivia Owner', true),
        ('20000000-0000-4000-8000-000000000002'::uuid, 'demo.admin@procureflow.test', 'Amir Admin', true),
        ('20000000-0000-4000-8000-000000000003'::uuid, 'demo.requester@procureflow.test', 'Ruth Requester', true),
        ('20000000-0000-4000-8000-000000000004'::uuid, 'demo.approver@procureflow.test', 'Abel Approver', true),
        ('20000000-0000-4000-8000-000000000005'::uuid, 'demo.procurement@procureflow.test', 'Priya Procurement', true),
        ('20000000-0000-4000-8000-000000000006'::uuid, 'demo.viewer@procureflow.test', 'Victor Viewer', true),
        ('20000000-0000-4000-8000-000000000007'::uuid, 'demo.invited@procureflow.test', 'Imani Invited', true),
        ('20000000-0000-4000-8000-000000000008'::uuid, 'demo.suspended@procureflow.test', 'Sam Suspended', true),
        ('20000000-0000-4000-8000-000000000009'::uuid, 'demo.removed@procureflow.test', 'Riley Removed', true),
        ('20000000-0000-4000-8000-000000000010'::uuid, 'demo.inactive@procureflow.test', 'Ines Inactive', false)
)
INSERT INTO users (id, email, password_hash, full_name, is_active, created_at, updated_at)
SELECT id, email, '` + demoPasswordHash + `', full_name, is_active, now(), now()
FROM demo_users
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    full_name = EXCLUDED.full_name,
    is_active = EXCLUDED.is_active,
    updated_at = now();

INSERT INTO organizations (id, name, slug, created_by_user_id, created_at, updated_at)
VALUES (
    '10000000-0000-4000-8000-000000000001',
    'ProcureFlow Demo Organization',
    'procureflow-demo',
    '20000000-0000-4000-8000-000000000001',
    now(),
    now()
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    created_by_user_id = EXCLUDED.created_by_user_id,
    archived_at = NULL,
    updated_at = now();

WITH demo_memberships(id, user_id, role, status, activated_at, suspended_at, removed_at) AS (
    VALUES
        ('21000000-0000-4000-8000-000000000001'::uuid, '20000000-0000-4000-8000-000000000001'::uuid, 'owner'::membership_role, 'active'::membership_status, now() - interval '30 days', NULL::timestamptz, NULL::timestamptz),
        ('21000000-0000-4000-8000-000000000002'::uuid, '20000000-0000-4000-8000-000000000002'::uuid, 'admin'::membership_role, 'active'::membership_status, now() - interval '29 days', NULL, NULL),
        ('21000000-0000-4000-8000-000000000003'::uuid, '20000000-0000-4000-8000-000000000003'::uuid, 'requester'::membership_role, 'active'::membership_status, now() - interval '28 days', NULL, NULL),
        ('21000000-0000-4000-8000-000000000004'::uuid, '20000000-0000-4000-8000-000000000004'::uuid, 'approver'::membership_role, 'active'::membership_status, now() - interval '27 days', NULL, NULL),
        ('21000000-0000-4000-8000-000000000005'::uuid, '20000000-0000-4000-8000-000000000005'::uuid, 'procurement_officer'::membership_role, 'active'::membership_status, now() - interval '26 days', NULL, NULL),
        ('21000000-0000-4000-8000-000000000006'::uuid, '20000000-0000-4000-8000-000000000006'::uuid, 'viewer'::membership_role, 'active'::membership_status, now() - interval '25 days', NULL, NULL),
        ('21000000-0000-4000-8000-000000000007'::uuid, '20000000-0000-4000-8000-000000000007'::uuid, 'viewer'::membership_role, 'invited'::membership_status, NULL, NULL, NULL),
        ('21000000-0000-4000-8000-000000000008'::uuid, '20000000-0000-4000-8000-000000000008'::uuid, 'requester'::membership_role, 'suspended'::membership_status, now() - interval '24 days', now() - interval '2 days', NULL),
        ('21000000-0000-4000-8000-000000000009'::uuid, '20000000-0000-4000-8000-000000000009'::uuid, 'approver'::membership_role, 'removed'::membership_status, now() - interval '23 days', NULL, now() - interval '1 day')
)
INSERT INTO organization_memberships (
    id, organization_id, user_id, role, status, created_by_user_id,
    invited_at, activated_at, suspended_at, removed_at, created_at, updated_at
)
SELECT
    id,
    '10000000-0000-4000-8000-000000000001',
    user_id,
    role,
    status,
    '20000000-0000-4000-8000-000000000001',
    now() - interval '30 days',
    activated_at,
    suspended_at,
    removed_at,
    now(),
    now()
FROM demo_memberships
ON CONFLICT (id) DO UPDATE SET
    role = EXCLUDED.role,
    status = EXCLUDED.status,
    activated_at = EXCLUDED.activated_at,
    suspended_at = EXCLUDED.suspended_at,
    removed_at = EXCLUDED.removed_at,
    updated_at = now();

WITH demo_vendors(id, name, legal_name, contact_name, email, phone, tax_identifier, city, country, notes, status, archived_at) AS (
    VALUES
        ('30000000-0000-4000-8000-000000000001'::uuid, 'Blue Nile Office Supply', 'Blue Nile Office Supply PLC', 'Hana Bekele', 'sales@bluenile-demo.test', '+251911000001', 'TIN-BNOS-001', 'Addis Ababa', 'ET', 'Preferred stationery and furniture vendor.', 'active'::vendor_status, NULL::timestamptz),
        ('30000000-0000-4000-8000-000000000002'::uuid, 'Sheger Technology', 'Sheger Technology Share Company', 'Mikael Tadesse', 'quotes@shegertech-demo.test', '+251911000002', 'TIN-SHEG-002', 'Addis Ababa', 'ET', 'IT equipment and accessories.', 'active'::vendor_status, NULL),
        ('30000000-0000-4000-8000-000000000003'::uuid, 'Rift Logistics', 'Rift Logistics LLC', 'Liya Ahmed', 'ops@riftlogistics-demo.test', '+251911000003', 'TIN-RIFT-003', 'Adama', 'ET', 'Logistics vendor retained for validation.', 'archived'::vendor_status, now() - interval '5 days')
)
INSERT INTO vendors (
    id, organization_id, name, legal_name, contact_name, email, phone,
    tax_identifier, address_line1, city, country, notes, status, archived_at,
    created_by_user_id, updated_by_user_id, created_at, updated_at
)
SELECT
    id,
    '10000000-0000-4000-8000-000000000001',
    name,
    legal_name,
    contact_name,
    email,
    phone,
    tax_identifier,
    'Bole Road',
    city,
    country,
    notes,
    status,
    archived_at,
    '20000000-0000-4000-8000-000000000005',
    '20000000-0000-4000-8000-000000000005',
    now(),
    now()
FROM demo_vendors
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    legal_name = EXCLUDED.legal_name,
    contact_name = EXCLUDED.contact_name,
    email = EXCLUDED.email,
    phone = EXCLUDED.phone,
    tax_identifier = EXCLUDED.tax_identifier,
    city = EXCLUDED.city,
    country = EXCLUDED.country,
    notes = EXCLUDED.notes,
    status = EXCLUDED.status,
    archived_at = EXCLUDED.archived_at,
    updated_by_user_id = EXCLUDED.updated_by_user_id,
    updated_at = now();

WITH demo_requests(id, requester_user_id, title, description, justification, status, submitted_at, submitted_by_user_id, approved_at, approved_by_user_id, rejected_at, rejected_by_user_id, decision_comment, cancelled_at, cancelled_by_user_id, amount) AS (
    VALUES
        ('40000000-0000-4000-8000-000000000001'::uuid, '20000000-0000-4000-8000-000000000003'::uuid, 'Draft laptop request', 'Laptop refresh request still being prepared.', 'New engineers need assigned laptops.', 'draft'::procurement_request_status, NULL::timestamptz, NULL::uuid, NULL::timestamptz, NULL::uuid, NULL::timestamptz, NULL::uuid, NULL::text, NULL::timestamptz, NULL::uuid, 4200.00),
        ('40000000-0000-4000-8000-000000000002'::uuid, '20000000-0000-4000-8000-000000000003'::uuid, 'Submitted monitor request', 'Pending approval for monitor replacements.', 'Customer support team needs dual monitors.', 'submitted'::procurement_request_status, now() - interval '3 days', '20000000-0000-4000-8000-000000000003'::uuid, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 1800.00),
        ('40000000-0000-4000-8000-000000000003'::uuid, '20000000-0000-4000-8000-000000000003'::uuid, 'Approved office chair request', 'Approved request used for RFQ validation.', 'Replace worn chairs in the operations area.', 'approved'::procurement_request_status, now() - interval '12 days', '20000000-0000-4000-8000-000000000003'::uuid, now() - interval '11 days', '20000000-0000-4000-8000-000000000004'::uuid, NULL, NULL, 'Approved for sourcing.', NULL, NULL, 9500.00),
        ('40000000-0000-4000-8000-000000000004'::uuid, '20000000-0000-4000-8000-000000000003'::uuid, 'Rejected tablet request', 'Rejected because shared tablets are available.', 'Field team requested spare tablets.', 'rejected'::procurement_request_status, now() - interval '10 days', '20000000-0000-4000-8000-000000000003'::uuid, NULL, NULL, now() - interval '9 days', '20000000-0000-4000-8000-000000000004'::uuid, 'Use existing inventory first.', NULL, NULL, 2400.00),
        ('40000000-0000-4000-8000-000000000005'::uuid, '20000000-0000-4000-8000-000000000003'::uuid, 'Cancelled printer request', 'Cancelled after budget reprioritization.', 'Finance office requested a printer.', 'cancelled'::procurement_request_status, now() - interval '8 days', '20000000-0000-4000-8000-000000000003'::uuid, NULL, NULL, NULL, NULL, 'Cancelled by requester.', now() - interval '7 days', '20000000-0000-4000-8000-000000000003'::uuid, 600.00)
)
INSERT INTO procurement_requests (
    id, organization_id, requester_user_id, title, description, justification,
    status, currency_code, estimated_total_amount, submitted_at, submitted_by_user_id,
    approved_at, approved_by_user_id, rejected_at, rejected_by_user_id,
    decision_comment, cancelled_at, cancelled_by_user_id, created_at, updated_at
)
SELECT
    id,
    '10000000-0000-4000-8000-000000000001',
    requester_user_id,
    title,
    description,
    justification,
    status,
    'USD',
    amount,
    submitted_at,
    submitted_by_user_id,
    approved_at,
    approved_by_user_id,
    rejected_at,
    rejected_by_user_id,
    decision_comment,
    cancelled_at,
    cancelled_by_user_id,
    now(),
    now()
FROM demo_requests
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    justification = EXCLUDED.justification,
    status = EXCLUDED.status,
    estimated_total_amount = EXCLUDED.estimated_total_amount,
    submitted_at = EXCLUDED.submitted_at,
    submitted_by_user_id = EXCLUDED.submitted_by_user_id,
    approved_at = EXCLUDED.approved_at,
    approved_by_user_id = EXCLUDED.approved_by_user_id,
    rejected_at = EXCLUDED.rejected_at,
    rejected_by_user_id = EXCLUDED.rejected_by_user_id,
    decision_comment = EXCLUDED.decision_comment,
    cancelled_at = EXCLUDED.cancelled_at,
    cancelled_by_user_id = EXCLUDED.cancelled_by_user_id,
    updated_at = now();

WITH demo_request_items(id, procurement_request_id, line_number, item_name, description, quantity, unit, unit_price) AS (
    VALUES
        ('41000000-0000-4000-8000-000000000001'::uuid, '40000000-0000-4000-8000-000000000001'::uuid, 1, 'Business laptop', '14 inch laptop with 16GB RAM.', 3.00, 'pcs', 1400.00),
        ('41000000-0000-4000-8000-000000000002'::uuid, '40000000-0000-4000-8000-000000000002'::uuid, 1, '27 inch monitor', 'USB-C display.', 6.00, 'pcs', 300.00),
        ('41000000-0000-4000-8000-000000000003'::uuid, '40000000-0000-4000-8000-000000000003'::uuid, 1, 'Ergonomic chair', 'Adjustable office chair.', 10.00, 'pcs', 850.00),
        ('41000000-0000-4000-8000-000000000004'::uuid, '40000000-0000-4000-8000-000000000003'::uuid, 2, 'Chair mat', 'Floor protection mat.', 10.00, 'pcs', 100.00),
        ('41000000-0000-4000-8000-000000000005'::uuid, '40000000-0000-4000-8000-000000000004'::uuid, 1, 'Field tablet', 'Rugged tablet.', 4.00, 'pcs', 600.00),
        ('41000000-0000-4000-8000-000000000006'::uuid, '40000000-0000-4000-8000-000000000005'::uuid, 1, 'Office printer', 'Shared workgroup printer.', 1.00, 'pcs', 600.00)
)
INSERT INTO procurement_request_items (
    id, organization_id, procurement_request_id, line_number, item_name,
    description, quantity, unit, estimated_unit_price, needed_by_date,
    created_at, updated_at
)
SELECT
    id,
    '10000000-0000-4000-8000-000000000001',
    procurement_request_id,
    line_number,
    item_name,
    description,
    quantity,
    unit,
    unit_price,
    current_date + 21,
    now(),
    now()
FROM demo_request_items
ON CONFLICT (id) DO UPDATE SET
    item_name = EXCLUDED.item_name,
    description = EXCLUDED.description,
    quantity = EXCLUDED.quantity,
    unit = EXCLUDED.unit,
    estimated_unit_price = EXCLUDED.estimated_unit_price,
    needed_by_date = EXCLUDED.needed_by_date,
    updated_at = now();

WITH demo_rfqs(id, reference_number, title, status, published_at, closed_at, evaluated_at, cancelled_at, cancelled_by_user_id) AS (
    VALUES
        ('50000000-0000-4000-8000-000000000001'::uuid, 'DEMO-RFQ-DRAFT', 'Draft RFQ for office chairs', 'draft'::rfq_status, NULL::timestamptz, NULL::timestamptz, NULL::timestamptz, NULL::timestamptz, NULL::uuid),
        ('50000000-0000-4000-8000-000000000002'::uuid, 'DEMO-RFQ-PUBLISHED', 'Published RFQ for office chairs', 'published'::rfq_status, now() - interval '6 days', NULL, NULL, NULL, NULL),
        ('50000000-0000-4000-8000-000000000003'::uuid, 'DEMO-RFQ-CLOSED', 'Closed RFQ for office chairs', 'closed'::rfq_status, now() - interval '10 days', now() - interval '4 days', NULL, NULL, NULL),
        ('50000000-0000-4000-8000-000000000004'::uuid, 'DEMO-RFQ-EVALUATED', 'Evaluated RFQ for office chairs', 'evaluated'::rfq_status, now() - interval '14 days', now() - interval '8 days', now() - interval '3 days', NULL, NULL),
        ('50000000-0000-4000-8000-000000000005'::uuid, 'DEMO-RFQ-AWARDED', 'Awarded RFQ for office chairs', 'awarded'::rfq_status, now() - interval '20 days', now() - interval '15 days', now() - interval '12 days', NULL, NULL),
        ('50000000-0000-4000-8000-000000000006'::uuid, 'DEMO-RFQ-CANCELLED', 'Cancelled RFQ for office chairs', 'cancelled'::rfq_status, now() - interval '16 days', NULL, NULL, now() - interval '13 days', '20000000-0000-4000-8000-000000000005'::uuid)
)
INSERT INTO rfqs (
    id, organization_id, procurement_request_id, reference_number, title,
    description, status, created_by_user_id, published_at, closed_at,
    evaluated_at, cancelled_at, cancelled_by_user_id, created_at, updated_at
)
SELECT
    id,
    '10000000-0000-4000-8000-000000000001',
    '40000000-0000-4000-8000-000000000003',
    reference_number,
    title,
    'Demo RFQ seeded for workflow validation.',
    status,
    '20000000-0000-4000-8000-000000000005',
    published_at,
    closed_at,
    evaluated_at,
    cancelled_at,
    cancelled_by_user_id,
    now(),
    now()
FROM demo_rfqs
ON CONFLICT (id) DO UPDATE SET
    reference_number = EXCLUDED.reference_number,
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    published_at = EXCLUDED.published_at,
    closed_at = EXCLUDED.closed_at,
    evaluated_at = EXCLUDED.evaluated_at,
    cancelled_at = EXCLUDED.cancelled_at,
    cancelled_by_user_id = EXCLUDED.cancelled_by_user_id,
    updated_at = now();

WITH demo_rfq_items(rfq_id, suffix, line_number, source_request_item_id, item_name, description, quantity, unit) AS (
    SELECT rfq_id, suffix, line_number, source_request_item_id, item_name, description, quantity, unit
    FROM (VALUES
        ('50000000-0000-4000-8000-000000000001'::uuid, '001', 1, '41000000-0000-4000-8000-000000000003'::uuid, 'Ergonomic chair', 'Adjustable office chair.', 10.00, 'pcs'),
        ('50000000-0000-4000-8000-000000000001'::uuid, '002', 2, '41000000-0000-4000-8000-000000000004'::uuid, 'Chair mat', 'Floor protection mat.', 10.00, 'pcs'),
        ('50000000-0000-4000-8000-000000000002'::uuid, '003', 1, '41000000-0000-4000-8000-000000000003'::uuid, 'Ergonomic chair', 'Adjustable office chair.', 10.00, 'pcs'),
        ('50000000-0000-4000-8000-000000000002'::uuid, '004', 2, '41000000-0000-4000-8000-000000000004'::uuid, 'Chair mat', 'Floor protection mat.', 10.00, 'pcs'),
        ('50000000-0000-4000-8000-000000000003'::uuid, '005', 1, '41000000-0000-4000-8000-000000000003'::uuid, 'Ergonomic chair', 'Adjustable office chair.', 10.00, 'pcs'),
        ('50000000-0000-4000-8000-000000000004'::uuid, '006', 1, '41000000-0000-4000-8000-000000000003'::uuid, 'Ergonomic chair', 'Adjustable office chair.', 10.00, 'pcs'),
        ('50000000-0000-4000-8000-000000000005'::uuid, '007', 1, '41000000-0000-4000-8000-000000000003'::uuid, 'Ergonomic chair', 'Adjustable office chair.', 10.00, 'pcs'),
        ('50000000-0000-4000-8000-000000000006'::uuid, '008', 1, '41000000-0000-4000-8000-000000000003'::uuid, 'Ergonomic chair', 'Adjustable office chair.', 10.00, 'pcs')
    ) AS rows(rfq_id, suffix, line_number, source_request_item_id, item_name, description, quantity, unit)
)
INSERT INTO rfq_items (
    id, organization_id, rfq_id, source_request_item_id, line_number,
    item_name, description, quantity, unit, target_date, created_at, updated_at
)
SELECT
    ('51000000-0000-4000-8000-000000000' || suffix)::uuid,
    '10000000-0000-4000-8000-000000000001',
    rfq_id,
    source_request_item_id,
    line_number,
    item_name,
    description,
    quantity,
    unit,
    current_date + 30,
    now(),
    now()
FROM demo_rfq_items
ON CONFLICT (id) DO UPDATE SET
    source_request_item_id = EXCLUDED.source_request_item_id,
    line_number = EXCLUDED.line_number,
    item_name = EXCLUDED.item_name,
    description = EXCLUDED.description,
    quantity = EXCLUDED.quantity,
    unit = EXCLUDED.unit,
    target_date = EXCLUDED.target_date,
    updated_at = now();

WITH demo_rfq_vendors(id, rfq_id, vendor_id) AS (
    VALUES
        ('52000000-0000-4000-8000-000000000001'::uuid, '50000000-0000-4000-8000-000000000001'::uuid, '30000000-0000-4000-8000-000000000001'::uuid),
        ('52000000-0000-4000-8000-000000000002'::uuid, '50000000-0000-4000-8000-000000000002'::uuid, '30000000-0000-4000-8000-000000000001'::uuid),
        ('52000000-0000-4000-8000-000000000003'::uuid, '50000000-0000-4000-8000-000000000002'::uuid, '30000000-0000-4000-8000-000000000002'::uuid),
        ('52000000-0000-4000-8000-000000000004'::uuid, '50000000-0000-4000-8000-000000000003'::uuid, '30000000-0000-4000-8000-000000000001'::uuid),
        ('52000000-0000-4000-8000-000000000005'::uuid, '50000000-0000-4000-8000-000000000004'::uuid, '30000000-0000-4000-8000-000000000001'::uuid),
        ('52000000-0000-4000-8000-000000000006'::uuid, '50000000-0000-4000-8000-000000000004'::uuid, '30000000-0000-4000-8000-000000000002'::uuid),
        ('52000000-0000-4000-8000-000000000007'::uuid, '50000000-0000-4000-8000-000000000005'::uuid, '30000000-0000-4000-8000-000000000001'::uuid),
        ('52000000-0000-4000-8000-000000000008'::uuid, '50000000-0000-4000-8000-000000000006'::uuid, '30000000-0000-4000-8000-000000000001'::uuid)
)
INSERT INTO rfq_vendors (
    id, organization_id, rfq_id, vendor_id, attached_by_user_id,
    attached_at, created_at
)
SELECT
    id,
    '10000000-0000-4000-8000-000000000001',
    rfq_id,
    vendor_id,
    '20000000-0000-4000-8000-000000000005',
    now() - interval '6 days',
    now()
FROM demo_rfq_vendors
ON CONFLICT (id) DO UPDATE SET
    vendor_id = EXCLUDED.vendor_id,
    attached_by_user_id = EXCLUDED.attached_by_user_id,
    attached_at = EXCLUDED.attached_at;

WITH demo_quotations(id, rfq_id, rfq_vendor_id, status, lead_time_days, payment_terms, notes, submitted_at, rejected_at, rejection_reason) AS (
    VALUES
        ('60000000-0000-4000-8000-000000000001'::uuid, '50000000-0000-4000-8000-000000000002'::uuid, '52000000-0000-4000-8000-000000000002'::uuid, 'draft'::quotation_status, 14, 'Net 30', 'Draft quote still being edited.', NULL::timestamptz, NULL::timestamptz, NULL::text),
        ('60000000-0000-4000-8000-000000000002'::uuid, '50000000-0000-4000-8000-000000000002'::uuid, '52000000-0000-4000-8000-000000000003'::uuid, 'submitted'::quotation_status, 10, 'Net 15', 'Submitted quote for comparison.', now() - interval '2 days', NULL, NULL),
        ('60000000-0000-4000-8000-000000000003'::uuid, '50000000-0000-4000-8000-000000000004'::uuid, '52000000-0000-4000-8000-000000000005'::uuid, 'rejected'::quotation_status, 20, 'Due on receipt', 'Rejected after evaluation.', now() - interval '5 days', now() - interval '3 days', 'Lead time was too long.'),
        ('60000000-0000-4000-8000-000000000004'::uuid, '50000000-0000-4000-8000-000000000004'::uuid, '52000000-0000-4000-8000-000000000006'::uuid, 'submitted'::quotation_status, 7, 'Net 30', 'Shortlisted submitted quote.', now() - interval '4 days', NULL, NULL),
        ('60000000-0000-4000-8000-000000000005'::uuid, '50000000-0000-4000-8000-000000000005'::uuid, '52000000-0000-4000-8000-000000000007'::uuid, 'submitted'::quotation_status, 5, 'Net 30', 'Awarded quote.', now() - interval '14 days', NULL, NULL)
)
INSERT INTO quotations (
    id, organization_id, rfq_id, rfq_vendor_id, status, currency_code,
    lead_time_days, payment_terms, notes, submitted_at, submitted_by_user_id,
    rejected_at, rejected_by_user_id, rejection_reason, created_by_user_id,
    updated_by_user_id, created_at, updated_at
)
SELECT
    id,
    '10000000-0000-4000-8000-000000000001',
    rfq_id,
    rfq_vendor_id,
    status,
    'USD',
    lead_time_days,
    payment_terms,
    notes,
    submitted_at,
    CASE WHEN submitted_at IS NULL THEN NULL ELSE '20000000-0000-4000-8000-000000000005'::uuid END,
    rejected_at,
    CASE WHEN rejected_at IS NULL THEN NULL ELSE '20000000-0000-4000-8000-000000000005'::uuid END,
    rejection_reason,
    '20000000-0000-4000-8000-000000000005',
    '20000000-0000-4000-8000-000000000005',
    now(),
    now()
FROM demo_quotations
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    lead_time_days = EXCLUDED.lead_time_days,
    payment_terms = EXCLUDED.payment_terms,
    notes = EXCLUDED.notes,
    submitted_at = EXCLUDED.submitted_at,
    submitted_by_user_id = EXCLUDED.submitted_by_user_id,
    rejected_at = EXCLUDED.rejected_at,
    rejected_by_user_id = EXCLUDED.rejected_by_user_id,
    rejection_reason = EXCLUDED.rejection_reason,
    updated_by_user_id = EXCLUDED.updated_by_user_id,
    updated_at = now();

WITH demo_quotation_items(id, quotation_id, rfq_id, rfq_item_id, line_number, item_name, description, quantity, unit, unit_price, delivery_days, notes) AS (
    VALUES
        ('61000000-0000-4000-8000-000000000001'::uuid, '60000000-0000-4000-8000-000000000001'::uuid, '50000000-0000-4000-8000-000000000002'::uuid, '51000000-0000-4000-8000-000000000003'::uuid, 1, 'Ergonomic chair', 'Adjustable office chair.', 10.00, 'pcs', 820.00, 14, 'Draft unit pricing.'),
        ('61000000-0000-4000-8000-000000000002'::uuid, '60000000-0000-4000-8000-000000000001'::uuid, '50000000-0000-4000-8000-000000000002'::uuid, '51000000-0000-4000-8000-000000000004'::uuid, 2, 'Chair mat', 'Floor protection mat.', 10.00, 'pcs', 95.00, 14, 'Draft accessory pricing.'),
        ('61000000-0000-4000-8000-000000000003'::uuid, '60000000-0000-4000-8000-000000000002'::uuid, '50000000-0000-4000-8000-000000000002'::uuid, '51000000-0000-4000-8000-000000000003'::uuid, 1, 'Ergonomic chair', 'Adjustable office chair.', 10.00, 'pcs', 790.00, 10, 'Submitted chair pricing.'),
        ('61000000-0000-4000-8000-000000000004'::uuid, '60000000-0000-4000-8000-000000000002'::uuid, '50000000-0000-4000-8000-000000000002'::uuid, '51000000-0000-4000-8000-000000000004'::uuid, 2, 'Chair mat', 'Floor protection mat.', 10.00, 'pcs', 88.00, 10, 'Submitted mat pricing.'),
        ('61000000-0000-4000-8000-000000000005'::uuid, '60000000-0000-4000-8000-000000000003'::uuid, '50000000-0000-4000-8000-000000000004'::uuid, '51000000-0000-4000-8000-000000000006'::uuid, 1, 'Ergonomic chair', 'Adjustable office chair.', 10.00, 'pcs', 810.00, 20, 'Rejected quote pricing.'),
        ('61000000-0000-4000-8000-000000000006'::uuid, '60000000-0000-4000-8000-000000000004'::uuid, '50000000-0000-4000-8000-000000000004'::uuid, '51000000-0000-4000-8000-000000000006'::uuid, 1, 'Ergonomic chair', 'Adjustable office chair.', 10.00, 'pcs', 780.00, 7, 'Shortlisted pricing.'),
        ('61000000-0000-4000-8000-000000000007'::uuid, '60000000-0000-4000-8000-000000000005'::uuid, '50000000-0000-4000-8000-000000000005'::uuid, '51000000-0000-4000-8000-000000000007'::uuid, 1, 'Ergonomic chair', 'Adjustable office chair.', 10.00, 'pcs', 770.00, 5, 'Awarded pricing.')
)
INSERT INTO quotation_items (
    id, organization_id, quotation_id, rfq_id, rfq_item_id, line_number,
    item_name, description, quantity, unit, unit_price, delivery_days,
    notes, created_at, updated_at
)
SELECT
    id,
    '10000000-0000-4000-8000-000000000001',
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
    notes,
    now(),
    now()
FROM demo_quotation_items
ON CONFLICT (id) DO UPDATE SET
    item_name = EXCLUDED.item_name,
    description = EXCLUDED.description,
    quantity = EXCLUDED.quantity,
    unit = EXCLUDED.unit,
    unit_price = EXCLUDED.unit_price,
    delivery_days = EXCLUDED.delivery_days,
    notes = EXCLUDED.notes,
    updated_at = now();

INSERT INTO rfq_awards (
    id, organization_id, rfq_id, quotation_id, awarded_by_user_id,
    reason, awarded_at, created_at
)
VALUES (
    '70000000-0000-4000-8000-000000000001',
    '10000000-0000-4000-8000-000000000001',
    '50000000-0000-4000-8000-000000000005',
    '60000000-0000-4000-8000-000000000005',
    '20000000-0000-4000-8000-000000000005',
    'Best price and fastest delivery among submitted quotations.',
    now() - interval '11 days',
    now()
)
ON CONFLICT (id) DO UPDATE SET
    quotation_id = EXCLUDED.quotation_id,
    awarded_by_user_id = EXCLUDED.awarded_by_user_id,
    reason = EXCLUDED.reason,
    awarded_at = EXCLUDED.awarded_at;

WITH demo_logs(id, actor_user_id, entity_type, entity_id, action, summary, metadata, occurred_at) AS (
    VALUES
        ('80000000-0000-4000-8000-000000000001'::uuid, '20000000-0000-4000-8000-000000000001'::uuid, 'organization', '10000000-0000-4000-8000-000000000001'::uuid, 'organization.created', 'Demo organization created.', '{"source":"demo_seed"}'::jsonb, now() - interval '30 days'),
        ('80000000-0000-4000-8000-000000000002'::uuid, '20000000-0000-4000-8000-000000000001'::uuid, 'membership', '21000000-0000-4000-8000-000000000001'::uuid, 'membership.created', 'Owner membership created.', '{"role":"owner","status":"active"}'::jsonb, now() - interval '30 days'),
        ('80000000-0000-4000-8000-000000000003'::uuid, '20000000-0000-4000-8000-000000000005'::uuid, 'vendor', '30000000-0000-4000-8000-000000000001'::uuid, 'vendor.created', 'Active demo vendor created.', '{"status":"active"}'::jsonb, now() - interval '22 days'),
        ('80000000-0000-4000-8000-000000000004'::uuid, '20000000-0000-4000-8000-000000000003'::uuid, 'procurement_request', '40000000-0000-4000-8000-000000000003'::uuid, 'procurement_request.created', 'Approved demo request created.', '{"request_status":"approved"}'::jsonb, now() - interval '12 days'),
        ('80000000-0000-4000-8000-000000000005'::uuid, '20000000-0000-4000-8000-000000000004'::uuid, 'procurement_request', '40000000-0000-4000-8000-000000000003'::uuid, 'procurement_request.approved', 'Approved for sourcing.', '{"request_status":"approved"}'::jsonb, now() - interval '11 days'),
        ('80000000-0000-4000-8000-000000000006'::uuid, '20000000-0000-4000-8000-000000000005'::uuid, 'rfq', '50000000-0000-4000-8000-000000000002'::uuid, 'rfq.published', 'Published RFQ for validation.', '{"rfq_status":"published"}'::jsonb, now() - interval '6 days'),
        ('80000000-0000-4000-8000-000000000007'::uuid, '20000000-0000-4000-8000-000000000005'::uuid, 'quotation', '60000000-0000-4000-8000-000000000002'::uuid, 'quotation.submitted', 'Submitted comparison quotation.', '{"quotation_status":"submitted"}'::jsonb, now() - interval '2 days'),
        ('80000000-0000-4000-8000-000000000008'::uuid, '20000000-0000-4000-8000-000000000005'::uuid, 'award', '70000000-0000-4000-8000-000000000001'::uuid, 'award.created', 'Demo award created.', '{"rfq_id":"50000000-0000-4000-8000-000000000005","quotation_id":"60000000-0000-4000-8000-000000000005"}'::jsonb, now() - interval '11 days'),
        ('80000000-0000-4000-8000-000000000009'::uuid, '20000000-0000-4000-8000-000000000005'::uuid, 'rfq', '50000000-0000-4000-8000-000000000005'::uuid, 'rfq.awarded', 'Awarded RFQ marked for validation.', '{"rfq_status":"awarded"}'::jsonb, now() - interval '11 days')
)
INSERT INTO activity_logs (
    id, organization_id, actor_user_id, entity_type, entity_id,
    action, summary, metadata, occurred_at
)
SELECT
    id,
    '10000000-0000-4000-8000-000000000001',
    actor_user_id,
    entity_type,
    entity_id,
    action,
    summary,
    metadata,
    occurred_at
FROM demo_logs
ON CONFLICT (id) DO UPDATE SET
    actor_user_id = EXCLUDED.actor_user_id,
    entity_type = EXCLUDED.entity_type,
    entity_id = EXCLUDED.entity_id,
    action = EXCLUDED.action,
    summary = EXCLUDED.summary,
    metadata = EXCLUDED.metadata,
    occurred_at = EXCLUDED.occurred_at;
`
