package injector

import (
	"travel-api/infrastructure/database"
	"travel-api/infrastructure/repository"
	"travel-api/interface/handler"
	"travel-api/usecase"
)

func NewTripHandler(db database.DBTX) *handler.TripHandler {
	queries := database.New(db)

	tripRepository := repository.NewTripPostgresRepository(queries)
	tripUsecase := usecase.NewTripInteractor(tripRepository)
	return handler.NewTripHandler(tripUsecase)
}
