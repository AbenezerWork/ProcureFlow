package organization

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	"github.com/google/uuid"
)

var (
	ErrOrganizationSlugTaken = errors.New("organization slug already exists")
	ErrInvalidOrganization   = errors.New("invalid organization")
)

type CreateOrganizationParams struct {
	Name            string
	Slug            string
	CreatedByUserID uuid.UUID
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
	CreateMembership(ctx context.Context, params CreateMembershipParams) (domainorganization.Membership, error)
	ListUserOrganizations(ctx context.Context, userID uuid.UUID) ([]domainorganization.UserOrganization, error)
}

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(repo Repository) error) error
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

type Service struct {
	repo Repository
	tx   TransactionManager
}

func NewService(repo Repository, tx TransactionManager) Service {
	return Service{repo: repo, tx: tx}
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

var slugNonAlnumPattern = regexp.MustCompile(`[^a-z0-9]+`)

func normalizeSlug(value string) string {
	slug := strings.TrimSpace(strings.ToLower(value))
	slug = slugNonAlnumPattern.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
