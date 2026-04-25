package procurement

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainprocurement "github.com/AbenezerWork/ProcureFlow/internal/domain/procurement"
	"github.com/google/uuid"
)

var (
	ErrInvalidProcurementRequest  = errors.New("invalid procurement request")
	ErrProcurementRequestNotFound = errors.New("procurement request not found")
	ErrProcurementItemNotFound    = errors.New("procurement request item not found")
	ErrMembershipNotFound         = errors.New("organization membership not found")
	ErrForbiddenProcurement       = errors.New("forbidden procurement operation")
)

type CreateRequestParams struct {
	OrganizationID       uuid.UUID
	RequesterUserID      uuid.UUID
	Title                string
	Description          *string
	Justification        *string
	CurrencyCode         string
	EstimatedTotalAmount *string
}

type UpdateRequestParams struct {
	OrganizationID       uuid.UUID
	RequestID            uuid.UUID
	Title                string
	Description          *string
	Justification        *string
	CurrencyCode         string
	EstimatedTotalAmount *string
}

type CreateItemParams struct {
	OrganizationID       uuid.UUID
	ProcurementRequestID uuid.UUID
	LineNumber           int32
	ItemName             string
	Description          *string
	Quantity             string
	Unit                 string
	EstimatedUnitPrice   *string
	NeededByDate         *string
}

type UpdateItemParams struct {
	OrganizationID       uuid.UUID
	ProcurementRequestID uuid.UUID
	ItemID               uuid.UUID
	ItemName             string
	Description          *string
	Quantity             string
	Unit                 string
	EstimatedUnitPrice   *string
	NeededByDate         *string
}

type Repository interface {
	CreateRequest(ctx context.Context, params CreateRequestParams) (domainprocurement.Request, error)
	GetRequest(ctx context.Context, organizationID, requestID uuid.UUID) (domainprocurement.Request, error)
	ListRequests(ctx context.Context, organizationID uuid.UUID, status *domainprocurement.RequestStatus) ([]domainprocurement.Request, error)
	ListApprovalInbox(ctx context.Context, organizationID uuid.UUID) ([]domainprocurement.Request, error)
	UpdateDraftRequest(ctx context.Context, params UpdateRequestParams) (domainprocurement.Request, error)
	SubmitRequest(ctx context.Context, organizationID, requestID, submittedByUserID uuid.UUID) (domainprocurement.Request, error)
	ApproveRequest(ctx context.Context, organizationID, requestID, approvedByUserID uuid.UUID, decisionComment *string) (domainprocurement.Request, error)
	RejectRequest(ctx context.Context, organizationID, requestID, rejectedByUserID uuid.UUID, decisionComment *string) (domainprocurement.Request, error)
	CancelRequest(ctx context.Context, organizationID, requestID, cancelledByUserID uuid.UUID) (domainprocurement.Request, error)
	CreateItem(ctx context.Context, params CreateItemParams) (domainprocurement.Item, error)
	ListItems(ctx context.Context, organizationID, requestID uuid.UUID) ([]domainprocurement.Item, error)
	UpdateItem(ctx context.Context, params UpdateItemParams) (domainprocurement.Item, error)
	DeleteItem(ctx context.Context, organizationID, requestID, itemID uuid.UUID) error
	GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error)
	CreateActivityLog(ctx context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error)
}

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(repo Repository) error) error
}

type CreateRequestInput struct {
	OrganizationID       uuid.UUID
	CurrentUser          uuid.UUID
	Title                string
	Description          *string
	Justification        *string
	CurrencyCode         *string
	EstimatedTotalAmount *string
}

type ListRequestsInput struct {
	OrganizationID uuid.UUID
	CurrentUser    uuid.UUID
	Status         *domainprocurement.RequestStatus
}

type ListApprovalInboxInput struct {
	OrganizationID uuid.UUID
	CurrentUser    uuid.UUID
}

type UpdateRequestInput struct {
	OrganizationID       uuid.UUID
	RequestID            uuid.UUID
	CurrentUser          uuid.UUID
	Title                *string
	Description          *string
	Justification        *string
	CurrencyCode         *string
	EstimatedTotalAmount *string
}

type SubmitRequestInput struct {
	OrganizationID uuid.UUID
	RequestID      uuid.UUID
	CurrentUser    uuid.UUID
}

type DecisionInput struct {
	OrganizationID  uuid.UUID
	RequestID       uuid.UUID
	CurrentUser     uuid.UUID
	DecisionComment *string
}

type CancelRequestInput struct {
	OrganizationID uuid.UUID
	RequestID      uuid.UUID
	CurrentUser    uuid.UUID
}

type CreateItemInput struct {
	OrganizationID     uuid.UUID
	RequestID          uuid.UUID
	CurrentUser        uuid.UUID
	ItemName           string
	Description        *string
	Quantity           string
	Unit               string
	EstimatedUnitPrice *string
	NeededByDate       *string
}

type UpdateItemInput struct {
	OrganizationID     uuid.UUID
	RequestID          uuid.UUID
	ItemID             uuid.UUID
	CurrentUser        uuid.UUID
	ItemName           *string
	Description        *string
	Quantity           *string
	Unit               *string
	EstimatedUnitPrice *string
	NeededByDate       *string
}

type DeleteItemInput struct {
	OrganizationID uuid.UUID
	RequestID      uuid.UUID
	ItemID         uuid.UUID
	CurrentUser    uuid.UUID
}

type Service struct {
	repo Repository
	tx   TransactionManager
}

func NewService(repo Repository, tx TransactionManager) Service {
	return Service{repo: repo, tx: tx}
}

func (s Service) CreateRequest(ctx context.Context, input CreateRequestInput) (domainprocurement.Request, error) {
	if input.OrganizationID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainprocurement.Request{}, ErrInvalidProcurementRequest
	}

	membership, err := s.loadActiveMembership(ctx, input.OrganizationID, input.CurrentUser)
	if err != nil {
		return domainprocurement.Request{}, err
	}
	if !canCreateRequest(membership.Role) {
		return domainprocurement.Request{}, ErrForbiddenProcurement
	}

	title := strings.TrimSpace(input.Title)
	currency, amount, err := normalizeRequestFields(title, input.CurrencyCode, input.EstimatedTotalAmount)
	if err != nil {
		return domainprocurement.Request{}, err
	}

	var created domainprocurement.Request
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		request, err := repo.CreateRequest(ctx, CreateRequestParams{
			OrganizationID:       input.OrganizationID,
			RequesterUserID:      input.CurrentUser,
			Title:                title,
			Description:          normalizeOptional(input.Description),
			Justification:        normalizeOptional(input.Justification),
			CurrencyCode:         currency,
			EstimatedTotalAmount: amount,
		})
		if err != nil {
			return err
		}

		summary := "Created procurement request"
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeProcurementRequest),
			EntityID:       request.ID,
			Action:         domainactivitylog.ActionProcurementRequestCreated,
			Summary:        &summary,
			Metadata: map[string]any{
				"requester_user_id": input.CurrentUser.String(),
				"request_status":    string(request.Status),
			},
		}); err != nil {
			return err
		}

		created = request
		return nil
	}); err != nil {
		return domainprocurement.Request{}, fmt.Errorf("create procurement request: %w", err)
	}

	return created, nil
}

func (s Service) ListRequests(ctx context.Context, input ListRequestsInput) ([]domainprocurement.Request, error) {
	if input.OrganizationID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return nil, ErrInvalidProcurementRequest
	}
	if input.Status != nil && !isValidRequestStatus(*input.Status) {
		return nil, ErrInvalidProcurementRequest
	}

	if _, err := s.loadActiveMembership(ctx, input.OrganizationID, input.CurrentUser); err != nil {
		return nil, err
	}

	requests, err := s.repo.ListRequests(ctx, input.OrganizationID, input.Status)
	if err != nil {
		return nil, fmt.Errorf("list procurement requests: %w", err)
	}

	return requests, nil
}

func (s Service) ListApprovalInbox(ctx context.Context, input ListApprovalInboxInput) ([]domainprocurement.Request, error) {
	if input.OrganizationID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return nil, ErrInvalidProcurementRequest
	}

	membership, err := s.loadActiveMembership(ctx, input.OrganizationID, input.CurrentUser)
	if err != nil {
		return nil, err
	}
	if !canApproveRequest(membership.Role) {
		return nil, ErrForbiddenProcurement
	}

	requests, err := s.repo.ListApprovalInbox(ctx, input.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("list approval inbox: %w", err)
	}

	return requests, nil
}

func (s Service) GetRequest(ctx context.Context, organizationID, requestID, currentUser uuid.UUID) (domainprocurement.Request, error) {
	if organizationID == uuid.Nil || requestID == uuid.Nil || currentUser == uuid.Nil {
		return domainprocurement.Request{}, ErrInvalidProcurementRequest
	}

	if _, err := s.loadActiveMembership(ctx, organizationID, currentUser); err != nil {
		return domainprocurement.Request{}, err
	}

	request, err := s.repo.GetRequest(ctx, organizationID, requestID)
	if err != nil {
		if errors.Is(err, ErrProcurementRequestNotFound) {
			return domainprocurement.Request{}, err
		}

		return domainprocurement.Request{}, fmt.Errorf("get procurement request: %w", err)
	}

	return request, nil
}

func (s Service) UpdateRequest(ctx context.Context, input UpdateRequestInput) (domainprocurement.Request, error) {
	if input.OrganizationID == uuid.Nil || input.RequestID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainprocurement.Request{}, ErrInvalidProcurementRequest
	}
	if !hasRequestUpdate(input) {
		return domainprocurement.Request{}, ErrInvalidProcurementRequest
	}

	request, membership, err := s.loadMutableRequest(ctx, input.OrganizationID, input.RequestID, input.CurrentUser)
	if err != nil {
		return domainprocurement.Request{}, err
	}
	if !canWriteRequest(membership.Role, request, input.CurrentUser) {
		return domainprocurement.Request{}, ErrForbiddenProcurement
	}
	if request.Status != domainprocurement.RequestStatusDraft {
		return domainprocurement.Request{}, ErrInvalidProcurementRequest
	}

	title := request.Title
	if input.Title != nil {
		title = strings.TrimSpace(*input.Title)
	}
	currencyInput := &request.CurrencyCode
	if input.CurrencyCode != nil {
		currencyInput = input.CurrencyCode
	}
	amountInput := request.EstimatedTotalAmount
	if input.EstimatedTotalAmount != nil {
		amountInput = input.EstimatedTotalAmount
	}

	currency, amount, err := normalizeRequestFields(title, currencyInput, amountInput)
	if err != nil {
		return domainprocurement.Request{}, err
	}

	var updated domainprocurement.Request
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		next, err := repo.UpdateDraftRequest(ctx, UpdateRequestParams{
			OrganizationID:       input.OrganizationID,
			RequestID:            input.RequestID,
			Title:                title,
			Description:          mergeOptional(request.Description, input.Description),
			Justification:        mergeOptional(request.Justification, input.Justification),
			CurrencyCode:         currency,
			EstimatedTotalAmount: amount,
		})
		if err != nil {
			return err
		}

		summary := "Updated procurement request"
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeProcurementRequest),
			EntityID:       next.ID,
			Action:         domainactivitylog.ActionProcurementRequestUpdated,
			Summary:        &summary,
			Metadata: map[string]any{
				"request_status": string(next.Status),
			},
		}); err != nil {
			return err
		}

		updated = next
		return nil
	}); err != nil {
		if errors.Is(err, ErrProcurementRequestNotFound) {
			return domainprocurement.Request{}, err
		}

		return domainprocurement.Request{}, fmt.Errorf("update procurement request: %w", err)
	}

	return updated, nil
}

func (s Service) SubmitRequest(ctx context.Context, input SubmitRequestInput) (domainprocurement.Request, error) {
	if input.OrganizationID == uuid.Nil || input.RequestID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainprocurement.Request{}, ErrInvalidProcurementRequest
	}

	request, membership, err := s.loadMutableRequest(ctx, input.OrganizationID, input.RequestID, input.CurrentUser)
	if err != nil {
		return domainprocurement.Request{}, err
	}
	if !canFinishDraftRequest(membership.Role, request, input.CurrentUser) {
		return domainprocurement.Request{}, ErrForbiddenProcurement
	}
	if request.Status != domainprocurement.RequestStatusDraft {
		return domainprocurement.Request{}, ErrInvalidProcurementRequest
	}

	var submitted domainprocurement.Request
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		updated, err := repo.SubmitRequest(ctx, input.OrganizationID, input.RequestID, input.CurrentUser)
		if err != nil {
			return err
		}

		summary := "Submitted procurement request"
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeProcurementRequest),
			EntityID:       input.RequestID,
			Action:         domainactivitylog.ActionProcurementRequestSubmitted,
			Summary:        &summary,
			Metadata: map[string]any{
				"request_status": string(updated.Status),
			},
		}); err != nil {
			return err
		}

		submitted = updated
		return nil
	}); err != nil {
		if errors.Is(err, ErrProcurementRequestNotFound) {
			return domainprocurement.Request{}, err
		}

		return domainprocurement.Request{}, fmt.Errorf("submit procurement request: %w", err)
	}

	return submitted, nil
}

func (s Service) ApproveRequest(ctx context.Context, input DecisionInput) (domainprocurement.Request, error) {
	if input.OrganizationID == uuid.Nil || input.RequestID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainprocurement.Request{}, ErrInvalidProcurementRequest
	}

	request, membership, err := s.loadMutableRequest(ctx, input.OrganizationID, input.RequestID, input.CurrentUser)
	if err != nil {
		return domainprocurement.Request{}, err
	}
	if !canApproveRequest(membership.Role) {
		return domainprocurement.Request{}, ErrForbiddenProcurement
	}
	if request.Status != domainprocurement.RequestStatusSubmitted {
		return domainprocurement.Request{}, ErrInvalidProcurementRequest
	}

	comment := normalizeOptional(input.DecisionComment)
	var approved domainprocurement.Request
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		updated, err := repo.ApproveRequest(ctx, input.OrganizationID, input.RequestID, input.CurrentUser, comment)
		if err != nil {
			return err
		}

		summary := "Approved procurement request"
		metadata := map[string]any{
			"request_status": string(updated.Status),
		}
		if comment != nil {
			metadata["decision_comment"] = *comment
		}
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeProcurementRequest),
			EntityID:       input.RequestID,
			Action:         domainactivitylog.ActionProcurementRequestApproved,
			Summary:        &summary,
			Metadata:       metadata,
		}); err != nil {
			return err
		}

		approved = updated
		return nil
	}); err != nil {
		if errors.Is(err, ErrProcurementRequestNotFound) {
			return domainprocurement.Request{}, err
		}

		return domainprocurement.Request{}, fmt.Errorf("approve procurement request: %w", err)
	}

	return approved, nil
}

func (s Service) RejectRequest(ctx context.Context, input DecisionInput) (domainprocurement.Request, error) {
	if input.OrganizationID == uuid.Nil || input.RequestID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainprocurement.Request{}, ErrInvalidProcurementRequest
	}

	request, membership, err := s.loadMutableRequest(ctx, input.OrganizationID, input.RequestID, input.CurrentUser)
	if err != nil {
		return domainprocurement.Request{}, err
	}
	if !canApproveRequest(membership.Role) {
		return domainprocurement.Request{}, ErrForbiddenProcurement
	}
	if request.Status != domainprocurement.RequestStatusSubmitted {
		return domainprocurement.Request{}, ErrInvalidProcurementRequest
	}

	comment := normalizeOptional(input.DecisionComment)
	var rejected domainprocurement.Request
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		updated, err := repo.RejectRequest(ctx, input.OrganizationID, input.RequestID, input.CurrentUser, comment)
		if err != nil {
			return err
		}

		summary := "Rejected procurement request"
		metadata := map[string]any{
			"request_status": string(updated.Status),
		}
		if comment != nil {
			metadata["decision_comment"] = *comment
		}
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeProcurementRequest),
			EntityID:       input.RequestID,
			Action:         domainactivitylog.ActionProcurementRequestRejected,
			Summary:        &summary,
			Metadata:       metadata,
		}); err != nil {
			return err
		}

		rejected = updated
		return nil
	}); err != nil {
		if errors.Is(err, ErrProcurementRequestNotFound) {
			return domainprocurement.Request{}, err
		}

		return domainprocurement.Request{}, fmt.Errorf("reject procurement request: %w", err)
	}

	return rejected, nil
}

func (s Service) CancelRequest(ctx context.Context, input CancelRequestInput) (domainprocurement.Request, error) {
	if input.OrganizationID == uuid.Nil || input.RequestID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainprocurement.Request{}, ErrInvalidProcurementRequest
	}

	request, membership, err := s.loadMutableRequest(ctx, input.OrganizationID, input.RequestID, input.CurrentUser)
	if err != nil {
		return domainprocurement.Request{}, err
	}
	if !canWriteRequest(membership.Role, request, input.CurrentUser) {
		return domainprocurement.Request{}, ErrForbiddenProcurement
	}
	if request.Status != domainprocurement.RequestStatusDraft && request.Status != domainprocurement.RequestStatusSubmitted {
		return domainprocurement.Request{}, ErrInvalidProcurementRequest
	}

	var cancelled domainprocurement.Request
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		updated, err := repo.CancelRequest(ctx, input.OrganizationID, input.RequestID, input.CurrentUser)
		if err != nil {
			return err
		}

		summary := "Cancelled procurement request"
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeProcurementRequest),
			EntityID:       input.RequestID,
			Action:         domainactivitylog.ActionProcurementRequestCanceled,
			Summary:        &summary,
			Metadata: map[string]any{
				"request_status": string(updated.Status),
			},
		}); err != nil {
			return err
		}

		cancelled = updated
		return nil
	}); err != nil {
		if errors.Is(err, ErrProcurementRequestNotFound) {
			return domainprocurement.Request{}, err
		}

		return domainprocurement.Request{}, fmt.Errorf("cancel procurement request: %w", err)
	}

	return cancelled, nil
}

func (s Service) ListItems(ctx context.Context, organizationID, requestID, currentUser uuid.UUID) ([]domainprocurement.Item, error) {
	if organizationID == uuid.Nil || requestID == uuid.Nil || currentUser == uuid.Nil {
		return nil, ErrInvalidProcurementRequest
	}

	if _, err := s.loadActiveMembership(ctx, organizationID, currentUser); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetRequest(ctx, organizationID, requestID); err != nil {
		if errors.Is(err, ErrProcurementRequestNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("load procurement request: %w", err)
	}

	items, err := s.repo.ListItems(ctx, organizationID, requestID)
	if err != nil {
		return nil, fmt.Errorf("list procurement request items: %w", err)
	}

	return items, nil
}

func (s Service) CreateItem(ctx context.Context, input CreateItemInput) (domainprocurement.Item, error) {
	if input.OrganizationID == uuid.Nil || input.RequestID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainprocurement.Item{}, ErrInvalidProcurementRequest
	}

	request, membership, err := s.loadMutableRequest(ctx, input.OrganizationID, input.RequestID, input.CurrentUser)
	if err != nil {
		return domainprocurement.Item{}, err
	}
	if !canWriteRequest(membership.Role, request, input.CurrentUser) {
		return domainprocurement.Item{}, ErrForbiddenProcurement
	}
	if request.Status != domainprocurement.RequestStatusDraft {
		return domainprocurement.Item{}, ErrInvalidProcurementRequest
	}

	items, err := s.repo.ListItems(ctx, input.OrganizationID, input.RequestID)
	if err != nil {
		return domainprocurement.Item{}, fmt.Errorf("load procurement request items: %w", err)
	}

	nextLineNumber := int32(1)
	for _, item := range items {
		if item.LineNumber >= nextLineNumber {
			nextLineNumber = item.LineNumber + 1
		}
	}

	params, err := normalizeItemParams(input.ItemName, input.Description, input.Quantity, input.Unit, input.EstimatedUnitPrice, input.NeededByDate)
	if err != nil {
		return domainprocurement.Item{}, err
	}

	var created domainprocurement.Item
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		item, err := repo.CreateItem(ctx, CreateItemParams{
			OrganizationID:       input.OrganizationID,
			ProcurementRequestID: input.RequestID,
			LineNumber:           nextLineNumber,
			ItemName:             params.ItemName,
			Description:          params.Description,
			Quantity:             params.Quantity,
			Unit:                 params.Unit,
			EstimatedUnitPrice:   params.EstimatedUnitPrice,
			NeededByDate:         params.NeededByDate,
		})
		if err != nil {
			return err
		}

		summary := "Added procurement request item"
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeProcurementRequest),
			EntityID:       input.RequestID,
			Action:         domainactivitylog.ActionProcurementRequestItemAdded,
			Summary:        &summary,
			Metadata: map[string]any{
				"item_id":     item.ID.String(),
				"line_number": item.LineNumber,
				"item_name":   item.ItemName,
			},
		}); err != nil {
			return err
		}

		created = item
		return nil
	}); err != nil {
		return domainprocurement.Item{}, fmt.Errorf("create procurement request item: %w", err)
	}

	return created, nil
}

func (s Service) UpdateItem(ctx context.Context, input UpdateItemInput) (domainprocurement.Item, error) {
	if input.OrganizationID == uuid.Nil || input.RequestID == uuid.Nil || input.ItemID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return domainprocurement.Item{}, ErrInvalidProcurementRequest
	}
	if !hasItemUpdate(input) {
		return domainprocurement.Item{}, ErrInvalidProcurementRequest
	}

	request, membership, err := s.loadMutableRequest(ctx, input.OrganizationID, input.RequestID, input.CurrentUser)
	if err != nil {
		return domainprocurement.Item{}, err
	}
	if !canWriteRequest(membership.Role, request, input.CurrentUser) {
		return domainprocurement.Item{}, ErrForbiddenProcurement
	}
	if request.Status != domainprocurement.RequestStatusDraft {
		return domainprocurement.Item{}, ErrInvalidProcurementRequest
	}

	items, err := s.repo.ListItems(ctx, input.OrganizationID, input.RequestID)
	if err != nil {
		return domainprocurement.Item{}, fmt.Errorf("load procurement request items: %w", err)
	}

	var existing *domainprocurement.Item
	for _, item := range items {
		if item.ID == input.ItemID {
			itemCopy := item
			existing = &itemCopy
			break
		}
	}
	if existing == nil {
		return domainprocurement.Item{}, ErrProcurementItemNotFound
	}

	itemName := existing.ItemName
	if input.ItemName != nil {
		itemName = *input.ItemName
	}
	quantity := existing.Quantity
	if input.Quantity != nil {
		quantity = *input.Quantity
	}
	unit := existing.Unit
	if input.Unit != nil {
		unit = *input.Unit
	}
	estimatedUnitPrice := existing.EstimatedUnitPrice
	if input.EstimatedUnitPrice != nil {
		estimatedUnitPrice = input.EstimatedUnitPrice
	}
	neededByDate := existing.NeededByDate
	if input.NeededByDate != nil {
		neededByDate = input.NeededByDate
	}

	params, err := normalizeItemParams(itemName, mergeOptional(existing.Description, input.Description), quantity, unit, estimatedUnitPrice, neededByDate)
	if err != nil {
		return domainprocurement.Item{}, err
	}

	var updated domainprocurement.Item
	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		item, err := repo.UpdateItem(ctx, UpdateItemParams{
			OrganizationID:       input.OrganizationID,
			ProcurementRequestID: input.RequestID,
			ItemID:               input.ItemID,
			ItemName:             params.ItemName,
			Description:          params.Description,
			Quantity:             params.Quantity,
			Unit:                 params.Unit,
			EstimatedUnitPrice:   params.EstimatedUnitPrice,
			NeededByDate:         params.NeededByDate,
		})
		if err != nil {
			return err
		}

		summary := "Updated procurement request item"
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeProcurementRequest),
			EntityID:       input.RequestID,
			Action:         domainactivitylog.ActionProcurementRequestItemUpdated,
			Summary:        &summary,
			Metadata: map[string]any{
				"item_id":     item.ID.String(),
				"line_number": item.LineNumber,
				"item_name":   item.ItemName,
			},
		}); err != nil {
			return err
		}

		updated = item
		return nil
	}); err != nil {
		if errors.Is(err, ErrProcurementItemNotFound) {
			return domainprocurement.Item{}, err
		}

		return domainprocurement.Item{}, fmt.Errorf("update procurement request item: %w", err)
	}

	return updated, nil
}

func (s Service) DeleteItem(ctx context.Context, input DeleteItemInput) error {
	if input.OrganizationID == uuid.Nil || input.RequestID == uuid.Nil || input.ItemID == uuid.Nil || input.CurrentUser == uuid.Nil {
		return ErrInvalidProcurementRequest
	}

	request, membership, err := s.loadMutableRequest(ctx, input.OrganizationID, input.RequestID, input.CurrentUser)
	if err != nil {
		return err
	}
	if !canWriteRequest(membership.Role, request, input.CurrentUser) {
		return ErrForbiddenProcurement
	}
	if request.Status != domainprocurement.RequestStatusDraft {
		return ErrInvalidProcurementRequest
	}

	if err := s.tx.WithinTransaction(ctx, func(repo Repository) error {
		if err := repo.DeleteItem(ctx, input.OrganizationID, input.RequestID, input.ItemID); err != nil {
			return err
		}

		summary := "Deleted procurement request item"
		if _, err := repo.CreateActivityLog(ctx, applicationactivitylog.CreateParams{
			OrganizationID: input.OrganizationID,
			ActorUserID:    &input.CurrentUser,
			EntityType:     string(domainactivitylog.EntityTypeProcurementRequest),
			EntityID:       input.RequestID,
			Action:         domainactivitylog.ActionProcurementRequestItemDeleted,
			Summary:        &summary,
			Metadata: map[string]any{
				"item_id": input.ItemID.String(),
			},
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		if errors.Is(err, ErrProcurementItemNotFound) {
			return err
		}

		return fmt.Errorf("delete procurement request item: %w", err)
	}

	return nil
}

func (s Service) loadActiveMembership(ctx context.Context, organizationID, currentUser uuid.UUID) (domainorganization.Membership, error) {
	membership, err := s.repo.GetMembership(ctx, organizationID, currentUser)
	if err != nil {
		if errors.Is(err, ErrMembershipNotFound) {
			return domainorganization.Membership{}, ErrForbiddenProcurement
		}

		return domainorganization.Membership{}, fmt.Errorf("load membership: %w", err)
	}
	if membership.Status != domainorganization.MembershipStatusActive {
		return domainorganization.Membership{}, ErrForbiddenProcurement
	}

	return membership, nil
}

func (s Service) loadMutableRequest(ctx context.Context, organizationID, requestID, currentUser uuid.UUID) (domainprocurement.Request, domainorganization.Membership, error) {
	membership, err := s.loadActiveMembership(ctx, organizationID, currentUser)
	if err != nil {
		return domainprocurement.Request{}, domainorganization.Membership{}, err
	}

	request, err := s.repo.GetRequest(ctx, organizationID, requestID)
	if err != nil {
		if errors.Is(err, ErrProcurementRequestNotFound) {
			return domainprocurement.Request{}, domainorganization.Membership{}, err
		}

		return domainprocurement.Request{}, domainorganization.Membership{}, fmt.Errorf("load procurement request: %w", err)
	}

	return request, membership, nil
}

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

func normalizeRequestFields(title string, currencyInput *string, amountInput *string) (string, *string, error) {
	if strings.TrimSpace(title) == "" {
		return "", nil, ErrInvalidProcurementRequest
	}

	currency := "USD"
	if currencyInput != nil && strings.TrimSpace(*currencyInput) != "" {
		currency = strings.ToUpper(strings.TrimSpace(*currencyInput))
	}
	if !currencyPattern.MatchString(currency) {
		return "", nil, ErrInvalidProcurementRequest
	}

	amount, err := normalizeNonNegativeDecimal(amountInput)
	if err != nil {
		return "", nil, ErrInvalidProcurementRequest
	}

	return currency, amount, nil
}

type normalizedItemParams struct {
	ItemName           string
	Description        *string
	Quantity           string
	Unit               string
	EstimatedUnitPrice *string
	NeededByDate       *string
}

func normalizeItemParams(itemName string, description *string, quantity string, unit string, estimatedUnitPrice *string, neededByDate *string) (normalizedItemParams, error) {
	name := strings.TrimSpace(itemName)
	if name == "" {
		return normalizedItemParams{}, ErrInvalidProcurementRequest
	}

	qty := strings.TrimSpace(quantity)
	if qty == "" || !isPositiveDecimal(qty) {
		return normalizedItemParams{}, ErrInvalidProcurementRequest
	}

	normalizedUnit := strings.TrimSpace(unit)
	if normalizedUnit == "" {
		return normalizedItemParams{}, ErrInvalidProcurementRequest
	}

	price, err := normalizeNonNegativeDecimal(estimatedUnitPrice)
	if err != nil {
		return normalizedItemParams{}, ErrInvalidProcurementRequest
	}

	date, err := normalizeDate(neededByDate)
	if err != nil {
		return normalizedItemParams{}, ErrInvalidProcurementRequest
	}

	return normalizedItemParams{
		ItemName:           name,
		Description:        normalizeOptional(description),
		Quantity:           qty,
		Unit:               normalizedUnit,
		EstimatedUnitPrice: price,
		NeededByDate:       date,
	}, nil
}

func hasRequestUpdate(input UpdateRequestInput) bool {
	return input.Title != nil ||
		input.Description != nil ||
		input.Justification != nil ||
		input.CurrencyCode != nil ||
		input.EstimatedTotalAmount != nil
}

func hasItemUpdate(input UpdateItemInput) bool {
	return input.ItemName != nil ||
		input.Description != nil ||
		input.Quantity != nil ||
		input.Unit != nil ||
		input.EstimatedUnitPrice != nil ||
		input.NeededByDate != nil
}

func canCreateRequest(role domainorganization.MembershipRole) bool {
	switch role {
	case domainorganization.MembershipRoleOwner,
		domainorganization.MembershipRoleAdmin,
		domainorganization.MembershipRoleProcurementOfficer,
		domainorganization.MembershipRoleRequester,
		domainorganization.MembershipRoleApprover:
		return true
	default:
		return false
	}
}

func canManageAnyRequest(role domainorganization.MembershipRole) bool {
	switch role {
	case domainorganization.MembershipRoleOwner,
		domainorganization.MembershipRoleAdmin,
		domainorganization.MembershipRoleProcurementOfficer:
		return true
	default:
		return false
	}
}

func canApproveRequest(role domainorganization.MembershipRole) bool {
	switch role {
	case domainorganization.MembershipRoleOwner,
		domainorganization.MembershipRoleAdmin,
		domainorganization.MembershipRoleApprover:
		return true
	default:
		return false
	}
}

func canWriteRequest(role domainorganization.MembershipRole, request domainprocurement.Request, currentUser uuid.UUID) bool {
	if canManageAnyRequest(role) {
		return true
	}

	return canFinishDraftRequest(role, request, currentUser)
}

func canFinishDraftRequest(role domainorganization.MembershipRole, request domainprocurement.Request, currentUser uuid.UUID) bool {
	if request.RequesterUserID != currentUser {
		return false
	}

	switch role {
	case domainorganization.MembershipRoleRequester,
		domainorganization.MembershipRoleApprover:
		return true
	default:
		return false
	}
}

func isValidRequestStatus(status domainprocurement.RequestStatus) bool {
	switch status {
	case domainprocurement.RequestStatusDraft,
		domainprocurement.RequestStatusSubmitted,
		domainprocurement.RequestStatusApproved,
		domainprocurement.RequestStatusRejected,
		domainprocurement.RequestStatusCancelled:
		return true
	default:
		return false
	}
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

func normalizeNonNegativeDecimal(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseFloat(trimmed, 64)
	if err != nil || parsed < 0 {
		return nil, ErrInvalidProcurementRequest
	}

	return &trimmed, nil
}

func isPositiveDecimal(value string) bool {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return err == nil && parsed > 0
}

func normalizeDate(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, err
	}

	formatted := parsed.Format("2006-01-02")
	return &formatted, nil
}
