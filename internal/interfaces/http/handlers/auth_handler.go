package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	applicationidentity "github.com/AbenezerWork/ProcureFlow/internal/application/identity"
	domainidentity "github.com/AbenezerWork/ProcureFlow/internal/domain/identity"
	httpmiddleware "github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/middleware"
	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, input applicationidentity.RegisterInput) (domainidentity.Session, error)
	Login(ctx context.Context, input applicationidentity.LoginInput) (domainidentity.Session, error)
	CurrentUser(ctx context.Context, userID uuid.UUID) (domainidentity.User, error)
}

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	session, err := h.service.Register(r.Context(), applicationidentity.RegisterInput{
		Email:    request.Email,
		Password: request.Password,
		FullName: request.FullName,
	})
	if err != nil {
		switch {
		case errors.Is(err, applicationidentity.ErrEmailAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, applicationidentity.ErrInvalidCredentials):
			writeError(w, http.StatusBadRequest, "email, password, and full_name are required")
		default:
			writeError(w, http.StatusInternalServerError, "register user")
		}
		return
	}

	writeJSON(w, http.StatusCreated, sessionResponse(session))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	session, err := h.service.Login(r.Context(), applicationidentity.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, applicationidentity.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, err.Error())
		case errors.Is(err, applicationidentity.ErrUserInactive):
			writeError(w, http.StatusForbidden, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "login user")
		}
		return
	}

	writeJSON(w, http.StatusOK, sessionResponse(session))
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpmiddleware.AuthenticatedUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	user, err := h.service.CurrentUser(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, applicationidentity.ErrUserNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, applicationidentity.ErrUserInactive):
			writeError(w, http.StatusForbidden, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "load current user")
		}
		return
	}

	writeJSON(w, http.StatusOK, userResponse(user))
}

func sessionResponse(session domainidentity.Session) map[string]any {
	return map[string]any{
		"access_token": session.Token.AccessToken,
		"token_type":   "Bearer",
		"expires_at":   session.Token.ExpiresAt.UTC().Format(time.RFC3339),
		"user":         userResponse(session.User),
	}
}

func userResponse(user domainidentity.User) map[string]any {
	response := map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"full_name":  user.FullName,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at": user.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if user.LastLoginAt != nil {
		response["last_login_at"] = user.LastLoginAt.UTC().Format(time.RFC3339)
	}

	return response
}
