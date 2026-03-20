package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	applicationprocurement "github.com/AbenezerWork/ProcureFlow/internal/application/procurement"
	domainprocurement "github.com/AbenezerWork/ProcureFlow/internal/domain/procurement"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProcurementService interface {
	CreateRequest(ctx context.Context, input applicationprocurement.CreateRequestInput) (domainprocurement.Request, error)
	ListRequests(ctx context.Context, input applicationprocurement.ListRequestsInput) ([]domainprocurement.Request, error)
	GetRequest(ctx context.Context, organizationID, requestID, currentUser uuid.UUID) (domainprocurement.Request, error)
	UpdateRequest(ctx context.Context, input applicationprocurement.UpdateRequestInput) (domainprocurement.Request, error)
	SubmitRequest(ctx context.Context, input applicationprocurement.SubmitRequestInput) (domainprocurement.Request, error)
	CancelRequest(ctx context.Context, input applicationprocurement.CancelRequestInput) (domainprocurement.Request, error)
	ListItems(ctx context.Context, organizationID, requestID, currentUser uuid.UUID) ([]domainprocurement.Item, error)
	CreateItem(ctx context.Context, input applicationprocurement.CreateItemInput) (domainprocurement.Item, error)
	UpdateItem(ctx context.Context, input applicationprocurement.UpdateItemInput) (domainprocurement.Item, error)
	DeleteItem(ctx context.Context, input applicationprocurement.DeleteItemInput) error
}

type ProcurementHandler struct {
	service ProcurementService
}

func NewProcurementHandler(service ProcurementService) *ProcurementHandler {
	return &ProcurementHandler{service: service}
}

type createProcurementRequestRequest struct {
	Title                string  `json:"title"`
	Description          *string `json:"description"`
	Justification        *string `json:"justification"`
	CurrencyCode         *string `json:"currency_code"`
	EstimatedTotalAmount *string `json:"estimated_total_amount"`
}

type updateProcurementRequestRequest struct {
	Title                *string `json:"title"`
	Description          *string `json:"description"`
	Justification        *string `json:"justification"`
	CurrencyCode         *string `json:"currency_code"`
	EstimatedTotalAmount *string `json:"estimated_total_amount"`
}

type createProcurementItemRequest struct {
	ItemName           string  `json:"item_name"`
	Description        *string `json:"description"`
	Quantity           string  `json:"quantity"`
	Unit               string  `json:"unit"`
	EstimatedUnitPrice *string `json:"estimated_unit_price"`
	NeededByDate       *string `json:"needed_by_date"`
}

type updateProcurementItemRequest struct {
	ItemName           *string `json:"item_name"`
	Description        *string `json:"description"`
	Quantity           *string `json:"quantity"`
	Unit               *string `json:"unit"`
	EstimatedUnitPrice *string `json:"estimated_unit_price"`
	NeededByDate       *string `json:"needed_by_date"`
}

func (h *ProcurementHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return
	}

	var request createProcurementRequestRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	procurementRequest, err := h.service.CreateRequest(r.Context(), applicationprocurement.CreateRequestInput{
		OrganizationID:       organizationID,
		CurrentUser:          currentUser,
		Title:                request.Title,
		Description:          request.Description,
		Justification:        request.Justification,
		CurrencyCode:         request.CurrencyCode,
		EstimatedTotalAmount: request.EstimatedTotalAmount,
	})
	if err != nil {
		writeProcurementError(w, err, "create procurement request")
		return
	}

	writeJSON(w, http.StatusCreated, procurementRequestResponse(procurementRequest))
}

func (h *ProcurementHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return
	}

	var status *domainprocurement.RequestStatus
	if raw := r.URL.Query().Get("status"); raw != "" {
		value := domainprocurement.RequestStatus(raw)
		status = &value
	}

	requests, err := h.service.ListRequests(r.Context(), applicationprocurement.ListRequestsInput{
		OrganizationID: organizationID,
		CurrentUser:    currentUser,
		Status:         status,
	})
	if err != nil {
		writeProcurementError(w, err, "list procurement requests")
		return
	}

	response := make([]map[string]any, 0, len(requests))
	for _, request := range requests {
		response = append(response, procurementRequestResponse(request))
	}

	writeJSON(w, http.StatusOK, map[string]any{"procurement_requests": response})
}

func (h *ProcurementHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, requestID, ok := authenticatedProcurementRequest(w, r)
	if !ok {
		return
	}

	procurementRequest, err := h.service.GetRequest(r.Context(), organizationID, requestID, currentUser)
	if err != nil {
		writeProcurementError(w, err, "get procurement request")
		return
	}

	writeJSON(w, http.StatusOK, procurementRequestResponse(procurementRequest))
}

func (h *ProcurementHandler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, requestID, ok := authenticatedProcurementRequest(w, r)
	if !ok {
		return
	}

	var request updateProcurementRequestRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	procurementRequest, err := h.service.UpdateRequest(r.Context(), applicationprocurement.UpdateRequestInput{
		OrganizationID:       organizationID,
		RequestID:            requestID,
		CurrentUser:          currentUser,
		Title:                request.Title,
		Description:          request.Description,
		Justification:        request.Justification,
		CurrencyCode:         request.CurrencyCode,
		EstimatedTotalAmount: request.EstimatedTotalAmount,
	})
	if err != nil {
		writeProcurementError(w, err, "update procurement request")
		return
	}

	writeJSON(w, http.StatusOK, procurementRequestResponse(procurementRequest))
}

func (h *ProcurementHandler) SubmitRequest(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, requestID, ok := authenticatedProcurementRequest(w, r)
	if !ok {
		return
	}

	procurementRequest, err := h.service.SubmitRequest(r.Context(), applicationprocurement.SubmitRequestInput{
		OrganizationID: organizationID,
		RequestID:      requestID,
		CurrentUser:    currentUser,
	})
	if err != nil {
		writeProcurementError(w, err, "submit procurement request")
		return
	}

	writeJSON(w, http.StatusOK, procurementRequestResponse(procurementRequest))
}

func (h *ProcurementHandler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, requestID, ok := authenticatedProcurementRequest(w, r)
	if !ok {
		return
	}

	procurementRequest, err := h.service.CancelRequest(r.Context(), applicationprocurement.CancelRequestInput{
		OrganizationID: organizationID,
		RequestID:      requestID,
		CurrentUser:    currentUser,
	})
	if err != nil {
		writeProcurementError(w, err, "cancel procurement request")
		return
	}

	writeJSON(w, http.StatusOK, procurementRequestResponse(procurementRequest))
}

func (h *ProcurementHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, requestID, ok := authenticatedProcurementRequest(w, r)
	if !ok {
		return
	}

	items, err := h.service.ListItems(r.Context(), organizationID, requestID, currentUser)
	if err != nil {
		writeProcurementError(w, err, "list procurement request items")
		return
	}

	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, procurementItemResponse(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (h *ProcurementHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, requestID, ok := authenticatedProcurementRequest(w, r)
	if !ok {
		return
	}

	var request createProcurementItemRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.CreateItem(r.Context(), applicationprocurement.CreateItemInput{
		OrganizationID:     organizationID,
		RequestID:          requestID,
		CurrentUser:        currentUser,
		ItemName:           request.ItemName,
		Description:        request.Description,
		Quantity:           request.Quantity,
		Unit:               request.Unit,
		EstimatedUnitPrice: request.EstimatedUnitPrice,
		NeededByDate:       request.NeededByDate,
	})
	if err != nil {
		writeProcurementError(w, err, "create procurement request item")
		return
	}

	writeJSON(w, http.StatusCreated, procurementItemResponse(item))
}

func (h *ProcurementHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, requestID, itemID, ok := authenticatedProcurementItemRequest(w, r)
	if !ok {
		return
	}

	var request updateProcurementItemRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.UpdateItem(r.Context(), applicationprocurement.UpdateItemInput{
		OrganizationID:     organizationID,
		RequestID:          requestID,
		ItemID:             itemID,
		CurrentUser:        currentUser,
		ItemName:           request.ItemName,
		Description:        request.Description,
		Quantity:           request.Quantity,
		Unit:               request.Unit,
		EstimatedUnitPrice: request.EstimatedUnitPrice,
		NeededByDate:       request.NeededByDate,
	})
	if err != nil {
		writeProcurementError(w, err, "update procurement request item")
		return
	}

	writeJSON(w, http.StatusOK, procurementItemResponse(item))
}

func (h *ProcurementHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, requestID, itemID, ok := authenticatedProcurementItemRequest(w, r)
	if !ok {
		return
	}

	if err := h.service.DeleteItem(r.Context(), applicationprocurement.DeleteItemInput{
		OrganizationID: organizationID,
		RequestID:      requestID,
		ItemID:         itemID,
		CurrentUser:    currentUser,
	}); err != nil {
		writeProcurementError(w, err, "delete procurement request item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func authenticatedProcurementRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	requestID, err := uuid.Parse(chi.URLParam(r, "requestID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request_id")
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	return currentUser, organizationID, requestID, true
}

func authenticatedProcurementItemRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	currentUser, organizationID, requestID, ok := authenticatedProcurementRequest(w, r)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	itemID, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid item_id")
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	return currentUser, organizationID, requestID, itemID, true
}

func writeProcurementError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, applicationprocurement.ErrInvalidProcurementRequest):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, applicationprocurement.ErrProcurementRequestNotFound),
		errors.Is(err, applicationprocurement.ErrProcurementItemNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, applicationprocurement.ErrForbiddenProcurement):
		writeError(w, http.StatusForbidden, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
}

func procurementRequestResponse(request domainprocurement.Request) map[string]any {
	response := map[string]any{
		"id":                     request.ID,
		"organization_id":        request.OrganizationID,
		"requester_user_id":      request.RequesterUserID,
		"title":                  request.Title,
		"description":            request.Description,
		"justification":          request.Justification,
		"status":                 string(request.Status),
		"currency_code":          request.CurrencyCode,
		"estimated_total_amount": request.EstimatedTotalAmount,
		"decision_comment":       request.DecisionComment,
		"created_at":             request.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":             request.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if request.SubmittedAt != nil {
		response["submitted_at"] = request.SubmittedAt.UTC().Format(time.RFC3339)
	}
	if request.SubmittedByUserID != nil {
		response["submitted_by_user_id"] = request.SubmittedByUserID
	}
	if request.ApprovedAt != nil {
		response["approved_at"] = request.ApprovedAt.UTC().Format(time.RFC3339)
	}
	if request.ApprovedByUserID != nil {
		response["approved_by_user_id"] = request.ApprovedByUserID
	}
	if request.RejectedAt != nil {
		response["rejected_at"] = request.RejectedAt.UTC().Format(time.RFC3339)
	}
	if request.RejectedByUserID != nil {
		response["rejected_by_user_id"] = request.RejectedByUserID
	}
	if request.CancelledAt != nil {
		response["cancelled_at"] = request.CancelledAt.UTC().Format(time.RFC3339)
	}
	if request.CancelledByUserID != nil {
		response["cancelled_by_user_id"] = request.CancelledByUserID
	}

	return response
}

func procurementItemResponse(item domainprocurement.Item) map[string]any {
	response := map[string]any{
		"id":                     item.ID,
		"organization_id":        item.OrganizationID,
		"procurement_request_id": item.ProcurementRequestID,
		"line_number":            item.LineNumber,
		"item_name":              item.ItemName,
		"description":            item.Description,
		"quantity":               item.Quantity,
		"unit":                   item.Unit,
		"estimated_unit_price":   item.EstimatedUnitPrice,
		"needed_by_date":         item.NeededByDate,
		"created_at":             item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":             item.UpdatedAt.UTC().Format(time.RFC3339),
	}

	return response
}
