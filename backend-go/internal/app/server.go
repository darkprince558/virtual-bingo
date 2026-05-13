package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/config"
)

type databasePinger interface {
	Ping(context.Context) error
}

type Server struct {
	cfg      config.Config
	logger   *slog.Logger
	database databasePinger
}

func NewServer(cfg config.Config, logger *slog.Logger, database databasePinger) *http.Server {
	if logger == nil {
		logger = slog.Default()
	}

	appServer := &Server{
		cfg:      cfg,
		logger:   logger,
		database: database,
	}

	return &http.Server{
		Addr:              cfg.Addr(),
		Handler:           appServer.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
}
