package injector

import (
	"travel-api/infrastructure/postgres"
	"travel-api/interface/handler"
	"travel-api/usecase"
)

func NewTripHandler(db postgres.DBTX) *handler.TripHandler {
	tripRepository := postgres.NewTripPostgresRepository(db)
	tripUsecase := usecase.NewTripInteractor(tripRepository)
	return handler.NewTripHandler(tripUsecase)
}
