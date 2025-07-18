package injector

import (
	"travel-api/internal/domain"
	"travel-api/internal/infrastructure/postgres"
	"travel-api/internal/interface/handler"
	"travel-api/internal/usecase"
)

func NewTripHandler(db postgres.DBTX) *handler.TripHandler {
	tripRepository := postgres.NewTripPostgresRepository(db)
	clock := &domain.SystemClock{}
	uuidGenerator := &domain.DefaultUUIDGenerator{}
	tripUsecase := usecase.NewTripInteractor(tripRepository, clock, uuidGenerator)
	return handler.NewTripHandler(tripUsecase)
}
