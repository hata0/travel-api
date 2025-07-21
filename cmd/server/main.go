package main

import (
	"context"
	"log/slog"
	"os"
	"travel-api/internal/server"
)

func main() {
	application, err := server.NewApplication()
	if err != nil {
		slog.Error("Failed to create application", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := application.Run(ctx); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}
