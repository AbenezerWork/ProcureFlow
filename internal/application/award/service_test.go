package award

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainaward "github.com/AbenezerWork/ProcureFlow/internal/domain/award"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainquotation "github.com/AbenezerWork/ProcureFlow/internal/domain/quotation"
	domainrfq "github.com/AbenezerWork/ProcureFlow/internal/domain/rfq"
	"github.com/google/uuid"
)

type fakeRepository struct {
	createAwardFn       func(context.Context, CreateAwardParams) (domainaward.Award, error)
	getAwardByRFQFn     func(context.Context, uuid.UUID, uuid.UUID) (domainaward.Award, error)
	getMembershipFn     func(context.Context, uuid.UUID, uuid.UUID) (domainorganization.Membership, error)
	getRFQFn            func(context.Context, uuid.UUID, uuid.UUID) (domainrfq.RFQ, error)
	getQuotationFn      func(context.Context, uuid.UUID, uuid.UUID) (domainquotation.Quotation, error)
	markRFQAwardedFn    func(context.Context, uuid.UUID, uuid.UUID) (domainrfq.RFQ, error)
	createActivityLogFn func(context.Context, applicationactivitylog.CreateParams) (domainactivitylog.Entry, error)
}

func (f fakeRepository) CreateAward(ctx context.Context, params CreateAwardParams) (domainaward.Award, error) {
	return f.createAwardFn(ctx, params)
}

func (f fakeRepository) GetAwardByRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainaward.Award, error) {
	return f.getAwardByRFQFn(ctx, organizationID, rfqID)
}

func (f fakeRepository) GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error) {
	return f.getMembershipFn(ctx, organizationID, userID)
}

func (f fakeRepository) GetRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	return f.getRFQFn(ctx, organizationID, rfqID)
}

func (f fakeRepository) GetQuotation(ctx context.Context, organizationID, quotationID uuid.UUID) (domainquotation.Quotation, error) {
	return f.getQuotationFn(ctx, organizationID, quotationID)
}

func (f fakeRepository) MarkRFQAwarded(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	return f.markRFQAwardedFn(ctx, organizationID, rfqID)
}
func (f fakeRepository) CreateActivityLog(ctx context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error) {
	if f.createActivityLogFn == nil {
		return domainactivitylog.Entry{}, nil
	}
	return f.createActivityLogFn(ctx, params)
}

type fakeTxManager struct {
	withinFn func(context.Context, func(Repository) error) error
}

func (f fakeTxManager) WithinTransaction(ctx context.Context, fn func(Repository) error) error {
	return f.withinFn(ctx, fn)
}

func TestServiceCreateAwardsSubmittedQuotationAndMarksRFQ(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	rfqID := uuid.New()
	quotationID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()
	markedAwarded := false
	logged := false
	rfqLogged := false

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, gotOrgID, gotUserID uuid.UUID) (domainorganization.Membership, error) {
			if gotOrgID != orgID || gotUserID != userID {
				t.Fatalf("unexpected membership lookup")
			}
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleProcurementOfficer,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRFQFn: func(_ context.Context, gotOrgID, gotRFQID uuid.UUID) (domainrfq.RFQ, error) {
			if gotOrgID != orgID || gotRFQID != rfqID {
				t.Fatalf("unexpected rfq lookup")
			}
			return domainrfq.RFQ{
				ID:             rfqID,
				OrganizationID: orgID,
				Status:         domainrfq.StatusEvaluated,
			}, nil
		},
		getQuotationFn: func(_ context.Context, gotOrgID, gotQuotationID uuid.UUID) (domainquotation.Quotation, error) {
			if gotOrgID != orgID || gotQuotationID != quotationID {
				t.Fatalf("unexpected quotation lookup")
			}
			return domainquotation.Quotation{
				ID:        quotationID,
				RFQID:     rfqID,
				Status:    domainquotation.StatusSubmitted,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
		createAwardFn: func(_ context.Context, params CreateAwardParams) (domainaward.Award, error) {
			if params.Reason != "best commercial value" {
				t.Fatalf("unexpected reason: %s", params.Reason)
			}
			return domainaward.Award{
				ID:              uuid.New(),
				OrganizationID:  params.OrganizationID,
				RFQID:           params.RFQID,
				QuotationID:     params.QuotationID,
				AwardedByUserID: params.AwardedByUserID,
				Reason:          params.Reason,
				AwardedAt:       now,
				CreatedAt:       now,
			}, nil
		},
		markRFQAwardedFn: func(_ context.Context, gotOrgID, gotRFQID uuid.UUID) (domainrfq.RFQ, error) {
			if gotOrgID != orgID || gotRFQID != rfqID {
				t.Fatalf("unexpected mark awarded args")
			}
			markedAwarded = true
			return domainrfq.RFQ{
				ID:             rfqID,
				OrganizationID: orgID,
				Status:         domainrfq.StatusAwarded,
			}, nil
		},
		createActivityLogFn: func(_ context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error) {
			switch {
			case params.EntityType == string(domainactivitylog.EntityTypeAward) && params.Action == domainactivitylog.ActionAwardCreated:
				logged = true
			case params.EntityType == string(domainactivitylog.EntityTypeRFQ) && params.Action == domainactivitylog.ActionRFQAwarded:
				rfqLogged = true
			default:
				t.Fatalf("unexpected activity log payload: %#v", params)
			}
			return domainactivitylog.Entry{EntityID: params.EntityID, Action: params.Action}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(_ context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})

	award, err := service.Create(context.Background(), CreateInput{
		OrganizationID: orgID,
		RFQID:          rfqID,
		QuotationID:    quotationID,
		CurrentUser:    userID,
		Reason:         "best commercial value",
	})
	if err != nil {
		t.Fatalf("create award returned error: %v", err)
	}
	if award.QuotationID != quotationID {
		t.Fatalf("expected quotation %s, got %s", quotationID, award.QuotationID)
	}
	if !markedAwarded {
		t.Fatalf("expected rfq status to be marked awarded")
	}
	if !logged {
		t.Fatalf("expected award activity log to be written")
	}
	if !rfqLogged {
		t.Fatalf("expected rfq activity log to be written")
	}
}

func TestServiceCreateRejectsNonSubmittedQuotation(t *testing.T) {
	t.Parallel()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleOwner,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRFQFn: func(_ context.Context, _, _ uuid.UUID) (domainrfq.RFQ, error) {
			return domainrfq.RFQ{Status: domainrfq.StatusEvaluated}, nil
		},
		getQuotationFn: func(_ context.Context, _, _ uuid.UUID) (domainquotation.Quotation, error) {
			return domainquotation.Quotation{Status: domainquotation.StatusDraft}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(context.Context, func(Repository) error) error {
			t.Fatal("transaction should not start")
			return nil
		},
	})

	_, err := service.Create(context.Background(), CreateInput{
		OrganizationID: uuid.New(),
		RFQID:          uuid.New(),
		QuotationID:    uuid.New(),
		CurrentUser:    uuid.New(),
		Reason:         "reason",
	})
	if !errors.Is(err, ErrInvalidAward) {
		t.Fatalf("expected invalid award error, got %v", err)
	}
}

func TestServiceCreateRejectsRequesterRole(t *testing.T) {
	t.Parallel()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleRequester,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRFQFn: func(_ context.Context, _, _ uuid.UUID) (domainrfq.RFQ, error) {
			return domainrfq.RFQ{Status: domainrfq.StatusEvaluated}, nil
		},
		getQuotationFn: func(_ context.Context, _, _ uuid.UUID) (domainquotation.Quotation, error) {
			return domainquotation.Quotation{Status: domainquotation.StatusSubmitted}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(context.Context, func(Repository) error) error {
			t.Fatal("transaction should not start")
			return nil
		},
	})

	_, err := service.Create(context.Background(), CreateInput{
		OrganizationID: uuid.New(),
		RFQID:          uuid.New(),
		QuotationID:    uuid.New(),
		CurrentUser:    uuid.New(),
		Reason:         "reason",
	})
	if !errors.Is(err, ErrForbiddenAward) {
		t.Fatalf("expected forbidden award error, got %v", err)
	}
}

func TestServiceGetByRFQAllowsViewer(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	rfqID := uuid.New()
	userID := uuid.New()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, gotOrgID, gotUserID uuid.UUID) (domainorganization.Membership, error) {
			if gotOrgID != orgID || gotUserID != userID {
				t.Fatalf("unexpected membership lookup")
			}
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleViewer,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getAwardByRFQFn: func(_ context.Context, gotOrgID, gotRFQID uuid.UUID) (domainaward.Award, error) {
			if gotOrgID != orgID || gotRFQID != rfqID {
				t.Fatalf("unexpected award lookup")
			}
			return domainaward.Award{ID: uuid.New(), RFQID: rfqID}, nil
		},
	}

	service := NewService(repo, fakeTxManager{})
	award, err := service.GetByRFQ(context.Background(), orgID, rfqID, userID)
	if err != nil {
		t.Fatalf("get award returned error: %v", err)
	}
	if award.RFQID != rfqID {
		t.Fatalf("expected rfq %s, got %s", rfqID, award.RFQID)
	}
}
