package config

import "testing"

func TestLoadAuthModeAndEntraPlaceholders(t *testing.T) {
	t.Setenv("AUTH_MODE", "entra")
	t.Setenv("ENTRA_TENANT_ID", "tenant-1")
	t.Setenv("ENTRA_CLIENT_ID", "client-1")
	t.Setenv("ENTRA_AUDIENCE", "api://virtual-bingo")
	t.Setenv("ENTRA_ISSUER", "https://issuer.example")
	t.Setenv("ENTRA_JWKS_URL", "https://issuer.example/keys")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.AuthMode != "entra-ready" {
		t.Fatalf("expected entra-ready auth mode, got %q", cfg.AuthMode)
	}
	if cfg.EntraTenantID != "tenant-1" || cfg.EntraClientID != "client-1" || cfg.EntraAudience != "api://virtual-bingo" || cfg.EntraIssuer == "" || cfg.EntraJWKSURL == "" {
		t.Fatalf("unexpected entra placeholders: %+v", cfg)
	}
}

func TestLoadRejectsUnsupportedAuthMode(t *testing.T) {
	t.Setenv("AUTH_MODE", "magic")

	if _, err := Load(); err == nil {
		t.Fatal("expected unsupported AUTH_MODE to fail")
	}
}
