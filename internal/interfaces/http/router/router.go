package router

import (
	"net/http"

	httpmiddleware "github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/middleware"
)

func New(healthHandler http.Handler, tenantHeader string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthz", healthHandler)

	return httpmiddleware.WithTenant(mux, tenantHeader)
}
