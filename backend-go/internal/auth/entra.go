package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var ErrInvalidToken = errors.New("invalid token")

type EntraConfig struct {
	TenantID string
	ClientID string
	Audience string
	Issuer   string
	JWKSURL  string
}

type TokenClaims struct {
	Subject   string
	Email     string
	Name      string
	Preferred string
	Roles     []string
	Scopes    []string
	TenantID  string
	Audience  []string
	Issuer    string
	Raw       map[string]any
}

type TokenVerifier interface {
	VerifyToken(context.Context, string) (TokenClaims, error)
}

type RoleMapper interface {
	MapRoles(TokenClaims) []string
}

type EntraReadyAuthenticator struct {
	config   EntraConfig
	verifier TokenVerifier
	mapper   RoleMapper
}

func NewEntraReadyAuthenticator(config EntraConfig, verifier TokenVerifier, mapper RoleMapper) EntraReadyAuthenticator {
	if verifier == nil {
		verifier = UnconfiguredTokenVerifier{}
	}
	if mapper == nil {
		mapper = DefaultRoleMapper{}
	}

	return EntraReadyAuthenticator{
		config:   config,
		verifier: verifier,
		mapper:   mapper,
	}
}

func (a EntraReadyAuthenticator) Authenticate(r *http.Request) (Principal, error) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		return Principal{}, ErrUnauthenticated
	}

	claims, err := a.verifier.VerifyToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrUnauthenticated) || errors.Is(err, ErrInvalidToken) {
			return Principal{}, ErrUnauthenticated
		}
		return Principal{}, fmt.Errorf("verify token: %w", err)
	}

	principal, err := PrincipalFromTokenClaims(claims, a.mapper)
	if err != nil {
		return Principal{}, err
	}

	return principal, nil
}

func PrincipalFromTokenClaims(claims TokenClaims, mapper RoleMapper) (Principal, error) {
	subject := strings.TrimSpace(claims.Subject)
	if subject == "" {
		return Principal{}, ErrUnauthenticated
	}

	email := normalizeClaimEmail(claims)
	if email == "" {
		return Principal{}, ErrUnauthenticated
	}

	displayName := strings.TrimSpace(claims.Name)
	if displayName == "" {
		displayName = email
	}

	if mapper == nil {
		mapper = DefaultRoleMapper{}
	}
	roles := mapper.MapRoles(claims)
	if len(roles) == 0 {
		roles = []string{"player"}
	}

	return Principal{
		ID:          "entra:" + subject,
		DisplayName: displayName,
		Email:       strings.ToLower(email),
		Roles:       roles,
	}, nil
}

type DefaultRoleMapper struct{}

func (DefaultRoleMapper) MapRoles(claims TokenClaims) []string {
	seen := make(map[string]bool)
	roles := make([]string, 0, len(claims.Roles)+len(claims.Scopes)+1)
	for _, value := range append(claims.Roles, claims.Scopes...) {
		role := normalizeRole(value)
		if role == "" || seen[role] {
			continue
		}
		seen[role] = true
		roles = append(roles, role)
	}
	if len(roles) == 0 {
		return []string{"player"}
	}

	return roles
}

type UnconfiguredTokenVerifier struct{}

func (UnconfiguredTokenVerifier) VerifyToken(context.Context, string) (TokenClaims, error) {
	return TokenClaims{}, ErrUnauthenticated
}

func bearerToken(value string) (string, bool) {
	parts := strings.Fields(value)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}

	return parts[1], true
}

func normalizeClaimEmail(claims TokenClaims) string {
	for _, value := range []string{claims.Email, claims.Preferred} {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" {
			return value
		}
	}

	return ""
}

func normalizeRole(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "bingo.")
	value = strings.TrimPrefix(value, "virtual_bingo.")
	switch value {
	case "admin", "admins":
		return "admin"
	case "host", "hosts":
		return "host"
	case "player", "players", "user", "users":
		return "player"
	case "viewer", "viewers", "read", "read.all":
		return "viewer"
	default:
		return ""
	}
}
