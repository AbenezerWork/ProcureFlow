package organization

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationidentity "github.com/AbenezerWork/ProcureFlow/internal/application/identity"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainidentity "github.com/AbenezerWork/ProcureFlow/internal/domain/identity"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	"github.com/google/uuid"
)

type fakeRepository struct {
	createOrganizationFn     func(context.Context, CreateOrganizationParams) (domainorganization.Organization, error)
	getOrganizationFn        func(context.Context, uuid.UUID) (domainorganization.Organization, error)
	updateOrganizationFn     func(context.Context, UpdateOrganizationParams) (domainorganization.Organization, error)
	createMembershipFn       func(context.Context, CreateMembershipParams) (domainorganization.Membership, error)
	getMembershipFn          func(context.Context, uuid.UUID, uuid.UUID) (domainorganization.Membership, error)
	listMembershipsFn        func(context.Context, uuid.UUID) ([]domainorganization.Membership, error)
	updateMembershipRoleFn   func(context.Context, uuid.UUID, uuid.UUID, domainorganization.MembershipRole) (domainorganization.Membership, error)
	updateMembershipStatusFn func(context.Context, uuid.UUID, uuid.UUID, domainorganization.MembershipStatus) (domainorganization.Membership, error)
	listUserOrgsFn           func(context.Context, uuid.UUID) ([]domainorganization.UserOrganization, error)
	createActivityLogFn      func(context.Context, CreateActivityLogParams) (domainactivitylog.Entry, error)
}

func (f fakeRepository) CreateOrganization(ctx context.Context, params CreateOrganizationParams) (domainorganization.Organization, error) {
	return f.createOrganizationFn(ctx, params)
}

func (f fakeRepository) GetOrganization(ctx context.Context, organizationID uuid.UUID) (domainorganization.Organization, error) {
	return f.getOrganizationFn(ctx, organizationID)
}

func (f fakeRepository) UpdateOrganization(ctx context.Context, params UpdateOrganizationParams) (domainorganization.Organization, error) {
	return f.updateOrganizationFn(ctx, params)
}

func (f fakeRepository) CreateMembership(ctx context.Context, params CreateMembershipParams) (domainorganization.Membership, error) {
	return f.createMembershipFn(ctx, params)
}

func (f fakeRepository) GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error) {
	return f.getMembershipFn(ctx, organizationID, userID)
}

func (f fakeRepository) ListMemberships(ctx context.Context, organizationID uuid.UUID) ([]domainorganization.Membership, error) {
	return f.listMembershipsFn(ctx, organizationID)
}

func (f fakeRepository) UpdateMembershipRole(ctx context.Context, organizationID, userID uuid.UUID, role domainorganization.MembershipRole) (domainorganization.Membership, error) {
	return f.updateMembershipRoleFn(ctx, organizationID, userID, role)
}

func (f fakeRepository) UpdateMembershipStatus(ctx context.Context, organizationID, userID uuid.UUID, status domainorganization.MembershipStatus) (domainorganization.Membership, error) {
	return f.updateMembershipStatusFn(ctx, organizationID, userID, status)
}

func (f fakeRepository) ListUserOrganizations(ctx context.Context, userID uuid.UUID) ([]domainorganization.UserOrganization, error) {
	return f.listUserOrgsFn(ctx, userID)
}

func (f fakeRepository) CreateActivityLog(ctx context.Context, params CreateActivityLogParams) (domainactivitylog.Entry, error) {
	if f.createActivityLogFn == nil {
		return domainactivitylog.Entry{}, nil
	}
	return f.createActivityLogFn(ctx, params)
}

type fakeTxManager struct {
	withTransactionFn func(context.Context, func(Repository) error) error
}

func (f fakeTxManager) WithinTransaction(ctx context.Context, fn func(Repository) error) error {
	return f.withTransactionFn(ctx, fn)
}

func passthroughTx(repo Repository) fakeTxManager {
	return fakeTxManager{
		withTransactionFn: func(ctx context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	}
}

type fakeUserDirectory struct {
	getUserByIDFn    func(context.Context, uuid.UUID) (domainidentity.User, error)
	getUserByEmailFn func(context.Context, string) (domainidentity.User, error)
}

func (f fakeUserDirectory) GetUserByID(ctx context.Context, id uuid.UUID) (domainidentity.User, error) {
	return f.getUserByIDFn(ctx, id)
}

func (f fakeUserDirectory) GetUserByEmail(ctx context.Context, email string) (domainidentity.User, error) {
	return f.getUserByEmailFn(ctx, email)
}

func TestServiceCreate(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	userID := uuid.New()
	orgID := uuid.New()
	loggedOrganization := false
	loggedMembership := false

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
		createActivityLogFn: func(_ context.Context, params CreateActivityLogParams) (domainactivitylog.Entry, error) {
			switch {
			case params.EntityType == string(domainactivitylog.EntityTypeOrganization) && params.Action == domainactivitylog.ActionOrganizationCreated:
				loggedOrganization = true
			case params.EntityType == string(domainactivitylog.EntityTypeMembership) && params.Action == domainactivitylog.ActionMembershipCreated:
				loggedMembership = true
			default:
				t.Fatalf("unexpected activity log payload: %#v", params)
			}
			return domainactivitylog.Entry{EntityID: params.EntityID, Action: params.Action}, nil
		},
	}

	service := NewService(repo, fakeTxManager{
		withTransactionFn: func(ctx context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	}, fakeUserDirectory{})

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
	if !loggedOrganization || !loggedMembership {
		t.Fatalf("expected organization and membership activity logs to be written")
	}
}

func TestServiceGetReturnsOrganizationForActiveMember(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	orgID := uuid.New()
	userID := uuid.New()
	repo := fakeRepository{
		getOrganizationFn: func(_ context.Context, id uuid.UUID) (domainorganization.Organization, error) {
			return domainorganization.Organization{
				ID:              id,
				Name:            "Acme",
				Slug:            "acme",
				CreatedByUserID: userID,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
		getMembershipFn: func(_ context.Context, organizationID, memberID uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				ID:              uuid.New(),
				OrganizationID:  organizationID,
				UserID:          memberID,
				Role:            domainorganization.MembershipRoleAdmin,
				Status:          domainorganization.MembershipStatusActive,
				CreatedByUserID: userID,
				InvitedAt:       now,
				ActivatedAt:     &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
	}

	service := NewService(repo, passthroughTx(repo), fakeUserDirectory{})

	got, err := service.Get(context.Background(), orgID, userID)
	if err != nil {
		t.Fatalf("get returned error: %v", err)
	}

	if got.Organization.ID != orgID {
		t.Fatalf("expected org %s, got %s", orgID, got.Organization.ID)
	}
}

func TestServiceUpdateNormalizesSlugFromName(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	orgID := uuid.New()
	userID := uuid.New()

	repo := fakeRepository{
		getOrganizationFn: func(_ context.Context, _ uuid.UUID) (domainorganization.Organization, error) {
			return domainorganization.Organization{
				ID:              orgID,
				Name:            "Acme",
				Slug:            "acme",
				CreatedByUserID: userID,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				ID:              uuid.New(),
				OrganizationID:  orgID,
				UserID:          userID,
				Role:            domainorganization.MembershipRoleOwner,
				Status:          domainorganization.MembershipStatusActive,
				CreatedByUserID: userID,
				InvitedAt:       now,
				ActivatedAt:     &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
		updateOrganizationFn: func(_ context.Context, params UpdateOrganizationParams) (domainorganization.Organization, error) {
			if params.Name != "Acme & Sons" {
				t.Fatalf("unexpected name: %s", params.Name)
			}
			if params.Slug != "acme-sons" {
				t.Fatalf("unexpected slug: %s", params.Slug)
			}

			return domainorganization.Organization{
				ID:              params.ID,
				Name:            params.Name,
				Slug:            params.Slug,
				CreatedByUserID: userID,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
	}

	service := NewService(repo, passthroughTx(repo), fakeUserDirectory{})
	name := "Acme & Sons"

	updated, err := service.Update(context.Background(), UpdateInput{
		OrganizationID: orgID,
		CurrentUser:    userID,
		Name:           &name,
	})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}

	if updated.Organization.Slug != "acme-sons" {
		t.Fatalf("expected slug to be normalized, got %s", updated.Organization.Slug)
	}
}

func TestServiceAddMembershipByEmail(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	orgID := uuid.New()
	managerID := uuid.New()
	memberID := uuid.New()
	member := domainidentity.User{
		ID:        memberID,
		Email:     "member@example.com",
		FullName:  "Member User",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	repo := fakeRepository{
		getOrganizationFn: func(_ context.Context, _ uuid.UUID) (domainorganization.Organization, error) {
			return domainorganization.Organization{ID: orgID, Name: "Acme", Slug: "acme", CreatedByUserID: managerID, CreatedAt: now, UpdatedAt: now}, nil
		},
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				ID:              uuid.New(),
				OrganizationID:  orgID,
				UserID:          managerID,
				Role:            domainorganization.MembershipRoleAdmin,
				Status:          domainorganization.MembershipStatusActive,
				CreatedByUserID: managerID,
				InvitedAt:       now,
				ActivatedAt:     &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
		createMembershipFn: func(_ context.Context, params CreateMembershipParams) (domainorganization.Membership, error) {
			if params.UserID != memberID {
				t.Fatalf("unexpected user id: %s", params.UserID)
			}
			if params.Status != domainorganization.MembershipStatusInvited {
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
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
	}

	service := NewService(repo, passthroughTx(repo), fakeUserDirectory{
		getUserByEmailFn: func(_ context.Context, email string) (domainidentity.User, error) {
			if email != "member@example.com" {
				t.Fatalf("unexpected email: %s", email)
			}

			return member, nil
		},
	})

	created, err := service.AddMembership(context.Background(), AddMembershipInput{
		OrganizationID: orgID,
		CurrentUser:    managerID,
		Email:          " Member@Example.com ",
		Role:           domainorganization.MembershipRoleViewer,
	})
	if err != nil {
		t.Fatalf("add membership returned error: %v", err)
	}

	if created.User.ID != memberID {
		t.Fatalf("expected member %s, got %s", memberID, created.User.ID)
	}
}

func TestServiceAddMembershipRejectsAdminCreatingOwner(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	orgID := uuid.New()
	managerID := uuid.New()
	targetID := uuid.New()

	repo := fakeRepository{
		getOrganizationFn: func(_ context.Context, _ uuid.UUID) (domainorganization.Organization, error) {
			return domainorganization.Organization{ID: orgID, Name: "Acme", Slug: "acme", CreatedByUserID: managerID, CreatedAt: now, UpdatedAt: now}, nil
		},
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				ID:              uuid.New(),
				OrganizationID:  orgID,
				UserID:          managerID,
				Role:            domainorganization.MembershipRoleAdmin,
				Status:          domainorganization.MembershipStatusActive,
				CreatedByUserID: managerID,
				InvitedAt:       now,
				ActivatedAt:     &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
		createMembershipFn: func(context.Context, CreateMembershipParams) (domainorganization.Membership, error) {
			return domainorganization.Membership{}, errors.New("unexpected call")
		},
	}

	service := NewService(repo, passthroughTx(repo), fakeUserDirectory{
		getUserByIDFn: func(_ context.Context, id uuid.UUID) (domainidentity.User, error) {
			return domainidentity.User{ID: id, Email: "target@example.com", FullName: "Target", IsActive: true, CreatedAt: now, UpdatedAt: now}, nil
		},
	})

	_, err := service.AddMembership(context.Background(), AddMembershipInput{
		OrganizationID: orgID,
		CurrentUser:    managerID,
		UserID:         &targetID,
		Role:           domainorganization.MembershipRoleOwner,
		Status:         domainorganization.MembershipStatusActive,
	})
	if !errors.Is(err, ErrForbiddenOrganization) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestServiceListMembershipsRequiresManagerRole(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	orgID := uuid.New()
	userID := uuid.New()
	repo := fakeRepository{
		getOrganizationFn: func(_ context.Context, _ uuid.UUID) (domainorganization.Organization, error) {
			return domainorganization.Organization{ID: orgID, Name: "Acme", Slug: "acme", CreatedByUserID: userID, CreatedAt: now, UpdatedAt: now}, nil
		},
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				ID:              uuid.New(),
				OrganizationID:  orgID,
				UserID:          userID,
				Role:            domainorganization.MembershipRoleViewer,
				Status:          domainorganization.MembershipStatusActive,
				CreatedByUserID: userID,
				InvitedAt:       now,
				ActivatedAt:     &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
		listMembershipsFn: func(context.Context, uuid.UUID) ([]domainorganization.Membership, error) {
			return nil, errors.New("unexpected call")
		},
	}

	service := NewService(repo, passthroughTx(repo), fakeUserDirectory{})

	_, err := service.ListMemberships(context.Background(), orgID, userID)
	if !errors.Is(err, ErrForbiddenOrganization) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestServiceUpdateMembershipRejectsOwnMembership(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	orgID := uuid.New()
	userID := uuid.New()
	repo := fakeRepository{
		getOrganizationFn: func(_ context.Context, _ uuid.UUID) (domainorganization.Organization, error) {
			return domainorganization.Organization{ID: orgID, Name: "Acme", Slug: "acme", CreatedByUserID: userID, CreatedAt: now, UpdatedAt: now}, nil
		},
		getMembershipFn: func(_ context.Context, _, _ uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				ID:              uuid.New(),
				OrganizationID:  orgID,
				UserID:          userID,
				Role:            domainorganization.MembershipRoleOwner,
				Status:          domainorganization.MembershipStatusActive,
				CreatedByUserID: userID,
				InvitedAt:       now,
				ActivatedAt:     &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
	}

	service := NewService(repo, passthroughTx(repo), fakeUserDirectory{})
	role := domainorganization.MembershipRoleAdmin

	_, err := service.UpdateMembership(context.Background(), UpdateMembershipInput{
		OrganizationID: orgID,
		CurrentUser:    userID,
		TargetUserID:   userID,
		Role:           &role,
	})
	if !errors.Is(err, ErrCannotModifyOwnMembership) {
		t.Fatalf("expected self-modification error, got %v", err)
	}
}

func TestServiceUpdateMembershipUpdatesRoleAndStatus(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	orgID := uuid.New()
	managerID := uuid.New()
	targetID := uuid.New()

	repo := fakeRepository{
		getOrganizationFn: func(_ context.Context, _ uuid.UUID) (domainorganization.Organization, error) {
			return domainorganization.Organization{ID: orgID, Name: "Acme", Slug: "acme", CreatedByUserID: managerID, CreatedAt: now, UpdatedAt: now}, nil
		},
		getMembershipFn: func(_ context.Context, _, userID uuid.UUID) (domainorganization.Membership, error) {
			switch userID {
			case managerID:
				return domainorganization.Membership{
					ID:              uuid.New(),
					OrganizationID:  orgID,
					UserID:          managerID,
					Role:            domainorganization.MembershipRoleOwner,
					Status:          domainorganization.MembershipStatusActive,
					CreatedByUserID: managerID,
					InvitedAt:       now,
					ActivatedAt:     &now,
					CreatedAt:       now,
					UpdatedAt:       now,
				}, nil
			case targetID:
				return domainorganization.Membership{
					ID:              uuid.New(),
					OrganizationID:  orgID,
					UserID:          targetID,
					Role:            domainorganization.MembershipRoleViewer,
					Status:          domainorganization.MembershipStatusInvited,
					CreatedByUserID: managerID,
					InvitedAt:       now,
					CreatedAt:       now,
					UpdatedAt:       now,
				}, nil
			default:
				return domainorganization.Membership{}, ErrMembershipNotFound
			}
		},
		updateMembershipRoleFn: func(_ context.Context, _, _ uuid.UUID, role domainorganization.MembershipRole) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				ID:              uuid.New(),
				OrganizationID:  orgID,
				UserID:          targetID,
				Role:            role,
				Status:          domainorganization.MembershipStatusInvited,
				CreatedByUserID: managerID,
				InvitedAt:       now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
		updateMembershipStatusFn: func(_ context.Context, _, _ uuid.UUID, status domainorganization.MembershipStatus) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				ID:              uuid.New(),
				OrganizationID:  orgID,
				UserID:          targetID,
				Role:            domainorganization.MembershipRoleApprover,
				Status:          status,
				CreatedByUserID: managerID,
				InvitedAt:       now,
				ActivatedAt:     &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
	}

	service := NewService(repo, passthroughTx(repo), fakeUserDirectory{
		getUserByIDFn: func(_ context.Context, id uuid.UUID) (domainidentity.User, error) {
			return domainidentity.User{ID: id, Email: "target@example.com", FullName: "Target", IsActive: true, CreatedAt: now, UpdatedAt: now}, nil
		},
	})

	role := domainorganization.MembershipRoleApprover
	status := domainorganization.MembershipStatusActive
	updated, err := service.UpdateMembership(context.Background(), UpdateMembershipInput{
		OrganizationID: orgID,
		CurrentUser:    managerID,
		TargetUserID:   targetID,
		Role:           &role,
		Status:         &status,
	})
	if err != nil {
		t.Fatalf("update membership returned error: %v", err)
	}

	if updated.Membership.Role != domainorganization.MembershipRoleApprover {
		t.Fatalf("expected approver role, got %s", updated.Membership.Role)
	}
	if updated.Membership.Status != domainorganization.MembershipStatusActive {
		t.Fatalf("expected active status, got %s", updated.Membership.Status)
	}
}

func TestServiceTransferOwnership(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	orgID := uuid.New()
	currentOwnerID := uuid.New()
	targetID := uuid.New()

	repo := fakeRepository{
		getOrganizationFn: func(_ context.Context, _ uuid.UUID) (domainorganization.Organization, error) {
			return domainorganization.Organization{ID: orgID, Name: "Acme", Slug: "acme", CreatedByUserID: currentOwnerID, CreatedAt: now, UpdatedAt: now}, nil
		},
		getMembershipFn: func(_ context.Context, _, userID uuid.UUID) (domainorganization.Membership, error) {
			switch userID {
			case currentOwnerID:
				return domainorganization.Membership{
					ID:              uuid.New(),
					OrganizationID:  orgID,
					UserID:          currentOwnerID,
					Role:            domainorganization.MembershipRoleOwner,
					Status:          domainorganization.MembershipStatusActive,
					CreatedByUserID: currentOwnerID,
					InvitedAt:       now,
					ActivatedAt:     &now,
					CreatedAt:       now,
					UpdatedAt:       now,
				}, nil
			case targetID:
				return domainorganization.Membership{
					ID:              uuid.New(),
					OrganizationID:  orgID,
					UserID:          targetID,
					Role:            domainorganization.MembershipRoleAdmin,
					Status:          domainorganization.MembershipStatusActive,
					CreatedByUserID: currentOwnerID,
					InvitedAt:       now,
					ActivatedAt:     &now,
					CreatedAt:       now,
					UpdatedAt:       now,
				}, nil
			default:
				return domainorganization.Membership{}, ErrMembershipNotFound
			}
		},
		updateMembershipRoleFn: func(_ context.Context, _, userID uuid.UUID, role domainorganization.MembershipRole) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				ID:              uuid.New(),
				OrganizationID:  orgID,
				UserID:          userID,
				Role:            role,
				Status:          domainorganization.MembershipStatusActive,
				CreatedByUserID: currentOwnerID,
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
	}, fakeUserDirectory{
		getUserByIDFn: func(_ context.Context, id uuid.UUID) (domainidentity.User, error) {
			return domainidentity.User{
				ID:        id,
				Email:     "user@example.com",
				FullName:  "User",
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	})

	result, err := service.TransferOwnership(context.Background(), TransferOwnershipInput{
		OrganizationID: orgID,
		CurrentUser:    currentOwnerID,
		TargetUserID:   targetID,
	})
	if err != nil {
		t.Fatalf("transfer ownership returned error: %v", err)
	}

	if result.NewOwner.Membership.Role != domainorganization.MembershipRoleOwner {
		t.Fatalf("expected target to become owner, got %s", result.NewOwner.Membership.Role)
	}
	if result.PreviousOwner.Membership.Role != domainorganization.MembershipRoleAdmin {
		t.Fatalf("expected previous owner to become admin, got %s", result.PreviousOwner.Membership.Role)
	}
}

func TestServiceTransferOwnershipRequiresOwner(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	orgID := uuid.New()
	currentUserID := uuid.New()
	targetID := uuid.New()

	repo := fakeRepository{
		getOrganizationFn: func(_ context.Context, _ uuid.UUID) (domainorganization.Organization, error) {
			return domainorganization.Organization{ID: orgID, Name: "Acme", Slug: "acme", CreatedByUserID: currentUserID, CreatedAt: now, UpdatedAt: now}, nil
		},
		getMembershipFn: func(_ context.Context, _, userID uuid.UUID) (domainorganization.Membership, error) {
			return domainorganization.Membership{
				ID:              uuid.New(),
				OrganizationID:  orgID,
				UserID:          userID,
				Role:            domainorganization.MembershipRoleAdmin,
				Status:          domainorganization.MembershipStatusActive,
				CreatedByUserID: currentUserID,
				InvitedAt:       now,
				ActivatedAt:     &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
	}

	service := NewService(repo, passthroughTx(repo), fakeUserDirectory{})

	_, err := service.TransferOwnership(context.Background(), TransferOwnershipInput{
		OrganizationID: orgID,
		CurrentUser:    currentUserID,
		TargetUserID:   targetID,
	})
	if !errors.Is(err, ErrForbiddenOrganization) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestServiceTransferOwnershipRequiresActiveTargetMember(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	orgID := uuid.New()
	currentOwnerID := uuid.New()
	targetID := uuid.New()

	repo := fakeRepository{
		getOrganizationFn: func(_ context.Context, _ uuid.UUID) (domainorganization.Organization, error) {
			return domainorganization.Organization{ID: orgID, Name: "Acme", Slug: "acme", CreatedByUserID: currentOwnerID, CreatedAt: now, UpdatedAt: now}, nil
		},
		getMembershipFn: func(_ context.Context, _, userID uuid.UUID) (domainorganization.Membership, error) {
			switch userID {
			case currentOwnerID:
				return domainorganization.Membership{
					ID:              uuid.New(),
					OrganizationID:  orgID,
					UserID:          currentOwnerID,
					Role:            domainorganization.MembershipRoleOwner,
					Status:          domainorganization.MembershipStatusActive,
					CreatedByUserID: currentOwnerID,
					InvitedAt:       now,
					ActivatedAt:     &now,
					CreatedAt:       now,
					UpdatedAt:       now,
				}, nil
			case targetID:
				return domainorganization.Membership{
					ID:              uuid.New(),
					OrganizationID:  orgID,
					UserID:          targetID,
					Role:            domainorganization.MembershipRoleAdmin,
					Status:          domainorganization.MembershipStatusInvited,
					CreatedByUserID: currentOwnerID,
					InvitedAt:       now,
					CreatedAt:       now,
					UpdatedAt:       now,
				}, nil
			default:
				return domainorganization.Membership{}, ErrMembershipNotFound
			}
		},
	}

	service := NewService(repo, fakeTxManager{
		withTransactionFn: func(ctx context.Context, fn func(Repository) error) error {
			return fn(repo)
		},
	}, fakeUserDirectory{})

	_, err := service.TransferOwnership(context.Background(), TransferOwnershipInput{
		OrganizationID: orgID,
		CurrentUser:    currentOwnerID,
		TargetUserID:   targetID,
	})
	if !errors.Is(err, ErrInvalidOwnershipTransfer) {
		t.Fatalf("expected invalid ownership transfer error, got %v", err)
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
	}, fakeUserDirectory{})

	_, err := service.ListForUser(context.Background(), uuid.Nil)
	if !errors.Is(err, ErrInvalidOrganization) {
		t.Fatalf("expected invalid organization error, got %v", err)
	}
}

func TestResolveUserMapsMissingUser(t *testing.T) {
	t.Parallel()

	service := NewService(fakeRepository{}, fakeTxManager{}, fakeUserDirectory{
		getUserByEmailFn: func(context.Context, string) (domainidentity.User, error) {
			return domainidentity.User{}, applicationidentity.ErrUserNotFound
		},
	})

	_, err := service.resolveUser(context.Background(), nil, "missing@example.com")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected user not found, got %v", err)
	}
}
