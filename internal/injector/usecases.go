package injector

import (
	"travel-api/internal/usecase"
)

// Usecases はユースケースを提供する
type Usecases struct {
	repos    RepositoryProvider
	services ServiceProvider
	config   *Config

	tripUsecase *usecase.TripInteractor
	authUsecase *usecase.AuthInteractor
}

// NewUsecases はユースケースを初期化する
func NewUsecases(repos RepositoryProvider, services ServiceProvider, config *Config) *Usecases {
	return &Usecases{
		repos:    repos,
		services: services,
		config:   config,
	}
}

func (u *Usecases) TripUsecase() *usecase.TripInteractor {
	if u.tripUsecase == nil {
		u.tripUsecase = usecase.NewTripInteractor(
			u.repos.TripRepository(),
			u.services.Clock(),
			u.services.UUIDGenerator(),
		)
	}
	return u.tripUsecase
}

func (u *Usecases) AuthUsecase() *usecase.AuthInteractor {
	if u.authUsecase == nil {
		u.authUsecase = usecase.NewAuthInteractor(
			u.repos.UserRepository(),
			u.repos.RefreshTokenRepository(),
			u.repos.RevokedTokenRepository(),
			u.services.Clock(),
			u.services.UUIDGenerator(),
			u.services.TransactionManager(),
			u.config.JWTSecret,
		)
	}
	return u.authUsecase
}
