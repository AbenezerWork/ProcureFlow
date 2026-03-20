package organization

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	applicationidentity "github.com/AbenezerWork/ProcureFlow/internal/application/identity"
	domainidentity "github.com/AbenezerWork/ProcureFlow/internal/domain/identity"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	"github.com/google/uuid"
)

var (
	ErrOrganizationSlugTaken     = errors.New("organization slug already exists")
	ErrOrganizationNotFound      = errors.New("organization not found")
	ErrInvalidOrganization       = errors.New("invalid organization")
	ErrInvalidOwnershipTransfer  = errors.New("invalid organization ownership transfer")
	ErrMembershipNotFound        = errors.New("organization membership not found")
	ErrMembershipAlreadyExists   = errors.New("organization membership already exists")
	ErrInvalidMembership         = errors.New("invalid organization membership")
	ErrUserNotFound              = errors.New("user not found")
	ErrForbiddenOrganization     = errors.New("forbidden organization operation")
	ErrCannotModifyOwnMembership = errors.New("cannot modify your own membership")
	ErrOwnerMembershipImmutable  = errors.New("owner memberships must be managed through a dedicated ownership flow")
)

type CreateOrganizationParams struct {
	Name            string
	Slug            string
	CreatedByUserID uuid.UUID
}

type UpdateOrganizationParams struct {
	ID   uuid.UUID
	Name string
	Slug string
}

type CreateMembershipParams struct {
	OrganizationID  uuid.UUID
	UserID          uuid.UUID
	Role            domainorganization.MembershipRole
	Status          domainorganization.MembershipStatus
	CreatedByUserID uuid.UUID
}

type Repository interface {
	CreateOrganization(ctx context.Context, params CreateOrganizationParams) (domainorganization.Organization, error)
	GetOrganization(ctx context.Context, organizationID uuid.UUID) (domainorganization.Organization, error)
	UpdateOrganization(ctx context.Context, params UpdateOrganizationParams) (domainorganization.Organization, error)
	CreateMembership(ctx context.Context, params CreateMembershipParams) (domainorganization.Membership, error)
	GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error)
	ListMemberships(ctx context.Context, organizationID uuid.UUID) ([]domainorganization.Membership, error)
	UpdateMembershipRole(ctx context.Context, organizationID, userID uuid.UUID, role domainorganization.MembershipRole) (domainorganization.Membership, error)
	UpdateMembershipStatus(ctx context.Context, organizationID, userID uuid.UUID, status domainorganization.MembershipStatus) (domainorganization.Membership, error)
	ListUserOrganizations(ctx context.Context, userID uuid.UUID) ([]domainorganization.UserOrganization, error)
}

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(repo Repository) error) error
}

type UserDirectory interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (domainidentity.User, error)
	GetUserByEmail(ctx context.Context, email string) (domainidentity.User, error)
}

type CreateInput struct {
	Name        string
	Slug        string
	CurrentUser uuid.UUID
}

type CreatedOrganization struct {
	Organization domainorganization.Organization
	Membership   domainorganization.Membership
}

type OrganizationDetails struct {
	Organization domainorganization.Organization
	Membership   domainorganization.Membership
}

type OrganizationMember struct {
	User       domainidentity.User
	Membership domainorganization.Membership
}

type UpdateInput struct {
	OrganizationID uuid.UUID
	CurrentUser    uuid.UUID
	Name           *string
	Slug           *string
}

type AddMembershipInput struct {
	OrganizationID uuid.UUID
	CurrentUser    uuid.UUID
	UserID         *uuid.UUID
	Email          string
	Role           domainorganization.MembershipRole
	Status         domainorganization.MembershipStatus
}

type UpdateMembershipInput struct {
	OrganizationID uuid.UUID
	CurrentUser    uuid.UUID
	TargetUserID   uuid.UUID
	Role           *domainorganization.MembershipRole
	Status         *domainorganization.MembershipStatus
}

type TransferOwnershipInput struct {
	OrganizationID      uuid.UUID
	CurrentUser         uuid.UUID
	TargetUserID        uuid.UUID
	CurrentOwnerNewRole domainorganization.MembershipRole
}

type OwnershipTransferResult struct {
	Organization  domainorganization.Organization
	PreviousOwner OrganizationMember
	NewOwner      OrganizationMember
}

type Service struct {
	repo  Repository
	tx    TransactionManager
	users UserDirectory
}

func NewService(repo Repository, tx TransactionManager, users UserDirectory) Service {
	return Service{repo: repo, tx: tx, users: users}
}

func (s Service) Create(ctx context.Context, input CreateInput) (CreatedOrganization, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" || input.CurrentUser == uuid.Nil {
		return CreatedOrganization{}, ErrInvalidOrganization
	}

	slug := normalizeSlug(input.Slug)
	if slug == "" {
		slug = normalizeSlug(name)
	}
	if slug == "" {
		return CreatedOrganization{}, ErrInvalidOrganization
	}

	var result CreatedOrganization
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		org, err := repo.CreateOrganization(ctx, CreateOrganizationParams{
			Name:            name,
			Slug:            slug,
			CreatedByUserID: input.CurrentUser,
		})
		if err != nil {
			return err
		}

		membership, err := repo.CreateMembership(ctx, CreateMembershipParams{
			OrganizationID:  org.ID,
			UserID:          input.CurrentUser,
			Role:            domainorganization.MembershipRoleOwner,
			Status:          domainorganization.MembershipStatusActive,
			CreatedByUserID: input.CurrentUser,
		})
		if err != nil {
			return err
		}

		result = CreatedOrganization{
			Organization: org,
			Membership:   membership,
		}
		return nil
	}); err != nil {
		if errors.Is(err, ErrOrganizationSlugTaken) {
			return CreatedOrganization{}, err
		}

		return CreatedOrganization{}, fmt.Errorf("create organization: %w", err)
	}

	return result, nil
}

func (s Service) ListForUser(ctx context.Context, userID uuid.UUID) ([]domainorganization.UserOrganization, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidOrganization
	}

	organizations, err := s.repo.ListUserOrganizations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user organizations: %w", err)
	}

	return organizations, nil
}

func (s Service) Get(ctx context.Context, organizationID, currentUser uuid.UUID) (OrganizationDetails, error) {
	if organizationID == uuid.Nil || currentUser == uuid.Nil {
		return OrganizationDetails{}, ErrInvalidOrganization
	}

	org, membership, err := s.loadAccessibleOrganization(ctx, organizationID, currentUser)
	if err != nil {
		return OrganizationDetails{}, err
	}

	return OrganizationDetails{
		Organization: org,
		Membership:   membership,
	}, nil
}

func (s Service) Update(ctx context.Context, input UpdateInput) (OrganizationDetails, error) {
	if input.OrganizationID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return OrganizationDetails{}, ErrInvalidOrganization
	}

	org, membership, err := s.loadAccessibleOrganization(ctx, input.OrganizationID, input.CurrentUser)
	if err != nil {
		return OrganizationDetails{}, err
	}
	if !canManageOrganization(membership.Role) {
		return OrganizationDetails{}, ErrForbiddenOrganization
	}

	name := org.Name
	if input.Name != nil {
		name = strings.TrimSpace(*input.Name)
	}
	slug := org.Slug
	if input.Slug != nil {
		slug = normalizeSlug(*input.Slug)
	}
	if input.Name != nil && input.Slug == nil {
		slug = normalizeSlug(name)
	}
	if name == "" || slug == "" {
		return OrganizationDetails{}, ErrInvalidOrganization
	}

	updated, err := s.repo.UpdateOrganization(ctx, UpdateOrganizationParams{
		ID:   input.OrganizationID,
		Name: name,
		Slug: slug,
	})
	if err != nil {
		return OrganizationDetails{}, fmt.Errorf("update organization: %w", err)
	}

	return OrganizationDetails{
		Organization: updated,
		Membership:   membership,
	}, nil
}

func (s Service) ListMemberships(ctx context.Context, organizationID, currentUser uuid.UUID) ([]OrganizationMember, error) {
	if organizationID == uuid.Nil || currentUser == uuid.Nil {
		return nil, ErrInvalidOrganization
	}

	if _, membership, err := s.loadAccessibleOrganization(ctx, organizationID, currentUser); err != nil {
		return nil, err
	} else if !canManageOrganization(membership.Role) {
		return nil, ErrForbiddenOrganization
	}

	memberships, err := s.repo.ListMemberships(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list memberships: %w", err)
	}

	return s.enrichMembers(ctx, memberships)
}

func (s Service) AddMembership(ctx context.Context, input AddMembershipInput) (OrganizationMember, error) {
	if input.OrganizationID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return OrganizationMember{}, ErrInvalidMembership
	}
	if !isValidRole(input.Role) {
		return OrganizationMember{}, ErrInvalidMembership
	}

	status := input.Status
	if status == "" {
		status = domainorganization.MembershipStatusInvited
	}
	if !isValidStatus(status) {
		return OrganizationMember{}, ErrInvalidMembership
	}

	_, managerMembership, err := s.loadAccessibleOrganization(ctx, input.OrganizationID, input.CurrentUser)
	if err != nil {
		return OrganizationMember{}, err
	}
	if !canManageOrganization(managerMembership.Role) {
		return OrganizationMember{}, ErrForbiddenOrganization
	}
	if input.Role == domainorganization.MembershipRoleOwner && managerMembership.Role != domainorganization.MembershipRoleOwner {
		return OrganizationMember{}, ErrForbiddenOrganization
	}

	user, err := s.resolveUser(ctx, input.UserID, input.Email)
	if err != nil {
		return OrganizationMember{}, err
	}

	created, err := s.repo.CreateMembership(ctx, CreateMembershipParams{
		OrganizationID:  input.OrganizationID,
		UserID:          user.ID,
		Role:            input.Role,
		Status:          status,
		CreatedByUserID: input.CurrentUser,
	})
	if err != nil {
		return OrganizationMember{}, fmt.Errorf("create membership: %w", err)
	}

	return OrganizationMember{
		User:       user,
		Membership: created,
	}, nil
}

func (s Service) UpdateMembership(ctx context.Context, input UpdateMembershipInput) (OrganizationMember, error) {
	if input.OrganizationID == uuid.Nil || input.CurrentUser == uuid.Nil || input.TargetUserID == uuid.Nil {
		return OrganizationMember{}, ErrInvalidMembership
	}
	if input.Role == nil && input.Status == nil {
		return OrganizationMember{}, ErrInvalidMembership
	}

	_, managerMembership, err := s.loadAccessibleOrganization(ctx, input.OrganizationID, input.CurrentUser)
	if err != nil {
		return OrganizationMember{}, err
	}
	if !canManageOrganization(managerMembership.Role) {
		return OrganizationMember{}, ErrForbiddenOrganization
	}
	if input.TargetUserID == input.CurrentUser {
		return OrganizationMember{}, ErrCannotModifyOwnMembership
	}

	targetMembership, err := s.repo.GetMembership(ctx, input.OrganizationID, input.TargetUserID)
	if err != nil {
		return OrganizationMember{}, fmt.Errorf("load target membership: %w", err)
	}
	if targetMembership.Role == domainorganization.MembershipRoleOwner {
		return OrganizationMember{}, ErrOwnerMembershipImmutable
	}

	updated := targetMembership
	if input.Role != nil {
		if !isValidRole(*input.Role) {
			return OrganizationMember{}, ErrInvalidMembership
		}
		if *input.Role == domainorganization.MembershipRoleOwner && managerMembership.Role != domainorganization.MembershipRoleOwner {
			return OrganizationMember{}, ErrForbiddenOrganization
		}
		if updated.Role != *input.Role {
			updated, err = s.repo.UpdateMembershipRole(ctx, input.OrganizationID, input.TargetUserID, *input.Role)
			if err != nil {
				return OrganizationMember{}, fmt.Errorf("update membership role: %w", err)
			}
		}
	}

	if input.Status != nil {
		if !isValidStatus(*input.Status) {
			return OrganizationMember{}, ErrInvalidMembership
		}
		if updated.Status != *input.Status {
			updated, err = s.repo.UpdateMembershipStatus(ctx, input.OrganizationID, input.TargetUserID, *input.Status)
			if err != nil {
				return OrganizationMember{}, fmt.Errorf("update membership status: %w", err)
			}
		}
	}

	user, err := s.users.GetUserByID(ctx, input.TargetUserID)
	if err != nil {
		if errors.Is(err, applicationidentity.ErrUserNotFound) {
			return OrganizationMember{}, ErrUserNotFound
		}

		return OrganizationMember{}, fmt.Errorf("load membership user: %w", err)
	}

	return OrganizationMember{
		User:       user,
		Membership: updated,
	}, nil
}

func (s Service) TransferOwnership(ctx context.Context, input TransferOwnershipInput) (OwnershipTransferResult, error) {
	if input.OrganizationID == uuid.Nil || input.CurrentUser == uuid.Nil || input.TargetUserID == uuid.Nil {
		return OwnershipTransferResult{}, ErrInvalidOwnershipTransfer
	}
	if input.TargetUserID == input.CurrentUser {
		return OwnershipTransferResult{}, ErrInvalidOwnershipTransfer
	}

	newRole := input.CurrentOwnerNewRole
	if newRole == "" {
		newRole = domainorganization.MembershipRoleAdmin
	}
	if !isValidRole(newRole) || newRole == domainorganization.MembershipRoleOwner {
		return OwnershipTransferResult{}, ErrInvalidOwnershipTransfer
	}

	org, currentMembership, err := s.loadAccessibleOrganization(ctx, input.OrganizationID, input.CurrentUser)
	if err != nil {
		return OwnershipTransferResult{}, err
	}
	if currentMembership.Role != domainorganization.MembershipRoleOwner {
		return OwnershipTransferResult{}, ErrForbiddenOrganization
	}

	var updatedCurrent domainorganization.Membership
	var updatedTarget domainorganization.Membership
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		targetMembership, err := repo.GetMembership(ctx, input.OrganizationID, input.TargetUserID)
		if err != nil {
			return err
		}
		if targetMembership.Status != domainorganization.MembershipStatusActive || targetMembership.Role == domainorganization.MembershipRoleOwner {
			return ErrInvalidOwnershipTransfer
		}

		updatedTarget, err = repo.UpdateMembershipRole(ctx, input.OrganizationID, input.TargetUserID, domainorganization.MembershipRoleOwner)
		if err != nil {
			return err
		}

		updatedCurrent, err = repo.UpdateMembershipRole(ctx, input.OrganizationID, input.CurrentUser, newRole)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		switch {
		case errors.Is(err, ErrMembershipNotFound), errors.Is(err, ErrInvalidOwnershipTransfer):
			return OwnershipTransferResult{}, err
		default:
			return OwnershipTransferResult{}, fmt.Errorf("transfer ownership: %w", err)
		}
	}

	currentUser, err := s.users.GetUserByID(ctx, input.CurrentUser)
	if err != nil {
		if errors.Is(err, applicationidentity.ErrUserNotFound) {
			return OwnershipTransferResult{}, ErrUserNotFound
		}

		return OwnershipTransferResult{}, fmt.Errorf("load current owner: %w", err)
	}

	targetUser, err := s.users.GetUserByID(ctx, input.TargetUserID)
	if err != nil {
		if errors.Is(err, applicationidentity.ErrUserNotFound) {
			return OwnershipTransferResult{}, ErrUserNotFound
		}

		return OwnershipTransferResult{}, fmt.Errorf("load new owner: %w", err)
	}

	return OwnershipTransferResult{
		Organization: org,
		PreviousOwner: OrganizationMember{
			User:       currentUser,
			Membership: updatedCurrent,
		},
		NewOwner: OrganizationMember{
			User:       targetUser,
			Membership: updatedTarget,
		},
	}, nil
}

var slugNonAlnumPattern = regexp.MustCompile(`[^a-z0-9]+`)

func normalizeSlug(value string) string {
	slug := strings.TrimSpace(strings.ToLower(value))
	slug = slugNonAlnumPattern.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func (s Service) loadAccessibleOrganization(ctx context.Context, organizationID, currentUser uuid.UUID) (domainorganization.Organization, domainorganization.Membership, error) {
	org, err := s.repo.GetOrganization(ctx, organizationID)
	if err != nil {
		return domainorganization.Organization{}, domainorganization.Membership{}, fmt.Errorf("load organization: %w", err)
	}

	membership, err := s.repo.GetMembership(ctx, organizationID, currentUser)
	if err != nil {
		if errors.Is(err, ErrMembershipNotFound) {
			return domainorganization.Organization{}, domainorganization.Membership{}, ErrForbiddenOrganization
		}

		return domainorganization.Organization{}, domainorganization.Membership{}, fmt.Errorf("load membership: %w", err)
	}
	if membership.Status != domainorganization.MembershipStatusActive {
		return domainorganization.Organization{}, domainorganization.Membership{}, ErrForbiddenOrganization
	}

	return org, membership, nil
}

func (s Service) enrichMembers(ctx context.Context, memberships []domainorganization.Membership) ([]OrganizationMember, error) {
	members := make([]OrganizationMember, 0, len(memberships))
	for _, membership := range memberships {
		user, err := s.users.GetUserByID(ctx, membership.UserID)
		if err != nil {
			if errors.Is(err, applicationidentity.ErrUserNotFound) {
				return nil, ErrUserNotFound
			}

			return nil, fmt.Errorf("load membership user: %w", err)
		}

		members = append(members, OrganizationMember{
			User:       user,
			Membership: membership,
		})
	}

	return members, nil
}

func (s Service) resolveUser(ctx context.Context, userID *uuid.UUID, email string) (domainidentity.User, error) {
	switch {
	case userID != nil && *userID != uuid.Nil:
		user, err := s.users.GetUserByID(ctx, *userID)
		if err != nil {
			if errors.Is(err, applicationidentity.ErrUserNotFound) {
				return domainidentity.User{}, ErrUserNotFound
			}

			return domainidentity.User{}, fmt.Errorf("load user by id: %w", err)
		}

		return user, nil
	case strings.TrimSpace(email) != "":
		user, err := s.users.GetUserByEmail(ctx, strings.TrimSpace(strings.ToLower(email)))
		if err != nil {
			if errors.Is(err, applicationidentity.ErrUserNotFound) {
				return domainidentity.User{}, ErrUserNotFound
			}

			return domainidentity.User{}, fmt.Errorf("load user by email: %w", err)
		}

		return user, nil
	default:
		return domainidentity.User{}, ErrInvalidMembership
	}
}

func canManageOrganization(role domainorganization.MembershipRole) bool {
	return role == domainorganization.MembershipRoleOwner || role == domainorganization.MembershipRoleAdmin
}

func isValidRole(role domainorganization.MembershipRole) bool {
	switch role {
	case domainorganization.MembershipRoleOwner,
		domainorganization.MembershipRoleAdmin,
		domainorganization.MembershipRoleRequester,
		domainorganization.MembershipRoleApprover,
		domainorganization.MembershipRoleProcurementOfficer,
		domainorganization.MembershipRoleViewer:
		return true
	default:
		return false
	}
}

func isValidStatus(status domainorganization.MembershipStatus) bool {
	switch status {
	case domainorganization.MembershipStatusInvited,
		domainorganization.MembershipStatusActive,
		domainorganization.MembershipStatusSuspended,
		domainorganization.MembershipStatusRemoved:
		return true
	default:
		return false
	}
}
