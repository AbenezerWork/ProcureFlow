package activitylog

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	"github.com/google/uuid"
)

var (
	ErrInvalidActivityLog   = errors.New("invalid activity log")
	ErrForbiddenActivityLog = errors.New("forbidden activity log operation")
	ErrMembershipNotFound   = errors.New("organization membership not found")
)

type CreateParams struct {
	OrganizationID uuid.UUID
	ActorUserID    *uuid.UUID
	EntityType     string
	EntityID       uuid.UUID
	Action         string
	Summary        *string
	Metadata       map[string]any
}

type Hook interface {
	HandleActivityLog(ctx context.Context, entry domainactivitylog.Entry) error
}

type Repository interface {
	CreateActivityLog(ctx context.Context, params CreateParams) (domainactivitylog.Entry, error)
	ListByEntity(ctx context.Context, organizationID uuid.UUID, entityType string, entityID uuid.UUID) ([]domainactivitylog.Entry, error)
	GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error)
}

type ListByEntityInput struct {
	OrganizationID uuid.UUID
	CurrentUser    uuid.UUID
	EntityType     string
	EntityID       uuid.UUID
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) ListByEntity(ctx context.Context, input ListByEntityInput) ([]domainactivitylog.Entry, error) {
	if input.OrganizationID == uuid.Nil || input.CurrentUser == uuid.Nil || input.EntityID == uuid.Nil {
		return nil, ErrInvalidActivityLog
	}

	entityType := strings.TrimSpace(input.EntityType)
	if entityType == "" || !domainactivitylog.IsKnownEntityType(entityType) {
		return nil, ErrInvalidActivityLog
	}

	membership, err := s.repo.GetMembership(ctx, input.OrganizationID, input.CurrentUser)
	if err != nil {
		if errors.Is(err, ErrMembershipNotFound) {
			return nil, ErrForbiddenActivityLog
		}
		return nil, fmt.Errorf("load membership: %w", err)
	}
	if membership.Status != domainorganization.MembershipStatusActive {
		return nil, ErrForbiddenActivityLog
	}

	entries, err := s.repo.ListByEntity(ctx, input.OrganizationID, entityType, input.EntityID)
	if err != nil {
		return nil, fmt.Errorf("list activity logs: %w", err)
	}

	return entries, nil
}
