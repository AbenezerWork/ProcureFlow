package activitylog

import (
	"time"

	"github.com/google/uuid"
)

type Entry struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	ActorUserID    *uuid.UUID
	EntityType     string
	EntityID       uuid.UUID
	Action         string
	Summary        *string
	Metadata       map[string]any
	OccurredAt     time.Time
}
