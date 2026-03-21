package config

import (
	"strings"
	"testing"
)

func TestLoadAllowsDevelopmentDefaults(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("APP_ENV", defaultEnvironment)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load returned error: %v", err)
	}

	if cfg.Auth.JWTSecret != defaultJWTSecret {
		t.Fatalf("expected default JWT secret %q, got %q", defaultJWTSecret, cfg.Auth.JWTSecret)
	}

	if cfg.Database.Password != defaultDBPassword {
		t.Fatalf("expected default DB password %q, got %q", defaultDBPassword, cfg.Database.Password)
	}
}

func TestLoadRejectsProductionDefaultSecret(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("APP_ENV", productionEnvironment)
	t.Setenv("DB_USER", "procureflow-prod")
	t.Setenv("DB_PASSWORD", "super-secret")
	t.Setenv("DB_NAME", "procureflow_prod")

	_, err := Load()
	if err == nil {
		t.Fatal("expected load to fail for default production JWT secret")
	}

	if !strings.Contains(err.Error(), "AUTH_JWT_SECRET") {
		t.Fatalf("expected AUTH_JWT_SECRET error, got %v", err)
	}
}

func TestLoadRejectsProductionDefaultDatabaseCredentials(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("APP_ENV", productionEnvironment)
	t.Setenv("AUTH_JWT_SECRET", "prod-secret")

	_, err := Load()
	if err == nil {
		t.Fatal("expected load to fail for default production database credentials")
	}

	if !strings.Contains(err.Error(), "DB_USER") {
		t.Fatalf("expected DB_USER error, got %v", err)
	}
}

func TestLoadAcceptsProductionExplicitSensitiveConfig(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("APP_ENV", productionEnvironment)
	t.Setenv("AUTH_JWT_SECRET", "prod-secret")
	t.Setenv("DB_USER", "procureflow-prod")
	t.Setenv("DB_PASSWORD", "super-secret")
	t.Setenv("DB_NAME", "procureflow_prod")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load returned error: %v", err)
	}

	if cfg.Auth.JWTSecret != "prod-secret" {
		t.Fatalf("expected JWT secret prod-secret, got %q", cfg.Auth.JWTSecret)
	}

	if cfg.Database.User != "procureflow-prod" {
		t.Fatalf("expected DB user procureflow-prod, got %q", cfg.Database.User)
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
