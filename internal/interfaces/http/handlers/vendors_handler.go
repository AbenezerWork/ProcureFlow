package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	applicationvendor "github.com/AbenezerWork/ProcureFlow/internal/application/vendor"
	domainvendor "github.com/AbenezerWork/ProcureFlow/internal/domain/vendor"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type VendorService interface {
	Create(ctx context.Context, input applicationvendor.CreateInput) (domainvendor.Vendor, error)
	List(ctx context.Context, input applicationvendor.ListInput) ([]domainvendor.Vendor, error)
	Get(ctx context.Context, organizationID, vendorID, currentUser uuid.UUID) (domainvendor.Vendor, error)
	Update(ctx context.Context, input applicationvendor.UpdateInput) (domainvendor.Vendor, error)
	Archive(ctx context.Context, input applicationvendor.ArchiveInput) (domainvendor.Vendor, error)
}

type VendorHandler struct {
	service VendorService
}

func NewVendorHandler(service VendorService) *VendorHandler {
	return &VendorHandler{service: service}
}

type createVendorRequest struct {
	Name          string  `json:"name"`
	LegalName     *string `json:"legal_name"`
	ContactName   *string `json:"contact_name"`
	Email         *string `json:"email"`
	Phone         *string `json:"phone"`
	TaxIdentifier *string `json:"tax_identifier"`
	AddressLine1  *string `json:"address_line1"`
	AddressLine2  *string `json:"address_line2"`
	City          *string `json:"city"`
	StateRegion   *string `json:"state_region"`
	PostalCode    *string `json:"postal_code"`
	Country       *string `json:"country"`
	Notes         *string `json:"notes"`
}

type updateVendorRequest struct {
	Name          *string `json:"name"`
	LegalName     *string `json:"legal_name"`
	ContactName   *string `json:"contact_name"`
	Email         *string `json:"email"`
	Phone         *string `json:"phone"`
	TaxIdentifier *string `json:"tax_identifier"`
	AddressLine1  *string `json:"address_line1"`
	AddressLine2  *string `json:"address_line2"`
	City          *string `json:"city"`
	StateRegion   *string `json:"state_region"`
	PostalCode    *string `json:"postal_code"`
	Country       *string `json:"country"`
	Notes         *string `json:"notes"`
}

func (h *VendorHandler) Create(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return
	}

	var request createVendorRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	vendor, err := h.service.Create(r.Context(), applicationvendor.CreateInput{
		OrganizationID: organizationID,
		CurrentUser:    currentUser,
		Name:           request.Name,
		LegalName:      request.LegalName,
		ContactName:    request.ContactName,
		Email:          request.Email,
		Phone:          request.Phone,
		TaxIdentifier:  request.TaxIdentifier,
		AddressLine1:   request.AddressLine1,
		AddressLine2:   request.AddressLine2,
		City:           request.City,
		StateRegion:    request.StateRegion,
		PostalCode:     request.PostalCode,
		Country:        request.Country,
		Notes:          request.Notes,
	})
	if err != nil {
		writeVendorError(w, err, "create vendor")
		return
	}

	writeJSON(w, http.StatusCreated, vendorResponse(vendor))
}

func (h *VendorHandler) List(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return
	}

	var status *domainvendor.Status
	if raw := r.URL.Query().Get("status"); raw != "" {
		value := domainvendor.Status(raw)
		status = &value
	}

	vendors, err := h.service.List(r.Context(), applicationvendor.ListInput{
		OrganizationID: organizationID,
		CurrentUser:    currentUser,
		Status:         status,
		Search:         r.URL.Query().Get("search"),
	})
	if err != nil {
		writeVendorError(w, err, "list vendors")
		return
	}

	response := make([]map[string]any, 0, len(vendors))
	for _, vendor := range vendors {
		response = append(response, vendorResponse(vendor))
	}

	writeJSON(w, http.StatusOK, map[string]any{"vendors": response})
}

func (h *VendorHandler) Get(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, vendorID, ok := authenticatedVendorRequest(w, r)
	if !ok {
		return
	}

	vendor, err := h.service.Get(r.Context(), organizationID, vendorID, currentUser)
	if err != nil {
		writeVendorError(w, err, "get vendor")
		return
	}

	writeJSON(w, http.StatusOK, vendorResponse(vendor))
}

func (h *VendorHandler) Update(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, vendorID, ok := authenticatedVendorRequest(w, r)
	if !ok {
		return
	}

	var request updateVendorRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	vendor, err := h.service.Update(r.Context(), applicationvendor.UpdateInput{
		OrganizationID: organizationID,
		VendorID:       vendorID,
		CurrentUser:    currentUser,
		Name:           request.Name,
		LegalName:      request.LegalName,
		ContactName:    request.ContactName,
		Email:          request.Email,
		Phone:          request.Phone,
		TaxIdentifier:  request.TaxIdentifier,
		AddressLine1:   request.AddressLine1,
		AddressLine2:   request.AddressLine2,
		City:           request.City,
		StateRegion:    request.StateRegion,
		PostalCode:     request.PostalCode,
		Country:        request.Country,
		Notes:          request.Notes,
	})
	if err != nil {
		writeVendorError(w, err, "update vendor")
		return
	}

	writeJSON(w, http.StatusOK, vendorResponse(vendor))
}

func (h *VendorHandler) Archive(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, vendorID, ok := authenticatedVendorRequest(w, r)
	if !ok {
		return
	}

	vendor, err := h.service.Archive(r.Context(), applicationvendor.ArchiveInput{
		OrganizationID: organizationID,
		VendorID:       vendorID,
		CurrentUser:    currentUser,
	})
	if err != nil {
		writeVendorError(w, err, "archive vendor")
		return
	}

	writeJSON(w, http.StatusOK, vendorResponse(vendor))
}

func authenticatedVendorRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	vendorID, err := uuid.Parse(chi.URLParam(r, "vendorID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vendor_id")
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	return currentUser, organizationID, vendorID, true
}

func writeVendorError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, applicationvendor.ErrInvalidVendor):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, applicationvendor.ErrVendorNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, applicationvendor.ErrForbiddenVendor):
		writeError(w, http.StatusForbidden, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
}

func vendorResponse(vendor domainvendor.Vendor) map[string]any {
	response := map[string]any{
		"id":                 vendor.ID,
		"organization_id":    vendor.OrganizationID,
		"name":               vendor.Name,
		"legal_name":         vendor.LegalName,
		"contact_name":       vendor.ContactName,
		"email":              vendor.Email,
		"phone":              vendor.Phone,
		"tax_identifier":     vendor.TaxIdentifier,
		"address_line1":      vendor.AddressLine1,
		"address_line2":      vendor.AddressLine2,
		"city":               vendor.City,
		"state_region":       vendor.StateRegion,
		"postal_code":        vendor.PostalCode,
		"country":            vendor.Country,
		"notes":              vendor.Notes,
		"status":             string(vendor.Status),
		"created_by_user_id": vendor.CreatedByUserID,
		"created_at":         vendor.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":         vendor.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if vendor.ArchivedAt != nil {
		response["archived_at"] = vendor.ArchivedAt.UTC().Format(time.RFC3339)
	}

	if vendor.UpdatedByUserID != nil {
		response["updated_by_user_id"] = vendor.UpdatedByUserID
	}

	return response
}
