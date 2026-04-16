//go:build integration

package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

func openTestStore(t *testing.T) (*database.Store, func()) {
	t.Helper()

	databaseURL := os.Getenv("PROCUREFLOW_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set PROCUREFLOW_TEST_DATABASE_URL to run integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping test database: %v", err)
	}

	if err := database.NewMigrator(pool).Up(ctx); err != nil {
		pool.Close()
		t.Fatalf("apply migrations: %v", err)
	}

	store := database.NewStore(pool)
	cleanupDatabase(t, store)

	return store, func() {
		cleanupDatabase(t, store)
		pool.Close()
	}
}

func cleanupDatabase(t *testing.T, store *database.Store) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := store.Pool().Exec(ctx, `
TRUNCATE TABLE
    activity_logs,
    rfq_awards,
    quotation_items,
    quotations,
    rfq_vendors,
    rfq_items,
    rfqs,
    procurement_request_items,
    procurement_requests,
    vendors,
    organization_memberships,
    organizations,
    users
RESTART IDENTITY CASCADE
`)
	if err != nil {
		t.Fatalf("clean test database: %v", err)
	}
}
