package vendor

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

type Vendor struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	Name            string
	LegalName       *string
	ContactName     *string
	Email           *string
	Phone           *string
	TaxIdentifier   *string
	AddressLine1    *string
	AddressLine2    *string
	City            *string
	StateRegion     *string
	PostalCode      *string
	Country         *string
	Notes           *string
	Status          Status
	ArchivedAt      *time.Time
	CreatedByUserID uuid.UUID
	UpdatedByUserID *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
