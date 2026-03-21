package activitylog

type EntityType string

const (
	EntityTypeOrganization       EntityType = "organization"
	EntityTypeMembership         EntityType = "membership"
	EntityTypeVendor             EntityType = "vendor"
	EntityTypeProcurementRequest EntityType = "procurement_request"
	EntityTypeRFQ                EntityType = "rfq"
	EntityTypeQuotation          EntityType = "quotation"
	EntityTypeAward              EntityType = "award"
)

const (
	ActionOrganizationCreated           = "organization.created"
	ActionOrganizationUpdated           = "organization.updated"
	ActionOrganizationOwnershipMoved    = "organization.ownership_transferred"
	ActionMembershipCreated             = "membership.created"
	ActionMembershipUpdated             = "membership.updated"
	ActionMembershipOwnershipGranted    = "membership.ownership_granted"
	ActionMembershipOwnershipRevoked    = "membership.ownership_revoked"
	ActionVendorCreated                 = "vendor.created"
	ActionVendorUpdated                 = "vendor.updated"
	ActionVendorArchived                = "vendor.archived"
	ActionProcurementRequestCreated     = "procurement_request.created"
	ActionProcurementRequestUpdated     = "procurement_request.updated"
	ActionProcurementRequestItemAdded   = "procurement_request.item_added"
	ActionProcurementRequestItemUpdated = "procurement_request.item_updated"
	ActionProcurementRequestItemDeleted = "procurement_request.item_deleted"
	ActionProcurementRequestSubmitted   = "procurement_request.submitted"
	ActionProcurementRequestApproved    = "procurement_request.approved"
	ActionProcurementRequestRejected    = "procurement_request.rejected"
	ActionProcurementRequestCanceled    = "procurement_request.cancelled"
	ActionRFQCreated                    = "rfq.created"
	ActionRFQUpdated                    = "rfq.updated"
	ActionRFQPublished                  = "rfq.published"
	ActionRFQClosed                     = "rfq.closed"
	ActionRFQEvaluated                  = "rfq.evaluated"
	ActionRFQCanceled                   = "rfq.cancelled"
	ActionRFQAwarded                    = "rfq.awarded"
	ActionRFQVendorAttached             = "rfq.vendor_attached"
	ActionRFQVendorRemoved              = "rfq.vendor_removed"
	ActionQuotationCreated              = "quotation.created"
	ActionQuotationUpdated              = "quotation.updated"
	ActionQuotationSubmitted            = "quotation.submitted"
	ActionQuotationRejected             = "quotation.rejected"
	ActionQuotationItemUpdated          = "quotation.item_updated"
	ActionAwardCreated                  = "award.created"
)

func IsKnownEntityType(value string) bool {
	switch EntityType(value) {
	case EntityTypeOrganization,
		EntityTypeMembership,
		EntityTypeVendor,
		EntityTypeProcurementRequest,
		EntityTypeRFQ,
		EntityTypeQuotation,
		EntityTypeAward:
		return true
	default:
		return false
	}
}
