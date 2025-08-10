package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hata0/travel-api/internal/infrastructure/di"
)

func SetupProtectedRoutes(group *gin.RouterGroup, container *di.Container) {
	tripHandler := container.TripHandler()
	tripHandler.RegisterAPI(group)
}
