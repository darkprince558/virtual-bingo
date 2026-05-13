package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	defaultPort   = 8080
	defaultAppEnv = "development"
)

type Config struct {
	Port               int
	AppEnv             string
	DatabaseURL        string
	CORSAllowedOrigins []string
	AzureTenantID      string
	AzureClientID      string
	AzureServiceBusNS  string
	AppInsightsConn    string
}

func Load() (Config, error) {
	cfg := Config{
		Port:               defaultPort,
		AppEnv:             getEnv("APP_ENV", defaultAppEnv),
		DatabaseURL:        strings.TrimSpace(os.Getenv("DATABASE_URL")),
		CORSAllowedOrigins: splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS")),
		AzureTenantID:      strings.TrimSpace(os.Getenv("AZURE_TENANT_ID")),
		AzureClientID:      strings.TrimSpace(os.Getenv("AZURE_CLIENT_ID")),
		AzureServiceBusNS:  strings.TrimSpace(os.Getenv("AZURE_SERVICE_BUS_NAMESPACE")),
		AppInsightsConn:    strings.TrimSpace(os.Getenv("APPLICATIONINSIGHTS_CONNECTION_STRING")),
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
