package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/config"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		slog.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	migrator := database.NewMigrator(pool)

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "up":
		result, err := migrator.Up(ctx)
		if err != nil {
			slog.Error("apply migrations", "error", err)
			os.Exit(1)
		}

		if len(result.Applied) == 0 {
			fmt.Printf("schema already at version %d; no migrations pending\n", result.CurrentVersion)
			return
		}

		fmt.Printf("applied %d migration(s); schema now at version %d\n", len(result.Applied), result.CurrentVersion)
	case "version":
		version, err := migrator.Version(ctx)
		if err != nil {
			slog.Error("load schema version", "error", err)
			os.Exit(1)
		}

		fmt.Println(version)
	default:
		fmt.Fprintf(os.Stderr, "usage: %s [up|version]\n", os.Args[0])
		os.Exit(2)
	}
}
