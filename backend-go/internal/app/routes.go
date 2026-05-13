package app

import (
	"encoding/json"
	"net/http"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/health"
)

const apiVersion = "v0.1.0"

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	healthHandler := health.NewHandler(s.database)

	mux.HandleFunc("/healthz", getOnly(healthHandler.Healthz))
	mux.HandleFunc("/readyz", getOnly(healthHandler.Readyz))
	mux.HandleFunc("/api/v1/version", getOnly(s.version))

	return s.recoverPanic(s.withCORS(s.logRequests(mux)))
}

func (s *Server) version(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "virtual-bingo-api",
		"version": apiVersion,
		"env":     s.cfg.AppEnv,
	})
}

func getOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
