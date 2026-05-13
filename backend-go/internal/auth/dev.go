package auth

import "net/http"

type Principal struct {
	ID          string
	DisplayName string
	Email       string
	Roles       []string
}

type DevAuthenticator struct {
	Enabled bool
}

func (a DevAuthenticator) PrincipalFromRequest(r *http.Request) (Principal, bool) {
	if !a.Enabled {
		return Principal{}, false
	}

	return Principal{
		ID:          "dev-user",
		DisplayName: "Development User",
		Email:       "dev@example.local",
		Roles:       []string{"admin", "host", "player"},
	}, true
}
