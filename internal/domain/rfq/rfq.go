package rfq

import (
	"time"

	domainvendor "github.com/AbenezerWork/ProcureFlow/internal/domain/vendor"
	"github.com/google/uuid"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusClosed    Status = "closed"
	StatusEvaluated Status = "evaluated"
	StatusAwarded   Status = "awarded"
	StatusCancelled Status = "cancelled"
)

type RFQ struct {
	ID                   uuid.UUID
	OrganizationID       uuid.UUID
	ProcurementRequestID uuid.UUID
	ReferenceNumber      *string
	Title                string
	Description          *string
	Status               Status
	CreatedByUserID      uuid.UUID
	PublishedAt          *time.Time
	ClosedAt             *time.Time
	EvaluatedAt          *time.Time
	CancelledAt          *time.Time
	CancelledByUserID    *uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type Item struct {
	ID                  uuid.UUID
	OrganizationID      uuid.UUID
	RFQID               uuid.UUID
	SourceRequestItemID *uuid.UUID
	LineNumber          int32
	ItemName            string
	Description         *string
	Quantity            string
	Unit                string
	TargetDate          *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type VendorLink struct {
	ID               uuid.UUID
	OrganizationID   uuid.UUID
	RFQID            uuid.UUID
	VendorID         uuid.UUID
	AttachedByUserID uuid.UUID
	AttachedAt       time.Time
	CreatedAt        time.Time
	VendorName       string
	VendorStatus     domainvendor.Status
}
