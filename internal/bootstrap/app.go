package bootstrap

import (
	"context"
	"fmt"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	applicationaward "github.com/AbenezerWork/ProcureFlow/internal/application/award"
	applicationhealth "github.com/AbenezerWork/ProcureFlow/internal/application/health"
	applicationidentity "github.com/AbenezerWork/ProcureFlow/internal/application/identity"
	applicationorganization "github.com/AbenezerWork/ProcureFlow/internal/application/organization"
	applicationprocurement "github.com/AbenezerWork/ProcureFlow/internal/application/procurement"
	applicationquotation "github.com/AbenezerWork/ProcureFlow/internal/application/quotation"
	applicationrfq "github.com/AbenezerWork/ProcureFlow/internal/application/rfq"
	applicationvendor "github.com/AbenezerWork/ProcureFlow/internal/application/vendor"
	authinfra "github.com/AbenezerWork/ProcureFlow/internal/infrastructure/auth"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/config"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database"
	dbrepositories "github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database/repositories"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/httpserver"
	"github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/handlers"
	httpmiddleware "github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/middleware"
	httprouter "github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/router"
)

type App struct {
	server   *httpserver.Server
	store    *database.Store
	shutdown func()
}

func New(ctx context.Context, cfg config.Config, version string) (*App, error) {
	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("initialize database pool: %w", err)
	}

	store := database.NewStore(pool)
	passwordHasher := authinfra.NewPasswordHasher()
	tokenManager := authinfra.NewTokenManager(cfg.Auth)
	activityLogRepository := dbrepositories.NewActivityLogRepository(store)
	identityRepository := dbrepositories.NewIdentityRepository(store)
	organizationRepository := dbrepositories.NewOrganizationRepository(store)
	procurementRepository := dbrepositories.NewProcurementRepository(store)
	awardRepository := dbrepositories.NewAwardRepository(store)
	quotationRepository := dbrepositories.NewQuotationRepository(store)
	rfqRepository := dbrepositories.NewRFQRepository(store)
	vendorRepository := dbrepositories.NewVendorRepository(store)
	activityLogService := applicationactivitylog.NewService(activityLogRepository)
	healthService := applicationhealth.NewService(cfg.AppName, cfg.Environment, version)
	awardService := applicationaward.NewService(awardRepository, awardRepository)
	identityService := applicationidentity.NewService(identityRepository, passwordHasher, tokenManager)
	organizationService := applicationorganization.NewService(organizationRepository, organizationRepository, identityRepository)
	procurementService := applicationprocurement.NewService(procurementRepository, procurementRepository)
	quotationService := applicationquotation.NewService(quotationRepository, quotationRepository)
	rfqService := applicationrfq.NewService(rfqRepository, rfqRepository)
	vendorService := applicationvendor.NewService(vendorRepository, vendorRepository)
	activityLogHandler := handlers.NewActivityLogHandler(activityLogService)
	healthHandler := handlers.NewHealthHandler(healthService, httpmiddleware.TenantFromContext)
	awardHandler := handlers.NewAwardHandler(awardService)
	authHandler := handlers.NewAuthHandler(identityService)
	organizationHandler := handlers.NewOrganizationHandler(organizationService)
	procurementHandler := handlers.NewProcurementHandler(procurementService)
	quotationHandler := handlers.NewQuotationHandler(quotationService)
	rfqHandler := handlers.NewRFQHandler(rfqService)
	vendorHandler := handlers.NewVendorHandler(vendorService)
	router := httprouter.New(
		healthHandler,
		authHandler,
		organizationHandler,
		vendorHandler,
		procurementHandler,
		quotationHandler,
		awardHandler,
		activityLogHandler,
		rfqHandler,
		httpmiddleware.RequireAuthentication(identityService),
		cfg.TenantHeader,
	)
	server := httpserver.New(cfg.HTTPAddress, router, cfg.ShutdownTimeout)

	return &App{
		server:   server,
		store:    store,
		shutdown: pool.Close,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.shutdown()

	return a.server.Run(ctx)
}
