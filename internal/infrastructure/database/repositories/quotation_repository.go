package repositories

import (
	"context"
	"errors"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	applicationquotation "github.com/AbenezerWork/ProcureFlow/internal/application/quotation"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainquotation "github.com/AbenezerWork/ProcureFlow/internal/domain/quotation"
	domainrfq "github.com/AbenezerWork/ProcureFlow/internal/domain/rfq"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type QuotationRepository struct {
	store *database.Store
	hooks []applicationactivitylog.Hook
}

func NewQuotationRepository(store *database.Store, hooks ...applicationactivitylog.Hook) *QuotationRepository {
	return &QuotationRepository{store: store, hooks: hooks}
}

func (r *QuotationRepository) CreateQuotation(ctx context.Context, params applicationquotation.CreateQuotationParams) (domainquotation.Quotation, error) {
	quotation, err := r.store.CreateQuotation(ctx, sqlc.CreateQuotationParams{
		OrganizationID:  params.OrganizationID,
		RfqID:           params.RFQID,
		RfqVendorID:     params.RFQVendorID,
		CurrencyCode:    params.CurrencyCode,
		LeadTimeDays:    params.LeadTimeDays,
		PaymentTerms:    params.PaymentTerms,
		Notes:           params.Notes,
		CreatedByUserID: params.CreatedByUserID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domainquotation.Quotation{}, applicationquotation.ErrQuotationExists
		}
		return domainquotation.Quotation{}, err
	}

	return mapQuotation(quotation, nil), nil
}

func (r *QuotationRepository) CreateQuotationItem(ctx context.Context, params applicationquotation.CreateItemParams) (domainquotation.Item, error) {
	quantity, err := parseNumeric(&params.Quantity)
	if err != nil {
		return domainquotation.Item{}, err
	}
	unitPrice, err := parseNumeric(&params.UnitPrice)
	if err != nil {
		return domainquotation.Item{}, err
	}

	item, err := r.store.CreateQuotationItem(ctx, sqlc.CreateQuotationItemParams{
		OrganizationID: params.OrganizationID,
		QuotationID:    params.QuotationID,
		RfqID:          params.RFQID,
		RfqItemID:      params.RFQItemID,
		LineNumber:     params.LineNumber,
		ItemName:       params.ItemName,
		Description:    params.Description,
		Quantity:       quantity,
		Unit:           params.Unit,
		UnitPrice:      unitPrice,
		DeliveryDays:   params.DeliveryDays,
		Notes:          params.Notes,
	})
	if err != nil {
		return domainquotation.Item{}, err
	}

	return mapQuotationItem(item), nil
}

func (r *QuotationRepository) GetQuotation(ctx context.Context, organizationID, quotationID uuid.UUID) (domainquotation.Quotation, error) {
	quotation, err := r.store.GetQuotationByID(ctx, sqlc.GetQuotationByIDParams{
		OrganizationID: organizationID,
		ID:             quotationID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainquotation.Quotation{}, applicationquotation.ErrQuotationNotFound
		}
		return domainquotation.Quotation{}, err
	}

	return mapQuotation(quotation, nil), nil
}

func (r *QuotationRepository) ListQuotationsByRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainquotation.Quotation, error) {
	rows, err := r.store.ListQuotationsByRFQ(ctx, sqlc.ListQuotationsByRFQParams{
		OrganizationID: organizationID,
		RfqID:          rfqID,
	})
	if err != nil {
		return nil, err
	}

	quotations := make([]domainquotation.Quotation, 0, len(rows))
	for _, row := range rows {
		vendorName := row.VendorName
		quotations = append(quotations, mapQuotation(sqlc.Quotation{
			ID:                row.ID,
			OrganizationID:    row.OrganizationID,
			RfqID:             row.RfqID,
			RfqVendorID:       row.RfqVendorID,
			Status:            row.Status,
			CurrencyCode:      row.CurrencyCode,
			LeadTimeDays:      row.LeadTimeDays,
			PaymentTerms:      row.PaymentTerms,
			Notes:             row.Notes,
			SubmittedAt:       row.SubmittedAt,
			SubmittedByUserID: row.SubmittedByUserID,
			RejectedAt:        row.RejectedAt,
			RejectedByUserID:  row.RejectedByUserID,
			RejectionReason:   row.RejectionReason,
			CreatedByUserID:   row.CreatedByUserID,
			UpdatedByUserID:   row.UpdatedByUserID,
			CreatedAt:         row.CreatedAt,
			UpdatedAt:         row.UpdatedAt,
		}, &vendorName))
	}

	return quotations, nil
}

func (r *QuotationRepository) CompareRFQQuotations(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainquotation.ComparisonRow, error) {
	rows, err := r.store.CompareRFQQuotations(ctx, sqlc.CompareRFQQuotationsParams{
		OrganizationID: organizationID,
		RfqID:          rfqID,
	})
	if err != nil {
		return nil, err
	}

	comparisonRows := make([]domainquotation.ComparisonRow, 0, len(rows))
	for _, row := range rows {
		comparisonRows = append(comparisonRows, domainquotation.ComparisonRow{
			QuotationID:     row.QuotationID,
			RFQID:           row.RfqID,
			RFQVendorID:     row.RfqVendorID,
			Status:          domainquotation.Status(row.Status),
			CurrencyCode:    row.CurrencyCode,
			LeadTimeDays:    row.LeadTimeDays,
			PaymentTerms:    row.PaymentTerms,
			QuotationNotes:  row.QuotationNotes,
			VendorID:        row.VendorID,
			VendorName:      row.VendorName,
			TotalAmount:     requiredNumeric(row.TotalAmount),
			QuotationItemID: row.QuotationItemID,
			RFQItemID:       row.RfqItemID,
			LineNumber:      row.LineNumber,
			ItemName:        row.ItemName,
			Description:     row.Description,
			Quantity:        requiredNumeric(row.Quantity),
			Unit:            row.Unit,
			UnitPrice:       requiredNumeric(row.UnitPrice),
			LineTotal:       requiredNumeric(row.LineTotal),
			DeliveryDays:    row.DeliveryDays,
			ItemNotes:       row.ItemNotes,
		})
	}

	return comparisonRows, nil
}

func (r *QuotationRepository) UpdateDraftQuotation(ctx context.Context, params applicationquotation.UpdateQuotationParams) (domainquotation.Quotation, error) {
	updatedByUserID := params.UpdatedByUserID
	quotation, err := r.store.UpdateDraftQuotation(ctx, sqlc.UpdateDraftQuotationParams{
		OrganizationID:  params.OrganizationID,
		ID:              params.QuotationID,
		CurrencyCode:    params.CurrencyCode,
		LeadTimeDays:    params.LeadTimeDays,
		PaymentTerms:    params.PaymentTerms,
		Notes:           params.Notes,
		UpdatedByUserID: &updatedByUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainquotation.Quotation{}, applicationquotation.ErrQuotationNotFound
		}
		return domainquotation.Quotation{}, err
	}

	return mapQuotation(quotation, nil), nil
}

func (r *QuotationRepository) SubmitQuotation(ctx context.Context, organizationID, quotationID, submittedByUserID uuid.UUID) (domainquotation.Quotation, error) {
	quotation, err := r.store.SubmitQuotation(ctx, sqlc.SubmitQuotationParams{
		OrganizationID:    organizationID,
		ID:                quotationID,
		SubmittedByUserID: &submittedByUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainquotation.Quotation{}, applicationquotation.ErrQuotationNotFound
		}
		return domainquotation.Quotation{}, err
	}

	return mapQuotation(quotation, nil), nil
}

func (r *QuotationRepository) RejectQuotation(ctx context.Context, organizationID, quotationID, rejectedByUserID uuid.UUID, rejectionReason *string) (domainquotation.Quotation, error) {
	quotation, err := r.store.RejectQuotation(ctx, sqlc.RejectQuotationParams{
		OrganizationID:   organizationID,
		ID:               quotationID,
		RejectedByUserID: &rejectedByUserID,
		RejectionReason:  rejectionReason,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainquotation.Quotation{}, applicationquotation.ErrQuotationNotFound
		}
		return domainquotation.Quotation{}, err
	}

	return mapQuotation(quotation, nil), nil
}

func (r *QuotationRepository) ListQuotationItems(ctx context.Context, organizationID, quotationID uuid.UUID) ([]domainquotation.Item, error) {
	rows, err := r.store.ListQuotationItems(ctx, sqlc.ListQuotationItemsParams{
		OrganizationID: organizationID,
		QuotationID:    quotationID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]domainquotation.Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapQuotationItem(row))
	}

	return items, nil
}

func (r *QuotationRepository) UpdateQuotationItem(ctx context.Context, params applicationquotation.UpdateItemParams) (domainquotation.Item, error) {
	quantity, err := parseNumeric(&params.Quantity)
	if err != nil {
		return domainquotation.Item{}, err
	}
	unitPrice, err := parseNumeric(&params.UnitPrice)
	if err != nil {
		return domainquotation.Item{}, err
	}

	item, err := r.store.UpdateQuotationItem(ctx, sqlc.UpdateQuotationItemParams{
		OrganizationID: params.OrganizationID,
		QuotationID:    params.QuotationID,
		ID:             params.ItemID,
		ItemName:       params.ItemName,
		Description:    params.Description,
		Quantity:       quantity,
		Unit:           params.Unit,
		UnitPrice:      unitPrice,
		DeliveryDays:   params.DeliveryDays,
		Notes:          params.Notes,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainquotation.Item{}, applicationquotation.ErrQuotationItemNotFound
		}
		return domainquotation.Item{}, err
	}

	return mapQuotationItem(item), nil
}

func (r *QuotationRepository) GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error) {
	membership, err := r.store.GetOrganizationMembership(ctx, sqlc.GetOrganizationMembershipParams{
		OrganizationID: organizationID,
		UserID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainorganization.Membership{}, applicationquotation.ErrMembershipNotFound
		}
		return domainorganization.Membership{}, err
	}

	return mapMembership(membership), nil
}

func (r *QuotationRepository) GetRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error) {
	rfq, err := r.store.GetRFQByID(ctx, sqlc.GetRFQByIDParams{
		OrganizationID: organizationID,
		ID:             rfqID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrfq.RFQ{}, applicationquotation.ErrRFQNotFound
		}
		return domainrfq.RFQ{}, err
	}

	return mapRFQ(rfq), nil
}

func (r *QuotationRepository) ListRFQItems(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainrfq.Item, error) {
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

func (r *QuotationRepository) ListRFQVendors(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainrfq.VendorLink, error) {
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
		})
	}

	return vendors, nil
}

func (r *QuotationRepository) CreateActivityLog(ctx context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error) {
	return createActivityLog(ctx, r.store, params, r.hooks...)
}

func (r *QuotationRepository) WithinTransaction(ctx context.Context, fn func(repo applicationquotation.Repository) error) error {
	return r.store.InTx(ctx, func(txStore *database.Store) error {
		return fn(NewQuotationRepository(txStore, r.hooks...))
	})
}

func mapQuotation(quotation sqlc.Quotation, vendorName *string) domainquotation.Quotation {
	return domainquotation.Quotation{
		ID:                quotation.ID,
		OrganizationID:    quotation.OrganizationID,
		RFQID:             quotation.RfqID,
		RFQVendorID:       quotation.RfqVendorID,
		Status:            domainquotation.Status(quotation.Status),
		CurrencyCode:      quotation.CurrencyCode,
		LeadTimeDays:      quotation.LeadTimeDays,
		PaymentTerms:      quotation.PaymentTerms,
		Notes:             quotation.Notes,
		SubmittedAt:       optionalTime(quotation.SubmittedAt),
		SubmittedByUserID: quotation.SubmittedByUserID,
		RejectedAt:        optionalTime(quotation.RejectedAt),
		RejectedByUserID:  quotation.RejectedByUserID,
		RejectionReason:   quotation.RejectionReason,
		CreatedByUserID:   quotation.CreatedByUserID,
		UpdatedByUserID:   quotation.UpdatedByUserID,
		CreatedAt:         requiredTime(quotation.CreatedAt),
		UpdatedAt:         requiredTime(quotation.UpdatedAt),
		VendorName:        vendorName,
	}
}

func mapQuotationItem(item sqlc.QuotationItem) domainquotation.Item {
	return domainquotation.Item{
		ID:             item.ID,
		OrganizationID: item.OrganizationID,
		QuotationID:    item.QuotationID,
		RFQID:          item.RfqID,
		RFQItemID:      item.RfqItemID,
		LineNumber:     item.LineNumber,
		ItemName:       item.ItemName,
		Description:    item.Description,
		Quantity:       requiredNumeric(item.Quantity),
		Unit:           item.Unit,
		UnitPrice:      requiredNumeric(item.UnitPrice),
		DeliveryDays:   item.DeliveryDays,
		Notes:          item.Notes,
		CreatedAt:      requiredTime(item.CreatedAt),
		UpdatedAt:      requiredTime(item.UpdatedAt),
	}
}
