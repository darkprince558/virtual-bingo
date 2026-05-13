package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/auth"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/db"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/game"
)

const requestIDHeader = "X-Request-ID"

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiErrorResponse struct {
	Error apiError `json:"error"`
}

type apiDataResponse struct {
	Data any `json:"data"`
}

func writeData(w http.ResponseWriter, status int, payload any) {
	writeJSON(w, status, apiDataResponse{Data: payload})
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, apiErrorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}

	return nil
}

func requireDatabase(w http.ResponseWriter, service *game.Service) bool {
	if service != nil {
		return true
	}

	writeAPIError(w, http.StatusServiceUnavailable, "database_required", "database is not configured")
	return false
}

func mapServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, game.ErrForbidden):
		writeAPIError(w, http.StatusForbidden, "forbidden", "you do not have permission to perform this action")
	case errors.Is(err, game.ErrNotAllowed):
		writeAPIError(w, http.StatusForbidden, "not_allowed", "player is not allowed to join this game")
	case errors.Is(err, game.ErrValidation):
		writeAPIError(w, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, game.ErrNotFound), errors.Is(err, db.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, "not_found", "resource was not found")
	case errors.Is(err, game.ErrConflict):
		writeAPIError(w, http.StatusConflict, "conflict", err.Error())
	case errors.Is(err, auth.ErrUnauthenticated):
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
	default:
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func generateRequestID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "req-fallback"
	}

	return hex.EncodeToString(bytes[:])
}
