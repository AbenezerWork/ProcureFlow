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
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database/demo"
)

func main() {
	command, seedDemo := parseArgs(os.Args[1:])

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

	switch command {
	case "up":
		result, err := migrator.Up(ctx)
		if err != nil {
			slog.Error("apply migrations", "error", err)
			os.Exit(1)
		}

		if len(result.Applied) == 0 {
			fmt.Printf("schema already at version %d; no migrations pending\n", result.CurrentVersion)
		} else {
			fmt.Printf("applied %d migration(s); schema now at version %d\n", len(result.Applied), result.CurrentVersion)
		}

		if seedDemo {
			result, err := demo.Seed(ctx, pool)
			if err != nil {
				slog.Error("seed demo data", "error", err)
				os.Exit(1)
			}

			fmt.Printf(
				"seeded demo organization %s with %d users; demo password: %s\n",
				result.OrganizationID,
				result.UserCount,
				demo.DemoPassword,
			)
		}
	case "version":
		if seedDemo {
			fmt.Fprintln(os.Stderr, "-seed-demo can only be used with up")
			os.Exit(2)
		}

		version, err := migrator.Version(ctx)
		if err != nil {
			slog.Error("load schema version", "error", err)
			os.Exit(1)
		}

		fmt.Println(version)
	default:
		fmt.Fprintf(os.Stderr, "usage: %s [-seed-demo] [up|version]\n", os.Args[0])
		os.Exit(2)
	}
}

func parseArgs(args []string) (string, bool) {
	command := "up"
	commandSet := false
	seedDemo := false

	for _, arg := range args {
		switch arg {
		case "-seed-demo", "--seed-demo":
			seedDemo = true
		case "up", "version":
			if commandSet {
				fmt.Fprintf(os.Stderr, "usage: %s [-seed-demo] [up|version]\n", os.Args[0])
				os.Exit(2)
			}
			command = arg
			commandSet = true
		default:
			fmt.Fprintf(os.Stderr, "usage: %s [-seed-demo] [up|version]\n", os.Args[0])
			os.Exit(2)
		}
	}

	return command, seedDemo
}
