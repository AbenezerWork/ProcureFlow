package quotation

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainquotation "github.com/AbenezerWork/ProcureFlow/internal/domain/quotation"
	domainrfq "github.com/AbenezerWork/ProcureFlow/internal/domain/rfq"
	"github.com/google/uuid"
)

var (
	ErrInvalidQuotation      = errors.New("invalid quotation")
	ErrQuotationNotFound     = errors.New("quotation not found")
	ErrQuotationItemNotFound = errors.New("quotation item not found")
	ErrQuotationExists       = errors.New("quotation already exists")
	ErrMembershipNotFound    = errors.New("organization membership not found")
	ErrRFQNotFound           = errors.New("rfq not found")
	ErrRFQVendorNotFound     = errors.New("rfq vendor not found")
	ErrForbiddenQuotation    = errors.New("forbidden quotation operation")
)

type CreateQuotationParams struct {
	OrganizationID  uuid.UUID
	RFQID           uuid.UUID
	RFQVendorID     uuid.UUID
	CurrencyCode    string
	LeadTimeDays    *int32
	PaymentTerms    *string
	Notes           *string
	CreatedByUserID uuid.UUID
}

type UpdateQuotationParams struct {
	OrganizationID  uuid.UUID
	QuotationID     uuid.UUID
	CurrencyCode    string
	LeadTimeDays    *int32
	PaymentTerms    *string
	Notes           *string
	UpdatedByUserID uuid.UUID
}

type CreateItemParams struct {
	OrganizationID uuid.UUID
	QuotationID    uuid.UUID
	RFQID          uuid.UUID
	RFQItemID      uuid.UUID
	LineNumber     int32
	ItemName       string
	Description    *string
	Quantity       string
	Unit           string
	UnitPrice      string
	DeliveryDays   *int32
	Notes          *string
}

type UpdateItemParams struct {
	OrganizationID uuid.UUID
	QuotationID    uuid.UUID
	ItemID         uuid.UUID
	ItemName       string
	Description    *string
	Quantity       string
	Unit           string
	UnitPrice      string
	DeliveryDays   *int32
	Notes          *string
}

type Repository interface {
	CreateQuotation(ctx context.Context, params CreateQuotationParams) (domainquotation.Quotation, error)
	CreateQuotationItem(ctx context.Context, params CreateItemParams) (domainquotation.Item, error)
	GetQuotation(ctx context.Context, organizationID, quotationID uuid.UUID) (domainquotation.Quotation, error)
	ListQuotationsByRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainquotation.Quotation, error)
	CompareRFQQuotations(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainquotation.ComparisonRow, error)
	UpdateDraftQuotation(ctx context.Context, params UpdateQuotationParams) (domainquotation.Quotation, error)
	SubmitQuotation(ctx context.Context, organizationID, quotationID, submittedByUserID uuid.UUID) (domainquotation.Quotation, error)
	RejectQuotation(ctx context.Context, organizationID, quotationID, rejectedByUserID uuid.UUID, rejectionReason *string) (domainquotation.Quotation, error)
	ListQuotationItems(ctx context.Context, organizationID, quotationID uuid.UUID) ([]domainquotation.Item, error)
	UpdateQuotationItem(ctx context.Context, params UpdateItemParams) (domainquotation.Item, error)
	GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error)
	GetRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error)
	ListRFQItems(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainrfq.Item, error)
	ListRFQVendors(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainrfq.VendorLink, error)
	CreateActivityLog(ctx context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error)
}

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(repo Repository) error) error
}

type CreateInput struct {
	OrganizationID uuid.UUID
	RFQID          uuid.UUID
	RFQVendorID    uuid.UUID
	CurrentUser    uuid.UUID
	CurrencyCode   *string
	LeadTimeDays   *int32
	PaymentTerms   *string
	Notes          *string
}

type ListInput struct {
	OrganizationID uuid.UUID
	RFQID          uuid.UUID
	CurrentUser    uuid.UUID
}

type ComparisonResult struct {
	RFQ        domainrfq.RFQ
	Comparison domainquotation.Comparison
}

type UpdateInput struct {
	OrganizationID uuid.UUID
	RFQID          uuid.UUID
	QuotationID    uuid.UUID
	CurrentUser    uuid.UUID
	CurrencyCode   *string
	LeadTimeDays   *int32
	PaymentTerms   *string
	Notes          *string
}

type TransitionInput struct {
	OrganizationID uuid.UUID
	RFQID          uuid.UUID
	QuotationID    uuid.UUID
	CurrentUser    uuid.UUID
}

type RejectInput struct {
	OrganizationID  uuid.UUID
	RFQID           uuid.UUID
	QuotationID     uuid.UUID
	CurrentUser     uuid.UUID
	RejectionReason *string
}

type UpdateItemInput struct {
	OrganizationID uuid.UUID
	RFQID          uuid.UUID
	QuotationID    uuid.UUID
	ItemID         uuid.UUID
	CurrentUser    uuid.UUID
	UnitPrice      *string
	DeliveryDays   *int32
	Notes          *string
}

type Service struct {
	repo Repository
	tx   TransactionManager
}

func NewService(repo Repository, tx TransactionManager) Service {
	return Service{repo: repo, tx: tx}
}

func (s Service) Create(ctx context.Context, input CreateInput) (domainquotation.Quotation, error) {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.RFQVendorID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainquotation.Quotation{}, ErrInvalidQuotation
	}

	rfq, membership, err := s.loadManagedRFQ(ctx, input.OrganizationID, input.RFQID, input.CurrentUser)
	if err != nil {
		return domainquotation.Quotation{}, err
	}
	if !canManageQuotation(membership.Role) {
		return domainquotation.Quotation{}, ErrForbiddenQuotation
	}
	if rfq.Status != domainrfq.StatusPublished {
		return domainquotation.Quotation{}, ErrInvalidQuotation
	}

	vendors, err := s.repo.ListRFQVendors(ctx, input.OrganizationID, input.RFQID)
	if err != nil {
		return domainquotation.Quotation{}, fmt.Errorf("list rfq vendors: %w", err)
	}
	vendorLink, ok := findVendorLink(vendors, input.RFQVendorID)
	if !ok {
		return domainquotation.Quotation{}, ErrRFQVendorNotFound
	}

	items, err := s.repo.ListRFQItems(ctx, input.OrganizationID, input.RFQID)
	if err != nil {
		return domainquotation.Quotation{}, fmt.Errorf("list rfq items: %w", err)
	}
	if len(items) == 0 {
		return domainquotation.Quotation{}, ErrInvalidQuotation
	}

	currency, leadTimeDays, paymentTerms, notes, err := normalizeQuotationFields(input.CurrencyCode, input.LeadTimeDays, input.PaymentTerms, input.Notes)
	if err != nil {
		return domainquotation.Quotation{}, err
	}

	var created domainquotation.Quotation
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		quotation, err := repo.CreateQuotation(ctx, CreateQuotationParams{
			OrganizationID:  input.OrganizationID,
			RFQID:           input.RFQID,
			RFQVendorID:     input.RFQVendorID,
			CurrencyCode:    currency,
			LeadTimeDays:    leadTimeDays,
			PaymentTerms:    paymentTerms,
			Notes:           notes,
			CreatedByUserID: input.CurrentUser,
		})
		if err != nil {
			return err
		}

		for _, item := range items {
			if _, err := repo.CreateQuotationItem(ctx, CreateItemParams{
				OrganizationID: input.OrganizationID,
				QuotationID:    quotation.ID,
				RFQID:          input.RFQID,
				RFQItemID:      item.ID,
				LineNumber:     item.LineNumber,
				ItemName:       item.ItemName,
				Description:    item.Description,
				Quantity:       item.Quantity,
				Unit:           item.Unit,
				UnitPrice:      "0",
			}); err != nil {
				return err
			}
		}

		summary := "Created quotation"
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeQuotation),
			EntityID:       quotation.ID,
			Action:         domainactivitylog.ActionQuotationCreated,
			Summary:        &summary,
			Metadata: map[string]any{
				"rfq_id":        input.RFQID.String(),
				"rfq_vendor_id": input.RFQVendorID.String(),
				"item_count":    len(items),
			},
		}); err != nil {
			return err
		}

		created = quotation
		return nil
	}); err != nil {
		if errors.Is(err, ErrQuotationExists) {
			return domainquotation.Quotation{}, err
		}
		return domainquotation.Quotation{}, fmt.Errorf("create quotation: %w", err)
	}

	created.VendorName = &vendorLink.VendorName
	return created, nil
}

func (s Service) List(ctx context.Context, input ListInput) ([]domainquotation.Quotation, error) {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return nil, ErrInvalidQuotation
	}
	if _, err := s.loadActiveMembership(ctx, input.OrganizationID, input.CurrentUser); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetRFQ(ctx, input.OrganizationID, input.RFQID); err != nil {
		if errors.Is(err, ErrRFQNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("load rfq: %w", err)
	}

	quotations, err := s.repo.ListQuotationsByRFQ(ctx, input.OrganizationID, input.RFQID)
	if err != nil {
		return nil, fmt.Errorf("list quotations: %w", err)
	}

	return quotations, nil
}

func (s Service) Compare(ctx context.Context, input ListInput) (ComparisonResult, error) {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return ComparisonResult{}, ErrInvalidQuotation
	}
	if _, err := s.loadActiveMembership(ctx, input.OrganizationID, input.CurrentUser); err != nil {
		return ComparisonResult{}, err
	}

	rfq, err := s.repo.GetRFQ(ctx, input.OrganizationID, input.RFQID)
	if err != nil {
		if errors.Is(err, ErrRFQNotFound) {
			return ComparisonResult{}, err
		}
		return ComparisonResult{}, fmt.Errorf("load rfq: %w", err)
	}

	rows, err := s.repo.CompareRFQQuotations(ctx, input.OrganizationID, input.RFQID)
	if err != nil {
		return ComparisonResult{}, fmt.Errorf("compare quotations: %w", err)
	}

	return ComparisonResult{
		RFQ:        rfq,
		Comparison: buildComparison(input.RFQID, rows),
	}, nil
}

func (s Service) Get(ctx context.Context, organizationID, rfqID, quotationID, currentUser uuid.UUID) (domainquotation.Quotation, error) {
	if organizationID == uuid.Nil || rfqID == uuid.Nil || quotationID == uuid.Nil || currentUser == uuid.Nil {
		return domainquotation.Quotation{}, ErrInvalidQuotation
	}
	if _, err := s.loadActiveMembership(ctx, organizationID, currentUser); err != nil {
		return domainquotation.Quotation{}, err
	}

	quotation, err := s.repo.GetQuotation(ctx, organizationID, quotationID)
	if err != nil {
		if errors.Is(err, ErrQuotationNotFound) {
			return domainquotation.Quotation{}, err
		}
		return domainquotation.Quotation{}, fmt.Errorf("get quotation: %w", err)
	}
	if quotation.RFQID != rfqID {
		return domainquotation.Quotation{}, ErrQuotationNotFound
	}

	return s.enrichVendorName(ctx, quotation)
}

func (s Service) Update(ctx context.Context, input UpdateInput) (domainquotation.Quotation, error) {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.QuotationID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainquotation.Quotation{}, ErrInvalidQuotation
	}
	if !hasQuotationUpdate(input) {
		return domainquotation.Quotation{}, ErrInvalidQuotation
	}

	quotation, rfq, membership, err := s.loadManagedQuotation(ctx, input.OrganizationID, input.RFQID, input.QuotationID, input.CurrentUser)
	if err != nil {
		return domainquotation.Quotation{}, err
	}
	if !canManageQuotation(membership.Role) {
		return domainquotation.Quotation{}, ErrForbiddenQuotation
	}
	if rfq.Status != domainrfq.StatusPublished || quotation.Status != domainquotation.StatusDraft {
		return domainquotation.Quotation{}, ErrInvalidQuotation
	}

	currencyInput := &quotation.CurrencyCode
	if input.CurrencyCode != nil {
		currencyInput = input.CurrencyCode
	}
	leadTimeInput := quotation.LeadTimeDays
	if input.LeadTimeDays != nil {
		leadTimeInput = input.LeadTimeDays
	}
	paymentTermsInput := quotation.PaymentTerms
	if input.PaymentTerms != nil {
		paymentTermsInput = input.PaymentTerms
	}
	notesInput := quotation.Notes
	if input.Notes != nil {
		notesInput = input.Notes
	}

	currency, leadTimeDays, paymentTerms, notes, err := normalizeQuotationFields(currencyInput, leadTimeInput, paymentTermsInput, notesInput)
	if err != nil {
		return domainquotation.Quotation{}, err
	}

	var updated domainquotation.Quotation
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		next, err := repo.UpdateDraftQuotation(ctx, UpdateQuotationParams{
			OrganizationID:  input.OrganizationID,
			QuotationID:     input.QuotationID,
			CurrencyCode:    currency,
			LeadTimeDays:    leadTimeDays,
			PaymentTerms:    paymentTerms,
			Notes:           notes,
			UpdatedByUserID: input.CurrentUser,
		})
		if err != nil {
			return err
		}

		summary := "Updated quotation"
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeQuotation),
			EntityID:       input.QuotationID,
			Action:         domainactivitylog.ActionQuotationUpdated,
			Summary:        &summary,
			Metadata: map[string]any{
				"currency_code": next.CurrencyCode,
				"rfq_id":        input.RFQID.String(),
			},
		}); err != nil {
			return err
		}

		updated = next
		return nil
	}); err != nil {
		if errors.Is(err, ErrQuotationNotFound) {
			return domainquotation.Quotation{}, err
		}
		return domainquotation.Quotation{}, fmt.Errorf("update quotation: %w", err)
	}

	return s.enrichVendorName(ctx, updated)
}

func (s Service) Submit(ctx context.Context, input TransitionInput) (domainquotation.Quotation, error) {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.QuotationID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainquotation.Quotation{}, ErrInvalidQuotation
	}

	quotation, rfq, membership, err := s.loadManagedQuotation(ctx, input.OrganizationID, input.RFQID, input.QuotationID, input.CurrentUser)
	if err != nil {
		return domainquotation.Quotation{}, err
	}
	if !canManageQuotation(membership.Role) {
		return domainquotation.Quotation{}, ErrForbiddenQuotation
	}
	if rfq.Status != domainrfq.StatusPublished || quotation.Status != domainquotation.StatusDraft {
		return domainquotation.Quotation{}, ErrInvalidQuotation
	}

	items, err := s.repo.ListQuotationItems(ctx, input.OrganizationID, input.QuotationID)
	if err != nil {
		return domainquotation.Quotation{}, fmt.Errorf("list quotation items: %w", err)
	}
	if len(items) == 0 {
		return domainquotation.Quotation{}, ErrInvalidQuotation
	}
	for _, item := range items {
		if item.UnitPrice == "" {
			return domainquotation.Quotation{}, ErrInvalidQuotation
		}
	}

	var submitted domainquotation.Quotation
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		updated, err := repo.SubmitQuotation(ctx, input.OrganizationID, input.QuotationID, input.CurrentUser)
		if err != nil {
			return err
		}

		summary := "Submitted quotation"
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeQuotation),
			EntityID:       input.QuotationID,
			Action:         domainactivitylog.ActionQuotationSubmitted,
			Summary:        &summary,
			Metadata: map[string]any{
				"rfq_id":           input.RFQID.String(),
				"rfq_vendor_id":    quotation.RFQVendorID.String(),
				"quotation_status": string(updated.Status),
			},
		}); err != nil {
			return err
		}

		submitted = updated
		return nil
	}); err != nil {
		if errors.Is(err, ErrQuotationNotFound) {
			return domainquotation.Quotation{}, err
		}
		return domainquotation.Quotation{}, fmt.Errorf("submit quotation: %w", err)
	}

	return s.enrichVendorName(ctx, submitted)
}

func (s Service) Reject(ctx context.Context, input RejectInput) (domainquotation.Quotation, error) {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.QuotationID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainquotation.Quotation{}, ErrInvalidQuotation
	}

	quotation, rfq, membership, err := s.loadManagedQuotation(ctx, input.OrganizationID, input.RFQID, input.QuotationID, input.CurrentUser)
	if err != nil {
		return domainquotation.Quotation{}, err
	}
	if !canManageQuotation(membership.Role) {
		return domainquotation.Quotation{}, ErrForbiddenQuotation
	}
	if quotation.Status == domainquotation.StatusRejected || !canRejectForRFQStatus(rfq.Status) {
		return domainquotation.Quotation{}, ErrInvalidQuotation
	}

	reason := normalizeOptional(input.RejectionReason)
	var rejected domainquotation.Quotation
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		updated, err := repo.RejectQuotation(ctx, input.OrganizationID, input.QuotationID, input.CurrentUser, reason)
		if err != nil {
			return err
		}

		summary := "Rejected quotation"
		metadata := map[string]any{
			"rfq_id":           input.RFQID.String(),
			"rfq_vendor_id":    quotation.RFQVendorID.String(),
			"quotation_status": string(updated.Status),
		}
		if reason != nil {
			metadata["rejection_reason"] = *reason
		}
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeQuotation),
			EntityID:       input.QuotationID,
			Action:         domainactivitylog.ActionQuotationRejected,
			Summary:        &summary,
			Metadata:       metadata,
		}); err != nil {
			return err
		}

		rejected = updated
		return nil
	}); err != nil {
		if errors.Is(err, ErrQuotationNotFound) {
			return domainquotation.Quotation{}, err
		}
		return domainquotation.Quotation{}, fmt.Errorf("reject quotation: %w", err)
	}

	return s.enrichVendorName(ctx, rejected)
}

func (s Service) ListItems(ctx context.Context, organizationID, rfqID, quotationID, currentUser uuid.UUID) ([]domainquotation.Item, error) {
	if organizationID == uuid.Nil || rfqID == uuid.Nil || quotationID == uuid.Nil || currentUser == uuid.Nil {
		return nil, ErrInvalidQuotation
	}
	if _, err := s.Get(ctx, organizationID, rfqID, quotationID, currentUser); err != nil {
		return nil, err
	}

	items, err := s.repo.ListQuotationItems(ctx, organizationID, quotationID)
	if err != nil {
		return nil, fmt.Errorf("list quotation items: %w", err)
	}

	return items, nil
}

func (s Service) UpdateItem(ctx context.Context, input UpdateItemInput) (domainquotation.Item, error) {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.QuotationID == uuid.Nil || input.ItemID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainquotation.Item{}, ErrInvalidQuotation
	}
	if !hasQuotationItemUpdate(input) {
		return domainquotation.Item{}, ErrInvalidQuotation
	}

	quotation, rfq, membership, err := s.loadManagedQuotation(ctx, input.OrganizationID, input.RFQID, input.QuotationID, input.CurrentUser)
	if err != nil {
		return domainquotation.Item{}, err
	}
	if !canManageQuotation(membership.Role) {
		return domainquotation.Item{}, ErrForbiddenQuotation
	}
	if rfq.Status != domainrfq.StatusPublished || quotation.Status != domainquotation.StatusDraft {
		return domainquotation.Item{}, ErrInvalidQuotation
	}

	items, err := s.repo.ListQuotationItems(ctx, input.OrganizationID, input.QuotationID)
	if err != nil {
		return domainquotation.Item{}, fmt.Errorf("list quotation items: %w", err)
	}

	var existing *domainquotation.Item
	for _, item := range items {
		if item.ID == input.ItemID {
			itemCopy := item
			existing = &itemCopy
			break
		}
	}
	if existing == nil {
		return domainquotation.Item{}, ErrQuotationItemNotFound
	}

	unitPrice := existing.UnitPrice
	if input.UnitPrice != nil {
		unitPrice = strings.TrimSpace(*input.UnitPrice)
	}
	if !isNonNegativeDecimal(unitPrice) {
		return domainquotation.Item{}, ErrInvalidQuotation
	}

	deliveryDays := existing.DeliveryDays
	if input.DeliveryDays != nil {
		deliveryDays = input.DeliveryDays
	}
	if deliveryDays != nil && *deliveryDays < 0 {
		return domainquotation.Item{}, ErrInvalidQuotation
	}

	var updated domainquotation.Item
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		item, err := repo.UpdateQuotationItem(ctx, UpdateItemParams{
			OrganizationID: input.OrganizationID,
			QuotationID:    input.QuotationID,
			ItemID:         input.ItemID,
			ItemName:       existing.ItemName,
			Description:    existing.Description,
			Quantity:       existing.Quantity,
			Unit:           existing.Unit,
			UnitPrice:      unitPrice,
			DeliveryDays:   deliveryDays,
			Notes:          mergeOptional(existing.Notes, input.Notes),
		})
		if err != nil {
			return err
		}

		summary := "Updated quotation item"
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeQuotation),
			EntityID:       input.QuotationID,
			Action:         domainactivitylog.ActionQuotationItemUpdated,
			Summary:        &summary,
			Metadata: map[string]any{
				"item_id":     item.ID.String(),
				"line_number": item.LineNumber,
				"unit_price":  item.UnitPrice,
			},
		}); err != nil {
			return err
		}

		updated = item
		return nil
	}); err != nil {
		if errors.Is(err, ErrQuotationItemNotFound) {
			return domainquotation.Item{}, err
		}
		return domainquotation.Item{}, fmt.Errorf("update quotation item: %w", err)
	}

	return updated, nil
}

func (s Service) enrichVendorName(ctx context.Context, quotation domainquotation.Quotation) (domainquotation.Quotation, error) {
	vendors, err := s.repo.ListRFQVendors(ctx, quotation.OrganizationID, quotation.RFQID)
	if err != nil {
		return domainquotation.Quotation{}, fmt.Errorf("list rfq vendors: %w", err)
	}

	if vendor, ok := findVendorLink(vendors, quotation.RFQVendorID); ok {
		quotation.VendorName = &vendor.VendorName
	}

	return quotation, nil
}

func (s Service) loadActiveMembership(ctx context.Context, organizationID, currentUser uuid.UUID) (domainorganization.Membership, error) {
	membership, err := s.repo.GetMembership(ctx, organizationID, currentUser)
	if err != nil {
		if errors.Is(err, ErrMembershipNotFound) {
			return domainorganization.Membership{}, ErrForbiddenQuotation
		}
		return domainorganization.Membership{}, fmt.Errorf("load membership: %w", err)
	}
	if membership.Status != domainorganization.MembershipStatusActive {
		return domainorganization.Membership{}, ErrForbiddenQuotation
	}

	return membership, nil
}

func (s Service) loadManagedRFQ(ctx context.Context, organizationID, rfqID, currentUser uuid.UUID) (domainrfq.RFQ, domainorganization.Membership, error) {
	membership, err := s.loadActiveMembership(ctx, organizationID, currentUser)
	if err != nil {
		return domainrfq.RFQ{}, domainorganization.Membership{}, err
	}

	rfq, err := s.repo.GetRFQ(ctx, organizationID, rfqID)
	if err != nil {
		if errors.Is(err, ErrRFQNotFound) {
			return domainrfq.RFQ{}, domainorganization.Membership{}, err
		}
		return domainrfq.RFQ{}, domainorganization.Membership{}, fmt.Errorf("load rfq: %w", err)
	}

	return rfq, membership, nil
}

func (s Service) loadManagedQuotation(ctx context.Context, organizationID, rfqID, quotationID, currentUser uuid.UUID) (domainquotation.Quotation, domainrfq.RFQ, domainorganization.Membership, error) {
	rfq, membership, err := s.loadManagedRFQ(ctx, organizationID, rfqID, currentUser)
	if err != nil {
		return domainquotation.Quotation{}, domainrfq.RFQ{}, domainorganization.Membership{}, err
	}

	quotation, err := s.repo.GetQuotation(ctx, organizationID, quotationID)
	if err != nil {
		if errors.Is(err, ErrQuotationNotFound) {
			return domainquotation.Quotation{}, domainrfq.RFQ{}, domainorganization.Membership{}, err
		}
		return domainquotation.Quotation{}, domainrfq.RFQ{}, domainorganization.Membership{}, fmt.Errorf("load quotation: %w", err)
	}
	if quotation.RFQID != rfqID {
		return domainquotation.Quotation{}, domainrfq.RFQ{}, domainorganization.Membership{}, ErrQuotationNotFound
	}

	return quotation, rfq, membership, nil
}

func findVendorLink(vendors []domainrfq.VendorLink, rfqVendorID uuid.UUID) (domainrfq.VendorLink, bool) {
	for _, vendor := range vendors {
		if vendor.ID == rfqVendorID {
			return vendor, true
		}
	}

	return domainrfq.VendorLink{}, false
}

func buildComparison(rfqID uuid.UUID, rows []domainquotation.ComparisonRow) domainquotation.Comparison {
	comparison := domainquotation.Comparison{
		RFQID:      rfqID,
		Quotations: make([]domainquotation.ComparisonQuotation, 0),
		LineItems:  make([]domainquotation.ComparisonLineItem, 0),
	}
	quotationSeen := make(map[uuid.UUID]bool)
	lineItemIndex := make(map[uuid.UUID]int)

	for _, row := range rows {
		if !quotationSeen[row.QuotationID] {
			comparison.Quotations = append(comparison.Quotations, domainquotation.ComparisonQuotation{
				QuotationID:  row.QuotationID,
				RFQVendorID:  row.RFQVendorID,
				VendorID:     row.VendorID,
				VendorName:   row.VendorName,
				Status:       row.Status,
				CurrencyCode: row.CurrencyCode,
				LeadTimeDays: row.LeadTimeDays,
				PaymentTerms: row.PaymentTerms,
				Notes:        row.QuotationNotes,
				TotalAmount:  row.TotalAmount,
			})
			quotationSeen[row.QuotationID] = true
		}

		index, ok := lineItemIndex[row.RFQItemID]
		if !ok {
			comparison.LineItems = append(comparison.LineItems, domainquotation.ComparisonLineItem{
				RFQItemID:   row.RFQItemID,
				LineNumber:  row.LineNumber,
				ItemName:    row.ItemName,
				Description: row.Description,
				Quantity:    row.Quantity,
				Unit:        row.Unit,
				Prices:      make([]domainquotation.ComparisonPrice, 0),
			})
			index = len(comparison.LineItems) - 1
			lineItemIndex[row.RFQItemID] = index
		}

		comparison.LineItems[index].Prices = append(comparison.LineItems[index].Prices, domainquotation.ComparisonPrice{
			QuotationID:     row.QuotationID,
			QuotationItemID: row.QuotationItemID,
			VendorID:        row.VendorID,
			VendorName:      row.VendorName,
			UnitPrice:       row.UnitPrice,
			LineTotal:       row.LineTotal,
			DeliveryDays:    row.DeliveryDays,
			Notes:           row.ItemNotes,
		})
	}

	return comparison
}

func normalizeQuotationFields(currencyInput *string, leadTimeDays *int32, paymentTermsInput *string, notesInput *string) (string, *int32, *string, *string, error) {
	currency := "USD"
	if currencyInput != nil && strings.TrimSpace(*currencyInput) != "" {
		currency = strings.ToUpper(strings.TrimSpace(*currencyInput))
	}
	if len(currency) != 3 {
		return "", nil, nil, nil, ErrInvalidQuotation
	}
	if leadTimeDays != nil && *leadTimeDays < 0 {
		return "", nil, nil, nil, ErrInvalidQuotation
	}

	return currency, leadTimeDays, normalizeOptional(paymentTermsInput), normalizeOptional(notesInput), nil
}

func canManageQuotation(role domainorganization.MembershipRole) bool {
	switch role {
	case domainorganization.MembershipRoleOwner,
		domainorganization.MembershipRoleAdmin,
		domainorganization.MembershipRoleProcurementOfficer:
		return true
	default:
		return false
	}
}

func canRejectForRFQStatus(status domainrfq.Status) bool {
	switch status {
	case domainrfq.StatusPublished,
		domainrfq.StatusClosed,
		domainrfq.StatusEvaluated:
		return true
	default:
		return false
	}
}

func hasQuotationUpdate(input UpdateInput) bool {
	return input.CurrencyCode != nil ||
		input.LeadTimeDays != nil ||
		input.PaymentTerms != nil ||
		input.Notes != nil
}

func hasQuotationItemUpdate(input UpdateItemInput) bool {
	return input.UnitPrice != nil || input.DeliveryDays != nil || input.Notes != nil
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

func isNonNegativeDecimal(value string) bool {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return err == nil && parsed >= 0
}
