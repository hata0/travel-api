package main

import (
	"context"
	"log/slog"
	"os"
	"travel-api/internal/config"
	"travel-api/internal/injector"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn, err := config.DSN()
	if err != nil {
		slog.Error("Failed to get DSN", "error", err)
		os.Exit(1)
	}

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		slog.Error("connection failed", "error", err)
		os.Exit(1)
	}

	router := gin.Default()

	injector.NewTripHandler(db).RegisterAPI(router)

	slog.Info("Server starting on port 8080")
	if err := router.Run(); err != nil {
		slog.Error("Failed to run server", "error", err)
		os.Exit(1)
	}
}
