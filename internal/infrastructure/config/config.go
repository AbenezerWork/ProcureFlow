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
	defaultAccessTokenTTL  = 24 * time.Hour
	defaultDBHost          = "localhost"
	defaultDBPort          = 5432
	defaultDBSSLMode       = "disable"
)

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
	jwtSecret, err := requiredEnv("AUTH_JWT_SECRET")
	if err != nil {
		return Config{}, err
	}

	dbUser, err := requiredEnv("DB_USER")
	if err != nil {
		return Config{}, err
	}

	dbPassword, err := requiredEnv("DB_PASSWORD")
	if err != nil {
		return Config{}, err
	}

	dbName, err := requiredEnv("DB_NAME")
	if err != nil {
		return Config{}, err
	}

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

	return Config{
		AppName:         valueOrDefault("APP_NAME", defaultAppName),
		Environment:     environment,
		HTTPAddress:     valueOrDefault("APP_HTTP_ADDRESS", defaultHTTPAddress),
		ShutdownTimeout: shutdownTimeout,
		TenantHeader:    valueOrDefault("APP_TENANT_HEADER", defaultTenantHeader),
		Auth: AuthConfig{
			JWTIssuer:      valueOrDefault("AUTH_JWT_ISSUER", defaultJWTIssuer),
			JWTSecret:      jwtSecret,
			AccessTokenTTL: accessTokenTTL,
		},
		Database: DatabaseConfig{
			Host:     valueOrDefault("DB_HOST", defaultDBHost),
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			Name:     dbName,
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

func requiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s must be set", key)
	}

	return value, nil
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
