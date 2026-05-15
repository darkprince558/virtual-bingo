package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort   = 8080
	defaultAppEnv = "development"
)

type Config struct {
	Port                           int
	AppEnv                         string
	DatabaseURL                    string
	CORSAllowedOrigins             []string
	AuthMode                       string
	PlayerConnectionTimeout        time.Duration
	PlayerConnectionSweepInterval  time.Duration
	PlayerConnectionSweepBatchSize int
	AzureTenantID                  string
	AzureClientID                  string
	EntraTenantID                  string
	EntraClientID                  string
	EntraAudience                  string
	EntraIssuer                    string
	EntraJWKSURL                   string
	AzureServiceBusNS              string
	AppInsightsConn                string
}

func Load() (Config, error) {
	cfg := Config{
		Port:                           defaultPort,
		AppEnv:                         getEnv("APP_ENV", defaultAppEnv),
		DatabaseURL:                    strings.TrimSpace(os.Getenv("DATABASE_URL")),
		CORSAllowedOrigins:             splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS")),
		AuthMode:                       normalizeAuthMode(getEnv("AUTH_MODE", "dev")),
		PlayerConnectionTimeout:        90 * time.Second,
		PlayerConnectionSweepInterval:  30 * time.Second,
		PlayerConnectionSweepBatchSize: 100,
		AzureTenantID:                  strings.TrimSpace(os.Getenv("AZURE_TENANT_ID")),
		AzureClientID:                  strings.TrimSpace(os.Getenv("AZURE_CLIENT_ID")),
		EntraTenantID:                  strings.TrimSpace(firstNonEmpty(os.Getenv("ENTRA_TENANT_ID"), os.Getenv("AZURE_TENANT_ID"))),
		EntraClientID:                  strings.TrimSpace(firstNonEmpty(os.Getenv("ENTRA_CLIENT_ID"), os.Getenv("AZURE_CLIENT_ID"))),
		EntraAudience:                  strings.TrimSpace(firstNonEmpty(os.Getenv("ENTRA_AUDIENCE"), os.Getenv("AZURE_CLIENT_ID"))),
		EntraIssuer:                    strings.TrimSpace(os.Getenv("ENTRA_ISSUER")),
		EntraJWKSURL:                   strings.TrimSpace(os.Getenv("ENTRA_JWKS_URL")),
		AzureServiceBusNS:              strings.TrimSpace(os.Getenv("AZURE_SERVICE_BUS_NAMESPACE")),
		AppInsightsConn:                strings.TrimSpace(os.Getenv("APPLICATIONINSIGHTS_CONNECTION_STRING")),
	}

	if rawPort := strings.TrimSpace(os.Getenv("PORT")); rawPort != "" {
		port, err := strconv.Atoi(rawPort)
		if err != nil {
			return Config{}, fmt.Errorf("PORT must be a number: %w", err)
		}
		cfg.Port = port
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return Config{}, errors.New("PORT must be between 1 and 65535")
	}
	if cfg.AuthMode != "dev" && cfg.AuthMode != "entra-ready" {
		return Config{}, errors.New("AUTH_MODE must be dev or entra-ready")
	}
	if timeout, err := parseSecondsEnv("PLAYER_CONNECTION_TIMEOUT_SECONDS", cfg.PlayerConnectionTimeout); err != nil {
		return Config{}, err
	} else {
		cfg.PlayerConnectionTimeout = timeout
	}
	if interval, err := parseSecondsEnv("PLAYER_CONNECTION_SWEEP_INTERVAL_SECONDS", cfg.PlayerConnectionSweepInterval); err != nil {
		return Config{}, err
	} else {
		cfg.PlayerConnectionSweepInterval = interval
	}
	if batchSize, err := parseIntEnv("PLAYER_CONNECTION_SWEEP_BATCH_SIZE", cfg.PlayerConnectionSweepBatchSize); err != nil {
		return Config{}, err
	} else {
		cfg.PlayerConnectionSweepBatchSize = batchSize
	}
	if cfg.PlayerConnectionTimeout < 0 || cfg.PlayerConnectionSweepInterval < 0 {
		return Config{}, errors.New("player connection durations must be zero or positive")
	}
	if cfg.PlayerConnectionSweepBatchSize < 1 {
		return Config{}, errors.New("PLAYER_CONNECTION_SWEEP_BATCH_SIZE must be at least 1")
	}

	return cfg, nil
}

func (c Config) Addr() string {
	return net.JoinHostPort("", strconv.Itoa(c.Port))
}

func (c Config) IsDevelopment() bool {
	return c.AppEnv == "development" || c.AppEnv == "local" || c.AppEnv == "test"
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}

	return out
}

func normalizeAuthMode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "entra", "microsoft", "microsoft-entra":
		return "entra-ready"
	default:
		return value
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

func parseSecondsEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a whole number of seconds: %w", key, err)
	}

	return time.Duration(seconds) * time.Second, nil
}

func parseIntEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number: %w", key, err)
	}

	return parsed, nil
}
