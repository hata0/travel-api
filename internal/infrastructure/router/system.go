package router

import (
	"net/http"
	"time"

	"travel-api/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
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
