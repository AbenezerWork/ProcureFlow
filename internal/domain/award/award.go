package award

import (
	"time"

	"github.com/google/uuid"
)

type Award struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	RFQID           uuid.UUID
	QuotationID     uuid.UUID
	AwardedByUserID uuid.UUID
	Reason          string
	AwardedAt       time.Time
	CreatedAt       time.Time
}
