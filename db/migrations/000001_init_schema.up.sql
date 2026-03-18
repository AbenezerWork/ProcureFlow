BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE membership_role AS ENUM (
    'owner',
    'admin',
    'requester',
    'approver',
    'procurement_officer',
    'viewer'
);

CREATE TYPE membership_status AS ENUM (
    'invited',
    'active',
    'suspended',
    'removed'
);

CREATE TYPE vendor_status AS ENUM (
    'active',
    'archived'
);

CREATE TYPE procurement_request_status AS ENUM (
    'draft',
    'submitted',
    'approved',
    'rejected',
    'cancelled'
);

CREATE TYPE rfq_status AS ENUM (
    'draft',
    'published',
    'closed',
    'evaluated',
    'awarded',
    'cancelled'
);

CREATE TYPE quotation_status AS ENUM (
    'draft',
    'submitted',
    'rejected'
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (btrim(email) <> ''),
    CHECK (btrim(full_name) <> '')
);

CREATE UNIQUE INDEX users_email_unique_idx ON users (lower(email));

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at TIMESTAMPTZ,
    CHECK (btrim(name) <> ''),
    CHECK (btrim(slug) <> '')
);

CREATE UNIQUE INDEX organizations_slug_unique_idx ON organizations (lower(slug));

CREATE TABLE organization_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role membership_role NOT NULL,
    status membership_status NOT NULL DEFAULT 'invited',
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    invited_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    activated_at TIMESTAMPTZ,
    suspended_at TIMESTAMPTZ,
    removed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organization_id, user_id),
    UNIQUE (organization_id, id)
);

CREATE INDEX organization_memberships_user_idx ON organization_memberships (user_id, status);
CREATE INDEX organization_memberships_org_role_idx ON organization_memberships (organization_id, role, status);

CREATE TABLE vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    legal_name TEXT,
    contact_name TEXT,
    email TEXT,
    phone TEXT,
    tax_identifier TEXT,
    address_line1 TEXT,
    address_line2 TEXT,
    city TEXT,
    state_region TEXT,
    postal_code TEXT,
    country TEXT,
    notes TEXT,
    status vendor_status NOT NULL DEFAULT 'active',
    archived_at TIMESTAMPTZ,
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    updated_by_user_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (btrim(name) <> ''),
    UNIQUE (organization_id, id)
);

CREATE INDEX vendors_org_status_idx ON vendors (organization_id, status);
CREATE INDEX vendors_org_name_idx ON vendors (organization_id, lower(name));

CREATE TABLE procurement_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    requester_user_id UUID NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    description TEXT,
    justification TEXT,
    status procurement_request_status NOT NULL DEFAULT 'draft',
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    estimated_total_amount NUMERIC(18, 2),
    submitted_at TIMESTAMPTZ,
    submitted_by_user_id UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    approved_by_user_id UUID REFERENCES users(id),
    rejected_at TIMESTAMPTZ,
    rejected_by_user_id UUID REFERENCES users(id),
    decision_comment TEXT,
    cancelled_at TIMESTAMPTZ,
    cancelled_by_user_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (btrim(title) <> ''),
    CHECK (estimated_total_amount IS NULL OR estimated_total_amount >= 0),
    UNIQUE (organization_id, id)
);

CREATE INDEX procurement_requests_org_status_idx ON procurement_requests (organization_id, status, created_at DESC);
CREATE INDEX procurement_requests_org_requester_idx ON procurement_requests (organization_id, requester_user_id, created_at DESC);

CREATE TABLE procurement_request_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    procurement_request_id UUID NOT NULL,
    line_number INTEGER NOT NULL,
    item_name TEXT NOT NULL,
    description TEXT,
    quantity NUMERIC(14, 2) NOT NULL,
    unit TEXT NOT NULL,
    estimated_unit_price NUMERIC(18, 2),
    needed_by_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (btrim(item_name) <> ''),
    CHECK (btrim(unit) <> ''),
    CHECK (quantity > 0),
    CHECK (estimated_unit_price IS NULL OR estimated_unit_price >= 0),
    UNIQUE (organization_id, id),
    UNIQUE (procurement_request_id, line_number),
    FOREIGN KEY (organization_id, procurement_request_id)
        REFERENCES procurement_requests (organization_id, id)
        ON DELETE CASCADE
);

CREATE INDEX procurement_request_items_request_idx ON procurement_request_items (procurement_request_id, line_number);

CREATE TABLE rfqs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    procurement_request_id UUID NOT NULL,
    reference_number TEXT,
    title TEXT NOT NULL,
    description TEXT,
    status rfq_status NOT NULL DEFAULT 'draft',
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    published_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    evaluated_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    cancelled_by_user_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (btrim(title) <> ''),
    UNIQUE (organization_id, id),
    FOREIGN KEY (organization_id, procurement_request_id)
        REFERENCES procurement_requests (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX rfqs_org_status_idx ON rfqs (organization_id, status, created_at DESC);
CREATE INDEX rfqs_org_request_idx ON rfqs (organization_id, procurement_request_id);
CREATE UNIQUE INDEX rfqs_org_reference_number_idx
    ON rfqs (organization_id, reference_number)
    WHERE reference_number IS NOT NULL;

CREATE TABLE rfq_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    rfq_id UUID NOT NULL,
    source_request_item_id UUID,
    line_number INTEGER NOT NULL,
    item_name TEXT NOT NULL,
    description TEXT,
    quantity NUMERIC(14, 2) NOT NULL,
    unit TEXT NOT NULL,
    target_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (btrim(item_name) <> ''),
    CHECK (btrim(unit) <> ''),
    CHECK (quantity > 0),
    UNIQUE (organization_id, id),
    UNIQUE (rfq_id, id),
    UNIQUE (rfq_id, line_number),
    FOREIGN KEY (organization_id, rfq_id)
        REFERENCES rfqs (organization_id, id)
        ON DELETE CASCADE,
    FOREIGN KEY (organization_id, source_request_item_id)
        REFERENCES procurement_request_items (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX rfq_items_rfq_idx ON rfq_items (rfq_id, line_number);

CREATE TABLE rfq_vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    rfq_id UUID NOT NULL,
    vendor_id UUID NOT NULL,
    attached_by_user_id UUID NOT NULL REFERENCES users(id),
    attached_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organization_id, id),
    UNIQUE (rfq_id, id),
    UNIQUE (rfq_id, vendor_id),
    FOREIGN KEY (organization_id, rfq_id)
        REFERENCES rfqs (organization_id, id)
        ON DELETE CASCADE,
    FOREIGN KEY (organization_id, vendor_id)
        REFERENCES vendors (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX rfq_vendors_vendor_idx ON rfq_vendors (vendor_id);

CREATE TABLE quotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    rfq_id UUID NOT NULL,
    rfq_vendor_id UUID NOT NULL,
    status quotation_status NOT NULL DEFAULT 'draft',
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    lead_time_days INTEGER,
    payment_terms TEXT,
    notes TEXT,
    submitted_at TIMESTAMPTZ,
    submitted_by_user_id UUID REFERENCES users(id),
    rejected_at TIMESTAMPTZ,
    rejected_by_user_id UUID REFERENCES users(id),
    rejection_reason TEXT,
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    updated_by_user_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (lead_time_days IS NULL OR lead_time_days >= 0),
    UNIQUE (organization_id, id),
    UNIQUE (rfq_id, id),
    UNIQUE (rfq_vendor_id),
    FOREIGN KEY (organization_id, rfq_id)
        REFERENCES rfqs (organization_id, id)
        ON DELETE CASCADE,
    FOREIGN KEY (rfq_id, rfq_vendor_id)
        REFERENCES rfq_vendors (rfq_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX quotations_rfq_status_idx ON quotations (rfq_id, status, created_at DESC);

CREATE TABLE quotation_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    quotation_id UUID NOT NULL,
    rfq_id UUID NOT NULL,
    rfq_item_id UUID NOT NULL,
    line_number INTEGER NOT NULL,
    item_name TEXT NOT NULL,
    description TEXT,
    quantity NUMERIC(14, 2) NOT NULL,
    unit TEXT NOT NULL,
    unit_price NUMERIC(18, 2) NOT NULL,
    delivery_days INTEGER,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (btrim(item_name) <> ''),
    CHECK (btrim(unit) <> ''),
    CHECK (quantity > 0),
    CHECK (unit_price >= 0),
    CHECK (delivery_days IS NULL OR delivery_days >= 0),
    UNIQUE (organization_id, id),
    UNIQUE (quotation_id, line_number),
    UNIQUE (quotation_id, rfq_item_id),
    FOREIGN KEY (organization_id, quotation_id)
        REFERENCES quotations (organization_id, id)
        ON DELETE CASCADE,
    FOREIGN KEY (rfq_id, quotation_id)
        REFERENCES quotations (rfq_id, id)
        ON DELETE CASCADE,
    FOREIGN KEY (rfq_id, rfq_item_id)
        REFERENCES rfq_items (rfq_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX quotation_items_quotation_idx ON quotation_items (quotation_id, line_number);

CREATE TABLE rfq_awards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    rfq_id UUID NOT NULL,
    quotation_id UUID NOT NULL,
    awarded_by_user_id UUID NOT NULL REFERENCES users(id),
    reason TEXT NOT NULL,
    awarded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (btrim(reason) <> ''),
    UNIQUE (organization_id, id),
    UNIQUE (rfq_id),
    UNIQUE (quotation_id),
    FOREIGN KEY (organization_id, rfq_id)
        REFERENCES rfqs (organization_id, id)
        ON DELETE CASCADE,
    FOREIGN KEY (rfq_id, quotation_id)
        REFERENCES quotations (rfq_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX rfq_awards_rfq_idx ON rfq_awards (rfq_id);

CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    action TEXT NOT NULL,
    summary TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (btrim(entity_type) <> ''),
    CHECK (btrim(action) <> ''),
    CHECK (jsonb_typeof(metadata) = 'object')
);

CREATE INDEX activity_logs_entity_timeline_idx
    ON activity_logs (organization_id, entity_type, entity_id, occurred_at DESC);
CREATE INDEX activity_logs_org_occurred_idx
    ON activity_logs (organization_id, occurred_at DESC);

COMMIT;
