package router

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"travel-api/internal/adapter/middleware"
	"travel-api/internal/config"
	"travel-api/internal/injector"
)

func SetupRouter(cfg config.Config, container *injector.Container, logger *slog.Logger) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// router.Use(
	// 	middleware.RequestIDMiddleware(),
	// 	middleware.StructuredLoggingMiddleware(logger),
	// 	middleware.RecoveryMiddleware(logger),
	// 	middleware.CORSMiddleware(),
	// 	middleware.SecurityHeadersMiddleware(),
	// )

	SetupSystemEndpoints(router, cfg)

	v1 := router.Group("/api/v1")

	public := v1.Group("/public")
	SetupPublicRoutes(public, container)

	protected := v1.Group("/")
	protected.Use(middleware.RateLimitMiddleware(100, time.Minute))
	protected.Use(middleware.AuthMiddleware(cfg.JWT().Secret()))
	SetupProtectedRoutes(protected, container)

	return router
}
