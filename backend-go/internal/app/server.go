package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/config"
)

type Server struct {
	cfg    config.Config
	logger *slog.Logger
}

func NewServer(cfg config.Config, logger *slog.Logger) *http.Server {
	if logger == nil {
		logger = slog.Default()
	}

	appServer := &Server{
		cfg:    cfg,
		logger: logger,
	}

	return &http.Server{
		Addr:              cfg.Addr(),
		Handler:           appServer.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
}
