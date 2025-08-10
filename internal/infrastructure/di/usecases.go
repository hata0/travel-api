package di

import (
	"github.com/hata0/travel-api/internal/infrastructure/config"
	"github.com/hata0/travel-api/internal/usecase"
)

// Usecases はユースケースを提供する
type Usecases struct {
	repos    RepositoryProvider
	services ServiceProvider
	config   config.Config

	tripUsecase *usecase.TripInteractor
	authUsecase *usecase.AuthInteractor
}

// NewUsecases はユースケースを初期化する
func NewUsecases(repos RepositoryProvider, services ServiceProvider, cfg config.Config) *Usecases {
	return &Usecases{
		repos:    repos,
		services: services,
		config:   cfg,
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
			u.config.JWT().Secret(),
		)
	}
	return u.authUsecase
}
