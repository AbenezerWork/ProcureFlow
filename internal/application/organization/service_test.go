package organization

import (
	"context"
	"errors"
	"testing"
	"time"

	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	"github.com/google/uuid"
)

type fakeRepository struct {
	createOrganizationFn func(context.Context, CreateOrganizationParams) (domainorganization.Organization, error)
	createMembershipFn   func(context.Context, CreateMembershipParams) (domainorganization.Membership, error)
	listUserOrgsFn       func(context.Context, uuid.UUID) ([]domainorganization.UserOrganization, error)
}

func (f fakeRepository) CreateOrganization(ctx context.Context, params CreateOrganizationParams) (domainorganization.Organization, error) {
	return f.createOrganizationFn(ctx, params)
}

func (f fakeRepository) CreateMembership(ctx context.Context, params CreateMembershipParams) (domainorganization.Membership, error) {
	return f.createMembershipFn(ctx, params)
}

func (f fakeRepository) ListUserOrganizations(ctx context.Context, userID uuid.UUID) ([]domainorganization.UserOrganization, error) {
	return f.listUserOrgsFn(ctx, userID)
}

type fakeTxManager struct {
	withTransactionFn func(context.Context, func(Repository) error) error
}

func (f fakeTxManager) WithinTransaction(ctx context.Context, fn func(Repository) error) error {
	return f.withTransactionFn(ctx, fn)
}

func TestServiceCreate(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	userID := uuid.New()
	orgID := uuid.New()

	repo := fakeRepository{
		createOrganizationFn: func(_ context.Context, params CreateOrganizationParams) (domainorganization.Organization, error) {
			if params.Slug != "acme-corporation" {
				t.Fatalf("unexpected slug: %s", params.Slug)
			}

			return domainorganization.Organization{
				ID:              orgID,
				Name:            params.Name,
				Slug:            params.Slug,
				CreatedByUserID: params.CreatedByUserID,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
		createMembershipFn: func(_ context.Context, params CreateMembershipParams) (domainorganization.Membership, error) {
			if params.Role != domainorganization.MembershipRoleOwner {
				t.Fatalf("unexpected role: %s", params.Role)
			}
			if params.Status != domainorganization.MembershipStatusActive {
				t.Fatalf("unexpected status: %s", params.Status)
			}

			return domainorganization.Membership{
				ID:              uuid.New(),
				OrganizationID:  params.OrganizationID,
				UserID:          params.UserID,
				Role:            params.Role,
				Status:          params.Status,
				CreatedByUserID: params.CreatedByUserID,
				InvitedAt:       now,
				ActivatedAt:     &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withTransactionFn: func(ctx context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	})

	created, err := service.Create(context.Background(), CreateInput{
		Name:        "Acme Corporation",
		CurrentUser: userID,
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	if created.Organization.ID != orgID {
		t.Fatalf("expected organization ID %s, got %s", orgID, created.Organization.ID)
	}
}

func TestNormalizeSlug(t *testing.T) {
	t.Parallel()

	if got, want := normalizeSlug("  ACME & Sons, Inc.  "), "acme-sons-inc"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestServiceListForUserInvalid(t *testing.T) {
	t.Parallel()

	service := NewService(fakeRepository{
		listUserOrgsFn: func(context.Context, uuid.UUID) ([]domainorganization.UserOrganization, error) {
			return nil, errors.New("unexpected call")
		},
	}, fakeTxManager{
		withTransactionFn: func(context.Context, func(Repository) error) error {
			return nil
		},
	})

	_, err := service.ListForUser(context.Background(), uuid.Nil)
	if !errors.Is(err, ErrInvalidOrganization) {
		t.Fatalf("expected invalid organization error, got %v", err)
	}
}
