package repositories

import (
	"context"
	"encoding/json"
	"errors"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ActivityLogRepository struct {
	store *database.Store
	hooks []applicationactivitylog.Hook
}

func NewActivityLogRepository(store *database.Store, hooks ...applicationactivitylog.Hook) *ActivityLogRepository {
	return &ActivityLogRepository{store: store, hooks: hooks}
}

func (r *ActivityLogRepository) CreateActivityLog(ctx context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error) {
	return createActivityLog(ctx, r.store, params, r.hooks...)
}

func createActivityLog(ctx context.Context, store *database.Store, params applicationactivitylog.CreateParams, hooks ...applicationactivitylog.Hook) (domainactivitylog.Entry, error) {
	metadata, err := json.Marshal(params.Metadata)
	if err != nil {
		return domainactivitylog.Entry{}, err
	}
	if metadata == nil {
		metadata = []byte(`{}`)
	}

	entry, err := store.CreateActivityLog(ctx, sqlc.CreateActivityLogParams{
		OrganizationID: params.OrganizationID,
		ActorUserID:    params.ActorUserID,
		EntityType:     params.EntityType,
		EntityID:       params.EntityID,
		Action:         params.Action,
		Summary:        params.Summary,
		Metadata:       metadata,
	})
	if err != nil {
		return domainactivitylog.Entry{}, err
	}

	mapped := mapActivityLog(entry)
	if err := applicationactivitylog.NotifyHooks(ctx, mapped, hooks...); err != nil {
		return domainactivitylog.Entry{}, err
	}

	return mapped, nil
}

func (r *ActivityLogRepository) ListByEntity(ctx context.Context, organizationID uuid.UUID, entityType string, entityID uuid.UUID) ([]domainactivitylog.Entry, error) {
	rows, err := r.store.ListActivityLogsByEntity(ctx, sqlc.ListActivityLogsByEntityParams{
		OrganizationID: organizationID,
		EntityType:     entityType,
		EntityID:       entityID,
	})
	if err != nil {
		return nil, err
	}

	entries := make([]domainactivitylog.Entry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, mapActivityLog(row))
	}

	return entries, nil
}

func (r *ActivityLogRepository) GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error) {
	membership, err := r.store.GetOrganizationMembership(ctx, sqlc.GetOrganizationMembershipParams{
		OrganizationID: organizationID,
		UserID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainorganization.Membership{}, applicationactivitylog.ErrMembershipNotFound
		}
		return domainorganization.Membership{}, err
	}

	return mapMembership(membership), nil
}

func mapActivityLog(entry sqlc.ActivityLog) domainactivitylog.Entry {
	metadata := map[string]any{}
	if len(entry.Metadata) > 0 {
		_ = json.Unmarshal(entry.Metadata, &metadata)
	}

	return domainactivitylog.Entry{
		ID:             entry.ID,
		OrganizationID: entry.OrganizationID,
		ActorUserID:    entry.ActorUserID,
		EntityType:     entry.EntityType,
		EntityID:       entry.EntityID,
		Action:         entry.Action,
		Summary:        entry.Summary,
		Metadata:       metadata,
		OccurredAt:     requiredTime(entry.OccurredAt),
	}
}
