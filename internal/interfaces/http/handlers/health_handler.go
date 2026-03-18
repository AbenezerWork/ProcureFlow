package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	domainhealth "github.com/AbenezerWork/ProcureFlow/internal/domain/health"
	"github.com/AbenezerWork/ProcureFlow/internal/domain/tenant"
)

type HealthChecker interface {
	Check(ctx context.Context) domainhealth.Status
}

type TenantLookup func(ctx context.Context) (tenant.ID, bool)

type HealthHandler struct {
	checker      HealthChecker
	tenantLookup TenantLookup
}

func NewHealthHandler(checker HealthChecker, tenantLookup TenantLookup) *HealthHandler {
	return &HealthHandler{
		checker:      checker,
		tenantLookup: tenantLookup,
	}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := struct {
		domainhealth.Status
		TenantID string `json:"tenant_id,omitempty"`
	}{
		Status: h.checker.Check(r.Context()),
	}

	if tenantID, ok := h.tenantLookup(r.Context()); ok {
		response.TenantID = tenantID.String()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
	}
}
