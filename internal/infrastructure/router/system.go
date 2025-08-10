package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hata0/travel-api/internal/infrastructure/config"
)

func SetupSystemEndpoints(router *gin.Engine, cfg config.Config) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   cfg.Version(),
			"env":       cfg.Environment(),
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	if !cfg.IsProduction() {
		router.GET("/metrics", func(c *gin.Context) {
			c.String(http.StatusOK, "# TODO: Implement metrics")
		})
	}
}
