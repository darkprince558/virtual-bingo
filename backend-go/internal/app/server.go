package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

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
		service = game.NewService(game.ServiceConfig{
			Store:         store,
			Authenticator: newAuthenticator(cfg),
			Publisher:     events.NoopPublisher{},
			AuditLogger:   audit.NewStoreLogger(store),
			Clock:         clock.SystemClock{},
		})
	}

	appServer := &Server{
		cfg:      cfg,
		logger:   logger,
		database: pinger,
		service:  service,
	}

	return &http.Server{
		Addr:              cfg.Addr(),
		Handler:           appServer.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
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
