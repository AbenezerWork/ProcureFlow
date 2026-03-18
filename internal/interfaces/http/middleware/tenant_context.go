package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/AbenezerWork/ProcureFlow/internal/domain/tenant"
)

type tenantContextKey struct{}

func WithTenant(next http.Handler, headerName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := strings.TrimSpace(r.Header.Get(headerName))
		if tenantID == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), tenantContextKey{}, tenant.ID(tenantID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TenantFromContext(ctx context.Context) (tenant.ID, bool) {
	tenantID, ok := ctx.Value(tenantContextKey{}).(tenant.ID)
	return tenantID, ok
}
