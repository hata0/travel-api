package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"travel-api/internal/config"
	"travel-api/internal/injector"
	"travel-api/internal/interface/middleware"

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
	defer db.Close()

	// 認証が必要なAPIグループ
	jwtSecret, err := config.JWTSecret()
	if err != nil {
		slog.Error("Failed to get JWT secret", "error", err)
		os.Exit(1)
	}

	router := gin.Default()

	injector.NewAuthHandler(db, jwtSecret).RegisterAPI(router)
	injector.NewTripHandler(db).RegisterAPI(router)

	// 認証が必要なAPIグループ
	authRequired := router.Group("/")
	authRequired.Use(middleware.RateLimitMiddleware(100, time.Minute))
	authRequired.Use(middleware.AuthMiddleware(jwtSecret))
	{
		// ここに認証が必要なAPIを登録
		// 例: authRequired.GET("/protected", handler.ProtectedHandler)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		slog.Info("Server starting on port 8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to run server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exiting")
}
