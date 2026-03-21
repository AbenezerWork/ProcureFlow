package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultAppName         = "procureflow-api"
	defaultEnvironment     = "development"
	defaultHTTPAddress     = ":8080"
	defaultShutdownTimeout = 10 * time.Second
	defaultTenantHeader    = "X-Tenant-ID"
	defaultJWTIssuer       = "procureflow-api"
	defaultJWTSecret       = "procureflow-development-secret"
	defaultAccessTokenTTL  = 24 * time.Hour
	defaultDBHost          = "localhost"
	defaultDBPort          = 5432
	defaultDBUser          = "procureflow"
	defaultDBPassword      = "procureflow"
	defaultDBName          = "procureflow"
	defaultDBSSLMode       = "disable"
)

const productionEnvironment = "production"

type Config struct {
	AppName         string
	Environment     string
	HTTPAddress     string
	ShutdownTimeout time.Duration
	TenantHeader    string
	Auth            AuthConfig
	Database        DatabaseConfig
}

type AuthConfig struct {
	JWTIssuer      string
	JWTSecret      string
	AccessTokenTTL time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (c DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

func Load() (Config, error) {
	environment := valueOrDefault("APP_ENV", defaultEnvironment)

	shutdownTimeout := defaultShutdownTimeout
	if raw := os.Getenv("APP_SHUTDOWN_TIMEOUT"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("parse APP_SHUTDOWN_TIMEOUT: %w", err)
		}
		shutdownTimeout = parsed
	}

	accessTokenTTL := defaultAccessTokenTTL
	if raw := os.Getenv("AUTH_ACCESS_TOKEN_TTL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("parse AUTH_ACCESS_TOKEN_TTL: %w", err)
		}
		accessTokenTTL = parsed
	}

	dbPort, err := intFromEnv("DB_PORT", defaultDBPort)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppName:         valueOrDefault("APP_NAME", defaultAppName),
		Environment:     environment,
		HTTPAddress:     valueOrDefault("APP_HTTP_ADDRESS", defaultHTTPAddress),
		ShutdownTimeout: shutdownTimeout,
		TenantHeader:    valueOrDefault("APP_TENANT_HEADER", defaultTenantHeader),
		Auth: AuthConfig{
			JWTIssuer:      valueOrDefault("AUTH_JWT_ISSUER", defaultJWTIssuer),
			JWTSecret:      valueOrDefault("AUTH_JWT_SECRET", defaultJWTSecret),
			AccessTokenTTL: accessTokenTTL,
		},
		Database: DatabaseConfig{
			Host:     valueOrDefault("DB_HOST", defaultDBHost),
			Port:     dbPort,
			User:     valueOrDefault("DB_USER", defaultDBUser),
			Password: valueOrDefault("DB_PASSWORD", defaultDBPassword),
			Name:     valueOrDefault("DB_NAME", defaultDBName),
			SSLMode:  valueOrDefault("DB_SSLMODE", defaultDBSSLMode),
		},
	}

	if environment == productionEnvironment {
		if err := validateProductionConfig(cfg); err != nil {
			return Config{}, err
		}
	}

	return cfg, nil
}

func validateProductionConfig(cfg Config) error {
	if cfg.Auth.JWTSecret == defaultJWTSecret {
		return fmt.Errorf("AUTH_JWT_SECRET must be set to a non-default value in %s", productionEnvironment)
	}

	if cfg.Database.User == defaultDBUser {
		return fmt.Errorf("DB_USER must be set to a non-default value in %s", productionEnvironment)
	}

	if cfg.Database.Password == defaultDBPassword {
		return fmt.Errorf("DB_PASSWORD must be set to a non-default value in %s", productionEnvironment)
	}

	if cfg.Database.Name == defaultDBName {
		return fmt.Errorf("DB_NAME must be set to a non-default value in %s", productionEnvironment)
	}

	return nil
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func intFromEnv(key string, fallback int) (int, error) {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("parse %s: %w", key, err)
		}

		return parsed, nil
	}

	return fallback, nil
}
