package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"travel-api/internal/config"
	"travel-api/internal/injector"
	"travel-api/internal/router"
)

type Server struct {
	config    config.Config
	server    *http.Server
	container *injector.Container
	logger    *slog.Logger
}

func NewServer() (*Server, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger := SetupLogger(cfg.Log())
	slog.SetDefault(logger)

	factory := injector.NewFactory()
	container, err := factory.CreateProductionContainer(cfg)
	if err != nil {
		logger.Error("Failed to create DI container", "error", err)
		return nil, err
	}

	r := router.SetupRouter(cfg, container, logger)

	server := &http.Server{
		Addr:         cfg.Server().Address(),
		Handler:      r,
		ReadTimeout:  cfg.Server().ReadTimeout(),
		WriteTimeout: cfg.Server().WriteTimeout(),
		IdleTimeout:  cfg.Server().IdleTimeout(),
	}

	return &Server{
		config:    cfg,
		server:    server,
		container: container,
		logger:    logger,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	s.logStartupInfo()

	serverErrors := make(chan error, 1)
	go func() {
		s.logger.Info("Starting HTTP server",
			"address", s.server.Addr,
			"env", s.config.Environment())
		serverErrors <- s.server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case sig := <-shutdown:
		s.logger.Info("Received shutdown signal", "signal", sig)
		return s.shutdown(ctx)
	case <-ctx.Done():
		s.logger.Info("Context cancelled, shutting down")
		return s.shutdown(ctx)
	}

	return nil
}

func (s *Server) shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.Server().ShutdownTimeout())
	defer cancel()

	s.logger.Info("Shutting down server", "timeout", s.config.Server().ShutdownTimeout())

	if s.container != nil {
		if err := s.container.Close(); err != nil {
			s.logger.Error("Failed to close container", "error", err)
		}
	}

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server shutdown failed", "error", err)
		return err
	}

	s.logger.Info("Server shutdown completed")
	return nil
}

func (s *Server) logStartupInfo() {
	s.logger.Info("Application starting",
		"env", s.config.Environment(),
		"version", s.config.Version(),
	)
}
