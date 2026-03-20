package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	applicationrfq "github.com/AbenezerWork/ProcureFlow/internal/application/rfq"
	domainrfq "github.com/AbenezerWork/ProcureFlow/internal/domain/rfq"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RFQService interface {
	Create(ctx context.Context, input applicationrfq.CreateInput) (domainrfq.RFQ, error)
	List(ctx context.Context, input applicationrfq.ListInput) ([]domainrfq.RFQ, error)
	Get(ctx context.Context, organizationID, rfqID, currentUser uuid.UUID) (domainrfq.RFQ, error)
	Update(ctx context.Context, input applicationrfq.UpdateInput) (domainrfq.RFQ, error)
	Publish(ctx context.Context, input applicationrfq.TransitionInput) (domainrfq.RFQ, error)
	Close(ctx context.Context, input applicationrfq.TransitionInput) (domainrfq.RFQ, error)
	Evaluate(ctx context.Context, input applicationrfq.TransitionInput) (domainrfq.RFQ, error)
	Cancel(ctx context.Context, input applicationrfq.TransitionInput) (domainrfq.RFQ, error)
	ListItems(ctx context.Context, organizationID, rfqID, currentUser uuid.UUID) ([]domainrfq.Item, error)
	ListVendors(ctx context.Context, organizationID, rfqID, currentUser uuid.UUID) ([]domainrfq.VendorLink, error)
	AttachVendor(ctx context.Context, input applicationrfq.AttachVendorInput) (domainrfq.VendorLink, error)
	RemoveVendor(ctx context.Context, input applicationrfq.RemoveVendorInput) error
}

type RFQHandler struct {
	service RFQService
}

func NewRFQHandler(service RFQService) *RFQHandler {
	return &RFQHandler{service: service}
}

type createRFQRequest struct {
	ProcurementRequestID string  `json:"procurement_request_id"`
	ReferenceNumber      *string `json:"reference_number"`
	Title                *string `json:"title"`
	Description          *string `json:"description"`
}

type updateRFQRequest struct {
	ReferenceNumber *string `json:"reference_number"`
	Title           *string `json:"title"`
	Description     *string `json:"description"`
}

type attachRFQVendorRequest struct {
	VendorID string `json:"vendor_id"`
}

func (h *RFQHandler) Create(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return
	}

	var request createRFQRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	procurementRequestID, err := uuid.Parse(request.ProcurementRequestID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid procurement_request_id")
		return
	}

	rfq, err := h.service.Create(r.Context(), applicationrfq.CreateInput{
		OrganizationID:       organizationID,
		ProcurementRequestID: procurementRequestID,
		CurrentUser:          currentUser,
		ReferenceNumber:      request.ReferenceNumber,
		Title:                request.Title,
		Description:          request.Description,
	})
	if err != nil {
		writeRFQError(w, err, "create rfq")
		return
	}

	writeJSON(w, http.StatusCreated, rfqResponse(rfq))
}

func (h *RFQHandler) List(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return
	}

	var status *domainrfq.Status
	if raw := r.URL.Query().Get("status"); raw != "" {
		value := domainrfq.Status(raw)
		status = &value
	}

	rfqs, err := h.service.List(r.Context(), applicationrfq.ListInput{
		OrganizationID: organizationID,
		CurrentUser:    currentUser,
		Status:         status,
	})
	if err != nil {
		writeRFQError(w, err, "list rfqs")
		return
	}

	response := make([]map[string]any, 0, len(rfqs))
	for _, rfq := range rfqs {
		response = append(response, rfqResponse(rfq))
	}

	writeJSON(w, http.StatusOK, map[string]any{"rfqs": response})
}

func (h *RFQHandler) Get(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, ok := authenticatedRFQRequest(w, r)
	if !ok {
		return
	}

	rfq, err := h.service.Get(r.Context(), organizationID, rfqID, currentUser)
	if err != nil {
		writeRFQError(w, err, "get rfq")
		return
	}

	writeJSON(w, http.StatusOK, rfqResponse(rfq))
}

func (h *RFQHandler) Update(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, ok := authenticatedRFQRequest(w, r)
	if !ok {
		return
	}

	var request updateRFQRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rfq, err := h.service.Update(r.Context(), applicationrfq.UpdateInput{
		OrganizationID:  organizationID,
		RFQID:           rfqID,
		CurrentUser:     currentUser,
		ReferenceNumber: request.ReferenceNumber,
		Title:           request.Title,
		Description:     request.Description,
	})
	if err != nil {
		writeRFQError(w, err, "update rfq")
		return
	}

	writeJSON(w, http.StatusOK, rfqResponse(rfq))
}

func (h *RFQHandler) Publish(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, h.service.Publish, "publish rfq")
}

func (h *RFQHandler) Close(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, h.service.Close, "close rfq")
}

func (h *RFQHandler) Evaluate(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, h.service.Evaluate, "evaluate rfq")
}

func (h *RFQHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, h.service.Cancel, "cancel rfq")
}

func (h *RFQHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, ok := authenticatedRFQRequest(w, r)
	if !ok {
		return
	}

	items, err := h.service.ListItems(r.Context(), organizationID, rfqID, currentUser)
	if err != nil {
		writeRFQError(w, err, "list rfq items")
		return
	}

	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, rfqItemResponse(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (h *RFQHandler) ListVendors(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, ok := authenticatedRFQRequest(w, r)
	if !ok {
		return
	}

	vendors, err := h.service.ListVendors(r.Context(), organizationID, rfqID, currentUser)
	if err != nil {
		writeRFQError(w, err, "list rfq vendors")
		return
	}

	response := make([]map[string]any, 0, len(vendors))
	for _, vendor := range vendors {
		response = append(response, rfqVendorResponse(vendor))
	}

	writeJSON(w, http.StatusOK, map[string]any{"vendors": response})
}

func (h *RFQHandler) AttachVendor(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, ok := authenticatedRFQRequest(w, r)
	if !ok {
		return
	}

	var request attachRFQVendorRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	vendorID, err := uuid.Parse(request.VendorID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vendor_id")
		return
	}

	vendor, err := h.service.AttachVendor(r.Context(), applicationrfq.AttachVendorInput{
		OrganizationID: organizationID,
		RFQID:          rfqID,
		VendorID:       vendorID,
		CurrentUser:    currentUser,
	})
	if err != nil {
		writeRFQError(w, err, "attach rfq vendor")
		return
	}

	writeJSON(w, http.StatusCreated, rfqVendorResponse(vendor))
}

func (h *RFQHandler) RemoveVendor(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, vendorID, ok := authenticatedRFQVendorRequest(w, r)
	if !ok {
		return
	}

	if err := h.service.RemoveVendor(r.Context(), applicationrfq.RemoveVendorInput{
		OrganizationID: organizationID,
		RFQID:          rfqID,
		VendorID:       vendorID,
		CurrentUser:    currentUser,
	}); err != nil {
		writeRFQError(w, err, "remove rfq vendor")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RFQHandler) transition(w http.ResponseWriter, r *http.Request, fn func(context.Context, applicationrfq.TransitionInput) (domainrfq.RFQ, error), fallback string) {
	currentUser, organizationID, rfqID, ok := authenticatedRFQRequest(w, r)
	if !ok {
		return
	}

	rfq, err := fn(r.Context(), applicationrfq.TransitionInput{
		OrganizationID: organizationID,
		RFQID:          rfqID,
		CurrentUser:    currentUser,
	})
	if err != nil {
		writeRFQError(w, err, fallback)
		return
	}

	writeJSON(w, http.StatusOK, rfqResponse(rfq))
}

func authenticatedRFQRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	rfqID, err := uuid.Parse(chi.URLParam(r, "rfqID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rfq_id")
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	return currentUser, organizationID, rfqID, true
}

func authenticatedRFQVendorRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	currentUser, organizationID, rfqID, ok := authenticatedRFQRequest(w, r)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	vendorID, err := uuid.Parse(chi.URLParam(r, "vendorID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vendor_id")
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	return currentUser, organizationID, rfqID, vendorID, true
}

func writeRFQError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, applicationrfq.ErrInvalidRFQ):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, applicationrfq.ErrRFQReferenceTaken),
		errors.Is(err, applicationrfq.ErrRFQVendorAlreadyAttached):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, applicationrfq.ErrRFQNotFound),
		errors.Is(err, applicationrfq.ErrRFQVendorNotFound),
		errors.Is(err, applicationrfq.ErrProcurementRequestNotFound),
		errors.Is(err, applicationrfq.ErrVendorNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, applicationrfq.ErrForbiddenRFQ):
		writeError(w, http.StatusForbidden, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
}

func rfqResponse(rfq domainrfq.RFQ) map[string]any {
	response := map[string]any{
		"id":                     rfq.ID,
		"organization_id":        rfq.OrganizationID,
		"procurement_request_id": rfq.ProcurementRequestID,
		"reference_number":       rfq.ReferenceNumber,
		"title":                  rfq.Title,
		"description":            rfq.Description,
		"status":                 string(rfq.Status),
		"created_by_user_id":     rfq.CreatedByUserID,
		"created_at":             rfq.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":             rfq.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if rfq.PublishedAt != nil {
		response["published_at"] = rfq.PublishedAt.UTC().Format(time.RFC3339)
	}
	if rfq.ClosedAt != nil {
		response["closed_at"] = rfq.ClosedAt.UTC().Format(time.RFC3339)
	}
	if rfq.EvaluatedAt != nil {
		response["evaluated_at"] = rfq.EvaluatedAt.UTC().Format(time.RFC3339)
	}
	if rfq.CancelledAt != nil {
		response["cancelled_at"] = rfq.CancelledAt.UTC().Format(time.RFC3339)
	}
	if rfq.CancelledByUserID != nil {
		response["cancelled_by_user_id"] = rfq.CancelledByUserID
	}

	return response
}

func rfqItemResponse(item domainrfq.Item) map[string]any {
	return map[string]any{
		"id":                     item.ID,
		"organization_id":        item.OrganizationID,
		"rfq_id":                 item.RFQID,
		"source_request_item_id": item.SourceRequestItemID,
		"line_number":            item.LineNumber,
		"item_name":              item.ItemName,
		"description":            item.Description,
		"quantity":               item.Quantity,
		"unit":                   item.Unit,
		"target_date":            item.TargetDate,
		"created_at":             item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":             item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func rfqVendorResponse(vendor domainrfq.VendorLink) map[string]any {
	return map[string]any{
		"id":                  vendor.ID,
		"organization_id":     vendor.OrganizationID,
		"rfq_id":              vendor.RFQID,
		"vendor_id":           vendor.VendorID,
		"attached_by_user_id": vendor.AttachedByUserID,
		"attached_at":         vendor.AttachedAt.UTC().Format(time.RFC3339),
		"created_at":          vendor.CreatedAt.UTC().Format(time.RFC3339),
		"vendor_name":         vendor.VendorName,
		"vendor_status":       string(vendor.VendorStatus),
	}
}
