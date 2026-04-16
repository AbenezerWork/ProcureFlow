package quotation

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusSubmitted Status = "submitted"
	StatusRejected  Status = "rejected"
)

type Quotation struct {
	ID                uuid.UUID
	OrganizationID    uuid.UUID
	RFQID             uuid.UUID
	RFQVendorID       uuid.UUID
	Status            Status
	CurrencyCode      string
	LeadTimeDays      *int32
	PaymentTerms      *string
	Notes             *string
	SubmittedAt       *time.Time
	SubmittedByUserID *uuid.UUID
	RejectedAt        *time.Time
	RejectedByUserID  *uuid.UUID
	RejectionReason   *string
	CreatedByUserID   uuid.UUID
	UpdatedByUserID   *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	VendorName        *string
}

type Item struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	QuotationID    uuid.UUID
	RFQID          uuid.UUID
	RFQItemID      uuid.UUID
	LineNumber     int32
	ItemName       string
	Description    *string
	Quantity       string
	Unit           string
	UnitPrice      string
	DeliveryDays   *int32
	Notes          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Comparison struct {
	RFQID      uuid.UUID
	Quotations []ComparisonQuotation
	LineItems  []ComparisonLineItem
}

type ComparisonQuotation struct {
	QuotationID  uuid.UUID
	RFQVendorID  uuid.UUID
	VendorID     uuid.UUID
	VendorName   string
	Status       Status
	CurrencyCode string
	LeadTimeDays *int32
	PaymentTerms *string
	Notes        *string
	TotalAmount  string
}

type ComparisonLineItem struct {
	RFQItemID   uuid.UUID
	LineNumber  int32
	ItemName    string
	Description *string
	Quantity    string
	Unit        string
	Prices      []ComparisonPrice
}

type ComparisonPrice struct {
	QuotationID     uuid.UUID
	QuotationItemID uuid.UUID
	VendorID        uuid.UUID
	VendorName      string
	UnitPrice       string
	LineTotal       string
	DeliveryDays    *int32
	Notes           *string
}

type ComparisonRow struct {
	QuotationID     uuid.UUID
	RFQID           uuid.UUID
	RFQVendorID     uuid.UUID
	Status          Status
	CurrencyCode    string
	LeadTimeDays    *int32
	PaymentTerms    *string
	QuotationNotes  *string
	VendorID        uuid.UUID
	VendorName      string
	TotalAmount     string
	QuotationItemID uuid.UUID
	RFQItemID       uuid.UUID
	LineNumber      int32
	ItemName        string
	Description     *string
	Quantity        string
	Unit            string
	UnitPrice       string
	LineTotal       string
	DeliveryDays    *int32
	ItemNotes       *string
}
