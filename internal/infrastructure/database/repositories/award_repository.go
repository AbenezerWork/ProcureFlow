package repositories

import (
	"context"
	"errors"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	applicationaward "github.com/AbenezerWork/ProcureFlow/internal/application/award"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainaward "github.com/AbenezerWork/ProcureFlow/internal/domain/award"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainquotation "github.com/AbenezerWork/ProcureFlow/internal/domain/quotation"
	domainrfq "github.com/AbenezerWork/ProcureFlow/internal/domain/rfq"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AwardRepository struct {
	store *database.Store
	hooks []applicationactivitylog.Hook
}

func NewAwardRepository(store *database.Store, hooks ...applicationactivitylog.Hook) *AwardRepository {
	return &AwardRepository{store: store, hooks: hooks}
}

func (r *AwardRepository) CreateAward(ctx context.Context, params applicationaward.CreateAwardParams) (domainaward.Award, error) {
	award, err := r.store.CreateRFQAward(ctx, sqlc.CreateRFQAwardParams{
		OrganizationID:  params.OrganizationID,
		RfqID:           params.RFQID,
		QuotationID:     params.QuotationID,
		AwardedByUserID: params.AwardedByUserID,
		Reason:          params.Reason,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domainaward.Award{}, applicationaward.ErrAwardExists
		}
		return domainaward.Award{}, err
	}

	return mapAward(award), nil
}

func (r *AwardRepository) GetAwardByRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainaward.Award, error) {
	award, err := r.store.GetRFQAwardByRFQ(ctx, sqlc.GetRFQAwardByRFQParams{
		OrganizationID: organizationID,
		RfqID:          rfqID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainaward.Award{}, applicationaward.ErrAwardNotFound
		}
		return domainaward.Award{}, err
	}

	return mapAward(award), nil
}

func (r *AwardRepository) GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error) {
	membership, err := r.store.GetOrganizationMembership(ctx, sqlc.GetOrganizationMembershipParams{
		OrganizationID: organizationID,
		UserID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainorganization.Membership{}, applicationaward.ErrMembershipNotFound
		}
		return domainorganization.Membership{}, err
	}

	return mapMembership(membership), nil
}

func (r *AwardRepository) GetRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	rfq, err := r.store.GetRFQByID(ctx, sqlc.GetRFQByIDParams{
		OrganizationID: organizationID,
		ID:             rfqID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrfq.RFQ{}, applicationaward.ErrRFQNotFound
		}
		return domainrfq.RFQ{}, err
	}

	return mapRFQ(rfq), nil
}

func (r *AwardRepository) GetQuotation(ctx context.Context, organizationID, quotationID uuid.UUID) (domainquotation.Quotation, error) {
	quotation, err := r.store.GetQuotationByID(ctx, sqlc.GetQuotationByIDParams{
		OrganizationID: organizationID,
		ID:             quotationID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainquotation.Quotation{}, applicationaward.ErrQuotationNotFound
		}
		return domainquotation.Quotation{}, err
	}

	return mapQuotation(quotation, nil), nil
}

func (r *AwardRepository) MarkRFQAwarded(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	rfq, err := r.store.MarkRFQAwarded(ctx, sqlc.MarkRFQAwardedParams{
		OrganizationID: organizationID,
		ID:             rfqID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrfq.RFQ{}, applicationaward.ErrRFQNotFound
		}
		return domainrfq.RFQ{}, err
	}

	return mapRFQ(rfq), nil
}

func (r *AwardRepository) CreateActivityLog(ctx context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error) {
	return createActivityLog(ctx, r.store, params, r.hooks...)
}

func (r *AwardRepository) WithinTransaction(ctx context.Context, fn func(repo applicationaward.Repository) error) error {
	return r.store.InTx(ctx, func(txStore *database.Store) error {
		return fn(NewAwardRepository(txStore, r.hooks...))
	})
}

func mapAward(award sqlc.RfqAward) domainaward.Award {
	return domainaward.Award{
		ID:              award.ID,
		OrganizationID:  award.OrganizationID,
		RFQID:           award.RfqID,
		QuotationID:     award.QuotationID,
		AwardedByUserID: award.AwardedByUserID,
		Reason:          award.Reason,
		AwardedAt:       requiredTime(award.AwardedAt),
		CreatedAt:       requiredTime(award.CreatedAt),
	}
}
