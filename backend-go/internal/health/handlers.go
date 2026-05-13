package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type Handler struct {
	startedAt time.Time
}

func NewHandler() Handler {
	return Handler{startedAt: time.Now().UTC()}
}

func (h Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":     "ready",
		"database":   "not_configured",
		"started_at": h.startedAt.Format(time.RFC3339),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
