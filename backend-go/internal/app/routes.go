package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/health"
)

const apiVersion = "v0.1.0"

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	healthHandler := health.NewHandler(s.database)

	mux.HandleFunc("/healthz", getOnly(healthHandler.Healthz))
	mux.HandleFunc("/readyz", getOnly(healthHandler.Readyz))
	mux.HandleFunc("/api/v1/version", getOnly(s.version))
	mux.HandleFunc("/api/v1/config", getOnly(s.getConfig))
	mux.HandleFunc("GET /api/v1/me", s.getCurrentUser)
	mux.HandleFunc("GET /api/v1/word-sets", s.listWordSets)
	mux.HandleFunc("POST /api/v1/word-sets", s.createWordSet)
	mux.HandleFunc("GET /api/v1/word-sets/{wordSetID}", s.getWordSet)
	mux.HandleFunc("PATCH /api/v1/word-sets/{wordSetID}", s.updateWordSet)
	mux.HandleFunc("POST /api/v1/word-sets/{wordSetID}/words", s.createWordSetWord)
	mux.HandleFunc("PATCH /api/v1/word-sets/{wordSetID}/words/{wordID}", s.updateWordSetWord)
	mux.HandleFunc("DELETE /api/v1/word-sets/{wordSetID}/words/{wordID}", s.deleteWordSetWord)
	mux.HandleFunc("POST /api/v1/games", s.createGame)
	mux.HandleFunc("GET /api/v1/games", s.listGames)
	mux.HandleFunc("/api/v1/games/", s.dispatchGameRoute)

	return s.recoverPanic(s.withRequestID(s.withCORS(s.logRequests(mux))))
}

func (s *Server) version(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, map[string]string{
		"service": "virtual-bingo-api",
		"version": apiVersion,
		"env":     s.cfg.AppEnv,
	})
}

func (s *Server) dispatchGameRoute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/games/")
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) == 0 || segments[0] == "" {
		writeAPIError(w, http.StatusNotFound, "not_found", "resource was not found")
		return
	}

	if segments[0] == "code" && len(segments) == 2 && r.Method == http.MethodGet {
		r.SetPathValue("code", segments[1])
		s.getGameByCode(w, r)
		return
	}

	r.SetPathValue("gameID", segments[0])
	if len(segments) == 1 {
		switch r.Method {
		case http.MethodGet:
			s.getGame(w, r)
		case http.MethodPatch:
			s.updateGame(w, r)
		default:
			writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}

	switch segments[1] {
	case "start":
		if r.Method == http.MethodPost && len(segments) == 2 {
			s.startGame(w, r)
			return
		}
	case "pause":
		if r.Method == http.MethodPost && len(segments) == 2 {
			s.pauseGame(w, r)
			return
		}
	case "resume":
		if r.Method == http.MethodPost && len(segments) == 2 {
			s.resumeGame(w, r)
			return
		}
	case "finish":
		if r.Method == http.MethodPost && len(segments) == 2 {
			s.finishGame(w, r)
			return
		}
	case "cancel":
		if r.Method == http.MethodPost && len(segments) == 2 {
			s.cancelGame(w, r)
			return
		}
	case "host-snapshot":
		if r.Method == http.MethodGet && len(segments) == 2 {
			s.getHostSnapshot(w, r)
			return
		}
	case "activity":
		if r.Method == http.MethodGet && len(segments) == 2 {
			s.listActivityEvents(w, r)
			return
		}
	case "events":
		if r.Method == http.MethodGet && len(segments) == 2 {
			s.streamGameEvents(w, r)
			return
		}
	case "calls":
		if len(segments) == 2 {
			if r.Method == http.MethodPost {
				s.callNextWord(w, r)
				return
			}
			if r.Method == http.MethodGet {
				s.listCalledWords(w, r)
				return
			}
		}
	case "allowed-players":
		if len(segments) == 2 {
			if r.Method == http.MethodPost {
				s.addAllowedPlayer(w, r)
				return
			}
			if r.Method == http.MethodGet {
				s.listAllowedPlayers(w, r)
				return
			}
		}
		if len(segments) == 3 && segments[2] == "bulk" && r.Method == http.MethodPost {
			s.bulkAddAllowedPlayers(w, r)
			return
		}
		if len(segments) == 3 {
			r.SetPathValue("allowedPlayerID", segments[2])
			if r.Method == http.MethodPatch {
				s.updateAllowedPlayer(w, r)
				return
			}
			if r.Method == http.MethodDelete {
				s.deleteAllowedPlayer(w, r)
				return
			}
		}
	case "players":
		if len(segments) == 2 && r.Method == http.MethodPost {
			s.joinPlayer(w, r)
			return
		}
		if len(segments) >= 3 && segments[2] == "me" {
			s.dispatchCurrentPlayerRoute(w, r, segments)
			return
		}
		if len(segments) >= 3 {
			r.SetPathValue("playerID", segments[2])
			s.dispatchPlayerRoute(w, r, segments)
			return
		}
	case "claims":
		if len(segments) == 2 {
			if r.Method == http.MethodPost {
				s.submitBingoClaim(w, r)
				return
			}
			if r.Method == http.MethodGet {
				s.listBingoClaims(w, r)
				return
			}
		}
		if len(segments) == 3 && r.Method == http.MethodGet {
			r.SetPathValue("claimID", segments[2])
			s.getBingoClaim(w, r)
			return
		}
	case "summary":
		if r.Method == http.MethodGet && len(segments) == 2 {
			s.getGameSummary(w, r)
			return
		}
	}

	writeAPIError(w, http.StatusNotFound, "not_found", "resource was not found")
}

func (s *Server) dispatchPlayerRoute(w http.ResponseWriter, r *http.Request, segments []string) {
	if len(segments) == 4 && segments[3] == "snapshot" && r.Method == http.MethodGet {
		s.getPlayerSnapshot(w, r)
		return
	}
	if len(segments) == 4 && segments[3] == "heartbeat" && r.Method == http.MethodPost {
		s.heartbeatPlayer(w, r)
		return
	}
	if len(segments) == 4 && segments[3] == "card" {
		if r.Method == http.MethodPost {
			s.assignPlayerCard(w, r)
			return
		}
		if r.Method == http.MethodGet {
			s.getPlayerCard(w, r)
			return
		}
	}
	if len(segments) == 6 && segments[3] == "card" && segments[4] == "cells" && r.Method == http.MethodPatch {
		r.SetPathValue("cellID", segments[5])
		s.markCardCell(w, r)
		return
	}

	writeAPIError(w, http.StatusNotFound, "not_found", "resource was not found")
}

func (s *Server) dispatchCurrentPlayerRoute(w http.ResponseWriter, r *http.Request, segments []string) {
	if len(segments) == 4 && segments[3] == "snapshot" && r.Method == http.MethodGet {
		s.getCurrentPlayerSnapshot(w, r)
		return
	}
	if len(segments) == 4 && segments[3] == "heartbeat" && r.Method == http.MethodPost {
		s.heartbeatCurrentPlayer(w, r)
		return
	}
	if len(segments) == 4 && segments[3] == "card" {
		if r.Method == http.MethodPost {
			s.assignCurrentPlayerCard(w, r)
			return
		}
		if r.Method == http.MethodGet {
			s.getCurrentPlayerCard(w, r)
			return
		}
	}
	if len(segments) == 6 && segments[3] == "card" && segments[4] == "cells" && r.Method == http.MethodPatch {
		r.SetPathValue("cellID", segments[5])
		s.markCurrentPlayerCard(w, r)
		return
	}

	writeAPIError(w, http.StatusNotFound, "not_found", "resource was not found")
}

func (s *Server) getConfig(w http.ResponseWriter, r *http.Request) {
	authMode := s.authMode()
	writeData(w, http.StatusOK, map[string]any{
		"service":  "virtual-bingo-api",
		"version":  apiVersion,
		"appEnv":   s.appEnv(),
		"authMode": authMode,
		"capabilities": map[string]bool{
			"sseEvents":      true,
			"devAuth":        authMode == "dev",
			"entraReadyAuth": authMode == "entra-ready",
			"rewards":        false,
			"automation":     false,
			"aiContent":      false,
			"voice":          false,
		},
		"playerConnection": map[string]any{
			"timeoutSeconds":       int(s.cfg.PlayerConnectionTimeout.Seconds()),
			"sweepIntervalSeconds": int(s.cfg.PlayerConnectionSweepInterval.Seconds()),
			"sweepBatchSize":       s.cfg.PlayerConnectionSweepBatchSize,
		},
	})
}

func (s *Server) authMode() string {
	if s.cfg.AuthMode == "" {
		return "dev"
	}

	return s.cfg.AuthMode
}

func (s *Server) appEnv() string {
	if s.cfg.AppEnv == "" {
		return "development"
	}

	return s.cfg.AppEnv
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
