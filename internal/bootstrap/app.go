package bootstrap

import (
	"context"
	"fmt"

	applicationhealth "github.com/AbenezerWork/ProcureFlow/internal/application/health"
	applicationidentity "github.com/AbenezerWork/ProcureFlow/internal/application/identity"
	applicationorganization "github.com/AbenezerWork/ProcureFlow/internal/application/organization"
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
	identityRepository := dbrepositories.NewIdentityRepository(store)
	organizationRepository := dbrepositories.NewOrganizationRepository(store)
	healthService := applicationhealth.NewService(cfg.AppName, cfg.Environment, version)
	identityService := applicationidentity.NewService(identityRepository, passwordHasher, tokenManager)
	organizationService := applicationorganization.NewService(organizationRepository, organizationRepository)
	healthHandler := handlers.NewHealthHandler(healthService, httpmiddleware.TenantFromContext)
	authHandler := handlers.NewAuthHandler(identityService)
	organizationHandler := handlers.NewOrganizationHandler(organizationService)
	router := httprouter.New(
		healthHandler,
		authHandler,
		organizationHandler,
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
