package router

import (
	"travel-api/internal/infrastructure/di"

	"github.com/gin-gonic/gin"
)

func SetupProtectedRoutes(group *gin.RouterGroup, container *di.Container) {
	tripHandler := container.TripHandler()
	tripHandler.RegisterAPI(group)
}
