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

func NewAuthHandler(db postgres.DBTX) *handler.AuthHandler {
	userRepository := postgres.NewUserPostgresRepository(db)
	refreshTokenRepository := postgres.NewRefreshTokenPostgresRepository(db)
	clock := &domain.SystemClock{}
	uuidGenerator := &domain.DefaultUUIDGenerator{}
	authUsecase := usecase.NewAuthInteractor(userRepository, refreshTokenRepository, clock, uuidGenerator)
	return handler.NewAuthHandler(authUsecase)
}
