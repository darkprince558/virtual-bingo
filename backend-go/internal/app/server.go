package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/ai"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/audit"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/auth"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/clock"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/config"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/db"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/events"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/game"
)

type databasePinger interface {
	Ping(context.Context) error
}

type Server struct {
	cfg      config.Config
	logger   *slog.Logger
	database databasePinger
	service  *game.Service
}

func NewServer(cfg config.Config, logger *slog.Logger, database *db.Pool) *http.Server {
	if logger == nil {
		logger = slog.Default()
	}

	var service *game.Service
	var pinger databasePinger
	if database != nil {
		pinger = database
		store := db.NewStore(database)
		var aiClient ai.Client = ai.DisabledClient{}
		if cfg.AIServiceEnabled {
			aiClient = ai.NewHTTPClient(cfg.AIServiceBaseURL, cfg.AIServiceTimeout)
		}
		service = game.NewService(game.ServiceConfig{
			Store:         store,
			Authenticator: newAuthenticator(cfg),
			Publisher:     events.NoopPublisher{},
			AuditLogger:   audit.NewStoreLogger(store),
			Clock:         clock.SystemClock{},
			AIClient:      aiClient,
		})
	}

	appServer := &Server{
		cfg:      cfg,
		logger:   logger,
		database: pinger,
		service:  service,
	}

	httpServer := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           appServer.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	if service != nil && cfg.PlayerConnectionTimeout > 0 && cfg.PlayerConnectionSweepInterval > 0 {
		ctx, cancel := context.WithCancel(context.Background())
		httpServer.RegisterOnShutdown(cancel)
		go appServer.runPlayerConnectionSweeper(ctx)
	}

	return httpServer
}

func newAuthenticator(cfg config.Config) auth.Authenticator {
	return auth.NewAuthenticator(auth.Options{
		Mode: cfg.AuthMode,
		EntraConfig: auth.EntraConfig{
			TenantID: cfg.EntraTenantID,
			ClientID: cfg.EntraClientID,
			Audience: cfg.EntraAudience,
			Issuer:   cfg.EntraIssuer,
			JWKSURL:  cfg.EntraJWKSURL,
		},
	})
}

func (s *Server) runPlayerConnectionSweeper(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.PlayerConnectionSweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			players, err := s.service.SweepStalePlayerConnections(ctx, s.cfg.PlayerConnectionTimeout, s.cfg.PlayerConnectionSweepBatchSize)
			if err != nil {
				s.logger.Error("player connection sweep failed", "error", err)
				continue
			}
			if len(players) > 0 {
				s.logger.Info("marked stale players disconnected", "count", len(players))
			}
		}
	}
}
