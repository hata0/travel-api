package injector

import (
	"travel-api/internal/domain"
	"travel-api/internal/domain/shared/clock"
	"travel-api/internal/interface/handler"
	"travel-api/internal/usecase"
)

// HandlerProvider はハンドラー生成のインターフェース
type HandlerProvider interface {
	TripHandler() *handler.TripHandler
	AuthHandler() *handler.AuthHandler
}

// ServiceProvider はドメインサービスのインターフェース
type ServiceProvider interface {
	Clock() clock.Clock
	UUIDGenerator() domain.UUIDGenerator
	TransactionManager() usecase.TransactionManager
}

// RepositoryProvider はリポジトリのインターフェース
type RepositoryProvider interface {
	TripRepository() domain.TripRepository
	UserRepository() domain.UserRepository
	RefreshTokenRepository() domain.RefreshTokenRepository
	RevokedTokenRepository() domain.RevokedTokenRepository
}
