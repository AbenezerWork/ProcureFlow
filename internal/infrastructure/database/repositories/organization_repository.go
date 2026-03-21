package repositories

import (
	"context"
	"errors"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	applicationorganization "github.com/AbenezerWork/ProcureFlow/internal/application/organization"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type OrganizationRepository struct {
	store *database.Store
	hooks []applicationactivitylog.Hook
}

func NewOrganizationRepository(store *database.Store, hooks ...applicationactivitylog.Hook) *OrganizationRepository {
	return &OrganizationRepository{store: store, hooks: hooks}
}

func (r *OrganizationRepository) CreateOrganization(ctx context.Context, params applicationorganization.CreateOrganizationParams) (domainorganization.Organization, error) {
	org, err := r.store.CreateOrganization(ctx, sqlc.CreateOrganizationParams{
		Name:            params.Name,
		Slug:            params.Slug,
		CreatedByUserID: params.CreatedByUserID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainorganization.Organization{}, applicationorganization.ErrOrganizationSlugTaken
		}

		return domainorganization.Organization{}, err
	}

	return mapOrganization(org), nil
}

func (r *OrganizationRepository) GetOrganization(ctx context.Context, organizationID uuid.UUID) (domainorganization.Organization, error) {
	org, err := r.store.GetOrganizationByID(ctx, organizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainorganization.Organization{}, applicationorganization.ErrOrganizationNotFound
		}

		return domainorganization.Organization{}, err
	}

	return mapOrganization(org), nil
}

func (r *OrganizationRepository) UpdateOrganization(ctx context.Context, params applicationorganization.UpdateOrganizationParams) (domainorganization.Organization, error) {
	org, err := r.store.UpdateOrganization(ctx, sqlc.UpdateOrganizationParams{
		ID:   params.ID,
		Name: params.Name,
		Slug: params.Slug,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainorganization.Organization{}, applicationorganization.ErrOrganizationNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainorganization.Organization{}, applicationorganization.ErrOrganizationSlugTaken
		}

		return domainorganization.Organization{}, err
	}

	return mapOrganization(org), nil
}

func (r *OrganizationRepository) CreateMembership(ctx context.Context, params applicationorganization.CreateMembershipParams) (domainorganization.Membership, error) {
	membership, err := r.store.CreateOrganizationMembership(ctx, sqlc.CreateOrganizationMembershipParams{
		OrganizationID:  params.OrganizationID,
		UserID:          params.UserID,
		Role:            sqlc.MembershipRole(params.Role),
		Status:          sqlc.MembershipStatus(params.Status),
		CreatedByUserID: params.CreatedByUserID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainorganization.Membership{}, applicationorganization.ErrMembershipAlreadyExists
		}

		return domainorganization.Membership{}, err
	}

	return mapMembership(membership), nil
}

func (r *OrganizationRepository) GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error) {
	membership, err := r.store.GetOrganizationMembership(ctx, sqlc.GetOrganizationMembershipParams{
		OrganizationID: organizationID,
		UserID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainorganization.Membership{}, applicationorganization.ErrMembershipNotFound
		}

		return domainorganization.Membership{}, err
	}

	return mapMembership(membership), nil
}

func (r *OrganizationRepository) ListMemberships(ctx context.Context, organizationID uuid.UUID) ([]domainorganization.Membership, error) {
	rows, err := r.store.ListOrganizationMemberships(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	memberships := make([]domainorganization.Membership, 0, len(rows))
	for _, membership := range rows {
		memberships = append(memberships, mapMembership(membership))
	}

	return memberships, nil
}

func (r *OrganizationRepository) UpdateMembershipRole(ctx context.Context, organizationID, userID uuid.UUID, role domainorganization.MembershipRole) (domainorganization.Membership, error) {
	membership, err := r.store.UpdateMembershipRole(ctx, sqlc.UpdateMembershipRoleParams{
		OrganizationID: organizationID,
		UserID:         userID,
		Role:           sqlc.MembershipRole(role),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainorganization.Membership{}, applicationorganization.ErrMembershipNotFound
		}

		return domainorganization.Membership{}, err
	}

	return mapMembership(membership), nil
}

func (r *OrganizationRepository) UpdateMembershipStatus(ctx context.Context, organizationID, userID uuid.UUID, status domainorganization.MembershipStatus) (domainorganization.Membership, error) {
	membership, err := r.store.UpdateMembershipStatus(ctx, sqlc.UpdateMembershipStatusParams{
		OrganizationID: organizationID,
		UserID:         userID,
		Status:         sqlc.MembershipStatus(status),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainorganization.Membership{}, applicationorganization.ErrMembershipNotFound
		}

		return domainorganization.Membership{}, err
	}

	return mapMembership(membership), nil
}

func (r *OrganizationRepository) ListUserOrganizations(ctx context.Context, userID uuid.UUID) ([]domainorganization.UserOrganization, error) {
	rows, err := r.store.ListUserOrganizations(ctx, userID)
	if err != nil {
		return nil, err
	}

	organizations := make([]domainorganization.UserOrganization, 0, len(rows))
	for _, row := range rows {
		organizations = append(organizations, domainorganization.UserOrganization{
			Organization: domainorganization.Organization{
				ID:              row.ID,
				Name:            row.Name,
				Slug:            row.Slug,
				CreatedByUserID: row.CreatedByUserID,
				CreatedAt:       requiredTime(row.CreatedAt),
				UpdatedAt:       requiredTime(row.UpdatedAt),
				ArchivedAt:      optionalTime(row.ArchivedAt),
			},
			Role:   domainorganization.MembershipRole(row.MembershipRole),
			Status: domainorganization.MembershipStatus(row.MembershipStatus),
		})
	}

	return organizations, nil
}

func (r *OrganizationRepository) CreateActivityLog(ctx context.Context, params applicationorganization.CreateActivityLogParams) (domainactivitylog.Entry, error) {
	return createActivityLog(ctx, r.store, applicationactivitylog.CreateParams(params), r.hooks...)
}

func (r *OrganizationRepository) WithinTransaction(ctx context.Context, fn func(repo applicationorganization.Repository) error) error {
	return r.store.InTx(ctx, func(txStore *database.Store) error {
		return fn(NewOrganizationRepository(txStore, r.hooks...))
	})
}

func mapOrganization(org sqlc.Organization) domainorganization.Organization {
	return domainorganization.Organization{
		ID:              org.ID,
		Name:            org.Name,
		Slug:            org.Slug,
		CreatedByUserID: org.CreatedByUserID,
		CreatedAt:       requiredTime(org.CreatedAt),
		UpdatedAt:       requiredTime(org.UpdatedAt),
		ArchivedAt:      optionalTime(org.ArchivedAt),
	}
}

func mapMembership(membership sqlc.OrganizationMembership) domainorganization.Membership {
	return domainorganization.Membership{
		ID:              membership.ID,
		OrganizationID:  membership.OrganizationID,
		UserID:          membership.UserID,
		Role:            domainorganization.MembershipRole(membership.Role),
		Status:          domainorganization.MembershipStatus(membership.Status),
		CreatedByUserID: membership.CreatedByUserID,
		InvitedAt:       requiredTime(membership.InvitedAt),
		ActivatedAt:     optionalTime(membership.ActivatedAt),
		SuspendedAt:     optionalTime(membership.SuspendedAt),
		RemovedAt:       optionalTime(membership.RemovedAt),
		CreatedAt:       requiredTime(membership.CreatedAt),
		UpdatedAt:       requiredTime(membership.UpdatedAt),
	}
}
