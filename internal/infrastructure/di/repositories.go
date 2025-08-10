package di

import (
	"travel-api/internal/domain"
	"travel-api/internal/infrastructure/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositories はリポジトリの実装を提供する
type Repositories struct {
	db                     *pgxpool.Pool
	tripRepository         domain.TripRepository
	userRepository         domain.UserRepository
	refreshTokenRepository domain.RefreshTokenRepository
	revokedTokenRepository domain.RevokedTokenRepository
}

// NewRepositories はリポジトリを初期化する
func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		db:                     db,
		tripRepository:         postgres.NewTripPostgresRepository(db),
		userRepository:         postgres.NewUserPostgresRepository(db),
		refreshTokenRepository: postgres.NewRefreshTokenPostgresRepository(db),
		revokedTokenRepository: postgres.NewRevokedTokenPostgresRepository(db),
	}
}

func (r *Repositories) TripRepository() domain.TripRepository {
	return r.tripRepository
}

func (r *Repositories) UserRepository() domain.UserRepository {
	return r.userRepository
}

func (r *Repositories) RefreshTokenRepository() domain.RefreshTokenRepository {
	return r.refreshTokenRepository
}

func (r *Repositories) RevokedTokenRepository() domain.RevokedTokenRepository {
	return r.revokedTokenRepository
}
