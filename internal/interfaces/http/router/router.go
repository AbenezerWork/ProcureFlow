package router

import (
	"net/http"

	"github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/apidocs"
	httpmiddleware "github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/middleware"
	"github.com/go-chi/chi/v5"
)

type AuthRoutes interface {
	Register(http.ResponseWriter, *http.Request)
	Login(http.ResponseWriter, *http.Request)
	Me(http.ResponseWriter, *http.Request)
}

type OrganizationRoutes interface {
	Create(http.ResponseWriter, *http.Request)
	ListMine(http.ResponseWriter, *http.Request)
	Get(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	ListMemberships(http.ResponseWriter, *http.Request)
	CreateMembership(http.ResponseWriter, *http.Request)
	UpdateMembership(http.ResponseWriter, *http.Request)
	TransferOwnership(http.ResponseWriter, *http.Request)
}

func New(
	healthHandler http.Handler,
	authHandler AuthRoutes,
	organizationHandler OrganizationRoutes,
	authMiddleware func(http.Handler) http.Handler,
	tenantHeader string,
) http.Handler {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return httpmiddleware.WithTenant(next, tenantHeader)
	})
	r.Get("/openapi.yaml", apidocs.SpecHandler().ServeHTTP)
	r.Get("/swagger", apidocs.SwaggerUIHandler("/openapi.yaml").ServeHTTP)
	r.Get("/healthz", healthHandler.ServeHTTP)
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)

			r.Group(func(r chi.Router) {
				r.Use(authMiddleware)
				r.Get("/me", authHandler.Me)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Route("/organizations", func(r chi.Router) {
				r.Get("/", organizationHandler.ListMine)
				r.Post("/", organizationHandler.Create)
				r.Route("/{organizationID}", func(r chi.Router) {
					r.Get("/", organizationHandler.Get)
					r.Patch("/", organizationHandler.Update)
					r.Post("/ownership-transfer", organizationHandler.TransferOwnership)
					r.Route("/memberships", func(r chi.Router) {
						r.Get("/", organizationHandler.ListMemberships)
						r.Post("/", organizationHandler.CreateMembership)
						r.Patch("/{userID}", organizationHandler.UpdateMembership)
					})
				})
			})
		})
	})

	return r
}
