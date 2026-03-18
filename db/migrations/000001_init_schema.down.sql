BEGIN;

DROP TABLE IF EXISTS activity_logs;
DROP TABLE IF EXISTS rfq_awards;
DROP TABLE IF EXISTS quotation_items;
DROP TABLE IF EXISTS quotations;
DROP TABLE IF EXISTS rfq_vendors;
DROP TABLE IF EXISTS rfq_items;
DROP TABLE IF EXISTS rfqs;
DROP TABLE IF EXISTS procurement_request_items;
DROP TABLE IF EXISTS procurement_requests;
DROP TABLE IF EXISTS vendors;
DROP TABLE IF EXISTS organization_memberships;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS quotation_status;
DROP TYPE IF EXISTS rfq_status;
DROP TYPE IF EXISTS procurement_request_status;
DROP TYPE IF EXISTS vendor_status;
DROP TYPE IF EXISTS membership_status;
DROP TYPE IF EXISTS membership_role;

COMMIT;
