package auth

import (
	"errors"
	"net/http"
	"strings"
)

var ErrUnauthenticated = errors.New("unauthenticated")

type Principal struct {
	ID          string
	DisplayName string
	Email       string
	Roles       []string
}

type Authenticator interface {
	Authenticate(*http.Request) (Principal, error)
}

type DevAuthenticator struct {
	Enabled bool
}

func (a DevAuthenticator) Authenticate(r *http.Request) (Principal, error) {
	principal, ok := a.PrincipalFromRequest(r)
	if !ok {
		return Principal{}, ErrUnauthenticated
	}

	return principal, nil
}

func (a DevAuthenticator) PrincipalFromRequest(r *http.Request) (Principal, bool) {
	if !a.Enabled {
		return Principal{}, false
	}

	email := strings.TrimSpace(r.Header.Get("X-Dev-User-Email"))
	if email == "" {
		email = "host@example.local"
	}

	displayName := strings.TrimSpace(r.Header.Get("X-Dev-User-Name"))
	if displayName == "" {
		displayName = "Local Demo Host"
	}

	role := strings.TrimSpace(r.Header.Get("X-Dev-User-Role"))
	if role == "" {
		role = "host"
	}

	externalSubject := strings.TrimSpace(r.Header.Get("X-Dev-User-ID"))
	if externalSubject == "" {
		externalSubject = "dev:" + strings.ToLower(email)
	}

	return Principal{
		ID:          externalSubject,
		DisplayName: displayName,
		Email:       strings.ToLower(email),
		Roles:       []string{strings.ToLower(role)},
	}, true
}

func HasRole(principal Principal, allowedRoles ...string) bool {
	for _, role := range principal.Roles {
		for _, allowed := range allowedRoles {
			if strings.EqualFold(role, allowed) {
				return true
			}
		}
	}

	return false
}
