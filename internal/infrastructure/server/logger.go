package server

import (
	"log/slog"
	"os"

	"travel-api/internal/infrastructure/config"
)

func SetupLogger(cfg config.LogConfig) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     cfg.Level(),
		AddSource: cfg.AddSource(),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Time(slog.TimeKey, a.Value.Time().UTC())
			}
			return a
		},
	}

	var handler slog.Handler
	switch cfg.Format() {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
