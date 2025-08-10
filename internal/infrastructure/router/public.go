package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hata0/travel-api/internal/infrastructure/di"
)

func SetupPublicRoutes(group *gin.RouterGroup, container *di.Container) {
	authHandler := container.AuthHandler()
	authHandler.RegisterAPI(group)
}
