package repositories

import (
	"context"
	"errors"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	applicationvendor "github.com/AbenezerWork/ProcureFlow/internal/application/vendor"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainvendor "github.com/AbenezerWork/ProcureFlow/internal/domain/vendor"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type VendorRepository struct {
	store *database.Store
	hooks []applicationactivitylog.Hook
}

func NewVendorRepository(store *database.Store, hooks ...applicationactivitylog.Hook) *VendorRepository {
	return &VendorRepository{store: store, hooks: hooks}
}

func (r *VendorRepository) CreateVendor(ctx context.Context, params applicationvendor.CreateVendorParams) (domainvendor.Vendor, error) {
	vendor, err := r.store.CreateVendor(ctx, sqlc.CreateVendorParams{
		OrganizationID:  params.OrganizationID,
		Name:            params.Name,
		LegalName:       params.LegalName,
		ContactName:     params.ContactName,
		Email:           params.Email,
		Phone:           params.Phone,
		TaxIdentifier:   params.TaxIdentifier,
		AddressLine1:    params.AddressLine1,
		AddressLine2:    params.AddressLine2,
		City:            params.City,
		StateRegion:     params.StateRegion,
		PostalCode:      params.PostalCode,
		Country:         params.Country,
		Notes:           params.Notes,
		CreatedByUserID: params.CreatedByUserID,
	})
	if err != nil {
		return domainvendor.Vendor{}, err
	}

	return mapVendor(vendor), nil
}

func (r *VendorRepository) GetVendor(ctx context.Context, organizationID, vendorID uuid.UUID) (domainvendor.Vendor, error) {
	vendor, err := r.store.GetVendorByID(ctx, sqlc.GetVendorByIDParams{
		OrganizationID: organizationID,
		ID:             vendorID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainvendor.Vendor{}, applicationvendor.ErrVendorNotFound
		}

		return domainvendor.Vendor{}, err
	}

	return mapVendor(vendor), nil
}

func (r *VendorRepository) ListVendors(ctx context.Context, organizationID uuid.UUID, status *domainvendor.Status) ([]domainvendor.Vendor, error) {
	params := sqlc.ListVendorsParams{
		OrganizationID: organizationID,
	}
	if status != nil {
		params.Status = sqlc.NullVendorStatus{
			VendorStatus: sqlc.VendorStatus(*status),
			Valid:        true,
		}
	}

	rows, err := r.store.ListVendors(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapVendors(rows), nil
}

func (r *VendorRepository) SearchVendors(ctx context.Context, organizationID uuid.UUID, search string) ([]domainvendor.Vendor, error) {
	rows, err := r.store.SearchVendors(ctx, sqlc.SearchVendorsParams{
		OrganizationID: organizationID,
		Search:         search,
	})
	if err != nil {
		return nil, err
	}

	return mapVendors(rows), nil
}

func (r *VendorRepository) UpdateVendor(ctx context.Context, params applicationvendor.UpdateVendorParams) (domainvendor.Vendor, error) {
	updatedBy := params.UpdatedByUserID
	vendor, err := r.store.UpdateVendor(ctx, sqlc.UpdateVendorParams{
		OrganizationID:  params.OrganizationID,
		ID:              params.VendorID,
		Name:            params.Name,
		LegalName:       params.LegalName,
		ContactName:     params.ContactName,
		Email:           params.Email,
		Phone:           params.Phone,
		TaxIdentifier:   params.TaxIdentifier,
		AddressLine1:    params.AddressLine1,
		AddressLine2:    params.AddressLine2,
		City:            params.City,
		StateRegion:     params.StateRegion,
		PostalCode:      params.PostalCode,
		Country:         params.Country,
		Notes:           params.Notes,
		UpdatedByUserID: &updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainvendor.Vendor{}, applicationvendor.ErrVendorNotFound
		}

		return domainvendor.Vendor{}, err
	}

	return mapVendor(vendor), nil
}

func (r *VendorRepository) ArchiveVendor(ctx context.Context, organizationID, vendorID, updatedByUserID uuid.UUID) (domainvendor.Vendor, error) {
	vendor, err := r.store.ArchiveVendor(ctx, sqlc.ArchiveVendorParams{
		OrganizationID:  organizationID,
		ID:              vendorID,
		UpdatedByUserID: &updatedByUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainvendor.Vendor{}, applicationvendor.ErrVendorNotFound
		}

		return domainvendor.Vendor{}, err
	}

	return mapVendor(vendor), nil
}

func (r *VendorRepository) GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error) {
	membership, err := r.store.GetOrganizationMembership(ctx, sqlc.GetOrganizationMembershipParams{
		OrganizationID: organizationID,
		UserID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainorganization.Membership{}, applicationvendor.ErrMembershipNotFound
		}

		return domainorganization.Membership{}, err
	}

	return mapMembership(membership), nil
}

func (r *VendorRepository) CreateActivityLog(ctx context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error) {
	return createActivityLog(ctx, r.store, params, r.hooks...)
}

func (r *VendorRepository) WithinTransaction(ctx context.Context, fn func(repo applicationvendor.Repository) error) error {
	return r.store.InTx(ctx, func(txStore *database.Store) error {
		return fn(NewVendorRepository(txStore, r.hooks...))
	})
}

func mapVendor(vendor sqlc.Vendor) domainvendor.Vendor {
	return domainvendor.Vendor{
		ID:              vendor.ID,
		OrganizationID:  vendor.OrganizationID,
		Name:            vendor.Name,
		LegalName:       vendor.LegalName,
		ContactName:     vendor.ContactName,
		Email:           vendor.Email,
		Phone:           vendor.Phone,
		TaxIdentifier:   vendor.TaxIdentifier,
		AddressLine1:    vendor.AddressLine1,
		AddressLine2:    vendor.AddressLine2,
		City:            vendor.City,
		StateRegion:     vendor.StateRegion,
		PostalCode:      vendor.PostalCode,
		Country:         vendor.Country,
		Notes:           vendor.Notes,
		Status:          domainvendor.Status(vendor.Status),
		ArchivedAt:      optionalTime(vendor.ArchivedAt),
		CreatedByUserID: vendor.CreatedByUserID,
		UpdatedByUserID: vendor.UpdatedByUserID,
		CreatedAt:       requiredTime(vendor.CreatedAt),
		UpdatedAt:       requiredTime(vendor.UpdatedAt),
	}
}

func mapVendors(rows []sqlc.Vendor) []domainvendor.Vendor {
	vendors := make([]domainvendor.Vendor, 0, len(rows))
	for _, row := range rows {
		vendors = append(vendors, mapVendor(row))
	}

	return vendors
}
