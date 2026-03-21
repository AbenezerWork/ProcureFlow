package vendor

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainvendor "github.com/AbenezerWork/ProcureFlow/internal/domain/vendor"
	"github.com/google/uuid"
)

type fakeRepository struct {
	createVendorFn      func(context.Context, CreateVendorParams) (domainvendor.Vendor, error)
	getVendorFn         func(context.Context, uuid.UUID, uuid.UUID) (domainvendor.Vendor, error)
	listVendorsFn       func(context.Context, uuid.UUID, *domainvendor.Status) ([]domainvendor.Vendor, error)
	searchVendorsFn     func(context.Context, uuid.UUID, string) ([]domainvendor.Vendor, error)
	updateVendorFn      func(context.Context, UpdateVendorParams) (domainvendor.Vendor, error)
	archiveVendorFn     func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domainvendor.Vendor, error)
	getMembershipFn     func(context.Context, uuid.UUID, uuid.UUID) (domainorganization.Membership, error)
	createActivityLogFn func(context.Context, applicationactivitylog.CreateParams) (domainactivitylog.Entry, error)
}

func (f fakeRepository) CreateVendor(ctx context.Context, params CreateVendorParams) (domainvendor.Vendor, error) {
	return f.createVendorFn(ctx, params)
}

func (f fakeRepository) GetVendor(ctx context.Context, organizationID, vendorID uuid.UUID) (domainvendor.Vendor, error) {
	return f.getVendorFn(ctx, organizationID, vendorID)
}

func (f fakeRepository) ListVendors(ctx context.Context, organizationID uuid.UUID, status *domainvendor.Status) ([]domainvendor.Vendor, error) {
	return f.listVendorsFn(ctx, organizationID, status)
}

func (f fakeRepository) SearchVendors(ctx context.Context, organizationID uuid.UUID, search string) ([]domainvendor.Vendor, error) {
	return f.searchVendorsFn(ctx, organizationID, search)
}

func (f fakeRepository) UpdateVendor(ctx context.Context, params UpdateVendorParams) (domainvendor.Vendor, error) {
	return f.updateVendorFn(ctx, params)
}

func (f fakeRepository) ArchiveVendor(ctx context.Context, organizationID, vendorID, updatedByUserID uuid.UUID) (domainvendor.Vendor, error) {
	return f.archiveVendorFn(ctx, organizationID, vendorID, updatedByUserID)
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

func passthroughTx(repo Repository) fakeTxManager {
	return fakeTxManager{
		withinFn: func(ctx context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	}
}

func TestServiceCreateVendor(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()
	logged := false
	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, gotOrgID, gotUserID uuid.UUID) (domainorganization.Membership, error) {
			if gotOrgID != orgID || gotUserID != userID {
				t.Fatalf("unexpected membership lookup")
			}
			return domainorganization.Membership{
				ID:             uuid.New(),
				OrganizationID: orgID,
				UserID:         userID,
				Role:           domainorganization.MembershipRoleProcurementOfficer,
				Status:         domainorganization.MembershipStatusActive,
				InvitedAt:      now,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil
		},
		createVendorFn: func(_ context.Context, params CreateVendorParams) (domainvendor.Vendor, error) {
			if params.Name != "Acme Supplies" {
				t.Fatalf("unexpected vendor name: %s", params.Name)
			}
			if params.CreatedByUserID != userID {
				t.Fatalf("unexpected creator: %s", params.CreatedByUserID)
			}
			return domainvendor.Vendor{
				ID:              uuid.New(),
				OrganizationID:  params.OrganizationID,
				Name:            params.Name,
				Status:          domainvendor.StatusActive,
				CreatedByUserID: params.CreatedByUserID,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
		createActivityLogFn: func(_ context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error) {
			logged = true
			if params.EntityType != string(domainactivitylog.EntityTypeVendor) || params.Action != domainactivitylog.ActionVendorCreated {
				t.Fatalf("unexpected activity log payload: %#v", params)
			}
			return domainactivitylog.Entry{EntityID: params.EntityID, Action: params.Action}, nil
		},
	}

	service := NewService(repo, passthroughTx(repo))
	created, err := service.Create(context.Background(), CreateInput{
		OrganizationID: orgID,
		CurrentUser:    userID,
		Name:           "  Acme Supplies  ",
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	if created.Name != "Acme Supplies" {
		t.Fatalf("expected normalized name, got %q", created.Name)
	}
	if !logged {
		t.Fatalf("expected vendor activity log to be written")
	}
}

func TestServiceCreateVendorRejectsUnauthorizedRole(t *testing.T) {
	t.Parallel()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleRequester,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		createVendorFn: func(context.Context, CreateVendorParams) (domainvendor.Vendor, error) {
			t.Fatal("create should not be called")
			return domainvendor.Vendor{}, nil
		},
	}

	service := NewService(repo, passthroughTx(repo))
	_, err := service.Create(context.Background(), CreateInput{
		OrganizationID: uuid.New(),
		CurrentUser:    uuid.New(),
		Name:           "Acme",
	})
	if !errors.Is(err, ErrForbiddenVendor) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestServiceListSearchFiltersByStatus(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	userID := uuid.New()
	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleViewer,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		searchVendorsFn: func(_ context.Context, gotOrgID uuid.UUID, search string) ([]domainvendor.Vendor, error) {
			if gotOrgID != orgID {
				t.Fatalf("unexpected organization ID: %s", gotOrgID)
			}
			if search != "acme" {
				t.Fatalf("unexpected search: %s", search)
			}
			return []domainvendor.Vendor{
				{ID: uuid.New(), OrganizationID: orgID, Name: "Acme Active", Status: domainvendor.StatusActive},
				{ID: uuid.New(), OrganizationID: orgID, Name: "Acme Archived", Status: domainvendor.StatusArchived},
			}, nil
		},
	}

	service := NewService(repo, passthroughTx(repo))
	status := domainvendor.StatusActive
	vendors, err := service.List(context.Background(), ListInput{
		OrganizationID: orgID,
		CurrentUser:    userID,
		Status:         &status,
		Search:         "acme",
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	if len(vendors) != 1 || vendors[0].Status != domainvendor.StatusActive {
		t.Fatalf("expected one active vendor, got %#v", vendors)
	}
}

func TestServiceUpdateMergesOptionalFields(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	userID := uuid.New()
	vendorID := uuid.New()
	notes := "Updated notes"
	empty := "   "

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleAdmin,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		getVendorFn: func(_ context.Context, _, _ uuid.UUID) (domainvendor.Vendor, error) {
			contact := "Alice"
			return domainvendor.Vendor{
				ID:             vendorID,
				OrganizationID: orgID,
				Name:           "Acme",
				ContactName:    &contact,
				Status:         domainvendor.StatusActive,
			}, nil
		},
		updateVendorFn: func(_ context.Context, params UpdateVendorParams) (domainvendor.Vendor, error) {
			if params.Name != "Acme Updated" {
				t.Fatalf("unexpected name: %s", params.Name)
			}
			if params.ContactName != nil {
				t.Fatalf("expected contact name to be cleared")
			}
			if params.Notes == nil || *params.Notes != notes {
				t.Fatalf("expected notes to be set")
			}
			return domainvendor.Vendor{
				ID:             vendorID,
				OrganizationID: orgID,
				Name:           params.Name,
				Notes:          params.Notes,
				Status:         domainvendor.StatusActive,
			}, nil
		},
	}

	service := NewService(repo, passthroughTx(repo))
	name := "Acme Updated"
	updated, err := service.Update(context.Background(), UpdateInput{
		OrganizationID: orgID,
		VendorID:       vendorID,
		CurrentUser:    userID,
		Name:           &name,
		ContactName:    &empty,
		Notes:          &notes,
	})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}

	if updated.Name != "Acme Updated" {
		t.Fatalf("unexpected updated vendor name: %s", updated.Name)
	}
}

func TestServiceArchiveVendor(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	userID := uuid.New()
	vendorID := uuid.New()
	now := time.Now().UTC()

	repo := fakeRepository{
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				Role:   domainorganization.MembershipRoleOwner,
				Status: domainorganization.MembershipStatusActive,
			}, nil
		},
		archiveVendorFn: func(_ context.Context, gotOrgID, gotVendorID, gotUserID uuid.UUID) (domainvendor.Vendor, error) {
			if gotOrgID != orgID || gotVendorID != vendorID || gotUserID != userID {
				t.Fatalf("unexpected archive args")
			}
			return domainvendor.Vendor{
				ID:              vendorID,
				OrganizationID:  orgID,
				Name:            "Acme",
				Status:          domainvendor.StatusArchived,
				ArchivedAt:      &now,
				UpdatedByUserID: &userID,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
	}

	service := NewService(repo, passthroughTx(repo))
	archived, err := service.Archive(context.Background(), ArchiveInput{
		OrganizationID: orgID,
		VendorID:       vendorID,
		CurrentUser:    userID,
	})
	if err != nil {
		t.Fatalf("archive returned error: %v", err)
	}

	if archived.Status != domainvendor.StatusArchived {
		t.Fatalf("expected archived status, got %s", archived.Status)
	}
}
