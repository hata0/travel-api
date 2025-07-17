package injector

import (
	"travel-api/internal/infrastructure/postgres"
	"travel-api/internal/interface/handler"
	"travel-api/internal/usecase"
)

func NewTripHandler(db postgres.DBTX) *handler.TripHandler {
	tripRepository := postgres.NewTripPostgresRepository(db)
	tripUsecase := usecase.NewTripInteractor(tripRepository)
	return handler.NewTripHandler(tripUsecase)
}
