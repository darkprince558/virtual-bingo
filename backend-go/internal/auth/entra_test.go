package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDevAuthenticatorStillUsesHeaders(t *testing.T) {
	authenticator := NewAuthenticator(Options{Mode: "dev"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Dev-User-Email", "HOST@Example.Local")
	req.Header.Set("X-Dev-User-Name", "Host Person")
	req.Header.Set("X-Dev-User-Role", "host")

	principal, err := authenticator.Authenticate(req)
	if err != nil {
		t.Fatalf("authenticate dev request: %v", err)
	}
	if principal.Email != "host@example.local" || principal.DisplayName != "Host Person" || !HasRole(principal, "host") {
		t.Fatalf("unexpected dev principal: %+v", principal)
	}
}

func TestEntraReadyAuthenticatorMapsVerifiedClaimsToPrincipal(t *testing.T) {
	authenticator := NewEntraReadyAuthenticator(EntraConfig{
		TenantID: "tenant-1",
		ClientID: "client-1",
		Audience: "api://virtual-bingo",
		Issuer:   "https://login.microsoftonline.com/tenant-1/v2.0",
		JWKSURL:  "https://login.microsoftonline.com/tenant-1/discovery/v2.0/keys",
	}, VerifierFunc(func(ctx context.Context, token string) (TokenClaims, error) {
		if token != "good-token" {
			return TokenClaims{}, ErrInvalidToken
		}
		return TokenClaims{
			Subject: "entra-subject-1",
			Email:   "Host@Example.Local",
			Name:    "Entra Host",
			Roles:   []string{"Bingo.Host"},
		}, nil
	}), nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	principal, err := authenticator.Authenticate(req)
	if err != nil {
		t.Fatalf("authenticate entra-ready request: %v", err)
	}
	if principal.ID != "entra:entra-subject-1" || principal.Email != "host@example.local" || !HasRole(principal, "host") {
		t.Fatalf("unexpected entra principal: %+v", principal)
	}
}

func TestEntraReadyAuthenticatorRejectsMissingOrInvalidBearer(t *testing.T) {
	authenticator := NewEntraReadyAuthenticator(EntraConfig{}, VerifierFunc(func(context.Context, string) (TokenClaims, error) {
		return TokenClaims{}, ErrInvalidToken
	}), nil)

	if _, err := authenticator.Authenticate(httptest.NewRequest(http.MethodGet, "/", nil)); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected missing token to be unauthenticated, got %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	if _, err := authenticator.Authenticate(req); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected invalid token to be unauthenticated, got %v", err)
	}
}

func TestDefaultRoleMapperDefaultsToPlayer(t *testing.T) {
	principal, err := PrincipalFromTokenClaims(TokenClaims{
		Subject: "subject-1",
		Email:   "player@example.local",
		Name:    "Player",
	}, nil)
	if err != nil {
		t.Fatalf("map principal: %v", err)
	}
	if !HasRole(principal, "player") {
		t.Fatalf("expected default player role, got %+v", principal)
	}
}
