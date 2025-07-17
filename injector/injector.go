package injector

import (
	"travel-api/infrastructure/repository"
	"travel-api/interface/controller"
	"travel-api/usecase"
)

func NewTripController(queries repository.TripQuerier) *controller.TripController {
	tripRepository := repository.NewTripPostgresRepository(queries)
	tripUsecase := usecase.NewTripInteractor(tripRepository)
	return controller.NewTripController(tripUsecase)
}
