package repositories

import (
	"context"
	"errors"

	applicationorganization "github.com/AbenezerWork/ProcureFlow/internal/application/organization"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type OrganizationRepository struct {
	store *database.Store
}

func NewOrganizationRepository(store *database.Store) *OrganizationRepository {
	return &OrganizationRepository{store: store}
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

func (r *OrganizationRepository) CreateMembership(ctx context.Context, params applicationorganization.CreateMembershipParams) (domainorganization.Membership, error) {
	membership, err := r.store.CreateOrganizationMembership(ctx, sqlc.CreateOrganizationMembershipParams{
		OrganizationID:  params.OrganizationID,
		UserID:          params.UserID,
		Role:            sqlc.MembershipRole(params.Role),
		Status:          sqlc.MembershipStatus(params.Status),
		CreatedByUserID: params.CreatedByUserID,
	})
	if err != nil {
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

func (r *OrganizationRepository) WithinTransaction(ctx context.Context, fn func(repo applicationorganization.Repository) error) error {
	return r.store.InTx(ctx, func(txStore *database.Store) error {
		return fn(NewOrganizationRepository(txStore))
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
