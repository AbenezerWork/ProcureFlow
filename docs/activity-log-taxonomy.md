# Activity Log Taxonomy

This document defines the canonical event vocabulary for ProcureFlow activity logs.

The goal is to keep timeline rendering, filtering, and analytics stable across slices by standardizing:

- `entity_type`
- `action`
- `summary`
- expected `metadata` keys

## Naming rules

- `entity_type` is a singular snake_case business object name.
- `action` is `<entity_type>.<past_tense_event>`.
- `summary` is a short human-readable sentence describing what happened.
- `metadata` contains machine-friendly supporting fields and should use snake_case keys.

## Canonical entity types

- `organization`
- `membership`
- `vendor`
- `procurement_request`
- `rfq`
- `quotation`
- `award`

## Event catalog

### Organizations

- `organization.created`
  - Summary example: `Created organization`
  - Metadata: `slug`
- `organization.updated`
  - Summary example: `Updated organization details`
  - Metadata: changed fields as needed
- `organization.ownership_transferred`
  - Summary example: `Transferred organization ownership`
  - Metadata: `previous_owner_user_id`, `new_owner_user_id`

### Memberships

- `membership.created`
  - Summary example: `Created organization membership`
  - Metadata: `user_id`, `role`, `status`
- `membership.updated`
  - Summary example: `Updated organization membership`
  - Metadata: changed `role` and/or `status`
- `membership.ownership_granted`
  - Summary example: `Granted organization ownership`
  - Metadata: `user_id`
- `membership.ownership_revoked`
  - Summary example: `Revoked organization ownership`
  - Metadata: `user_id`, `new_role`

### Vendors

- `vendor.created`
  - Summary example: `Created vendor`
  - Metadata: `vendor_name`
- `vendor.updated`
  - Summary example: `Updated vendor`
  - Metadata: changed fields as needed
- `vendor.archived`
  - Summary example: `Archived vendor`
  - Metadata: `vendor_name`

### Procurement Requests

- `procurement_request.created`
  - Summary example: `Created procurement request`
  - Metadata: `requester_user_id`, `request_status`
- `procurement_request.updated`
  - Summary example: `Updated procurement request`
  - Metadata: `request_status`
- `procurement_request.item_added`
  - Summary example: `Added procurement request item`
  - Metadata: `item_id`, `line_number`, `item_name`
- `procurement_request.item_updated`
  - Summary example: `Updated procurement request item`
  - Metadata: `item_id`, `line_number`, `item_name`
- `procurement_request.item_deleted`
  - Summary example: `Deleted procurement request item`
  - Metadata: `item_id`
- `procurement_request.submitted`
  - Summary example: `Submitted procurement request`
  - Metadata: `request_status`
- `procurement_request.approved`
  - Summary example: `Approved procurement request`
  - Metadata: `request_status`, optional `decision_comment`
- `procurement_request.rejected`
  - Summary example: `Rejected procurement request`
  - Metadata: `request_status`, optional `decision_comment`
- `procurement_request.cancelled`
  - Summary example: `Cancelled procurement request`
  - Metadata: `request_status`

### RFQs

- `rfq.created`
  - Summary example: `Created RFQ from approved procurement request`
  - Metadata: `procurement_request_id`, `item_count`
- `rfq.updated`
  - Summary example: `Updated RFQ`
  - Metadata: `reference_number`, `title`
- `rfq.published`
  - Summary example: `Published RFQ`
  - Metadata: `rfq_status`
- `rfq.closed`
  - Summary example: `Closed RFQ`
  - Metadata: `rfq_status`
- `rfq.evaluated`
  - Summary example: `Evaluated RFQ`
  - Metadata: `rfq_status`
- `rfq.cancelled`
  - Summary example: `Cancelled RFQ`
  - Metadata: `rfq_status`
- `rfq.awarded`
  - Summary example: `Awarded RFQ`
  - Metadata: `award_id`, `quotation_id`, `award_reason`
- `rfq.vendor_attached`
  - Summary example: `Attached vendor to RFQ`
  - Metadata: `vendor_id`, `vendor_name`
- `rfq.vendor_removed`
  - Summary example: `Removed vendor from RFQ`
  - Metadata: `vendor_id`

### Quotations

- `quotation.created`
  - Summary example: `Created quotation`
  - Metadata: `rfq_id`, `rfq_vendor_id`
- `quotation.updated`
  - Summary example: `Updated quotation`
  - Metadata: `currency_code`, `rfq_id`
- `quotation.submitted`
  - Summary example: `Submitted quotation`
  - Metadata: `rfq_id`, `rfq_vendor_id`, `quotation_status`
- `quotation.rejected`
  - Summary example: `Rejected quotation`
  - Metadata: `rfq_id`, `rfq_vendor_id`, `quotation_status`, optional `rejection_reason`
- `quotation.item_updated`
  - Summary example: `Updated quotation item`
  - Metadata: `item_id`, `line_number`, `unit_price`

### Awards

- `award.created`
  - Summary example: `Created award`
  - Metadata: `rfq_id`, `quotation_id`

## Current implementation status

The codebase currently emits these taxonomy events:

- `organization.created`
- `membership.created`
- `organization.updated`
- `membership.updated`
- `organization.ownership_transferred`
- `membership.ownership_granted`
- `membership.ownership_revoked`
- `vendor.created`
- `vendor.updated`
- `vendor.archived`
- `procurement_request.created`
- `procurement_request.updated`
- `procurement_request.item_added`
- `procurement_request.item_updated`
- `procurement_request.item_deleted`
- `procurement_request.submitted`
- `procurement_request.approved`
- `procurement_request.rejected`
- `procurement_request.cancelled`
- `rfq.created`
- `rfq.updated`
- `rfq.published`
- `rfq.closed`
- `rfq.evaluated`
- `rfq.cancelled`
- `rfq.vendor_attached`
- `rfq.vendor_removed`
- `quotation.created`
- `quotation.updated`
- `quotation.submitted`
- `quotation.rejected`
- `quotation.item_updated`
- `award.created`
- `rfq.awarded`

## Integration model

Activity-log writes currently persist to Postgres first and then pass through optional hooks.

- Hook contract: `internal/application/activitylog.Hook`
- Hook helper: `internal/application/activitylog.NotifyHooks`
- Database write gateway: `internal/infrastructure/database/repositories/activity_log_repository.go`

Repository constructors now accept optional activity-log hooks, so future integrations can subscribe without changing the application-service contracts again. That makes the current Postgres timeline the primary sink while leaving room for forwarding to systems such as message buses, webhooks, search indexes, or analytics pipelines.
