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

type Application struct {
	config    config.Config
	server    *http.Server
	container *injector.Container
	logger    *slog.Logger
}

func NewApplication() (*Application, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger := SetupLogger(cfg.Log())
	slog.SetDefault(logger)

	container, err := CreateContainer(cfg)
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

	return &Application{
		config:    cfg,
		server:    server,
		container: container,
		logger:    logger,
	}, nil
}

func (a *Application) Run(ctx context.Context) error {
	a.logStartupInfo()

	serverErrors := make(chan error, 1)
	go func() {
		a.logger.Info("Starting HTTP server",
			"address", a.server.Addr,
			"env", a.config.Environment())
		serverErrors <- a.server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case sig := <-shutdown:
		a.logger.Info("Received shutdown signal", "signal", sig)
		return a.shutdown(ctx)
	case <-ctx.Done():
		a.logger.Info("Context cancelled, shutting down")
		return a.shutdown(ctx)
	}

	return nil
}

func (a *Application) shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, a.config.Server().ShutdownTimeout())
	defer cancel()

	a.logger.Info("Shutting down server", "timeout", a.config.Server().ShutdownTimeout())

	if a.container != nil {
		if err := a.container.Close(); err != nil {
			a.logger.Error("Failed to close container", "error", err)
		}
	}

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("Server shutdown failed", "error", err)
		return err
	}

	a.logger.Info("Server shutdown completed")
	return nil
}

func (a *Application) logStartupInfo() {
	a.logger.Info("Application starting",
		"env", a.config.Environment(),
		"version", a.config.Version(),
	)
}
