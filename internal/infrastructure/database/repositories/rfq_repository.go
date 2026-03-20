package repositories

import (
	"context"
	"errors"

	applicationrfq "github.com/AbenezerWork/ProcureFlow/internal/application/rfq"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainprocurement "github.com/AbenezerWork/ProcureFlow/internal/domain/procurement"
	domainrfq "github.com/AbenezerWork/ProcureFlow/internal/domain/rfq"
	domainvendor "github.com/AbenezerWork/ProcureFlow/internal/domain/vendor"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type RFQRepository struct {
	store *database.Store
}

func NewRFQRepository(store *database.Store) *RFQRepository {
	return &RFQRepository{store: store}
}

func (r *RFQRepository) CreateRFQ(ctx context.Context, params applicationrfq.CreateRFQParams) (domainrfq.RFQ, error) {
	rfq, err := r.store.CreateRFQ(ctx, sqlc.CreateRFQParams{
		OrganizationID:       params.OrganizationID,
		ProcurementRequestID: params.ProcurementRequestID,
		ReferenceNumber:      params.ReferenceNumber,
		Title:                params.Title,
		Description:          params.Description,
		CreatedByUserID:      params.CreatedByUserID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domainrfq.RFQ{}, applicationrfq.ErrRFQReferenceTaken
		}
		return domainrfq.RFQ{}, err
	}

	return mapRFQ(rfq), nil
}

func (r *RFQRepository) CreateRFQItem(ctx context.Context, params applicationrfq.CreateRFQItemParams) (domainrfq.Item, error) {
	quantity, err := parseNumeric(&params.Quantity)
	if err != nil {
		return domainrfq.Item{}, err
	}
	targetDate, err := parseDate(params.TargetDate)
	if err != nil {
		return domainrfq.Item{}, err
	}

	item, err := r.store.CreateRFQItem(ctx, sqlc.CreateRFQItemParams{
		OrganizationID:      params.OrganizationID,
		RfqID:               params.RFQID,
		SourceRequestItemID: params.SourceRequestItemID,
		LineNumber:          params.LineNumber,
		ItemName:            params.ItemName,
		Description:         params.Description,
		Quantity:            quantity,
		Unit:                params.Unit,
		TargetDate:          targetDate,
	})
	if err != nil {
		return domainrfq.Item{}, err
	}

	return mapRFQItem(item), nil
}

func (r *RFQRepository) GetRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	rfq, err := r.store.GetRFQByID(ctx, sqlc.GetRFQByIDParams{
		OrganizationID: organizationID,
		ID:             rfqID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrfq.RFQ{}, applicationrfq.ErrRFQNotFound
		}
		return domainrfq.RFQ{}, err
	}

	return mapRFQ(rfq), nil
}

func (r *RFQRepository) ListRFQs(ctx context.Context, organizationID uuid.UUID, status *domainrfq.Status) ([]domainrfq.RFQ, error) {
	params := sqlc.ListRFQsParams{OrganizationID: organizationID}
	if status != nil {
		params.Status = sqlc.NullRfqStatus{
			RfqStatus: sqlc.RfqStatus(*status),
			Valid:     true,
		}
	}

	rows, err := r.store.ListRFQs(ctx, params)
	if err != nil {
		return nil, err
	}

	rfqs := make([]domainrfq.RFQ, 0, len(rows))
	for _, row := range rows {
		rfqs = append(rfqs, mapRFQ(row))
	}

	return rfqs, nil
}

func (r *RFQRepository) UpdateDraftRFQ(ctx context.Context, params applicationrfq.UpdateRFQParams) (domainrfq.RFQ, error) {
	rfq, err := r.store.UpdateDraftRFQ(ctx, sqlc.UpdateDraftRFQParams{
		OrganizationID:  params.OrganizationID,
		ID:              params.RFQID,
		ReferenceNumber: params.ReferenceNumber,
		Title:           params.Title,
		Description:     params.Description,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrfq.RFQ{}, applicationrfq.ErrRFQNotFound
		}
		if isUniqueViolation(err) {
			return domainrfq.RFQ{}, applicationrfq.ErrRFQReferenceTaken
		}
		return domainrfq.RFQ{}, err
	}

	return mapRFQ(rfq), nil
}

func (r *RFQRepository) PublishRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	rfq, err := r.store.PublishRFQ(ctx, sqlc.PublishRFQParams{
		OrganizationID: organizationID,
		ID:             rfqID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrfq.RFQ{}, applicationrfq.ErrRFQNotFound
		}
		return domainrfq.RFQ{}, err
	}

	return mapRFQ(rfq), nil
}

func (r *RFQRepository) CloseRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	rfq, err := r.store.CloseRFQ(ctx, sqlc.CloseRFQParams{
		OrganizationID: organizationID,
		ID:             rfqID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrfq.RFQ{}, applicationrfq.ErrRFQNotFound
		}
		return domainrfq.RFQ{}, err
	}

	return mapRFQ(rfq), nil
}

func (r *RFQRepository) EvaluateRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	rfq, err := r.store.EvaluateRFQ(ctx, sqlc.EvaluateRFQParams{
		OrganizationID: organizationID,
		ID:             rfqID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrfq.RFQ{}, applicationrfq.ErrRFQNotFound
		}
		return domainrfq.RFQ{}, err
	}

	return mapRFQ(rfq), nil
}

func (r *RFQRepository) CancelRFQ(ctx context.Context, organizationID, rfqID, cancelledByUserID uuid.UUID) (domainrfq.RFQ, error) {
	rfq, err := r.store.CancelRFQ(ctx, sqlc.CancelRFQParams{
		OrganizationID:    organizationID,
		ID:                rfqID,
		CancelledByUserID: &cancelledByUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrfq.RFQ{}, applicationrfq.ErrRFQNotFound
		}
		return domainrfq.RFQ{}, err
	}

	return mapRFQ(rfq), nil
}

func (r *RFQRepository) ListRFQItems(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainrfq.Item, error) {
	rows, err := r.store.ListRFQItems(ctx, sqlc.ListRFQItemsParams{
		OrganizationID: organizationID,
		RfqID:          rfqID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]domainrfq.Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapRFQItem(row))
	}

	return items, nil
}

func (r *RFQRepository) AttachVendor(ctx context.Context, organizationID, rfqID, vendorID, attachedByUserID uuid.UUID) (domainrfq.VendorLink, error) {
	attached, err := r.store.AttachVendorToRFQ(ctx, sqlc.AttachVendorToRFQParams{
		OrganizationID:   organizationID,
		RfqID:            rfqID,
		VendorID:         vendorID,
		AttachedByUserID: attachedByUserID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domainrfq.VendorLink{}, applicationrfq.ErrRFQVendorAlreadyAttached
		}
		return domainrfq.VendorLink{}, err
	}

	vendor, err := r.GetVendor(ctx, organizationID, vendorID)
	if err != nil {
		return domainrfq.VendorLink{}, err
	}

	return domainrfq.VendorLink{
		ID:               attached.ID,
		OrganizationID:   attached.OrganizationID,
		RFQID:            attached.RfqID,
		VendorID:         attached.VendorID,
		AttachedByUserID: attached.AttachedByUserID,
		AttachedAt:       requiredTime(attached.AttachedAt),
		CreatedAt:        requiredTime(attached.CreatedAt),
		VendorName:       vendor.Name,
		VendorStatus:     vendor.Status,
	}, nil
}

func (r *RFQRepository) RemoveVendor(ctx context.Context, organizationID, rfqID, vendorID uuid.UUID) error {
	rows, err := r.store.RemoveVendorFromRFQ(ctx, sqlc.RemoveVendorFromRFQParams{
		OrganizationID: organizationID,
		RfqID:          rfqID,
		VendorID:       vendorID,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return applicationrfq.ErrRFQVendorNotFound
	}

	return nil
}

func (r *RFQRepository) ListRFQVendors(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainrfq.VendorLink, error) {
	rows, err := r.store.ListRFQVendors(ctx, sqlc.ListRFQVendorsParams{
		OrganizationID: organizationID,
		RfqID:          rfqID,
	})
	if err != nil {
		return nil, err
	}

	vendors := make([]domainrfq.VendorLink, 0, len(rows))
	for _, row := range rows {
		vendors = append(vendors, domainrfq.VendorLink{
			ID:               row.ID,
			OrganizationID:   row.OrganizationID,
			RFQID:            row.RfqID,
			VendorID:         row.VendorID,
			AttachedByUserID: row.AttachedByUserID,
			AttachedAt:       requiredTime(row.AttachedAt),
			CreatedAt:        requiredTime(row.CreatedAt),
			VendorName:       row.VendorName,
			VendorStatus:     domainvendor.Status(row.VendorStatus),
		})
	}

	return vendors, nil
}

func (r *RFQRepository) GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error) {
	membership, err := r.store.GetOrganizationMembership(ctx, sqlc.GetOrganizationMembershipParams{
		OrganizationID: organizationID,
		UserID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainorganization.Membership{}, applicationrfq.ErrMembershipNotFound
		}
		return domainorganization.Membership{}, err
	}

	return mapMembership(membership), nil
}

func (r *RFQRepository) GetProcurementRequest(ctx context.Context, organizationID, requestID uuid.UUID) (domainprocurement.Request, error) {
	request, err := r.store.GetProcurementRequestByID(ctx, sqlc.GetProcurementRequestByIDParams{
		OrganizationID: organizationID,
		ID:             requestID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainprocurement.Request{}, applicationrfq.ErrProcurementRequestNotFound
		}
		return domainprocurement.Request{}, err
	}

	return mapProcurementRequest(request), nil
}

func (r *RFQRepository) ListProcurementRequestItems(ctx context.Context, organizationID, requestID uuid.UUID) ([]domainprocurement.Item, error) {
	rows, err := r.store.ListProcurementRequestItems(ctx, sqlc.ListProcurementRequestItemsParams{
		OrganizationID:       organizationID,
		ProcurementRequestID: requestID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]domainprocurement.Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapProcurementRequestItem(row))
	}

	return items, nil
}

func (r *RFQRepository) GetVendor(ctx context.Context, organizationID, vendorID uuid.UUID) (domainvendor.Vendor, error) {
	vendor, err := r.store.GetVendorByID(ctx, sqlc.GetVendorByIDParams{
		OrganizationID: organizationID,
		ID:             vendorID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainvendor.Vendor{}, applicationrfq.ErrVendorNotFound
		}
		return domainvendor.Vendor{}, err
	}

	return mapVendor(vendor), nil
}

func (r *RFQRepository) WithinTransaction(ctx context.Context, fn func(repo applicationrfq.Repository) error) error {
	return r.store.InTx(ctx, func(txStore *database.Store) error {
		return fn(NewRFQRepository(txStore))
	})
}

func mapRFQ(rfq sqlc.Rfq) domainrfq.RFQ {
	return domainrfq.RFQ{
		ID:                   rfq.ID,
		OrganizationID:       rfq.OrganizationID,
		ProcurementRequestID: rfq.ProcurementRequestID,
		ReferenceNumber:      rfq.ReferenceNumber,
		Title:                rfq.Title,
		Description:          rfq.Description,
		Status:               domainrfq.Status(rfq.Status),
		CreatedByUserID:      rfq.CreatedByUserID,
		PublishedAt:          optionalTime(rfq.PublishedAt),
		ClosedAt:             optionalTime(rfq.ClosedAt),
		EvaluatedAt:          optionalTime(rfq.EvaluatedAt),
		CancelledAt:          optionalTime(rfq.CancelledAt),
		CancelledByUserID:    rfq.CancelledByUserID,
		CreatedAt:            requiredTime(rfq.CreatedAt),
		UpdatedAt:            requiredTime(rfq.UpdatedAt),
	}
}

func mapRFQItem(item sqlc.RfqItem) domainrfq.Item {
	return domainrfq.Item{
		ID:                  item.ID,
		OrganizationID:      item.OrganizationID,
		RFQID:               item.RfqID,
		SourceRequestItemID: item.SourceRequestItemID,
		LineNumber:          item.LineNumber,
		ItemName:            item.ItemName,
		Description:         item.Description,
		Quantity:            requiredNumeric(item.Quantity),
		Unit:                item.Unit,
		TargetDate:          optionalDate(item.TargetDate),
		CreatedAt:           requiredTime(item.CreatedAt),
		UpdatedAt:           requiredTime(item.UpdatedAt),
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
