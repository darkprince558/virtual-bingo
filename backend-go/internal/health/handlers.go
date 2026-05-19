package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Pinger interface {
	Ping(context.Context) error
}

type Handler struct {
	startedAt time.Time
	database  Pinger
}

func NewHandler(database Pinger) Handler {
	return Handler{
		startedAt: time.Now().UTC(),
		database:  database,
	}
}

func (h Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	if h.database == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":     "ready",
			"database":   "not_configured",
			"started_at": h.startedAt.Format(time.RFC3339),
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.database.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status":     "not_ready",
			"database":   "unavailable",
			"started_at": h.startedAt.Format(time.RFC3339),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":     "ready",
		"database":   "ready",
		"started_at": h.startedAt.Format(time.RFC3339),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
