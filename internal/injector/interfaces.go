package injector

import (
	"travel-api/internal/domain"
	"travel-api/internal/interface/handler"
)

// HandlerProvider はハンドラー生成のインターフェース
type HandlerProvider interface {
	TripHandler() *handler.TripHandler
	AuthHandler() *handler.AuthHandler
}

// ServiceProvider はドメインサービスのインターフェース
type ServiceProvider interface {
	Clock() domain.Clock
	UUIDGenerator() domain.UUIDGenerator
	TransactionManager() domain.TransactionManager
}

// RepositoryProvider はリポジトリのインターフェース
type RepositoryProvider interface {
	TripRepository() domain.TripRepository
	UserRepository() domain.UserRepository
	RefreshTokenRepository() domain.RefreshTokenRepository
	RevokedTokenRepository() domain.RevokedTokenRepository
}
