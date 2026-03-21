package activitylog

import (
	"context"
	"errors"
	"testing"
	"time"

	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	"github.com/google/uuid"
)

type fakeRepository struct {
	createActivityLogFn func(context.Context, CreateParams) (domainactivitylog.Entry, error)
	listByEntityFn      func(context.Context, uuid.UUID, string, uuid.UUID) ([]domainactivitylog.Entry, error)
	getMembershipFn     func(context.Context, uuid.UUID, uuid.UUID) (domainorganization.Membership, error)
}

func (f fakeRepository) CreateActivityLog(ctx context.Context, params CreateParams) (domainactivitylog.Entry, error) {
	if f.createActivityLogFn == nil {
		return domainactivitylog.Entry{}, nil
	}
	return f.createActivityLogFn(ctx, params)
}

func (f fakeRepository) ListByEntity(ctx context.Context, organizationID uuid.UUID, entityType string, entityID uuid.UUID) ([]domainactivitylog.Entry, error) {
	return f.listByEntityFn(ctx, organizationID, entityType, entityID)
}

func (f fakeRepository) GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error) {
	return f.getMembershipFn(ctx, organizationID, userID)
}

func TestServiceListByEntityRejectsUnknownEntityType(t *testing.T) {
	t.Parallel()

	service := NewService(fakeRepository{
		getMembershipFn: func(context.Context, uuid.UUID, uuid.UUID) (domainorganization.Membership, error) {
			t.Fatal("membership lookup should not be called")
			return domainorganization.Membership{}, nil
		},
		listByEntityFn: func(context.Context, uuid.UUID, string, uuid.UUID) ([]domainactivitylog.Entry, error) {
			t.Fatal("list should not be called")
			return nil, nil
		},
	})

	_, err := service.ListByEntity(context.Background(), ListByEntityInput{
		OrganizationID: uuid.New(),
		CurrentUser:    uuid.New(),
		EntityType:     "invoice",
		EntityID:       uuid.New(),
	})
	if !errors.Is(err, ErrInvalidActivityLog) {
		t.Fatalf("expected invalid activity log error, got %v", err)
	}
}

func TestServiceListByEntityAllowsKnownEntityType(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	userID := uuid.New()
	entityID := uuid.New()
	now := time.Now().UTC()

	service := NewService(fakeRepository{
		getMembershipFn: func(_ context.Context, gotOrgID, gotUserID uuid.UUID) (domainorganization.Membership, error) {
			if gotOrgID != orgID || gotUserID != userID {
				t.Fatalf("unexpected membership lookup")
			}
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleViewer,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		listByEntityFn: func(_ context.Context, gotOrgID uuid.UUID, gotEntityType string, gotEntityID uuid.UUID) ([]domainactivitylog.Entry, error) {
			if gotOrgID != orgID || gotEntityID != entityID {
				t.Fatalf("unexpected list args")
			}
			if gotEntityType != string(domainactivitylog.EntityTypeProcurementRequest) {
				t.Fatalf("unexpected entity type: %s", gotEntityType)
			}
			return []domainactivitylog.Entry{{
				ID:             uuid.New(),
				OrganizationID: orgID,
				EntityType:     gotEntityType,
				EntityID:       entityID,
				Action:         domainactivitylog.ActionProcurementRequestSubmitted,
				OccurredAt:     now,
			}}, nil
		},
	})

	entries, err := service.ListByEntity(context.Background(), ListByEntityInput{
		OrganizationID: orgID,
		CurrentUser:    userID,
		EntityType:     string(domainactivitylog.EntityTypeProcurementRequest),
		EntityID:       entityID,
	})
	if err != nil {
		t.Fatalf("list activity logs returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 activity log entry, got %d", len(entries))
	}
}
