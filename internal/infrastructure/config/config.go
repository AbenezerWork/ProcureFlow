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
	defaultDBHost          = "localhost"
	defaultDBPort          = 5432
	defaultDBUser          = "procureflow"
	defaultDBPassword      = "procureflow"
	defaultDBName          = "procureflow"
	defaultDBSSLMode       = "disable"
)

type Config struct {
	AppName         string
	Environment     string
	HTTPAddress     string
	ShutdownTimeout time.Duration
	TenantHeader    string
	Database        DatabaseConfig
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
	shutdownTimeout := defaultShutdownTimeout
	if raw := os.Getenv("APP_SHUTDOWN_TIMEOUT"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("parse APP_SHUTDOWN_TIMEOUT: %w", err)
		}
		shutdownTimeout = parsed
	}

	dbPort, err := intFromEnv("DB_PORT", defaultDBPort)
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppName:         valueOrDefault("APP_NAME", defaultAppName),
		Environment:     valueOrDefault("APP_ENV", defaultEnvironment),
		HTTPAddress:     valueOrDefault("APP_HTTP_ADDRESS", defaultHTTPAddress),
		ShutdownTimeout: shutdownTimeout,
		TenantHeader:    valueOrDefault("APP_TENANT_HEADER", defaultTenantHeader),
		Database: DatabaseConfig{
			Host:     valueOrDefault("DB_HOST", defaultDBHost),
			Port:     dbPort,
			User:     valueOrDefault("DB_USER", defaultDBUser),
			Password: valueOrDefault("DB_PASSWORD", defaultDBPassword),
			Name:     valueOrDefault("DB_NAME", defaultDBName),
			SSLMode:  valueOrDefault("DB_SSLMODE", defaultDBSSLMode),
		},
	}, nil
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
