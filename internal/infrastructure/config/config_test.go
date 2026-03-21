package config

import (
	"strings"
	"testing"
)

func TestLoadRequiresJWTSecret(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("DB_USER", "procureflow")
	t.Setenv("DB_PASSWORD", "procureflow")
	t.Setenv("DB_NAME", "procureflow")

	_, err := Load()
	if err == nil {
		t.Fatal("expected load to fail when AUTH_JWT_SECRET is missing")
	}

	if !strings.Contains(err.Error(), "AUTH_JWT_SECRET") {
		t.Fatalf("expected AUTH_JWT_SECRET error, got %v", err)
	}
}

func TestLoadRequiresDatabaseUser(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("AUTH_JWT_SECRET", "test-secret")
	t.Setenv("DB_PASSWORD", "procureflow")
	t.Setenv("DB_NAME", "procureflow")

	_, err := Load()
	if err == nil {
		t.Fatal("expected load to fail when DB_USER is missing")
	}

	if !strings.Contains(err.Error(), "DB_USER") {
		t.Fatalf("expected DB_USER error, got %v", err)
	}
}

func TestLoadRequiresDatabasePassword(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("AUTH_JWT_SECRET", "test-secret")
	t.Setenv("DB_USER", "procureflow")
	t.Setenv("DB_NAME", "procureflow")

	_, err := Load()
	if err == nil {
		t.Fatal("expected load to fail when DB_PASSWORD is missing")
	}

	if !strings.Contains(err.Error(), "DB_PASSWORD") {
		t.Fatalf("expected DB_PASSWORD error, got %v", err)
	}
}

func TestLoadRequiresDatabaseName(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("AUTH_JWT_SECRET", "test-secret")
	t.Setenv("DB_USER", "procureflow")
	t.Setenv("DB_PASSWORD", "procureflow")

	_, err := Load()
	if err == nil {
		t.Fatal("expected load to fail when DB_NAME is missing")
	}

	if !strings.Contains(err.Error(), "DB_NAME") {
		t.Fatalf("expected DB_NAME error, got %v", err)
	}
}

func TestLoadAcceptsExplicitSensitiveConfig(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("APP_ENV", defaultEnvironment)
	t.Setenv("AUTH_JWT_SECRET", "test-secret")
	t.Setenv("DB_USER", "procureflow")
	t.Setenv("DB_PASSWORD", "procureflow")
	t.Setenv("DB_NAME", "procureflow")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load returned error: %v", err)
	}

	if cfg.Auth.JWTSecret != "test-secret" {
		t.Fatalf("expected JWT secret test-secret, got %q", cfg.Auth.JWTSecret)
	}

	if cfg.Database.User != "procureflow" {
		t.Fatalf("expected DB user procureflow, got %q", cfg.Database.User)
	}
}

func resetConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"APP_NAME",
		"APP_ENV",
		"APP_HTTP_ADDRESS",
		"APP_SHUTDOWN_TIMEOUT",
		"APP_TENANT_HEADER",
		"AUTH_JWT_ISSUER",
		"AUTH_JWT_SECRET",
		"AUTH_ACCESS_TOKEN_TTL",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"DB_SSLMODE",
	} {
		t.Setenv(key, "")
	}
}
