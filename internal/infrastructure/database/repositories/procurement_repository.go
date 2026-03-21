package repositories

import (
	"context"
	"errors"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	applicationprocurement "github.com/AbenezerWork/ProcureFlow/internal/application/procurement"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	domainprocurement "github.com/AbenezerWork/ProcureFlow/internal/domain/procurement"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProcurementRepository struct {
	store *database.Store
	hooks []applicationactivitylog.Hook
}

func NewProcurementRepository(store *database.Store, hooks ...applicationactivitylog.Hook) *ProcurementRepository {
	return &ProcurementRepository{store: store, hooks: hooks}
}

func (r *ProcurementRepository) CreateRequest(ctx context.Context, params applicationprocurement.CreateRequestParams) (domainprocurement.Request, error) {
	amount, err := parseNumeric(params.EstimatedTotalAmount)
	if err != nil {
		return domainprocurement.Request{}, err
	}

	request, err := r.store.CreateProcurementRequest(ctx, sqlc.CreateProcurementRequestParams{
		OrganizationID:       params.OrganizationID,
		RequesterUserID:      params.RequesterUserID,
		Title:                params.Title,
		Description:          params.Description,
		Justification:        params.Justification,
		CurrencyCode:         params.CurrencyCode,
		EstimatedTotalAmount: amount,
	})
	if err != nil {
		return domainprocurement.Request{}, err
	}

	return mapProcurementRequest(request), nil
}

func (r *ProcurementRepository) GetRequest(ctx context.Context, organizationID, requestID uuid.UUID) (domainprocurement.Request, error) {
	request, err := r.store.GetProcurementRequestByID(ctx, sqlc.GetProcurementRequestByIDParams{
		OrganizationID: organizationID,
		ID:             requestID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainprocurement.Request{}, applicationprocurement.ErrProcurementRequestNotFound
		}

		return domainprocurement.Request{}, err
	}

	return mapProcurementRequest(request), nil
}

func (r *ProcurementRepository) ListRequests(ctx context.Context, organizationID uuid.UUID, status *domainprocurement.RequestStatus) ([]domainprocurement.Request, error) {
	params := sqlc.ListProcurementRequestsParams{
		OrganizationID: organizationID,
	}
	if status != nil {
		params.Status = sqlc.NullProcurementRequestStatus{
			ProcurementRequestStatus: sqlc.ProcurementRequestStatus(*status),
			Valid:                    true,
		}
	}

	rows, err := r.store.ListProcurementRequests(ctx, params)
	if err != nil {
		return nil, err
	}

	requests := make([]domainprocurement.Request, 0, len(rows))
	for _, row := range rows {
		requests = append(requests, mapProcurementRequest(row))
	}

	return requests, nil
}

func (r *ProcurementRepository) ListApprovalInbox(ctx context.Context, organizationID uuid.UUID) ([]domainprocurement.Request, error) {
	rows, err := r.store.ListApprovalInboxRequests(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	requests := make([]domainprocurement.Request, 0, len(rows))
	for _, row := range rows {
		requests = append(requests, mapProcurementRequest(row))
	}

	return requests, nil
}

func (r *ProcurementRepository) UpdateDraftRequest(ctx context.Context, params applicationprocurement.UpdateRequestParams) (domainprocurement.Request, error) {
	amount, err := parseNumeric(params.EstimatedTotalAmount)
	if err != nil {
		return domainprocurement.Request{}, err
	}

	request, err := r.store.UpdateDraftProcurementRequest(ctx, sqlc.UpdateDraftProcurementRequestParams{
		OrganizationID:       params.OrganizationID,
		ID:                   params.RequestID,
		Title:                params.Title,
		Description:          params.Description,
		Justification:        params.Justification,
		CurrencyCode:         params.CurrencyCode,
		EstimatedTotalAmount: amount,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainprocurement.Request{}, applicationprocurement.ErrProcurementRequestNotFound
		}

		return domainprocurement.Request{}, err
	}

	return mapProcurementRequest(request), nil
}

func (r *ProcurementRepository) SubmitRequest(ctx context.Context, organizationID, requestID, submittedByUserID uuid.UUID) (domainprocurement.Request, error) {
	request, err := r.store.SubmitProcurementRequest(ctx, sqlc.SubmitProcurementRequestParams{
		OrganizationID:    organizationID,
		ID:                requestID,
		SubmittedByUserID: &submittedByUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainprocurement.Request{}, applicationprocurement.ErrProcurementRequestNotFound
		}

		return domainprocurement.Request{}, err
	}

	return mapProcurementRequest(request), nil
}

func (r *ProcurementRepository) ApproveRequest(ctx context.Context, organizationID, requestID, approvedByUserID uuid.UUID, decisionComment *string) (domainprocurement.Request, error) {
	request, err := r.store.ApproveProcurementRequest(ctx, sqlc.ApproveProcurementRequestParams{
		OrganizationID:   organizationID,
		ID:               requestID,
		ApprovedByUserID: &approvedByUserID,
		DecisionComment:  decisionComment,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainprocurement.Request{}, applicationprocurement.ErrProcurementRequestNotFound
		}

		return domainprocurement.Request{}, err
	}

	return mapProcurementRequest(request), nil
}

func (r *ProcurementRepository) RejectRequest(ctx context.Context, organizationID, requestID, rejectedByUserID uuid.UUID, decisionComment *string) (domainprocurement.Request, error) {
	request, err := r.store.RejectProcurementRequest(ctx, sqlc.RejectProcurementRequestParams{
		OrganizationID:   organizationID,
		ID:               requestID,
		RejectedByUserID: &rejectedByUserID,
		DecisionComment:  decisionComment,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainprocurement.Request{}, applicationprocurement.ErrProcurementRequestNotFound
		}

		return domainprocurement.Request{}, err
	}

	return mapProcurementRequest(request), nil
}

func (r *ProcurementRepository) CancelRequest(ctx context.Context, organizationID, requestID, cancelledByUserID uuid.UUID) (domainprocurement.Request, error) {
	request, err := r.store.CancelProcurementRequest(ctx, sqlc.CancelProcurementRequestParams{
		OrganizationID:    organizationID,
		ID:                requestID,
		CancelledByUserID: &cancelledByUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainprocurement.Request{}, applicationprocurement.ErrProcurementRequestNotFound
		}

		return domainprocurement.Request{}, err
	}

	return mapProcurementRequest(request), nil
}

func (r *ProcurementRepository) CreateItem(ctx context.Context, params applicationprocurement.CreateItemParams) (domainprocurement.Item, error) {
	quantity, err := parseNumeric(&params.Quantity)
	if err != nil {
		return domainprocurement.Item{}, err
	}
	estimatedUnitPrice, err := parseNumeric(params.EstimatedUnitPrice)
	if err != nil {
		return domainprocurement.Item{}, err
	}
	neededByDate, err := parseDate(params.NeededByDate)
	if err != nil {
		return domainprocurement.Item{}, err
	}

	item, err := r.store.CreateProcurementRequestItem(ctx, sqlc.CreateProcurementRequestItemParams{
		OrganizationID:       params.OrganizationID,
		ProcurementRequestID: params.ProcurementRequestID,
		LineNumber:           params.LineNumber,
		ItemName:             params.ItemName,
		Description:          params.Description,
		Quantity:             quantity,
		Unit:                 params.Unit,
		EstimatedUnitPrice:   estimatedUnitPrice,
		NeededByDate:         neededByDate,
	})
	if err != nil {
		return domainprocurement.Item{}, err
	}

	return mapProcurementRequestItem(item), nil
}

func (r *ProcurementRepository) ListItems(ctx context.Context, organizationID, requestID uuid.UUID) ([]domainprocurement.Item, error) {
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

func (r *ProcurementRepository) UpdateItem(ctx context.Context, params applicationprocurement.UpdateItemParams) (domainprocurement.Item, error) {
	quantity, err := parseNumeric(&params.Quantity)
	if err != nil {
		return domainprocurement.Item{}, err
	}
	estimatedUnitPrice, err := parseNumeric(params.EstimatedUnitPrice)
	if err != nil {
		return domainprocurement.Item{}, err
	}
	neededByDate, err := parseDate(params.NeededByDate)
	if err != nil {
		return domainprocurement.Item{}, err
	}

	item, err := r.store.UpdateProcurementRequestItem(ctx, sqlc.UpdateProcurementRequestItemParams{
		OrganizationID:       params.OrganizationID,
		ProcurementRequestID: params.ProcurementRequestID,
		ID:                   params.ItemID,
		ItemName:             params.ItemName,
		Description:          params.Description,
		Quantity:             quantity,
		Unit:                 params.Unit,
		EstimatedUnitPrice:   estimatedUnitPrice,
		NeededByDate:         neededByDate,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainprocurement.Item{}, applicationprocurement.ErrProcurementItemNotFound
		}

		return domainprocurement.Item{}, err
	}

	return mapProcurementRequestItem(item), nil
}

func (r *ProcurementRepository) DeleteItem(ctx context.Context, organizationID, requestID, itemID uuid.UUID) error {
	rows, err := r.store.DeleteProcurementRequestItem(ctx, sqlc.DeleteProcurementRequestItemParams{
		OrganizationID:       organizationID,
		ProcurementRequestID: requestID,
		ID:                   itemID,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return applicationprocurement.ErrProcurementItemNotFound
	}

	return nil
}

func (r *ProcurementRepository) GetMembership(ctx context.Context, organizationID, userID uuid.UUID) (domainorganization.Membership, error) {
	membership, err := r.store.GetOrganizationMembership(ctx, sqlc.GetOrganizationMembershipParams{
		OrganizationID: organizationID,
		UserID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainorganization.Membership{}, applicationprocurement.ErrMembershipNotFound
		}

		return domainorganization.Membership{}, err
	}

	return mapMembership(membership), nil
}

func (r *ProcurementRepository) CreateActivityLog(ctx context.Context, params applicationactivitylog.CreateParams) (domainactivitylog.Entry, error) {
	return createActivityLog(ctx, r.store, params, r.hooks...)
}

func (r *ProcurementRepository) WithinTransaction(ctx context.Context, fn func(repo applicationprocurement.Repository) error) error {
	return r.store.InTx(ctx, func(txStore *database.Store) error {
		return fn(NewProcurementRepository(txStore, r.hooks...))
	})
}

func mapProcurementRequest(request sqlc.ProcurementRequest) domainprocurement.Request {
	return domainprocurement.Request{
		ID:                   request.ID,
		OrganizationID:       request.OrganizationID,
		RequesterUserID:      request.RequesterUserID,
		Title:                request.Title,
		Description:          request.Description,
		Justification:        request.Justification,
		Status:               domainprocurement.RequestStatus(request.Status),
		CurrencyCode:         request.CurrencyCode,
		EstimatedTotalAmount: optionalNumeric(request.EstimatedTotalAmount),
		SubmittedAt:          optionalTime(request.SubmittedAt),
		SubmittedByUserID:    request.SubmittedByUserID,
		ApprovedAt:           optionalTime(request.ApprovedAt),
		ApprovedByUserID:     request.ApprovedByUserID,
		RejectedAt:           optionalTime(request.RejectedAt),
		RejectedByUserID:     request.RejectedByUserID,
		DecisionComment:      request.DecisionComment,
		CancelledAt:          optionalTime(request.CancelledAt),
		CancelledByUserID:    request.CancelledByUserID,
		CreatedAt:            requiredTime(request.CreatedAt),
		UpdatedAt:            requiredTime(request.UpdatedAt),
	}
}

func mapProcurementRequestItem(item sqlc.ProcurementRequestItem) domainprocurement.Item {
	return domainprocurement.Item{
		ID:                   item.ID,
		OrganizationID:       item.OrganizationID,
		ProcurementRequestID: item.ProcurementRequestID,
		LineNumber:           item.LineNumber,
		ItemName:             item.ItemName,
		Description:          item.Description,
		Quantity:             requiredNumeric(item.Quantity),
		Unit:                 item.Unit,
		EstimatedUnitPrice:   optionalNumeric(item.EstimatedUnitPrice),
		NeededByDate:         optionalDate(item.NeededByDate),
		CreatedAt:            requiredTime(item.CreatedAt),
		UpdatedAt:            requiredTime(item.UpdatedAt),
	}
}
