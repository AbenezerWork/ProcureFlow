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

type VendorRoutes interface {
	Create(http.ResponseWriter, *http.Request)
	List(http.ResponseWriter, *http.Request)
	Get(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Archive(http.ResponseWriter, *http.Request)
}

type ProcurementRoutes interface {
	CreateRequest(http.ResponseWriter, *http.Request)
	ListRequests(http.ResponseWriter, *http.Request)
	ListApprovalInbox(http.ResponseWriter, *http.Request)
	GetRequest(http.ResponseWriter, *http.Request)
	UpdateRequest(http.ResponseWriter, *http.Request)
	SubmitRequest(http.ResponseWriter, *http.Request)
	ApproveRequest(http.ResponseWriter, *http.Request)
	RejectRequest(http.ResponseWriter, *http.Request)
	CancelRequest(http.ResponseWriter, *http.Request)
	ListItems(http.ResponseWriter, *http.Request)
	CreateItem(http.ResponseWriter, *http.Request)
	UpdateItem(http.ResponseWriter, *http.Request)
	DeleteItem(http.ResponseWriter, *http.Request)
}

type RFQRoutes interface {
	Create(http.ResponseWriter, *http.Request)
	List(http.ResponseWriter, *http.Request)
	Get(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Publish(http.ResponseWriter, *http.Request)
	Close(http.ResponseWriter, *http.Request)
	Evaluate(http.ResponseWriter, *http.Request)
	Cancel(http.ResponseWriter, *http.Request)
	ListItems(http.ResponseWriter, *http.Request)
	ListVendors(http.ResponseWriter, *http.Request)
	AttachVendor(http.ResponseWriter, *http.Request)
	RemoveVendor(http.ResponseWriter, *http.Request)
}

type QuotationRoutes interface {
	Compare(http.ResponseWriter, *http.Request)
	Create(http.ResponseWriter, *http.Request)
	List(http.ResponseWriter, *http.Request)
	Get(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Submit(http.ResponseWriter, *http.Request)
	Reject(http.ResponseWriter, *http.Request)
	ListItems(http.ResponseWriter, *http.Request)
	UpdateItem(http.ResponseWriter, *http.Request)
}

type AwardRoutes interface {
	Create(http.ResponseWriter, *http.Request)
	Get(http.ResponseWriter, *http.Request)
}

type ActivityLogRoutes interface {
	ListByEntity(http.ResponseWriter, *http.Request)
}

func New(
	healthHandler http.Handler,
	authHandler AuthRoutes,
	organizationHandler OrganizationRoutes,
	vendorHandler VendorRoutes,
	procurementHandler ProcurementRoutes,
	quotationHandler QuotationRoutes,
	awardHandler AwardRoutes,
	activityLogHandler ActivityLogRoutes,
	rfqHandler RFQRoutes,
	authMiddleware func(http.Handler) http.Handler,
	tenantHeader string,
) http.Handler {
	r := chi.NewRouter()
	r.Use(httpmiddleware.WithRequestID)
	r.Use(func(next http.Handler) http.Handler {
		return httpmiddleware.WithTenant(next, tenantHeader)
	})
	r.Use(httpmiddleware.AccessLog)
	r.Use(httpmiddleware.Recover)
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
					r.Route("/activity-logs", func(r chi.Router) {
						r.Get("/", activityLogHandler.ListByEntity)
					})
					r.Route("/approvals", func(r chi.Router) {
						r.Route("/procurement-requests", func(r chi.Router) {
							r.Get("/", procurementHandler.ListApprovalInbox)
						})
					})
					r.Route("/memberships", func(r chi.Router) {
						r.Get("/", organizationHandler.ListMemberships)
						r.Post("/", organizationHandler.CreateMembership)
						r.Patch("/{userID}", organizationHandler.UpdateMembership)
					})
					r.Route("/vendors", func(r chi.Router) {
						r.Get("/", vendorHandler.List)
						r.Post("/", vendorHandler.Create)
						r.Route("/{vendorID}", func(r chi.Router) {
							r.Get("/", vendorHandler.Get)
							r.Patch("/", vendorHandler.Update)
							r.Post("/archive", vendorHandler.Archive)
						})
					})
					r.Route("/procurement-requests", func(r chi.Router) {
						r.Get("/", procurementHandler.ListRequests)
						r.Post("/", procurementHandler.CreateRequest)
						r.Route("/{requestID}", func(r chi.Router) {
							r.Get("/", procurementHandler.GetRequest)
							r.Patch("/", procurementHandler.UpdateRequest)
							r.Post("/submit", procurementHandler.SubmitRequest)
							r.Post("/approve", procurementHandler.ApproveRequest)
							r.Post("/reject", procurementHandler.RejectRequest)
							r.Post("/cancel", procurementHandler.CancelRequest)
							r.Route("/items", func(r chi.Router) {
								r.Get("/", procurementHandler.ListItems)
								r.Post("/", procurementHandler.CreateItem)
								r.Route("/{itemID}", func(r chi.Router) {
									r.Patch("/", procurementHandler.UpdateItem)
									r.Delete("/", procurementHandler.DeleteItem)
								})
							})
						})
					})
					r.Route("/rfqs", func(r chi.Router) {
						r.Get("/", rfqHandler.List)
						r.Post("/", rfqHandler.Create)
						r.Route("/{rfqID}", func(r chi.Router) {
							r.Get("/", rfqHandler.Get)
							r.Patch("/", rfqHandler.Update)
							r.Post("/publish", rfqHandler.Publish)
							r.Post("/close", rfqHandler.Close)
							r.Post("/evaluate", rfqHandler.Evaluate)
							r.Post("/cancel", rfqHandler.Cancel)
							r.Route("/items", func(r chi.Router) {
								r.Get("/", rfqHandler.ListItems)
							})
							r.Route("/vendors", func(r chi.Router) {
								r.Get("/", rfqHandler.ListVendors)
								r.Post("/", rfqHandler.AttachVendor)
								r.Delete("/{vendorID}", rfqHandler.RemoveVendor)
							})
							r.Get("/comparison", quotationHandler.Compare)
							r.Route("/quotations", func(r chi.Router) {
								r.Get("/", quotationHandler.List)
								r.Post("/", quotationHandler.Create)
								r.Route("/{quotationID}", func(r chi.Router) {
									r.Get("/", quotationHandler.Get)
									r.Patch("/", quotationHandler.Update)
									r.Post("/submit", quotationHandler.Submit)
									r.Post("/reject", quotationHandler.Reject)
									r.Route("/items", func(r chi.Router) {
										r.Get("/", quotationHandler.ListItems)
										r.Patch("/{itemID}", quotationHandler.UpdateItem)
									})
								})
							})
							r.Route("/award", func(r chi.Router) {
								r.Get("/", awardHandler.Get)
								r.Post("/", awardHandler.Create)
							})
						})
					})
				})
			})
		})
	})

	return r
}
