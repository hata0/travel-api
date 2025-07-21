package main

import (
	"context"
	"log/slog"
	"os"
	"travel-api/internal/server"
)

func main() {
	server, err := server.NewServer()
	if err != nil {
		slog.Error("Failed to create application", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}
