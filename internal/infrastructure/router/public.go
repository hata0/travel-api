package router

import (
	"travel-api/internal/infrastructure/di"

	"github.com/gin-gonic/gin"
)

func SetupPublicRoutes(group *gin.RouterGroup, container *di.Container) {
	authHandler := container.AuthHandler()
	authHandler.RegisterAPI(group)
}
