package rfq

import (
	"context"
	"errors"
	"testing"
	"time"

	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainprocurement "github.com/AbenezerWork/ProcureFlow/internal/domain/procurement"
	domainrfq "github.com/AbenezerWork/ProcureFlow/internal/domain/rfq"
	domainvendor "github.com/AbenezerWork/ProcureFlow/internal/domain/vendor"
	"github.com/google/uuid"
)

type fakeRepository struct {
	createRFQFn                   func(context.Context, CreateRFQParams) (domainrfq.RFQ, error)
	createRFQItemFn               func(context.Context, CreateRFQItemParams) (domainrfq.Item, error)
	getRFQFn                      func(context.Context, uuid.UUID, uuid.UUID) (domainrfq.RFQ, error)
	listRFQsFn                    func(context.Context, uuid.UUID, *domainrfq.Status) ([]domainrfq.RFQ, error)
	updateDraftRFQFn              func(context.Context, UpdateRFQParams) (domainrfq.RFQ, error)
	publishRFQFn                  func(context.Context, uuid.UUID, uuid.UUID) (domainrfq.RFQ, error)
	closeRFQFn                    func(context.Context, uuid.UUID, uuid.UUID) (domainrfq.RFQ, error)
	evaluateRFQFn                 func(context.Context, uuid.UUID, uuid.UUID) (domainrfq.RFQ, error)
	cancelRFQFn                   func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domainrfq.RFQ, error)
	listRFQItemsFn                func(context.Context, uuid.UUID, uuid.UUID) ([]domainrfq.Item, error)
	attachVendorFn                func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) (domainrfq.VendorLink, error)
	removeVendorFn                func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
	listRFQVendorsFn              func(context.Context, uuid.UUID, uuid.UUID) ([]domainrfq.VendorLink, error)
	getMembershipFn               func(context.Context, uuid.UUID, uuid.UUID) (domainorganization.Membership, error)
	getProcurementRequestFn       func(context.Context, uuid.UUID, uuid.UUID) (domainprocurement.Request, error)
	listProcurementRequestItemsFn func(context.Context, uuid.UUID, uuid.UUID) ([]domainprocurement.Item, error)
	getVendorFn                   func(context.Context, uuid.UUID, uuid.UUID) (domainvendor.Vendor, error)
}

func (f fakeRepository) CreateRFQ(ctx context.Context, params CreateRFQParams) (domainrfq.RFQ, error) {
	return f.createRFQFn(ctx, params)
}
func (f fakeRepository) CreateRFQItem(ctx context.Context, params CreateRFQItemParams) (domainrfq.Item, error) {
	return f.createRFQItemFn(ctx, params)
}
func (f fakeRepository) GetRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	return f.getRFQFn(ctx, organizationID, rfqID)
}
func (f fakeRepository) ListRFQs(ctx context.Context, organizationID uuid.UUID, status *domainrfq.Status) ([]domainrfq.RFQ, error) {
	return f.listRFQsFn(ctx, organizationID, status)
}
func (f fakeRepository) UpdateDraftRFQ(ctx context.Context, params UpdateRFQParams) (domainrfq.RFQ, error) {
	return f.updateDraftRFQFn(ctx, params)
}
func (f fakeRepository) PublishRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	return f.publishRFQFn(ctx, organizationID, rfqID)
}
func (f fakeRepository) CloseRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	return f.closeRFQFn(ctx, organizationID, rfqID)
}
func (f fakeRepository) EvaluateRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	return f.evaluateRFQFn(ctx, organizationID, rfqID)
}
func (f fakeRepository) CancelRFQ(ctx context.Context, organizationID, rfqID, cancelledByUserID uuid.UUID) (domainrfq.RFQ, error) {
	return f.cancelRFQFn(ctx, organizationID, rfqID, cancelledByUserID)
}
func (f fakeRepository) ListRFQItems(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainrfq.Item, error) {
	return f.listRFQItemsFn(ctx, organizationID, rfqID)
}
func (f fakeRepository) AttachVendor(ctx context.Context, organizationID, rfqID, vendorID, attachedByUserID uuid.UUID) (domainrfq.VendorLink, error) {
	return f.attachVendorFn(ctx, organizationID, rfqID, vendorID, attachedByUserID)
}
func (f fakeRepository) RemoveVendor(ctx context.Context, organizationID, rfqID, vendorID uuid.UUID) error {
	return f.removeVendorFn(ctx, organizationID, rfqID, vendorID)
}
func (f fakeRepository) ListRFQVendors(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainrfq.VendorLink, error) {
	return f.listRFQVendorsFn(ctx, organizationID, rfqID)
}
func (f fakeRepository) GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error) {
	return f.getMembershipFn(ctx, organizationID, userID)
}
func (f fakeRepository) GetProcurementRequest(ctx context.Context, organizationID, requestID uuid.UUID) (domainprocurement.Request, error) {
	return f.getProcurementRequestFn(ctx, organizationID, requestID)
}
func (f fakeRepository) ListProcurementRequestItems(ctx context.Context, organizationID, requestID uuid.UUID) ([]domainprocurement.Item, error) {
	return f.listProcurementRequestItemsFn(ctx, organizationID, requestID)
}
func (f fakeRepository) GetVendor(ctx context.Context, organizationID, vendorID uuid.UUID) (domainvendor.Vendor, error) {
	return f.getVendorFn(ctx, organizationID, vendorID)
}

type fakeTxManager struct {
	withTransactionFn func(context.Context, func(Repository) error) error
}

func (f fakeTxManager) WithinTransaction(ctx context.Context, fn func(Repository) error) error {
	return f.withTransactionFn(ctx, fn)
}

func TestServiceCreateSnapshotsApprovedRequestItems(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	requestID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()
	var createdItemCount int

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{Role: domainorganization.MembershipRoleProcurementOfficer, Status: domainorganization.MembershipStatusActive}, nil
		},
		getProcurementRequestFn: func(_ context.Context, gotOrgID, gotRequestID uuid.UUID) (domainprocurement.Request, error) {
			if gotOrgID != orgID || gotRequestID != requestID {
				t.Fatalf("unexpected procurement request lookup")
			}
			return domainprocurement.Request{ID: requestID, OrganizationID: orgID, Title: "Office Chairs", Status: domainprocurement.RequestStatusApproved}, nil
		},
		listProcurementRequestItemsFn: func(_ context.Context, _, _ uuid.UUID) ([]domainprocurement.Item, error) {
			return []domainprocurement.Item{
				{ID: uuid.New(), LineNumber: 1, ItemName: "Chair", Quantity: "10.00", Unit: "pcs"},
				{ID: uuid.New(), LineNumber: 2, ItemName: "Desk", Quantity: "4.00", Unit: "pcs"},
			}, nil
		},
		createRFQFn: func(_ context.Context, params CreateRFQParams) (domainrfq.RFQ, error) {
			if params.Title != "Office Chairs" {
				t.Fatalf("unexpected title: %s", params.Title)
			}
			return domainrfq.RFQ{
				ID:                   uuid.New(),
				OrganizationID:       params.OrganizationID,
				ProcurementRequestID: params.ProcurementRequestID,
				Title:                params.Title,
				Status:               domainrfq.StatusDraft,
				CreatedByUserID:      params.CreatedByUserID,
				CreatedAt:            now,
				UpdatedAt:            now,
			}, nil
		},
		createRFQItemFn: func(_ context.Context, params CreateRFQItemParams) (domainrfq.Item, error) {
			createdItemCount++
			return domainrfq.Item{
				ID:             uuid.New(),
				OrganizationID: params.OrganizationID,
				RFQID:          params.RFQID,
				LineNumber:     params.LineNumber,
				ItemName:       params.ItemName,
				Quantity:       params.Quantity,
				Unit:           params.Unit,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withTransactionFn: func(ctx context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})

	rfq, err := service.Create(context.Background(), CreateInput{
		OrganizationID:       orgID,
		ProcurementRequestID: requestID,
		CurrentUser:          userID,
	})
	if err != nil {
		t.Fatalf("create rfq returned error: %v", err)
	}
	if rfq.Status != domainrfq.StatusDraft {
		t.Fatalf("expected draft status, got %s", rfq.Status)
	}
	if createdItemCount != 2 {
		t.Fatalf("expected 2 snapshot items, got %d", createdItemCount)
	}
}

func TestServiceCreateRejectsNonApprovedRequest(t *testing.T) {
	t.Parallel()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{Role: domainorganization.MembershipRoleOwner, Status: domainorganization.MembershipStatusActive}, nil
		},
		getProcurementRequestFn: func(_ context.Context, _, _ uuid.UUID) (domainprocurement.Request, error) {
			return domainprocurement.Request{Status: domainprocurement.RequestStatusSubmitted}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withTransactionFn: func(context.Context, func(Repository) error) error {
			t.Fatal("transaction should not start")
			return nil
		},
	})

	_, err := service.Create(context.Background(), CreateInput{
		OrganizationID:       uuid.New(),
		ProcurementRequestID: uuid.New(),
		CurrentUser:          uuid.New(),
	})
	if !errors.Is(err, ErrInvalidRFQ) {
		t.Fatalf("expected invalid rfq error, got %v", err)
	}
}

func TestServicePublishRequiresAttachedVendors(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	rfqID := uuid.New()
	userID := uuid.New()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{Role: domainorganization.MembershipRoleAdmin, Status: domainorganization.MembershipStatusActive}, nil
		},
		getRFQFn: func(_ context.Context, _, _ uuid.UUID) (domainrfq.RFQ, error) {
			return domainrfq.RFQ{ID: rfqID, OrganizationID: orgID, Status: domainrfq.StatusDraft}, nil
		},
		listRFQItemsFn: func(_ context.Context, _, _ uuid.UUID) ([]domainrfq.Item, error) {
			return []domainrfq.Item{{ID: uuid.New()}}, nil
		},
		listRFQVendorsFn: func(_ context.Context, _, _ uuid.UUID) ([]domainrfq.VendorLink, error) {
			return nil, nil
		},
		publishRFQFn: func(context.Context, uuid.UUID, uuid.UUID) (domainrfq.RFQ, error) {
			t.Fatal("publish should not be called")
			return domainrfq.RFQ{}, nil
		},
	}

	service := NewService(repo, fakeTxManager{})
	_, err := service.Publish(context.Background(), TransitionInput{
		OrganizationID: orgID,
		RFQID:          rfqID,
		CurrentUser:    userID,
	})
	if !errors.Is(err, ErrInvalidRFQ) {
		t.Fatalf("expected invalid rfq error, got %v", err)
	}
}

func TestServiceAttachVendorRejectsArchivedVendor(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	rfqID := uuid.New()
	vendorID := uuid.New()
	userID := uuid.New()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{Role: domainorganization.MembershipRoleProcurementOfficer, Status: domainorganization.MembershipStatusActive}, nil
		},
		getRFQFn: func(_ context.Context, _, _ uuid.UUID) (domainrfq.RFQ, error) {
			return domainrfq.RFQ{ID: rfqID, OrganizationID: orgID, Status: domainrfq.StatusDraft}, nil
		},
		getVendorFn: func(_ context.Context, _, _ uuid.UUID) (domainvendor.Vendor, error) {
			return domainvendor.Vendor{ID: vendorID, Status: domainvendor.StatusArchived}, nil
		},
		attachVendorFn: func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) (domainrfq.VendorLink, error) {
			t.Fatal("attach vendor should not be called")
			return domainrfq.VendorLink{}, nil
		},
	}

	service := NewService(repo, fakeTxManager{})
	_, err := service.AttachVendor(context.Background(), AttachVendorInput{
		OrganizationID: orgID,
		RFQID:          rfqID,
		VendorID:       vendorID,
		CurrentUser:    userID,
	})
	if !errors.Is(err, ErrInvalidRFQ) {
		t.Fatalf("expected invalid rfq error, got %v", err)
	}
}

func TestServiceCancelRejectsAwardedRFQ(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	rfqID := uuid.New()
	userID := uuid.New()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{Role: domainorganization.MembershipRoleOwner, Status: domainorganization.MembershipStatusActive}, nil
		},
		getRFQFn: func(_ context.Context, _, _ uuid.UUID) (domainrfq.RFQ, error) {
			return domainrfq.RFQ{ID: rfqID, OrganizationID: orgID, Status: domainrfq.StatusAwarded}, nil
		},
		cancelRFQFn: func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domainrfq.RFQ, error) {
			t.Fatal("cancel should not be called")
			return domainrfq.RFQ{}, nil
		},
	}

	service := NewService(repo, fakeTxManager{})
	_, err := service.Cancel(context.Background(), TransitionInput{
		OrganizationID: orgID,
		RFQID:          rfqID,
		CurrentUser:    userID,
	})
	if !errors.Is(err, ErrInvalidRFQ) {
		t.Fatalf("expected invalid rfq error, got %v", err)
	}
}
