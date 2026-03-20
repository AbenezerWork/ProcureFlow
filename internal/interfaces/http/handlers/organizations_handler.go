package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	applicationorganization "github.com/AbenezerWork/ProcureFlow/internal/application/organization"
	domainidentity "github.com/AbenezerWork/ProcureFlow/internal/domain/identity"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	httpmiddleware "github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrganizationService interface {
	Create(ctx context.Context, input applicationorganization.CreateInput) (applicationorganization.CreatedOrganization, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]domainorganization.UserOrganization, error)
	Get(ctx context.Context, organizationID, currentUser uuid.UUID) (applicationorganization.OrganizationDetails, error)
	Update(ctx context.Context, input applicationorganization.UpdateInput) (applicationorganization.OrganizationDetails, error)
	ListMemberships(ctx context.Context, organizationID, currentUser uuid.UUID) ([]applicationorganization.OrganizationMember, error)
	AddMembership(ctx context.Context, input applicationorganization.AddMembershipInput) (applicationorganization.OrganizationMember, error)
	UpdateMembership(ctx context.Context, input applicationorganization.UpdateMembershipInput) (applicationorganization.OrganizationMember, error)
	TransferOwnership(ctx context.Context, input applicationorganization.TransferOwnershipInput) (applicationorganization.OwnershipTransferResult, error)
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

type updateOrganizationRequest struct {
	Name *string `json:"name"`
	Slug *string `json:"slug"`
}

type createMembershipRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

type updateMembershipRequest struct {
	Role   *string `json:"role"`
	Status *string `json:"status"`
}

type transferOwnershipRequest struct {
	TargetUserID        string `json:"target_user_id"`
	CurrentOwnerNewRole string `json:"current_owner_new_role"`
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

func (h *OrganizationHandler) Get(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return
	}

	details, err := h.service.Get(r.Context(), organizationID, currentUser)
	if err != nil {
		writeOrganizationError(w, err, "get organization")
		return
	}

	writeJSON(w, http.StatusOK, organizationDetailsResponse(details))
}

func (h *OrganizationHandler) Update(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return
	}

	var request updateOrganizationRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	details, err := h.service.Update(r.Context(), applicationorganization.UpdateInput{
		OrganizationID: organizationID,
		CurrentUser:    currentUser,
		Name:           request.Name,
		Slug:           request.Slug,
	})
	if err != nil {
		writeOrganizationError(w, err, "update organization")
		return
	}

	writeJSON(w, http.StatusOK, organizationDetailsResponse(details))
}

func (h *OrganizationHandler) ListMemberships(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return
	}

	members, err := h.service.ListMemberships(r.Context(), organizationID, currentUser)
	if err != nil {
		writeOrganizationError(w, err, "list organization memberships")
		return
	}

	response := make([]map[string]any, 0, len(members))
	for _, member := range members {
		response = append(response, organizationMemberResponse(member))
	}

	writeJSON(w, http.StatusOK, map[string]any{"memberships": response})
}

func (h *OrganizationHandler) CreateMembership(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return
	}

	var request createMembershipRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var targetUserID *uuid.UUID
	if request.UserID != "" {
		parsed, err := uuid.Parse(request.UserID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid user_id")
			return
		}

		targetUserID = &parsed
	}

	member, err := h.service.AddMembership(r.Context(), applicationorganization.AddMembershipInput{
		OrganizationID: organizationID,
		CurrentUser:    currentUser,
		UserID:         targetUserID,
		Email:          request.Email,
		Role:           domainorganization.MembershipRole(request.Role),
		Status:         domainorganization.MembershipStatus(request.Status),
	})
	if err != nil {
		writeOrganizationError(w, err, "create organization membership")
		return
	}

	writeJSON(w, http.StatusCreated, organizationMemberResponse(member))
}

func (h *OrganizationHandler) UpdateMembership(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return
	}

	targetUserID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	var request updateMembershipRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var role *domainorganization.MembershipRole
	if request.Role != nil {
		value := domainorganization.MembershipRole(*request.Role)
		role = &value
	}

	var status *domainorganization.MembershipStatus
	if request.Status != nil {
		value := domainorganization.MembershipStatus(*request.Status)
		status = &value
	}

	member, err := h.service.UpdateMembership(r.Context(), applicationorganization.UpdateMembershipInput{
		OrganizationID: organizationID,
		CurrentUser:    currentUser,
		TargetUserID:   targetUserID,
		Role:           role,
		Status:         status,
	})
	if err != nil {
		writeOrganizationError(w, err, "update organization membership")
		return
	}

	writeJSON(w, http.StatusOK, organizationMemberResponse(member))
}

func (h *OrganizationHandler) TransferOwnership(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return
	}

	var request transferOwnershipRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	targetUserID, err := uuid.Parse(request.TargetUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid target_user_id")
		return
	}

	result, err := h.service.TransferOwnership(r.Context(), applicationorganization.TransferOwnershipInput{
		OrganizationID:      organizationID,
		CurrentUser:         currentUser,
		TargetUserID:        targetUserID,
		CurrentOwnerNewRole: domainorganization.MembershipRole(request.CurrentOwnerNewRole),
	})
	if err != nil {
		writeOrganizationError(w, err, "transfer organization ownership")
		return
	}

	writeJSON(w, http.StatusOK, ownershipTransferResponse(result))
}

func organizationWithMembershipResponse(created applicationorganization.CreatedOrganization) map[string]any {
	return map[string]any{
		"organization": organizationResponse(created.Organization),
		"membership":   membershipResponse(created.Membership),
	}
}

func organizationDetailsResponse(details applicationorganization.OrganizationDetails) map[string]any {
	return map[string]any{
		"organization": organizationResponse(details.Organization),
		"membership":   membershipResponse(details.Membership),
	}
}

func organizationMemberResponse(member applicationorganization.OrganizationMember) map[string]any {
	return map[string]any{
		"user":       userSummaryResponse(member.User),
		"membership": membershipResponse(member.Membership),
	}
}

func ownershipTransferResponse(result applicationorganization.OwnershipTransferResult) map[string]any {
	return map[string]any{
		"organization":   organizationResponse(result.Organization),
		"previous_owner": organizationMemberResponse(result.PreviousOwner),
		"new_owner":      organizationMemberResponse(result.NewOwner),
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

func userSummaryResponse(user domainidentity.User) map[string]any {
	return map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"full_name":  user.FullName,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at": user.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func authenticatedOrganizationRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	currentUser, ok := httpmiddleware.AuthenticatedUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return uuid.Nil, uuid.Nil, false
	}

	organizationID, err := uuid.Parse(chi.URLParam(r, "organizationID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization_id")
		return uuid.Nil, uuid.Nil, false
	}

	tenantID, ok := httpmiddleware.TenantFromContext(r.Context())
	if !ok || strings.TrimSpace(tenantID.String()) == "" {
		writeError(w, http.StatusBadRequest, "missing tenant context")
		return uuid.Nil, uuid.Nil, false
	}

	tenantOrganizationID, err := uuid.Parse(tenantID.String())
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant_id")
		return uuid.Nil, uuid.Nil, false
	}

	if tenantOrganizationID != organizationID {
		writeError(w, http.StatusForbidden, "tenant does not match organization")
		return uuid.Nil, uuid.Nil, false
	}

	return currentUser, organizationID, true
}

func writeOrganizationError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, applicationorganization.ErrInvalidOrganization),
		errors.Is(err, applicationorganization.ErrInvalidMembership),
		errors.Is(err, applicationorganization.ErrInvalidOwnershipTransfer):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, applicationorganization.ErrOrganizationNotFound),
		errors.Is(err, applicationorganization.ErrMembershipNotFound),
		errors.Is(err, applicationorganization.ErrUserNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, applicationorganization.ErrForbiddenOrganization),
		errors.Is(err, applicationorganization.ErrOwnerMembershipImmutable),
		errors.Is(err, applicationorganization.ErrCannotModifyOwnMembership):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, applicationorganization.ErrOrganizationSlugTaken),
		errors.Is(err, applicationorganization.ErrMembershipAlreadyExists):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
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
