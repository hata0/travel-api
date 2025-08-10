package router

import (
	"travel-api/internal/injector"

	"github.com/gin-gonic/gin"
)

func SetupProtectedRoutes(group *gin.RouterGroup, container *injector.Container) {
	tripHandler := container.TripHandler()
	tripHandler.RegisterAPI(group)
}
