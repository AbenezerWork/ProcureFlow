package vendor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainvendor "github.com/AbenezerWork/ProcureFlow/internal/domain/vendor"
	"github.com/google/uuid"
)

var (
	ErrInvalidVendor      = errors.New("invalid vendor")
	ErrVendorNotFound     = errors.New("vendor not found")
	ErrMembershipNotFound = errors.New("organization membership not found")
	ErrForbiddenVendor    = errors.New("forbidden vendor operation")
)

type CreateVendorParams struct {
	OrganizationID  uuid.UUID
	Name            string
	LegalName       *string
	ContactName     *string
	Email           *string
	Phone           *string
	TaxIdentifier   *string
	AddressLine1    *string
	AddressLine2    *string
	City            *string
	StateRegion     *string
	PostalCode      *string
	Country         *string
	Notes           *string
	CreatedByUserID uuid.UUID
}

type UpdateVendorParams struct {
	OrganizationID  uuid.UUID
	VendorID        uuid.UUID
	Name            string
	LegalName       *string
	ContactName     *string
	Email           *string
	Phone           *string
	TaxIdentifier   *string
	AddressLine1    *string
	AddressLine2    *string
	City            *string
	StateRegion     *string
	PostalCode      *string
	Country         *string
	Notes           *string
	UpdatedByUserID uuid.UUID
}

type Repository interface {
	CreateVendor(ctx context.Context, params CreateVendorParams) (domainvendor.Vendor, error)
	GetVendor(ctx context.Context, organizationID, vendorID uuid.UUID) (domainvendor.Vendor, error)
	ListVendors(ctx context.Context, organizationID uuid.UUID, status *domainvendor.Status) ([]domainvendor.Vendor, error)
	SearchVendors(ctx context.Context, organizationID uuid.UUID, search string) ([]domainvendor.Vendor, error)
	UpdateVendor(ctx context.Context, params UpdateVendorParams) (domainvendor.Vendor, error)
	ArchiveVendor(ctx context.Context, organizationID, vendorID, updatedByUserID uuid.UUID) (domainvendor.Vendor, error)
	GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error)
}

type CreateInput struct {
	OrganizationID uuid.UUID
	CurrentUser    uuid.UUID
	Name           string
	LegalName      *string
	ContactName    *string
	Email          *string
	Phone          *string
	TaxIdentifier  *string
	AddressLine1   *string
	AddressLine2   *string
	City           *string
	StateRegion    *string
	PostalCode     *string
	Country        *string
	Notes          *string
}

type ListInput struct {
	OrganizationID uuid.UUID
	CurrentUser    uuid.UUID
	Status         *domainvendor.Status
	Search         string
}

type UpdateInput struct {
	OrganizationID uuid.UUID
	VendorID       uuid.UUID
	CurrentUser    uuid.UUID
	Name           *string
	LegalName      *string
	ContactName    *string
	Email          *string
	Phone          *string
	TaxIdentifier  *string
	AddressLine1   *string
	AddressLine2   *string
	City           *string
	StateRegion    *string
	PostalCode     *string
	Country        *string
	Notes          *string
}

type ArchiveInput struct {
	OrganizationID uuid.UUID
	VendorID       uuid.UUID
	CurrentUser    uuid.UUID
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) Create(ctx context.Context, input CreateInput) (domainvendor.Vendor, error) {
	if input.OrganizationID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainvendor.Vendor{}, ErrInvalidVendor
	}

	membership, err := s.loadActiveMembership(ctx, input.OrganizationID, input.CurrentUser)
	if err != nil {
		return domainvendor.Vendor{}, err
	}
	if !canManageVendors(membership.Role) {
		return domainvendor.Vendor{}, ErrForbiddenVendor
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domainvendor.Vendor{}, ErrInvalidVendor
	}

	created, err := s.repo.CreateVendor(ctx, CreateVendorParams{
		OrganizationID:  input.OrganizationID,
		Name:            name,
		LegalName:       normalizeOptional(input.LegalName),
		ContactName:     normalizeOptional(input.ContactName),
		Email:           normalizeOptional(input.Email),
		Phone:           normalizeOptional(input.Phone),
		TaxIdentifier:   normalizeOptional(input.TaxIdentifier),
		AddressLine1:    normalizeOptional(input.AddressLine1),
		AddressLine2:    normalizeOptional(input.AddressLine2),
		City:            normalizeOptional(input.City),
		StateRegion:     normalizeOptional(input.StateRegion),
		PostalCode:      normalizeOptional(input.PostalCode),
		Country:         normalizeOptional(input.Country),
		Notes:           normalizeOptional(input.Notes),
		CreatedByUserID: input.CurrentUser,
	})
	if err != nil {
		return domainvendor.Vendor{}, fmt.Errorf("create vendor: %w", err)
	}

	return created, nil
}

func (s Service) List(ctx context.Context, input ListInput) ([]domainvendor.Vendor, error) {
	if input.OrganizationID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return nil, ErrInvalidVendor
	}
	if input.Status != nil && !isValidStatus(*input.Status) {
		return nil, ErrInvalidVendor
	}

	if _, err := s.loadActiveMembership(ctx, input.OrganizationID, input.CurrentUser); err != nil {
		return nil, err
	}

	search := strings.TrimSpace(input.Search)
	if search != "" {
		vendors, err := s.repo.SearchVendors(ctx, input.OrganizationID, search)
		if err != nil {
			return nil, fmt.Errorf("search vendors: %w", err)
		}

		if input.Status == nil {
			return vendors, nil
		}

		filtered := make([]domainvendor.Vendor, 0, len(vendors))
		for _, vendor := range vendors {
			if vendor.Status == *input.Status {
				filtered = append(filtered, vendor)
			}
		}
		return filtered, nil
	}

	vendors, err := s.repo.ListVendors(ctx, input.OrganizationID, input.Status)
	if err != nil {
		return nil, fmt.Errorf("list vendors: %w", err)
	}

	return vendors, nil
}

func (s Service) Get(ctx context.Context, organizationID, vendorID, currentUser uuid.UUID) (domainvendor.Vendor, error) {
	if organizationID == uuid.Nil || vendorID == uuid.Nil || currentUser == uuid.Nil {
		return domainvendor.Vendor{}, ErrInvalidVendor
	}

	if _, err := s.loadActiveMembership(ctx, organizationID, currentUser); err != nil {
		return domainvendor.Vendor{}, err
	}

	vendor, err := s.repo.GetVendor(ctx, organizationID, vendorID)
	if err != nil {
		if errors.Is(err, ErrVendorNotFound) {
			return domainvendor.Vendor{}, err
		}

		return domainvendor.Vendor{}, fmt.Errorf("get vendor: %w", err)
	}

	return vendor, nil
}

func (s Service) Update(ctx context.Context, input UpdateInput) (domainvendor.Vendor, error) {
	if input.OrganizationID == uuid.Nil || input.VendorID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainvendor.Vendor{}, ErrInvalidVendor
	}
	if !hasVendorUpdate(input) {
		return domainvendor.Vendor{}, ErrInvalidVendor
	}

	membership, err := s.loadActiveMembership(ctx, input.OrganizationID, input.CurrentUser)
	if err != nil {
		return domainvendor.Vendor{}, err
	}
	if !canManageVendors(membership.Role) {
		return domainvendor.Vendor{}, ErrForbiddenVendor
	}

	existing, err := s.repo.GetVendor(ctx, input.OrganizationID, input.VendorID)
	if err != nil {
		if errors.Is(err, ErrVendorNotFound) {
			return domainvendor.Vendor{}, err
		}

		return domainvendor.Vendor{}, fmt.Errorf("load vendor: %w", err)
	}

	name := existing.Name
	if input.Name != nil {
		name = strings.TrimSpace(*input.Name)
	}
	if name == "" {
		return domainvendor.Vendor{}, ErrInvalidVendor
	}

	updated, err := s.repo.UpdateVendor(ctx, UpdateVendorParams{
		OrganizationID:  input.OrganizationID,
		VendorID:        input.VendorID,
		Name:            name,
		LegalName:       mergeOptional(existing.LegalName, input.LegalName),
		ContactName:     mergeOptional(existing.ContactName, input.ContactName),
		Email:           mergeOptional(existing.Email, input.Email),
		Phone:           mergeOptional(existing.Phone, input.Phone),
		TaxIdentifier:   mergeOptional(existing.TaxIdentifier, input.TaxIdentifier),
		AddressLine1:    mergeOptional(existing.AddressLine1, input.AddressLine1),
		AddressLine2:    mergeOptional(existing.AddressLine2, input.AddressLine2),
		City:            mergeOptional(existing.City, input.City),
		StateRegion:     mergeOptional(existing.StateRegion, input.StateRegion),
		PostalCode:      mergeOptional(existing.PostalCode, input.PostalCode),
		Country:         mergeOptional(existing.Country, input.Country),
		Notes:           mergeOptional(existing.Notes, input.Notes),
		UpdatedByUserID: input.CurrentUser,
	})
	if err != nil {
		if errors.Is(err, ErrVendorNotFound) {
			return domainvendor.Vendor{}, err
		}

		return domainvendor.Vendor{}, fmt.Errorf("update vendor: %w", err)
	}

	return updated, nil
}

func (s Service) Archive(ctx context.Context, input ArchiveInput) (domainvendor.Vendor, error) {
	if input.OrganizationID == uuid.Nil || input.VendorID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainvendor.Vendor{}, ErrInvalidVendor
	}

	membership, err := s.loadActiveMembership(ctx, input.OrganizationID, input.CurrentUser)
	if err != nil {
		return domainvendor.Vendor{}, err
	}
	if !canManageVendors(membership.Role) {
		return domainvendor.Vendor{}, ErrForbiddenVendor
	}

	archived, err := s.repo.ArchiveVendor(ctx, input.OrganizationID, input.VendorID, input.CurrentUser)
	if err != nil {
		if errors.Is(err, ErrVendorNotFound) {
			return domainvendor.Vendor{}, err
		}

		return domainvendor.Vendor{}, fmt.Errorf("archive vendor: %w", err)
	}

	return archived, nil
}

func (s Service) loadActiveMembership(ctx context.Context, organizationID, currentUser uuid.UUID) (domainorganization.Membership, error) {
	membership, err := s.repo.GetMembership(ctx, organizationID, currentUser)
	if err != nil {
		if errors.Is(err, ErrMembershipNotFound) {
			return domainorganization.Membership{}, ErrForbiddenVendor
		}

		return domainorganization.Membership{}, fmt.Errorf("load membership: %w", err)
	}
	if membership.Status != domainorganization.MembershipStatusActive {
		return domainorganization.Membership{}, ErrForbiddenVendor
	}

	return membership, nil
}

func canManageVendors(role domainorganization.MembershipRole) bool {
	switch role {
	case domainorganization.MembershipRoleOwner,
		domainorganization.MembershipRoleAdmin,
		domainorganization.MembershipRoleProcurementOfficer:
		return true
	default:
		return false
	}
}

func hasVendorUpdate(input UpdateInput) bool {
	return input.Name != nil ||
		input.LegalName != nil ||
		input.ContactName != nil ||
		input.Email != nil ||
		input.Phone != nil ||
		input.TaxIdentifier != nil ||
		input.AddressLine1 != nil ||
		input.AddressLine2 != nil ||
		input.City != nil ||
		input.StateRegion != nil ||
		input.PostalCode != nil ||
		input.Country != nil ||
		input.Notes != nil
}

func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func mergeOptional(current, update *string) *string {
	if update == nil {
		return current
	}

	return normalizeOptional(update)
}

func isValidStatus(status domainvendor.Status) bool {
	switch status {
	case domainvendor.StatusActive, domainvendor.StatusArchived:
		return true
	default:
		return false
	}
}
