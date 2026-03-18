package bootstrap

import (
	"context"

	applicationhealth "github.com/AbenezerWork/ProcureFlow/internal/application/health"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/config"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/httpserver"
	"github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/handlers"
	httpmiddleware "github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/middleware"
	httprouter "github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/router"
)

type App struct {
	server *httpserver.Server
}

func New(cfg config.Config, version string) *App {
	healthService := applicationhealth.NewService(cfg.AppName, cfg.Environment, version)
	healthHandler := handlers.NewHealthHandler(healthService, httpmiddleware.TenantFromContext)
	router := httprouter.New(healthHandler, cfg.TenantHeader)
	server := httpserver.New(cfg.HTTPAddress, router, cfg.ShutdownTimeout)

	return &App{server: server}
}

func (a *App) Run(ctx context.Context) error {
	return a.server.Run(ctx)
}
