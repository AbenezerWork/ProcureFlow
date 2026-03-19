package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/AbenezerWork/ProcureFlow/internal/bootstrap"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/config"
)

var version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := bootstrap.New(ctx, cfg, version)
	if err != nil {
		slog.Error("initialize application", "error", err)
		os.Exit(1)
	}

	if err := app.Run(ctx); err != nil {
		slog.Error("run application", "error", err)
		os.Exit(1)
	}
}
