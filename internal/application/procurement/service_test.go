package procurement

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainprocurement "github.com/AbenezerWork/ProcureFlow/internal/domain/procurement"
	"github.com/google/uuid"
)

type fakeRepository struct {
	createRequestFn     func(context.Context, CreateRequestParams) (domainprocurement.Request, error)
	getRequestFn        func(context.Context, uuid.UUID, uuid.UUID) (domainprocurement.Request, error)
	listRequestsFn      func(context.Context, uuid.UUID, *domainprocurement.RequestStatus) ([]domainprocurement.Request, error)
	listInboxFn         func(context.Context, uuid.UUID) ([]domainprocurement.Request, error)
	updateRequestFn     func(context.Context, UpdateRequestParams) (domainprocurement.Request, error)
	submitRequestFn     func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domainprocurement.Request, error)
	approveRequestFn    func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, *string) (domainprocurement.Request, error)
	rejectRequestFn     func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, *string) (domainprocurement.Request, error)
	cancelRequestFn     func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domainprocurement.Request, error)
	createItemFn        func(context.Context, CreateItemParams) (domainprocurement.Item, error)
	listItemsFn         func(context.Context, uuid.UUID, uuid.UUID) ([]domainprocurement.Item, error)
	updateItemFn        func(context.Context, UpdateItemParams) (domainprocurement.Item, error)
	deleteItemFn        func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
	getMembershipFn     func(context.Context, uuid.UUID, uuid.UUID) (domainorganization.Membership, error)
	createActivityLogFn func(context.Context, applicationactivitylog.CreateParams) (domainactivitylog.Entry, error)
}

func (f fakeRepository) CreateRequest(ctx context.Context, params CreateRequestParams) (domainprocurement.Request, error) {
	return f.createRequestFn(ctx, params)
}
func (f fakeRepository) GetRequest(ctx context.Context, organizationID, requestID uuid.UUID) (domainprocurement.Request, error) {
	return f.getRequestFn(ctx, organizationID, requestID)
}
func (f fakeRepository) ListRequests(ctx context.Context, organizationID uuid.UUID, status *domainprocurement.RequestStatus) ([]domainprocurement.Request, error) {
	return f.listRequestsFn(ctx, organizationID, status)
}
func (f fakeRepository) ListApprovalInbox(ctx context.Context, organizationID uuid.UUID) ([]domainprocurement.Request, error) {
	return f.listInboxFn(ctx, organizationID)
}
func (f fakeRepository) UpdateDraftRequest(ctx context.Context, params UpdateRequestParams) (domainprocurement.Request, error) {
	return f.updateRequestFn(ctx, params)
}
func (f fakeRepository) SubmitRequest(ctx context.Context, organizationID, requestID, submittedByUserID uuid.UUID) (domainprocurement.Request, error) {
	return f.submitRequestFn(ctx, organizationID, requestID, submittedByUserID)
}
func (f fakeRepository) ApproveRequest(ctx context.Context, organizationID, requestID, approvedByUserID uuid.UUID, decisionComment *string) (domainprocurement.Request, error) {
	return f.approveRequestFn(ctx, organizationID, requestID, approvedByUserID, decisionComment)
}
func (f fakeRepository) RejectRequest(ctx context.Context, organizationID, requestID, rejectedByUserID uuid.UUID, decisionComment *string) (domainprocurement.Request, error) {
	return f.rejectRequestFn(ctx, organizationID, requestID, rejectedByUserID, decisionComment)
}
func (f fakeRepository) CancelRequest(ctx context.Context, organizationID, requestID, cancelledByUserID uuid.UUID) (domainprocurement.Request, error) {
	return f.cancelRequestFn(ctx, organizationID, requestID, cancelledByUserID)
}
func (f fakeRepository) CreateItem(ctx context.Context, params CreateItemParams) (domainprocurement.Item, error) {
	return f.createItemFn(ctx, params)
}
func (f fakeRepository) ListItems(ctx context.Context, organizationID, requestID uuid.UUID) ([]domainprocurement.Item, error) {
	return f.listItemsFn(ctx, organizationID, requestID)
}
func (f fakeRepository) UpdateItem(ctx context.Context, params UpdateItemParams) (domainprocurement.Item, error) {
	return f.updateItemFn(ctx, params)
}
func (f fakeRepository) DeleteItem(ctx context.Context, organizationID, requestID, itemID uuid.UUID) error {
	return f.deleteItemFn(ctx, organizationID, requestID, itemID)
}
func (f fakeRepository) GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error) {
	return f.getMembershipFn(ctx, organizationID, userID)
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

func TestServiceCreateRequest(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, gotOrgID, gotUserID uuid.UUID) (domainorganization.Membership, error) {
			if gotOrgID != orgID || gotUserID != userID {
				t.Fatalf("unexpected membership lookup")
			}
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleRequester,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		createRequestFn: func(_ context.Context, params CreateRequestParams) (domainprocurement.Request, error) {
			if params.CurrencyCode != "ETB" {
				t.Fatalf("unexpected currency code: %s", params.CurrencyCode)
			}
			if params.EstimatedTotalAmount == nil || *params.EstimatedTotalAmount != "1000.50" {
				t.Fatalf("unexpected estimated amount: %#v", params.EstimatedTotalAmount)
			}
			return domainprocurement.Request{
				ID:                   uuid.New(),
				OrganizationID:       params.OrganizationID,
				RequesterUserID:      params.RequesterUserID,
				Title:                params.Title,
				Status:               domainprocurement.RequestStatusDraft,
				CurrencyCode:         params.CurrencyCode,
				EstimatedTotalAmount: params.EstimatedTotalAmount,
				CreatedAt:            now,
				UpdatedAt:            now,
			}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(_ context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})
	created, err := service.CreateRequest(context.Background(), CreateRequestInput{
		OrganizationID:       orgID,
		CurrentUser:          userID,
		Title:                "Office Chairs",
		CurrencyCode:         stringPtr("etb"),
		EstimatedTotalAmount: stringPtr("1000.50"),
	})
	if err != nil {
		t.Fatalf("create request returned error: %v", err)
	}
	if created.Status != domainprocurement.RequestStatusDraft {
		t.Fatalf("expected draft status, got %s", created.Status)
	}
}

func TestServiceUpdateRequestRejectsNonOwnerRequester(t *testing.T) {
	t.Parallel()

	requesterID := uuid.New()
	otherUserID := uuid.New()
	orgID := uuid.New()
	requestID := uuid.New()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, userID uuid.UUID) (domainorganization.Membership, error) {
			if userID != otherUserID {
				t.Fatalf("unexpected membership user: %s", userID)
			}
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleRequester,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRequestFn: func(_ context.Context, _, _ uuid.UUID) (domainprocurement.Request, error) {
			return domainprocurement.Request{
				ID:              requestID,
				OrganizationID:  orgID,
				RequesterUserID: requesterID,
				Title:           "Existing",
				Status:          domainprocurement.RequestStatusDraft,
				CurrencyCode:    "USD",
			}, nil
		},
		updateRequestFn: func(context.Context, UpdateRequestParams) (domainprocurement.Request, error) {
			t.Fatal("update should not be called")
			return domainprocurement.Request{}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(_ context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})
	title := "Updated"
	_, err := service.UpdateRequest(context.Background(), UpdateRequestInput{
		OrganizationID: orgID,
		RequestID:      requestID,
		CurrentUser:    otherUserID,
		Title:          &title,
	})
	if !errors.Is(err, ErrForbiddenProcurement) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestServiceCreateItemAssignsNextLineNumber(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	requestID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleRequester,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRequestFn: func(_ context.Context, _, _ uuid.UUID) (domainprocurement.Request, error) {
			return domainprocurement.Request{
				ID:              requestID,
				OrganizationID:  orgID,
				RequesterUserID: userID,
				Title:           "Existing",
				Status:          domainprocurement.RequestStatusDraft,
				CurrencyCode:    "USD",
			}, nil
		},
		listItemsFn: func(_ context.Context, _, _ uuid.UUID) ([]domainprocurement.Item, error) {
			return []domainprocurement.Item{
				{LineNumber: 1},
				{LineNumber: 3},
			}, nil
		},
		createItemFn: func(_ context.Context, params CreateItemParams) (domainprocurement.Item, error) {
			if params.LineNumber != 4 {
				t.Fatalf("expected line number 4, got %d", params.LineNumber)
			}
			return domainprocurement.Item{
				ID:                   uuid.New(),
				OrganizationID:       params.OrganizationID,
				ProcurementRequestID: params.ProcurementRequestID,
				LineNumber:           params.LineNumber,
				ItemName:             params.ItemName,
				Quantity:             params.Quantity,
				Unit:                 params.Unit,
				CreatedAt:            now,
				UpdatedAt:            now,
			}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(_ context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})
	item, err := service.CreateItem(context.Background(), CreateItemInput{
		OrganizationID: orgID,
		RequestID:      requestID,
		CurrentUser:    userID,
		ItemName:       "Chair",
		Quantity:       "2",
		Unit:           "pcs",
	})
	if err != nil {
		t.Fatalf("create item returned error: %v", err)
	}
	if item.LineNumber != 4 {
		t.Fatalf("expected line number 4, got %d", item.LineNumber)
	}
}

func TestServiceListApprovalInboxAllowsApprover(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	userID := uuid.New()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, gotOrgID, gotUserID uuid.UUID) (domainorganization.Membership, error) {
			if gotOrgID != orgID || gotUserID != userID {
				t.Fatalf("unexpected membership lookup")
			}
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleApprover,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		listInboxFn: func(_ context.Context, gotOrgID uuid.UUID) ([]domainprocurement.Request, error) {
			if gotOrgID != orgID {
				t.Fatalf("unexpected org id: %s", gotOrgID)
			}
			return []domainprocurement.Request{{ID: uuid.New(), Status: domainprocurement.RequestStatusSubmitted}}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(_ context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})
	requests, err := service.ListApprovalInbox(context.Background(), ListApprovalInboxInput{
		OrganizationID: orgID,
		CurrentUser:    userID,
	})
	if err != nil {
		t.Fatalf("list approval inbox returned error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}
}

func TestServiceApproveRequestRejectsRequesterRole(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	requestID := uuid.New()
	userID := uuid.New()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleRequester,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRequestFn: func(_ context.Context, _, _ uuid.UUID) (domainprocurement.Request, error) {
			return domainprocurement.Request{
				ID:             requestID,
				OrganizationID: orgID,
				Status:         domainprocurement.RequestStatusSubmitted,
			}, nil
		},
		approveRequestFn: func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, *string) (domainprocurement.Request, error) {
			t.Fatal("approve should not be called")
			return domainprocurement.Request{}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(_ context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})
	_, err := service.ApproveRequest(context.Background(), DecisionInput{
		OrganizationID: orgID,
		RequestID:      requestID,
		CurrentUser:    userID,
	})
	if !errors.Is(err, ErrForbiddenProcurement) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestServiceApproveRequestAllowsApprover(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	requestID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()
	comment := "approved for next stage"
	logged := false

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleApprover,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRequestFn: func(_ context.Context, _, _ uuid.UUID) (domainprocurement.Request, error) {
			return domainprocurement.Request{
				ID:             requestID,
				OrganizationID: orgID,
				Status:         domainprocurement.RequestStatusSubmitted,
			}, nil
		},
		approveRequestFn: func(_ context.Context, gotOrgID, gotRequestID, gotUserID uuid.UUID, gotComment *string) (domainprocurement.Request, error) {
			if gotOrgID != orgID || gotRequestID != requestID || gotUserID != userID {
				t.Fatalf("unexpected approve args")
			}
			if gotComment == nil || *gotComment != comment {
				t.Fatalf("unexpected comment: %#v", gotComment)
			}
			return domainprocurement.Request{
				ID:              requestID,
				OrganizationID:  orgID,
				Status:          domainprocurement.RequestStatusApproved,
				DecisionComment: gotComment,
				ApprovedAt:      &now,
			}, nil
		},
		createActivityLogFn: func(_ context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error) {
			logged = true
			if params.EntityType != string(domainactivitylog.EntityTypeProcurementRequest) || params.Action != domainactivitylog.ActionProcurementRequestApproved {
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
	request, err := service.ApproveRequest(context.Background(), DecisionInput{
		OrganizationID:  orgID,
		RequestID:       requestID,
		CurrentUser:     userID,
		DecisionComment: &comment,
	})
	if err != nil {
		t.Fatalf("approve request returned error: %v", err)
	}
	if request.Status != domainprocurement.RequestStatusApproved {
		t.Fatalf("expected approved status, got %s", request.Status)
	}
	if !logged {
		t.Fatalf("expected activity log to be written")
	}
}

func TestServiceRejectRequestRequiresSubmittedStatus(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	requestID := uuid.New()
	userID := uuid.New()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleApprover,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRequestFn: func(_ context.Context, _, _ uuid.UUID) (domainprocurement.Request, error) {
			return domainprocurement.Request{
				ID:             requestID,
				OrganizationID: orgID,
				Status:         domainprocurement.RequestStatusDraft,
			}, nil
		},
		rejectRequestFn: func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, *string) (domainprocurement.Request, error) {
			t.Fatal("reject should not be called")
			return domainprocurement.Request{}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(_ context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})
	_, err := service.RejectRequest(context.Background(), DecisionInput{
		OrganizationID: orgID,
		RequestID:      requestID,
		CurrentUser:    userID,
	})
	if !errors.Is(err, ErrInvalidProcurementRequest) {
		t.Fatalf("expected invalid request error, got %v", err)
	}
}

func TestServiceSubmitRequestAllowsManager(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	requestID := uuid.New()
	currentUser := uuid.New()
	logged := false

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleProcurementOfficer,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRequestFn: func(_ context.Context, _, _ uuid.UUID) (domainprocurement.Request, error) {
			return domainprocurement.Request{
				ID:              requestID,
				OrganizationID:  orgID,
				RequesterUserID: uuid.New(),
				Title:           "Existing",
				Status:          domainprocurement.RequestStatusDraft,
				CurrencyCode:    "USD",
			}, nil
		},
		submitRequestFn: func(_ context.Context, gotOrgID, gotRequestID, gotUserID uuid.UUID) (domainprocurement.Request, error) {
			if gotOrgID != orgID || gotRequestID != requestID || gotUserID != currentUser {
				t.Fatalf("unexpected submit args")
			}
			return domainprocurement.Request{
				ID:              requestID,
				OrganizationID:  orgID,
				RequesterUserID: currentUser,
				Title:           "Existing",
				Status:          domainprocurement.RequestStatusSubmitted,
				CurrencyCode:    "USD",
			}, nil
		},
		createActivityLogFn: func(_ context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error) {
			logged = true
			if params.EntityType != string(domainactivitylog.EntityTypeProcurementRequest) || params.Action != domainactivitylog.ActionProcurementRequestSubmitted {
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
	request, err := service.SubmitRequest(context.Background(), SubmitRequestInput{
		OrganizationID: orgID,
		RequestID:      requestID,
		CurrentUser:    currentUser,
	})
	if err != nil {
		t.Fatalf("submit request returned error: %v", err)
	}
	if request.Status != domainprocurement.RequestStatusSubmitted {
		t.Fatalf("expected submitted status, got %s", request.Status)
	}
	if !logged {
		t.Fatalf("expected activity log to be written")
	}
}

func TestServiceRejectRequestWritesActivityLog(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	requestID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()
	comment := "budget constraints"
	logged := false

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleApprover,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRequestFn: func(_ context.Context, _, _ uuid.UUID) (domainprocurement.Request, error) {
			return domainprocurement.Request{
				ID:             requestID,
				OrganizationID: orgID,
				Status:         domainprocurement.RequestStatusSubmitted,
			}, nil
		},
		rejectRequestFn: func(_ context.Context, gotOrgID, gotRequestID, gotUserID uuid.UUID, gotComment *string) (domainprocurement.Request, error) {
			if gotOrgID != orgID || gotRequestID != requestID || gotUserID != userID {
				t.Fatalf("unexpected reject args")
			}
			if gotComment == nil || *gotComment != comment {
				t.Fatalf("unexpected comment: %#v", gotComment)
			}
			return domainprocurement.Request{
				ID:              requestID,
				OrganizationID:  orgID,
				Status:          domainprocurement.RequestStatusRejected,
				DecisionComment: gotComment,
				RejectedAt:      &now,
			}, nil
		},
		createActivityLogFn: func(_ context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error) {
			logged = true
			if params.EntityType != string(domainactivitylog.EntityTypeProcurementRequest) || params.Action != domainactivitylog.ActionProcurementRequestRejected {
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
	request, err := service.RejectRequest(context.Background(), DecisionInput{
		OrganizationID:  orgID,
		RequestID:       requestID,
		CurrentUser:     userID,
		DecisionComment: &comment,
	})
	if err != nil {
		t.Fatalf("reject request returned error: %v", err)
	}
	if request.Status != domainprocurement.RequestStatusRejected {
		t.Fatalf("expected rejected status, got %s", request.Status)
	}
	if !logged {
		t.Fatalf("expected activity log to be written")
	}
}

func TestServiceDeleteItemRejectsNonDraftRequest(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	requestID := uuid.New()
	itemID := uuid.New()
	userID := uuid.New()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleOwner,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getRequestFn: func(_ context.Context, _, _ uuid.UUID) (domainprocurement.Request, error) {
			return domainprocurement.Request{
				ID:              requestID,
				OrganizationID:  orgID,
				RequesterUserID: userID,
				Title:           "Existing",
				Status:          domainprocurement.RequestStatusSubmitted,
				CurrencyCode:    "USD",
			}, nil
		},
		deleteItemFn: func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
			t.Fatal("delete should not be called")
			return nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withinFn: func(_ context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})
	err := service.DeleteItem(context.Background(), DeleteItemInput{
		OrganizationID: orgID,
		RequestID:      requestID,
		ItemID:         itemID,
		CurrentUser:    userID,
	})
	if !errors.Is(err, ErrInvalidProcurementRequest) {
		t.Fatalf("expected invalid request error, got %v", err)
	}
}

func stringPtr(value string) *string {
	return &value
}
