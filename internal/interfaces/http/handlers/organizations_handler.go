package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	applicationorganization "github.com/AbenezerWork/ProcureFlow/internal/application/organization"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	httpmiddleware "github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/middleware"
	"github.com/google/uuid"
)

type OrganizationService interface {
	Create(ctx context.Context, input applicationorganization.CreateInput) (applicationorganization.CreatedOrganization, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]domainorganization.UserOrganization, error)
}

type OrganizationHandler struct {
	service OrganizationService
}

func NewOrganizationHandler(service OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{service: service}
}

type createOrganizationRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *OrganizationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpmiddleware.AuthenticatedUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var request createOrganizationRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.service.Create(r.Context(), applicationorganization.CreateInput{
		Name:        request.Name,
		Slug:        request.Slug,
		CurrentUser: userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, applicationorganization.ErrInvalidOrganization):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, applicationorganization.ErrOrganizationSlugTaken):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "create organization")
		}
		return
	}

	writeJSON(w, http.StatusCreated, organizationWithMembershipResponse(created))
}

func (h *OrganizationHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpmiddleware.AuthenticatedUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	organizations, err := h.service.ListForUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, applicationorganization.ErrInvalidOrganization) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "list organizations")
		return
	}

	response := make([]map[string]any, 0, len(organizations))
	for _, organization := range organizations {
		response = append(response, userOrganizationResponse(organization))
	}

	writeJSON(w, http.StatusOK, map[string]any{"organizations": response})
}

func organizationWithMembershipResponse(created applicationorganization.CreatedOrganization) map[string]any {
	return map[string]any{
		"organization": organizationResponse(created.Organization),
		"membership":   membershipResponse(created.Membership),
	}
}

func userOrganizationResponse(userOrganization domainorganization.UserOrganization) map[string]any {
	return map[string]any{
		"organization": organizationResponse(userOrganization.Organization),
		"role":         string(userOrganization.Role),
		"status":       string(userOrganization.Status),
	}
}

func organizationResponse(organization domainorganization.Organization) map[string]any {
	response := map[string]any{
		"id":                 organization.ID,
		"name":               organization.Name,
		"slug":               organization.Slug,
		"created_by_user_id": organization.CreatedByUserID,
		"created_at":         organization.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":         organization.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if organization.ArchivedAt != nil {
		response["archived_at"] = organization.ArchivedAt.UTC().Format(time.RFC3339)
	}

	return response
}

func membershipResponse(membership domainorganization.Membership) map[string]any {
	response := map[string]any{
		"id":                 membership.ID,
		"organization_id":    membership.OrganizationID,
		"user_id":            membership.UserID,
		"role":               string(membership.Role),
		"status":             string(membership.Status),
		"created_by_user_id": membership.CreatedByUserID,
		"invited_at":         membership.InvitedAt.UTC().Format(time.RFC3339),
		"created_at":         membership.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":         membership.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if membership.ActivatedAt != nil {
		response["activated_at"] = membership.ActivatedAt.UTC().Format(time.RFC3339)
	}

	if membership.SuspendedAt != nil {
		response["suspended_at"] = membership.SuspendedAt.UTC().Format(time.RFC3339)
	}

	if membership.RemovedAt != nil {
		response["removed_at"] = membership.RemovedAt.UTC().Format(time.RFC3339)
	}

	return response
}
