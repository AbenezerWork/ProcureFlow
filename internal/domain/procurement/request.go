package procurement

import (
	"time"

	"github.com/google/uuid"
)

type RequestStatus string

const (
	RequestStatusDraft     RequestStatus = "draft"
	RequestStatusSubmitted RequestStatus = "submitted"
	RequestStatusApproved  RequestStatus = "approved"
	RequestStatusRejected  RequestStatus = "rejected"
	RequestStatusCancelled RequestStatus = "cancelled"
)

type Request struct {
	ID                   uuid.UUID
	OrganizationID       uuid.UUID
	RequesterUserID      uuid.UUID
	Title                string
	Description          *string
	Justification        *string
	Status               RequestStatus
	CurrencyCode         string
	EstimatedTotalAmount *string
	SubmittedAt          *time.Time
	SubmittedByUserID    *uuid.UUID
	ApprovedAt           *time.Time
	ApprovedByUserID     *uuid.UUID
	RejectedAt           *time.Time
	RejectedByUserID     *uuid.UUID
	DecisionComment      *string
	CancelledAt          *time.Time
	CancelledByUserID    *uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type Item struct {
	ID                   uuid.UUID
	OrganizationID       uuid.UUID
	ProcurementRequestID uuid.UUID
	LineNumber           int32
	ItemName             string
	Description          *string
	Quantity             string
	Unit                 string
	EstimatedUnitPrice   *string
	NeededByDate         *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
