package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	applicationquotation "github.com/AbenezerWork/ProcureFlow/internal/application/quotation"
	domainquotation "github.com/AbenezerWork/ProcureFlow/internal/domain/quotation"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type QuotationService interface {
	Create(ctx context.Context, input applicationquotation.CreateInput) (domainquotation.Quotation, error)
	List(ctx context.Context, input applicationquotation.ListInput) ([]domainquotation.Quotation, error)
	Get(ctx context.Context, organizationID, rfqID, quotationID, currentUser uuid.UUID) (domainquotation.Quotation, error)
	Update(ctx context.Context, input applicationquotation.UpdateInput) (domainquotation.Quotation, error)
	Submit(ctx context.Context, input applicationquotation.TransitionInput) (domainquotation.Quotation, error)
	Reject(ctx context.Context, input applicationquotation.RejectInput) (domainquotation.Quotation, error)
	ListItems(ctx context.Context, organizationID, rfqID, quotationID, currentUser uuid.UUID) ([]domainquotation.Item, error)
	UpdateItem(ctx context.Context, input applicationquotation.UpdateItemInput) (domainquotation.Item, error)
}

type QuotationHandler struct {
	service QuotationService
}

func NewQuotationHandler(service QuotationService) *QuotationHandler {
	return &QuotationHandler{service: service}
}

type createQuotationRequest struct {
	RFQVendorID  string  `json:"rfq_vendor_id"`
	CurrencyCode *string `json:"currency_code"`
	LeadTimeDays *int32  `json:"lead_time_days"`
	PaymentTerms *string `json:"payment_terms"`
	Notes        *string `json:"notes"`
}

type updateQuotationRequest struct {
	CurrencyCode *string `json:"currency_code"`
	LeadTimeDays *int32  `json:"lead_time_days"`
	PaymentTerms *string `json:"payment_terms"`
	Notes        *string `json:"notes"`
}

type rejectQuotationRequest struct {
	RejectionReason *string `json:"rejection_reason"`
}

type updateQuotationItemRequest struct {
	UnitPrice    *string `json:"unit_price"`
	DeliveryDays *int32  `json:"delivery_days"`
	Notes        *string `json:"notes"`
}

func (h *QuotationHandler) Create(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, ok := authenticatedRFQRequest(w, r)
	if !ok {
		return
	}

	var request createQuotationRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rfqVendorID, err := uuid.Parse(request.RFQVendorID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rfq_vendor_id")
		return
	}

	quotation, err := h.service.Create(r.Context(), applicationquotation.CreateInput{
		OrganizationID: organizationID,
		RFQID:          rfqID,
		RFQVendorID:    rfqVendorID,
		CurrentUser:    currentUser,
		CurrencyCode:   request.CurrencyCode,
		LeadTimeDays:   request.LeadTimeDays,
		PaymentTerms:   request.PaymentTerms,
		Notes:          request.Notes,
	})
	if err != nil {
		writeQuotationError(w, err, "create quotation")
		return
	}

	writeJSON(w, http.StatusCreated, quotationResponse(quotation))
}

func (h *QuotationHandler) List(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, ok := authenticatedRFQRequest(w, r)
	if !ok {
		return
	}

	quotations, err := h.service.List(r.Context(), applicationquotation.ListInput{
		OrganizationID: organizationID,
		RFQID:          rfqID,
		CurrentUser:    currentUser,
	})
	if err != nil {
		writeQuotationError(w, err, "list quotations")
		return
	}

	response := make([]map[string]any, 0, len(quotations))
	for _, quotation := range quotations {
		response = append(response, quotationResponse(quotation))
	}

	writeJSON(w, http.StatusOK, map[string]any{"quotations": response})
}

func (h *QuotationHandler) Get(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, quotationID, ok := authenticatedQuotationRequest(w, r)
	if !ok {
		return
	}

	quotation, err := h.service.Get(r.Context(), organizationID, rfqID, quotationID, currentUser)
	if err != nil {
		writeQuotationError(w, err, "get quotation")
		return
	}

	writeJSON(w, http.StatusOK, quotationResponse(quotation))
}

func (h *QuotationHandler) Update(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, quotationID, ok := authenticatedQuotationRequest(w, r)
	if !ok {
		return
	}

	var request updateQuotationRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	quotation, err := h.service.Update(r.Context(), applicationquotation.UpdateInput{
		OrganizationID: organizationID,
		RFQID:          rfqID,
		QuotationID:    quotationID,
		CurrentUser:    currentUser,
		CurrencyCode:   request.CurrencyCode,
		LeadTimeDays:   request.LeadTimeDays,
		PaymentTerms:   request.PaymentTerms,
		Notes:          request.Notes,
	})
	if err != nil {
		writeQuotationError(w, err, "update quotation")
		return
	}

	writeJSON(w, http.StatusOK, quotationResponse(quotation))
}

func (h *QuotationHandler) Submit(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, quotationID, ok := authenticatedQuotationRequest(w, r)
	if !ok {
		return
	}

	quotation, err := h.service.Submit(r.Context(), applicationquotation.TransitionInput{
		OrganizationID: organizationID,
		RFQID:          rfqID,
		QuotationID:    quotationID,
		CurrentUser:    currentUser,
	})
	if err != nil {
		writeQuotationError(w, err, "submit quotation")
		return
	}

	writeJSON(w, http.StatusOK, quotationResponse(quotation))
}

func (h *QuotationHandler) Reject(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, quotationID, ok := authenticatedQuotationRequest(w, r)
	if !ok {
		return
	}

	var request rejectQuotationRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	quotation, err := h.service.Reject(r.Context(), applicationquotation.RejectInput{
		OrganizationID:  organizationID,
		RFQID:           rfqID,
		QuotationID:     quotationID,
		CurrentUser:     currentUser,
		RejectionReason: request.RejectionReason,
	})
	if err != nil {
		writeQuotationError(w, err, "reject quotation")
		return
	}

	writeJSON(w, http.StatusOK, quotationResponse(quotation))
}

func (h *QuotationHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, quotationID, ok := authenticatedQuotationRequest(w, r)
	if !ok {
		return
	}

	items, err := h.service.ListItems(r.Context(), organizationID, rfqID, quotationID, currentUser)
	if err != nil {
		writeQuotationError(w, err, "list quotation items")
		return
	}

	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, quotationItemResponse(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (h *QuotationHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, quotationID, itemID, ok := authenticatedQuotationItemRequest(w, r)
	if !ok {
		return
	}

	var request updateQuotationItemRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.UpdateItem(r.Context(), applicationquotation.UpdateItemInput{
		OrganizationID: organizationID,
		RFQID:          rfqID,
		QuotationID:    quotationID,
		ItemID:         itemID,
		CurrentUser:    currentUser,
		UnitPrice:      request.UnitPrice,
		DeliveryDays:   request.DeliveryDays,
		Notes:          request.Notes,
	})
	if err != nil {
		writeQuotationError(w, err, "update quotation item")
		return
	}

	writeJSON(w, http.StatusOK, quotationItemResponse(item))
}

func authenticatedQuotationRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	currentUser, organizationID, rfqID, ok := authenticatedRFQRequest(w, r)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	quotationID, err := uuid.Parse(chi.URLParam(r, "quotationID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid quotation_id")
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	return currentUser, organizationID, rfqID, quotationID, true
}

func authenticatedQuotationItemRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	currentUser, organizationID, rfqID, quotationID, ok := authenticatedQuotationRequest(w, r)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	itemID, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid item_id")
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	return currentUser, organizationID, rfqID, quotationID, itemID, true
}

func writeQuotationError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, applicationquotation.ErrInvalidQuotation):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, applicationquotation.ErrQuotationExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, applicationquotation.ErrQuotationNotFound),
		errors.Is(err, applicationquotation.ErrQuotationItemNotFound),
		errors.Is(err, applicationquotation.ErrRFQNotFound),
		errors.Is(err, applicationquotation.ErrRFQVendorNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, applicationquotation.ErrForbiddenQuotation):
		writeError(w, http.StatusForbidden, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
}

func quotationResponse(quotation domainquotation.Quotation) map[string]any {
	response := map[string]any{
		"id":                 quotation.ID,
		"organization_id":    quotation.OrganizationID,
		"rfq_id":             quotation.RFQID,
		"rfq_vendor_id":      quotation.RFQVendorID,
		"status":             string(quotation.Status),
		"currency_code":      quotation.CurrencyCode,
		"lead_time_days":     quotation.LeadTimeDays,
		"payment_terms":      quotation.PaymentTerms,
		"notes":              quotation.Notes,
		"created_by_user_id": quotation.CreatedByUserID,
		"updated_by_user_id": quotation.UpdatedByUserID,
		"created_at":         quotation.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":         quotation.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if quotation.VendorName != nil {
		response["vendor_name"] = *quotation.VendorName
	}
	if quotation.SubmittedAt != nil {
		response["submitted_at"] = quotation.SubmittedAt.UTC().Format(time.RFC3339)
	}
	if quotation.SubmittedByUserID != nil {
		response["submitted_by_user_id"] = quotation.SubmittedByUserID
	}
	if quotation.RejectedAt != nil {
		response["rejected_at"] = quotation.RejectedAt.UTC().Format(time.RFC3339)
	}
	if quotation.RejectedByUserID != nil {
		response["rejected_by_user_id"] = quotation.RejectedByUserID
	}
	if quotation.RejectionReason != nil {
		response["rejection_reason"] = quotation.RejectionReason
	}

	return response
}

func quotationItemResponse(item domainquotation.Item) map[string]any {
	return map[string]any{
		"id":              item.ID,
		"organization_id": item.OrganizationID,
		"quotation_id":    item.QuotationID,
		"rfq_id":          item.RFQID,
		"rfq_item_id":     item.RFQItemID,
		"line_number":     item.LineNumber,
		"item_name":       item.ItemName,
		"description":     item.Description,
		"quantity":        item.Quantity,
		"unit":            item.Unit,
		"unit_price":      item.UnitPrice,
		"delivery_days":   item.DeliveryDays,
		"notes":           item.Notes,
		"created_at":      item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":      item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
