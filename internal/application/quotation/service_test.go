package quotation

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainquotation "github.com/AbenezerWork/ProcureFlow/internal/domain/quotation"
	domainrfq "github.com/AbenezerWork/ProcureFlow/internal/domain/rfq"
	"github.com/google/uuid"
)

type fakeRepository struct {
	createQuotationFn     func(context.Context, CreateQuotationParams) (domainquotation.Quotation, error)
	createQuotationItemFn func(context.Context, CreateItemParams) (domainquotation.Item, error)
	getQuotationFn        func(context.Context, uuid.UUID, uuid.UUID) (domainquotation.Quotation, error)
	listQuotationsByRFQFn func(context.Context, uuid.UUID, uuid.UUID) ([]domainquotation.Quotation, error)
	updateQuotationFn     func(context.Context, UpdateQuotationParams) (domainquotation.Quotation, error)
	submitQuotationFn     func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domainquotation.Quotation, error)
	rejectQuotationFn     func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, *string) (domainquotation.Quotation, error)
	listQuotationItemsFn  func(context.Context, uuid.UUID, uuid.UUID) ([]domainquotation.Item, error)
	updateQuotationItemFn func(context.Context, UpdateItemParams) (domainquotation.Item, error)
	getMembershipFn       func(context.Context, uuid.UUID, uuid.UUID) (domainorganization.Membership, error)
	getRFQFn              func(context.Context, uuid.UUID, uuid.UUID) (domainrfq.RFQ, error)
	listRFQItemsFn        func(context.Context, uuid.UUID, uuid.UUID) ([]domainrfq.Item, error)
	listRFQVendorsFn      func(context.Context, uuid.UUID, uuid.UUID) ([]domainrfq.VendorLink, error)
	createActivityLogFn   func(context.Context, applicationactivitylog.CreateParams) (domainactivitylog.Entry, error)
}

func (f fakeRepository) CreateQuotation(ctx context.Context, params CreateQuotationParams) (domainquotation.Quotation, error) {
	return f.createQuotationFn(ctx, params)
}

func (f fakeRepository) CreateQuotationItem(ctx context.Context, params CreateItemParams) (domainquotation.Item, error) {
	return f.createQuotationItemFn(ctx, params)
}

func (f fakeRepository) GetQuotation(ctx context.Context, organizationID, quotationID uuid.UUID) (domainquotation.Quotation, error) {
	return f.getQuotationFn(ctx, organizationID, quotationID)
}

func (f fakeRepository) ListQuotationsByRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainquotation.Quotation, error) {
	return f.listQuotationsByRFQFn(ctx, organizationID, rfqID)
}

func (f fakeRepository) UpdateDraftQuotation(ctx context.Context, params UpdateQuotationParams) (domainquotation.Quotation, error) {
	return f.updateQuotationFn(ctx, params)
}

func (f fakeRepository) SubmitQuotation(ctx context.Context, organizationID, quotationID, submittedByUserID uuid.UUID) (domainquotation.Quotation, error) {
	return f.submitQuotationFn(ctx, organizationID, quotationID, submittedByUserID)
}

func (f fakeRepository) RejectQuotation(ctx context.Context, organizationID, quotationID, rejectedByUserID uuid.UUID, rejectionReason *string) (domainquotation.Quotation, error) {
	return f.rejectQuotationFn(ctx, organizationID, quotationID, rejectedByUserID, rejectionReason)
}

func (f fakeRepository) ListQuotationItems(ctx context.Context, organizationID, quotationID uuid.UUID) ([]domainquotation.Item, error) {
	return f.listQuotationItemsFn(ctx, organizationID, quotationID)
}

func (f fakeRepository) UpdateQuotationItem(ctx context.Context, params UpdateItemParams) (domainquotation.Item, error) {
	return f.updateQuotationItemFn(ctx, params)
}

func (f fakeRepository) GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error) {
	return f.getMembershipFn(ctx, organizationID, userID)
}

func (f fakeRepository) GetRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	return f.getRFQFn(ctx, organizationID, rfqID)
}

func (f fakeRepository) ListRFQItems(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainrfq.Item, error) {
	return f.listRFQItemsFn(ctx, organizationID, rfqID)
}

func (f fakeRepository) ListRFQVendors(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainrfq.VendorLink, error) {
	return f.listRFQVendorsFn(ctx, organizationID, rfqID)
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

func TestServiceCreateSnapshotsRFQItems(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	rfqID := uuid.New()
	rfqVendorID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()
	createdItems := 0

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
				Status:         domainrfq.StatusPublished,
			}, nil
		},
		listRFQVendorsFn: func(_ context.Context, gotOrgID, gotRFQID uuid.UUID) ([]domainrfq.VendorLink, error) {
			if gotOrgID != orgID || gotRFQID != rfqID {
				t.Fatalf("unexpected rfq vendor lookup")
			}
			return []domainrfq.VendorLink{{
				ID:         rfqVendorID,
				RFQID:      rfqID,
				VendorName: "Blue Nile Supplies",
			}}, nil
		},
		listRFQItemsFn: func(_ context.Context, gotOrgID, gotRFQID uuid.UUID) ([]domainrfq.Item, error) {
			if gotOrgID != orgID || gotRFQID != rfqID {
				t.Fatalf("unexpected rfq item lookup")
			}
			return []domainrfq.Item{
				{ID: uuid.New(), LineNumber: 1, ItemName: "Chair", Quantity: "5", Unit: "pcs"},
				{ID: uuid.New(), LineNumber: 2, ItemName: "Desk", Quantity: "2", Unit: "pcs"},
			}, nil
		},
		createQuotationFn: func(_ context.Context, params CreateQuotationParams) (domainquotation.Quotation, error) {
			if params.CurrencyCode != "ETB" {
				t.Fatalf("unexpected currency: %s", params.CurrencyCode)
			}
			return domainquotation.Quotation{
				ID:             uuid.New(),
				OrganizationID: params.OrganizationID,
				RFQID:          params.RFQID,
				RFQVendorID:    params.RFQVendorID,
				Status:         domainquotation.StatusDraft,
				CurrencyCode:   params.CurrencyCode,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil
		},
		createQuotationItemFn: func(_ context.Context, params CreateItemParams) (domainquotation.Item, error) {
			createdItems++
			if params.UnitPrice != "0" {
				t.Fatalf("expected snapshotted item unit price 0, got %s", params.UnitPrice)
			}
			return domainquotation.Item{
				ID:          uuid.New(),
				QuotationID: params.QuotationID,
				LineNumber:  params.LineNumber,
				UnitPrice:   params.UnitPrice,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(_ context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})

	quotation, err := service.Create(context.Background(), CreateInput{
		OrganizationID: orgID,
		RFQID:          rfqID,
		RFQVendorID:    rfqVendorID,
		CurrentUser:    userID,
		CurrencyCode:   stringPtr("etb"),
	})
	if err != nil {
		t.Fatalf("create quotation returned error: %v", err)
	}
	if quotation.Status != domainquotation.StatusDraft {
		t.Fatalf("expected draft status, got %s", quotation.Status)
	}
	if quotation.VendorName == nil || *quotation.VendorName != "Blue Nile Supplies" {
		t.Fatalf("expected vendor name to be populated")
	}
	if createdItems != 2 {
		t.Fatalf("expected 2 snapshotted items, got %d", createdItems)
	}
}

func TestServiceSubmitRequiresPublishedRFQ(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	rfqID := uuid.New()
	quotationID := uuid.New()
	userID := uuid.New()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleProcurementOfficer,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRFQFn: func(_ context.Context, _, _ uuid.UUID) (domainrfq.RFQ, error) {
			return domainrfq.RFQ{
				ID:     rfqID,
				Status: domainrfq.StatusClosed,
			}, nil
		},
		getQuotationFn: func(_ context.Context, _, _ uuid.UUID) (domainquotation.Quotation, error) {
			return domainquotation.Quotation{
				ID:          quotationID,
				RFQID:       rfqID,
				Status:      domainquotation.StatusDraft,
				RFQVendorID: uuid.New(),
			}, nil
		},
		submitQuotationFn: func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domainquotation.Quotation, error) {
			t.Fatal("submit should not be called")
			return domainquotation.Quotation{}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(_ context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})

	_, err := service.Submit(context.Background(), TransitionInput{
		OrganizationID: orgID,
		RFQID:          rfqID,
		QuotationID:    quotationID,
		CurrentUser:    userID,
	})
	if !errors.Is(err, ErrInvalidQuotation) {
		t.Fatalf("expected invalid quotation error, got %v", err)
	}
}

func TestServiceSubmitWritesActivityLog(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	rfqID := uuid.New()
	quotationID := uuid.New()
	rfqVendorID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()
	logged := false

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleProcurementOfficer,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRFQFn: func(_ context.Context, _, _ uuid.UUID) (domainrfq.RFQ, error) {
			return domainrfq.RFQ{ID: rfqID, Status: domainrfq.StatusPublished}, nil
		},
		getQuotationFn: func(_ context.Context, _, _ uuid.UUID) (domainquotation.Quotation, error) {
			return domainquotation.Quotation{
				ID:          quotationID,
				RFQID:       rfqID,
				RFQVendorID: rfqVendorID,
				Status:      domainquotation.StatusDraft,
			}, nil
		},
		listQuotationItemsFn: func(_ context.Context, _, _ uuid.UUID) ([]domainquotation.Item, error) {
			return []domainquotation.Item{{ID: uuid.New(), UnitPrice: "10.00"}}, nil
		},
		submitQuotationFn: func(_ context.Context, _, _ uuid.UUID, _ uuid.UUID) (domainquotation.Quotation, error) {
			return domainquotation.Quotation{
				ID:             quotationID,
				OrganizationID: orgID,
				RFQID:          rfqID,
				RFQVendorID:    rfqVendorID,
				Status:         domainquotation.StatusSubmitted,
				SubmittedAt:    &now,
			}, nil
		},
		listRFQVendorsFn: func(_ context.Context, _, _ uuid.UUID) ([]domainrfq.VendorLink, error) {
			return []domainrfq.VendorLink{{ID: rfqVendorID, VendorName: "Blue Nile Supplies"}}, nil
		},
		createActivityLogFn: func(_ context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error) {
			logged = true
			if params.EntityType != string(domainactivitylog.EntityTypeQuotation) || params.Action != domainactivitylog.ActionQuotationSubmitted {
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

	quotation, err := service.Submit(context.Background(), TransitionInput{
		OrganizationID: orgID,
		RFQID:          rfqID,
		QuotationID:    quotationID,
		CurrentUser:    userID,
	})
	if err != nil {
		t.Fatalf("submit quotation returned error: %v", err)
	}
	if quotation.Status != domainquotation.StatusSubmitted {
		t.Fatalf("expected submitted status, got %s", quotation.Status)
	}
	if !logged {
		t.Fatalf("expected activity log to be written")
	}
}

func TestServiceUpdateItemRejectsNegativePrice(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	rfqID := uuid.New()
	quotationID := uuid.New()
	itemID := uuid.New()
	userID := uuid.New()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleOwner,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRFQFn: func(_ context.Context, _, _ uuid.UUID) (domainrfq.RFQ, error) {
			return domainrfq.RFQ{
				ID:     rfqID,
				Status: domainrfq.StatusPublished,
			}, nil
		},
		getQuotationFn: func(_ context.Context, _, _ uuid.UUID) (domainquotation.Quotation, error) {
			return domainquotation.Quotation{
				ID:          quotationID,
				RFQID:       rfqID,
				Status:      domainquotation.StatusDraft,
				RFQVendorID: uuid.New(),
			}, nil
		},
		listQuotationItemsFn: func(_ context.Context, _, _ uuid.UUID) ([]domainquotation.Item, error) {
			return []domainquotation.Item{{
				ID:          itemID,
				QuotationID: quotationID,
				Quantity:    "3",
				Unit:        "pcs",
				UnitPrice:   "10.00",
			}}, nil
		},
		updateQuotationItemFn: func(context.Context, UpdateItemParams) (domainquotation.Item, error) {
			t.Fatal("update item should not be called")
			return domainquotation.Item{}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(_ context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})

	_, err := service.UpdateItem(context.Background(), UpdateItemInput{
		OrganizationID: orgID,
		RFQID:          rfqID,
		QuotationID:    quotationID,
		ItemID:         itemID,
		CurrentUser:    userID,
		UnitPrice:      stringPtr("-1"),
	})
	if !errors.Is(err, ErrInvalidQuotation) {
		t.Fatalf("expected invalid quotation error, got %v", err)
	}
}

func TestServiceListAllowsViewer(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	rfqID := uuid.New()
	userID := uuid.New()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleViewer,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRFQFn: func(_ context.Context, _, _ uuid.UUID) (domainrfq.RFQ, error) {
			return domainrfq.RFQ{ID: rfqID}, nil
		},
		listQuotationsByRFQFn: func(_ context.Context, gotOrgID, gotRFQID uuid.UUID) ([]domainquotation.Quotation, error) {
			if gotOrgID != orgID || gotRFQID != rfqID {
				t.Fatalf("unexpected quotation list args")
			}
			return []domainquotation.Quotation{{ID: uuid.New(), RFQID: rfqID}}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(_ context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})

	quotations, err := service.List(context.Background(), ListInput{
		OrganizationID: orgID,
		RFQID:          rfqID,
		CurrentUser:    userID,
	})
	if err != nil {
		t.Fatalf("list quotations returned error: %v", err)
	}
	if len(quotations) != 1 {
		t.Fatalf("expected 1 quotation, got %d", len(quotations))
	}
}

func stringPtr(value string) *string {
	return &value
}
