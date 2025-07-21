package injector

import (
	"travel-api/internal/domain"
	"travel-api/internal/infrastructure/postgres"
	"travel-api/internal/interface/handler"
	"travel-api/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewTripHandler(db *pgxpool.Pool) *handler.TripHandler {
	tripRepository := postgres.NewTripPostgresRepository(db)
	clock := &domain.SystemClock{}
	uuidGenerator := &domain.DefaultUUIDGenerator{}
	tripUsecase := usecase.NewTripInteractor(tripRepository, clock, uuidGenerator)
	return handler.NewTripHandler(tripUsecase)
}

func NewAuthHandler(db *pgxpool.Pool, jwtSecret string) *handler.AuthHandler {
	userRepository := postgres.NewUserPostgresRepository(db)
	refreshTokenRepository := postgres.NewRefreshTokenPostgresRepository(db)
	clock := &domain.SystemClock{}
	uuidGenerator := &domain.DefaultUUIDGenerator{}
	transactionManager := postgres.NewTransactionManager(db)
	revokedTokenRepository := postgres.NewRevokedTokenPostgresRepository(db)
	authUsecase := usecase.NewAuthInteractor(userRepository, refreshTokenRepository, revokedTokenRepository, clock, uuidGenerator, transactionManager, jwtSecret)
	return handler.NewAuthHandler(authUsecase)
}
