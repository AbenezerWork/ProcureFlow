package rfq

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainprocurement "github.com/AbenezerWork/ProcureFlow/internal/domain/procurement"
	domainrfq "github.com/AbenezerWork/ProcureFlow/internal/domain/rfq"
	domainvendor "github.com/AbenezerWork/ProcureFlow/internal/domain/vendor"
	"github.com/google/uuid"
)

var (
	ErrInvalidRFQ                 = errors.New("invalid rfq")
	ErrRFQNotFound                = errors.New("rfq not found")
	ErrRFQReferenceTaken          = errors.New("rfq reference number already exists")
	ErrRFQVendorNotFound          = errors.New("rfq vendor not found")
	ErrRFQVendorAlreadyAttached   = errors.New("vendor already attached to rfq")
	ErrMembershipNotFound         = errors.New("organization membership not found")
	ErrProcurementRequestNotFound = errors.New("procurement request not found")
	ErrVendorNotFound             = errors.New("vendor not found")
	ErrForbiddenRFQ               = errors.New("forbidden rfq operation")
)

type CreateRFQParams struct {
	OrganizationID       uuid.UUID
	ProcurementRequestID uuid.UUID
	ReferenceNumber      *string
	Title                string
	Description          *string
	CreatedByUserID      uuid.UUID
}

type CreateRFQItemParams struct {
	OrganizationID      uuid.UUID
	RFQID               uuid.UUID
	SourceRequestItemID *uuid.UUID
	LineNumber          int32
	ItemName            string
	Description         *string
	Quantity            string
	Unit                string
	TargetDate          *string
}

type UpdateRFQParams struct {
	OrganizationID  uuid.UUID
	RFQID           uuid.UUID
	ReferenceNumber *string
	Title           string
	Description     *string
}

type Repository interface {
	CreateRFQ(ctx context.Context, params CreateRFQParams) (domainrfq.RFQ, error)
	CreateRFQItem(ctx context.Context, params CreateRFQItemParams) (domainrfq.Item, error)
	GetRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error)
	ListRFQs(ctx context.Context, organizationID uuid.UUID, status *domainrfq.Status) ([]domainrfq.RFQ, error)
	UpdateDraftRFQ(ctx context.Context, params UpdateRFQParams) (domainrfq.RFQ, error)
	PublishRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error)
	CloseRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error)
	EvaluateRFQ(ctx context.Context, organizationID, rfqID uuid.UUID) (domainrfq.RFQ, error)
	CancelRFQ(ctx context.Context, organizationID, rfqID, cancelledByUserID uuid.UUID) (domainrfq.RFQ, error)
	ListRFQItems(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainrfq.Item, error)
	AttachVendor(ctx context.Context, organizationID, rfqID, vendorID, attachedByUserID uuid.UUID) (domainrfq.VendorLink, error)
	RemoveVendor(ctx context.Context, organizationID, rfqID, vendorID uuid.UUID) error
	ListRFQVendors(ctx context.Context, organizationID, rfqID uuid.UUID) ([]domainrfq.VendorLink, error)
	GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error)
	GetProcurementRequest(ctx context.Context, organizationID, requestID uuid.UUID) (domainprocurement.Request, error)
	ListProcurementRequestItems(ctx context.Context, organizationID, requestID uuid.UUID) ([]domainprocurement.Item, error)
	GetVendor(ctx context.Context, organizationID, vendorID uuid.UUID) (domainvendor.Vendor, error)
}

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(repo Repository) error) error
}

type CreateInput struct {
	OrganizationID       uuid.UUID
	ProcurementRequestID uuid.UUID
	CurrentUser          uuid.UUID
	ReferenceNumber      *string
	Title                *string
	Description          *string
}

type ListInput struct {
	OrganizationID uuid.UUID
	CurrentUser    uuid.UUID
	Status         *domainrfq.Status
}

type UpdateInput struct {
	OrganizationID  uuid.UUID
	RFQID           uuid.UUID
	CurrentUser     uuid.UUID
	ReferenceNumber *string
	Title           *string
	Description     *string
}

type TransitionInput struct {
	OrganizationID uuid.UUID
	RFQID          uuid.UUID
	CurrentUser    uuid.UUID
}

type AttachVendorInput struct {
	OrganizationID uuid.UUID
	RFQID          uuid.UUID
	VendorID       uuid.UUID
	CurrentUser    uuid.UUID
}

type RemoveVendorInput struct {
	OrganizationID uuid.UUID
	RFQID          uuid.UUID
	VendorID       uuid.UUID
	CurrentUser    uuid.UUID
}

type Service struct {
	repo Repository
	tx   TransactionManager
}

func NewService(repo Repository, tx TransactionManager) Service {
	return Service{repo: repo, tx: tx}
}

func (s Service) Create(ctx context.Context, input CreateInput) (domainrfq.RFQ, error) {
	if input.OrganizationID == uuid.Nil || input.ProcurementRequestID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}

	membership, err := s.loadActiveMembership(ctx, input.OrganizationID, input.CurrentUser)
	if err != nil {
		return domainrfq.RFQ{}, err
	}
	if !canManageRFQ(membership.Role) {
		return domainrfq.RFQ{}, ErrForbiddenRFQ
	}

	request, err := s.repo.GetProcurementRequest(ctx, input.OrganizationID, input.ProcurementRequestID)
	if err != nil {
		if errors.Is(err, ErrProcurementRequestNotFound) {
			return domainrfq.RFQ{}, err
		}
		return domainrfq.RFQ{}, fmt.Errorf("load procurement request: %w", err)
	}
	if request.Status != domainprocurement.RequestStatusApproved {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}

	sourceItems, err := s.repo.ListProcurementRequestItems(ctx, input.OrganizationID, input.ProcurementRequestID)
	if err != nil {
		return domainrfq.RFQ{}, fmt.Errorf("list procurement request items: %w", err)
	}
	if len(sourceItems) == 0 {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}

	titleInput := request.Title
	if input.Title != nil {
		titleInput = *input.Title
	}
	descriptionInput := request.Description
	if input.Description != nil {
		descriptionInput = input.Description
	}

	referenceNumber, title, description, err := normalizeRFQFields(input.ReferenceNumber, titleInput, descriptionInput)
	if err != nil {
		return domainrfq.RFQ{}, err
	}

	var created domainrfq.RFQ
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		rfq, err := repo.CreateRFQ(ctx, CreateRFQParams{
			OrganizationID:       input.OrganizationID,
			ProcurementRequestID: input.ProcurementRequestID,
			ReferenceNumber:      referenceNumber,
			Title:                title,
			Description:          description,
			CreatedByUserID:      input.CurrentUser,
		})
		if err != nil {
			return err
		}

		for _, item := range sourceItems {
			if _, err := repo.CreateRFQItem(ctx, CreateRFQItemParams{
				OrganizationID:      input.OrganizationID,
				RFQID:               rfq.ID,
				SourceRequestItemID: &item.ID,
				LineNumber:          item.LineNumber,
				ItemName:            item.ItemName,
				Description:         item.Description,
				Quantity:            item.Quantity,
				Unit:                item.Unit,
				TargetDate:          item.NeededByDate,
			}); err != nil {
				return err
			}
		}

		created = rfq
		return nil
	}); err != nil {
		if errors.Is(err, ErrRFQReferenceTaken) {
			return domainrfq.RFQ{}, err
		}
		return domainrfq.RFQ{}, fmt.Errorf("create rfq: %w", err)
	}

	return created, nil
}

func (s Service) List(ctx context.Context, input ListInput) ([]domainrfq.RFQ, error) {
	if input.OrganizationID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return nil, ErrInvalidRFQ
	}
	if input.Status != nil && !isValidRFQStatus(*input.Status) {
		return nil, ErrInvalidRFQ
	}

	if _, err := s.loadActiveMembership(ctx, input.OrganizationID, input.CurrentUser); err != nil {
		return nil, err
	}

	rfqs, err := s.repo.ListRFQs(ctx, input.OrganizationID, input.Status)
	if err != nil {
		return nil, fmt.Errorf("list rfqs: %w", err)
	}

	return rfqs, nil
}

func (s Service) Get(ctx context.Context, organizationID, rfqID, currentUser uuid.UUID) (domainrfq.RFQ, error) {
	if organizationID == uuid.Nil || rfqID == uuid.Nil || currentUser == uuid.Nil {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}
	if _, err := s.loadActiveMembership(ctx, organizationID, currentUser); err != nil {
		return domainrfq.RFQ{}, err
	}

	rfq, err := s.repo.GetRFQ(ctx, organizationID, rfqID)
	if err != nil {
		if errors.Is(err, ErrRFQNotFound) {
			return domainrfq.RFQ{}, err
		}
		return domainrfq.RFQ{}, fmt.Errorf("get rfq: %w", err)
	}

	return rfq, nil
}

func (s Service) Update(ctx context.Context, input UpdateInput) (domainrfq.RFQ, error) {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}
	if !hasRFQUpdate(input) {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}

	rfq, membership, err := s.loadManagedRFQ(ctx, input.OrganizationID, input.RFQID, input.CurrentUser)
	if err != nil {
		return domainrfq.RFQ{}, err
	}
	if !canManageRFQ(membership.Role) {
		return domainrfq.RFQ{}, ErrForbiddenRFQ
	}
	if rfq.Status != domainrfq.StatusDraft {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}

	titleInput := rfq.Title
	if input.Title != nil {
		titleInput = *input.Title
	}
	descriptionInput := rfq.Description
	if input.Description != nil {
		descriptionInput = input.Description
	}
	referenceInput := rfq.ReferenceNumber
	if input.ReferenceNumber != nil {
		referenceInput = input.ReferenceNumber
	}

	referenceNumber, title, description, err := normalizeRFQFields(referenceInput, titleInput, descriptionInput)
	if err != nil {
		return domainrfq.RFQ{}, err
	}

	updated, err := s.repo.UpdateDraftRFQ(ctx, UpdateRFQParams{
		OrganizationID:  input.OrganizationID,
		RFQID:           input.RFQID,
		ReferenceNumber: referenceNumber,
		Title:           title,
		Description:     description,
	})
	if err != nil {
		if errors.Is(err, ErrRFQNotFound) || errors.Is(err, ErrRFQReferenceTaken) {
			return domainrfq.RFQ{}, err
		}
		return domainrfq.RFQ{}, fmt.Errorf("update rfq: %w", err)
	}

	return updated, nil
}

func (s Service) Publish(ctx context.Context, input TransitionInput) (domainrfq.RFQ, error) {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}

	rfq, membership, err := s.loadManagedRFQ(ctx, input.OrganizationID, input.RFQID, input.CurrentUser)
	if err != nil {
		return domainrfq.RFQ{}, err
	}
	if !canManageRFQ(membership.Role) {
		return domainrfq.RFQ{}, ErrForbiddenRFQ
	}
	if rfq.Status != domainrfq.StatusDraft {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}

	items, err := s.repo.ListRFQItems(ctx, input.OrganizationID, input.RFQID)
	if err != nil {
		return domainrfq.RFQ{}, fmt.Errorf("list rfq items: %w", err)
	}
	vendors, err := s.repo.ListRFQVendors(ctx, input.OrganizationID, input.RFQID)
	if err != nil {
		return domainrfq.RFQ{}, fmt.Errorf("list rfq vendors: %w", err)
	}
	if len(items) == 0 || len(vendors) == 0 {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}

	published, err := s.repo.PublishRFQ(ctx, input.OrganizationID, input.RFQID)
	if err != nil {
		if errors.Is(err, ErrRFQNotFound) {
			return domainrfq.RFQ{}, err
		}
		return domainrfq.RFQ{}, fmt.Errorf("publish rfq: %w", err)
	}

	return published, nil
}

func (s Service) Close(ctx context.Context, input TransitionInput) (domainrfq.RFQ, error) {
	return s.transition(ctx, input, domainrfq.StatusPublished, s.repo.CloseRFQ, "close rfq")
}

func (s Service) Evaluate(ctx context.Context, input TransitionInput) (domainrfq.RFQ, error) {
	return s.transition(ctx, input, domainrfq.StatusClosed, s.repo.EvaluateRFQ, "evaluate rfq")
}

func (s Service) Cancel(ctx context.Context, input TransitionInput) (domainrfq.RFQ, error) {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}

	rfq, membership, err := s.loadManagedRFQ(ctx, input.OrganizationID, input.RFQID, input.CurrentUser)
	if err != nil {
		return domainrfq.RFQ{}, err
	}
	if !canManageRFQ(membership.Role) {
		return domainrfq.RFQ{}, ErrForbiddenRFQ
	}
	if rfq.Status == domainrfq.StatusAwarded || rfq.Status == domainrfq.StatusCancelled {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}

	cancelled, err := s.repo.CancelRFQ(ctx, input.OrganizationID, input.RFQID, input.CurrentUser)
	if err != nil {
		if errors.Is(err, ErrRFQNotFound) {
			return domainrfq.RFQ{}, err
		}
		return domainrfq.RFQ{}, fmt.Errorf("cancel rfq: %w", err)
	}

	return cancelled, nil
}

func (s Service) ListItems(ctx context.Context, organizationID, rfqID, currentUser uuid.UUID) ([]domainrfq.Item, error) {
	if organizationID == uuid.Nil || rfqID == uuid.Nil || currentUser == uuid.Nil {
		return nil, ErrInvalidRFQ
	}
	if _, err := s.loadActiveMembership(ctx, organizationID, currentUser); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetRFQ(ctx, organizationID, rfqID); err != nil {
		if errors.Is(err, ErrRFQNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("load rfq: %w", err)
	}

	items, err := s.repo.ListRFQItems(ctx, organizationID, rfqID)
	if err != nil {
		return nil, fmt.Errorf("list rfq items: %w", err)
	}

	return items, nil
}

func (s Service) ListVendors(ctx context.Context, organizationID, rfqID, currentUser uuid.UUID) ([]domainrfq.VendorLink, error) {
	if organizationID == uuid.Nil || rfqID == uuid.Nil || currentUser == uuid.Nil {
		return nil, ErrInvalidRFQ
	}
	if _, err := s.loadActiveMembership(ctx, organizationID, currentUser); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetRFQ(ctx, organizationID, rfqID); err != nil {
		if errors.Is(err, ErrRFQNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("load rfq: %w", err)
	}

	vendors, err := s.repo.ListRFQVendors(ctx, organizationID, rfqID)
	if err != nil {
		return nil, fmt.Errorf("list rfq vendors: %w", err)
	}

	return vendors, nil
}

func (s Service) AttachVendor(ctx context.Context, input AttachVendorInput) (domainrfq.VendorLink, error) {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.VendorID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainrfq.VendorLink{}, ErrInvalidRFQ
	}

	rfq, membership, err := s.loadManagedRFQ(ctx, input.OrganizationID, input.RFQID, input.CurrentUser)
	if err != nil {
		return domainrfq.VendorLink{}, err
	}
	if !canManageRFQ(membership.Role) {
		return domainrfq.VendorLink{}, ErrForbiddenRFQ
	}
	if rfq.Status != domainrfq.StatusDraft {
		return domainrfq.VendorLink{}, ErrInvalidRFQ
	}

	vendor, err := s.repo.GetVendor(ctx, input.OrganizationID, input.VendorID)
	if err != nil {
		if errors.Is(err, ErrVendorNotFound) {
			return domainrfq.VendorLink{}, err
		}
		return domainrfq.VendorLink{}, fmt.Errorf("load vendor: %w", err)
	}
	if vendor.Status != domainvendor.StatusActive {
		return domainrfq.VendorLink{}, ErrInvalidRFQ
	}

	attached, err := s.repo.AttachVendor(ctx, input.OrganizationID, input.RFQID, input.VendorID, input.CurrentUser)
	if err != nil {
		if errors.Is(err, ErrRFQVendorAlreadyAttached) {
			return domainrfq.VendorLink{}, err
		}
		return domainrfq.VendorLink{}, fmt.Errorf("attach rfq vendor: %w", err)
	}

	return attached, nil
}

func (s Service) RemoveVendor(ctx context.Context, input RemoveVendorInput) error {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.VendorID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return ErrInvalidRFQ
	}

	rfq, membership, err := s.loadManagedRFQ(ctx, input.OrganizationID, input.RFQID, input.CurrentUser)
	if err != nil {
		return err
	}
	if !canManageRFQ(membership.Role) {
		return ErrForbiddenRFQ
	}
	if rfq.Status != domainrfq.StatusDraft {
		return ErrInvalidRFQ
	}

	if err := s.repo.RemoveVendor(ctx, input.OrganizationID, input.RFQID, input.VendorID); err != nil {
		if errors.Is(err, ErrRFQVendorNotFound) {
			return err
		}
		return fmt.Errorf("remove rfq vendor: %w", err)
	}

	return nil
}

func (s Service) transition(ctx context.Context, input TransitionInput, expected domainrfq.Status, fn func(context.Context, uuid.UUID, uuid.UUID) (domainrfq.RFQ, error), action string) (domainrfq.RFQ, error) {
	if input.OrganizationID == uuid.Nil || input.RFQID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}

	rfq, membership, err := s.loadManagedRFQ(ctx, input.OrganizationID, input.RFQID, input.CurrentUser)
	if err != nil {
		return domainrfq.RFQ{}, err
	}
	if !canManageRFQ(membership.Role) {
		return domainrfq.RFQ{}, ErrForbiddenRFQ
	}
	if rfq.Status != expected {
		return domainrfq.RFQ{}, ErrInvalidRFQ
	}

	updated, err := fn(ctx, input.OrganizationID, input.RFQID)
	if err != nil {
		if errors.Is(err, ErrRFQNotFound) {
			return domainrfq.RFQ{}, err
		}
		return domainrfq.RFQ{}, fmt.Errorf("%s: %w", action, err)
	}

	return updated, nil
}

func (s Service) loadActiveMembership(ctx context.Context, organizationID, currentUser uuid.UUID) (domainorganization.Membership, error) {
	membership, err := s.repo.GetMembership(ctx, organizationID, currentUser)
	if err != nil {
		if errors.Is(err, ErrMembershipNotFound) {
			return domainorganization.Membership{}, ErrForbiddenRFQ
		}
		return domainorganization.Membership{}, fmt.Errorf("load membership: %w", err)
	}
	if membership.Status != domainorganization.MembershipStatusActive {
		return domainorganization.Membership{}, ErrForbiddenRFQ
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

func hasRFQUpdate(input UpdateInput) bool {
	return input.ReferenceNumber != nil || input.Title != nil || input.Description != nil
}

func canManageRFQ(role domainorganization.MembershipRole) bool {
	switch role {
	case domainorganization.MembershipRoleOwner,
		domainorganization.MembershipRoleAdmin,
		domainorganization.MembershipRoleProcurementOfficer:
		return true
	default:
		return false
	}
}

func isValidRFQStatus(status domainrfq.Status) bool {
	switch status {
	case domainrfq.StatusDraft,
		domainrfq.StatusPublished,
		domainrfq.StatusClosed,
		domainrfq.StatusEvaluated,
		domainrfq.StatusAwarded,
		domainrfq.StatusCancelled:
		return true
	default:
		return false
	}
}

func normalizeRFQFields(referenceNumber *string, title string, description *string) (*string, string, *string, error) {
	normalizedTitle := strings.TrimSpace(title)
	if normalizedTitle == "" {
		return nil, "", nil, ErrInvalidRFQ
	}

	return normalizeOptional(referenceNumber), normalizedTitle, normalizeOptional(description), nil
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
