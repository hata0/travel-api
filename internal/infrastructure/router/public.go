package router

import (
	"travel-api/internal/injector"

	"github.com/gin-gonic/gin"
)

func SetupPublicRoutes(group *gin.RouterGroup, container *injector.Container) {
	authHandler := container.AuthHandler()
	authHandler.RegisterAPI(group)
}
