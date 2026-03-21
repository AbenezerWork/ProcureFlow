package award

import (
	"context"
	"errors"
	"fmt"
	"strings"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainaward "github.com/AbenezerWork/ProcureFlow/internal/domain/award"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainquotation "github.com/AbenezerWork/ProcureFlow/internal/domain/quotation"
	domainrfq "github.com/AbenezerWork/ProcureFlow/internal/domain/rfq"
	"github.com/google/uuid"
)

var (
	ErrInvalidAward       = errors.New("invalid award")
	ErrAwardNotFound      = errors.New("award not found")
	ErrAwardExists        = errors.New("award already exists")
	ErrMembershipNotFound = errors.New("organization membership not found")
	ErrRFQNotFound        = errors.New("rfq not found")
	ErrQuotationNotFound  = errors.New("quotation not found")
	ErrForbiddenAward     = errors.New("forbidden award operation")
)

type CreateAwardParams struct {
	OrganizationID  uuid.UUID
	RFQID           uuid.UUID
	QuotationID     uuid.UUID
	AwardedByUserID uuid.UUID
	Reason          string
}

type Repository interface {
	CreateAward(ctx context.Context, params CreateAwardParams) (domainaward.Award, error)
	GetAwardByRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainaward.Award, error)
	GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error)
	GetRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error)
	GetQuotation(ctx context.Context, organizationID, quotationID uuid.UUID) (domainquotation.Quotation, error)
	MarkRFQAwarded(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error)
	CreateActivityLog(ctx context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error)
}

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(repo Repository) error) error
}

type CreateInput struct {
	OrganizationID uuid.UUID
	RFQID          uuid.UUID
	QuotationID    uuid.UUID
	CurrentUser    uuid.UUID
	Reason         string
}

type Service struct {
	repo Repository
	tx   TransactionManager
}

func NewService(repo Repository, tx TransactionManager) Service {
	return Service{repo: repo, tx: tx}
}

func (s Service) Create(ctx context.Context, input CreateInput) (domainaward.Award, error) {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.QuotationID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainaward.Award{}, ErrInvalidAward
	}

	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return domainaward.Award{}, ErrInvalidAward
	}

	rfq, membership, err := s.loadManagedRFQ(ctx, input.OrganizationID, input.RFQID, input.CurrentUser)
	if err != nil {
		return domainaward.Award{}, err
	}
	if !canManageAward(membership.Role) {
		return domainaward.Award{}, ErrForbiddenAward
	}
	if rfq.Status != domainrfq.StatusEvaluated {
		return domainaward.Award{}, ErrInvalidAward
	}

	quotation, err := s.repo.GetQuotation(ctx, input.OrganizationID, input.QuotationID)
	if err != nil {
		if errors.Is(err, ErrQuotationNotFound) {
			return domainaward.Award{}, err
		}
		return domainaward.Award{}, fmt.Errorf("load quotation: %w", err)
	}
	if quotation.RFQID != input.RFQID || quotation.Status != domainquotation.StatusSubmitted {
		return domainaward.Award{}, ErrInvalidAward
	}

	var created domainaward.Award
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		award, err := repo.CreateAward(ctx, CreateAwardParams{
			OrganizationID:  input.OrganizationID,
			RFQID:           input.RFQID,
			QuotationID:     input.QuotationID,
			AwardedByUserID: input.CurrentUser,
			Reason:          reason,
		})
		if err != nil {
			return err
		}

		if _, err := repo.MarkRFQAwarded(ctx, input.OrganizationID, input.RFQID); err != nil {
			return err
		}

		summary := "Created award"
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeAward),
			EntityID:       award.ID,
			Action:         domainactivitylog.ActionAwardCreated,
			Summary:        &summary,
			Metadata: map[string]any{
				"rfq_id":       input.RFQID.String(),
				"quotation_id": input.QuotationID.String(),
			},
		}); err != nil {
			return err
		}

		rfqSummary := "Awarded RFQ"
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeRFQ),
			EntityID:       input.RFQID,
			Action:         domainactivitylog.ActionRFQAwarded,
			Summary:        &rfqSummary,
			Metadata: map[string]any{
				"award_id":     award.ID.String(),
				"quotation_id": input.QuotationID.String(),
				"award_reason": reason,
			},
		}); err != nil {
			return err
		}

		created = award
		return nil
	}); err != nil {
		if errors.Is(err, ErrAwardExists) {
			return domainaward.Award{}, err
		}
		if errors.Is(err, ErrRFQNotFound) {
			return domainaward.Award{}, err
		}
		return domainaward.Award{}, fmt.Errorf("create award: %w", err)
	}

	return created, nil
}

func (s Service) GetByRFQ(ctx context.Context, organizationID, rfqID, currentUser uuid.UUID) (domainaward.Award, error) {
	if organizationID == uuid.Nil || rfqID == uuid.Nil || currentUser == uuid.Nil {
		return domainaward.Award{}, ErrInvalidAward
	}
	if _, err := s.loadActiveMembership(ctx, organizationID, currentUser); err != nil {
		return domainaward.Award{}, err
	}

	award, err := s.repo.GetAwardByRFQ(ctx, organizationID, rfqID)
	if err != nil {
		if errors.Is(err, ErrAwardNotFound) {
			return domainaward.Award{}, err
		}
		return domainaward.Award{}, fmt.Errorf("get award: %w", err)
	}

	return award, nil
}

func (s Service) loadActiveMembership(ctx context.Context, organizationID, currentUser uuid.UUID) (domainorganization.Membership, error) {
	membership, err := s.repo.GetMembership(ctx, organizationID, currentUser)
	if err != nil {
		if errors.Is(err, ErrMembershipNotFound) {
			return domainorganization.Membership{}, ErrForbiddenAward
		}
		return domainorganization.Membership{}, fmt.Errorf("load membership: %w", err)
	}
	if membership.Status != domainorganization.MembershipStatusActive {
		return domainorganization.Membership{}, ErrForbiddenAward
	}

	return membership, nil
}

func (s Service) loadManagedRFQ(ctx context.Context, organizationID, rfqID, currentUser uuid.UUID) (domainrfq.RFQ, domainorganization.Membership, error) {
	membership, err := s.loadActiveMembership(ctx, organizationID, currentUser)
	if err != nil {
		return domainrfq.RFQ{}, domainorganization.Membership{}, err
	}

	rfq, err := s.repo.GetRFQ(ctx, organizationID, rfqID)
	if err != nil {
		if errors.Is(err, ErrRFQNotFound) {
			return domainrfq.RFQ{}, domainorganization.Membership{}, err
		}
		return domainrfq.RFQ{}, domainorganization.Membership{}, fmt.Errorf("load rfq: %w", err)
	}

	return rfq, membership, nil
}

func canManageAward(role domainorganization.MembershipRole) bool {
	switch role {
	case domainorganization.MembershipRoleOwner,
		domainorganization.MembershipRoleAdmin,
		domainorganization.MembershipRoleProcurementOfficer:
		return true
	default:
		return false
	}
}
