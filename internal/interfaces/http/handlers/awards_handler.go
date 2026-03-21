package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	applicationaward "github.com/AbenezerWork/ProcureFlow/internal/application/award"
	domainaward "github.com/AbenezerWork/ProcureFlow/internal/domain/award"
	"github.com/google/uuid"
)

type AwardService interface {
	Create(ctx context.Context, input applicationaward.CreateInput) (domainaward.Award, error)
	GetByRFQ(ctx context.Context, organizationID, rfqID, currentUser uuid.UUID) (domainaward.Award, error)
}

type AwardHandler struct {
	service AwardService
}

func NewAwardHandler(service AwardService) *AwardHandler {
	return &AwardHandler{service: service}
}

type createAwardRequest struct {
	QuotationID string `json:"quotation_id"`
	Reason      string `json:"reason"`
}

func (h *AwardHandler) Create(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, ok := authenticatedRFQRequest(w, r)
	if !ok {
		return
	}

	var request createAwardRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	quotationID, err := uuid.Parse(request.QuotationID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid quotation_id")
		return
	}

	award, err := h.service.Create(r.Context(), applicationaward.CreateInput{
		OrganizationID: organizationID,
		RFQID:          rfqID,
		QuotationID:    quotationID,
		CurrentUser:    currentUser,
		Reason:         request.Reason,
	})
	if err != nil {
		writeAwardError(w, err, "create award")
		return
	}

	writeJSON(w, http.StatusCreated, awardResponse(award))
}

func (h *AwardHandler) Get(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, rfqID, ok := authenticatedRFQRequest(w, r)
	if !ok {
		return
	}

	award, err := h.service.GetByRFQ(r.Context(), organizationID, rfqID, currentUser)
	if err != nil {
		writeAwardError(w, err, "get award")
		return
	}

	writeJSON(w, http.StatusOK, awardResponse(award))
}

func writeAwardError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, applicationaward.ErrInvalidAward):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, applicationaward.ErrAwardExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, applicationaward.ErrAwardNotFound),
		errors.Is(err, applicationaward.ErrRFQNotFound),
		errors.Is(err, applicationaward.ErrQuotationNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, applicationaward.ErrForbiddenAward):
		writeError(w, http.StatusForbidden, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
}

func awardResponse(award domainaward.Award) map[string]any {
	return map[string]any{
		"id":                 award.ID,
		"organization_id":    award.OrganizationID,
		"rfq_id":             award.RFQID,
		"quotation_id":       award.QuotationID,
		"awarded_by_user_id": award.AwardedByUserID,
		"reason":             award.Reason,
		"awarded_at":         award.AwardedAt.UTC().Format(time.RFC3339),
		"created_at":         award.CreatedAt.UTC().Format(time.RFC3339),
	}
}
