package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	"github.com/google/uuid"
)

type ActivityLogService interface {
	ListByEntity(ctx context.Context, input applicationactivitylog.ListByEntityInput) ([]domainactivitylog.Entry, error)
}

type ActivityLogHandler struct {
	service ActivityLogService
}

func NewActivityLogHandler(service ActivityLogService) *ActivityLogHandler {
	return &ActivityLogHandler{service: service}
}

func (h *ActivityLogHandler) ListByEntity(w http.ResponseWriter, r *http.Request) {
	currentUser, organizationID, ok := authenticatedOrganizationRequest(w, r)
	if !ok {
		return
	}

	entityType := r.URL.Query().Get("entity_type")
	entityIDRaw := r.URL.Query().Get("entity_id")
	entityID, err := uuid.Parse(entityIDRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entity_id")
		return
	}

	entries, err := h.service.ListByEntity(r.Context(), applicationactivitylog.ListByEntityInput{
		OrganizationID: organizationID,
		CurrentUser:    currentUser,
		EntityType:     entityType,
		EntityID:       entityID,
	})
	if err != nil {
		switch {
		case errors.Is(err, applicationactivitylog.ErrInvalidActivityLog):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, applicationactivitylog.ErrForbiddenActivityLog):
			writeError(w, http.StatusForbidden, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "list activity logs")
		}
		return
	}

	response := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		response = append(response, activityLogResponse(entry))
	}

	writeJSON(w, http.StatusOK, map[string]any{"activity_logs": response})
}

func activityLogResponse(entry domainactivitylog.Entry) map[string]any {
	response := map[string]any{
		"id":              entry.ID,
		"organization_id": entry.OrganizationID,
		"entity_type":     entry.EntityType,
		"entity_id":       entry.EntityID,
		"action":          entry.Action,
		"summary":         entry.Summary,
		"metadata":        entry.Metadata,
		"occurred_at":     entry.OccurredAt.UTC().Format(time.RFC3339),
	}

	if entry.ActorUserID != nil {
		response["actor_user_id"] = entry.ActorUserID
	}

	return response
}
