package di

import (
	"travel-api/internal/adapter/handler"
	"travel-api/internal/domain"
	"travel-api/internal/domain/shared/clock"
	"travel-api/internal/domain/shared/transaction_manager"
	"travel-api/internal/domain/shared/uuid"
)

// HandlerProvider はハンドラー生成のインターフェース
type HandlerProvider interface {
	TripHandler() *handler.TripHandler
	AuthHandler() *handler.AuthHandler
}

// ServiceProvider はドメインサービスのインターフェース
type ServiceProvider interface {
	Clock() clock.Clock
	UUIDGenerator() uuid.UUIDGenerator
	TransactionManager() transaction_manager.TransactionManager
}

// RepositoryProvider はリポジトリのインターフェース
type RepositoryProvider interface {
	TripRepository() domain.TripRepository
	UserRepository() domain.UserRepository
	RefreshTokenRepository() domain.RefreshTokenRepository
	RevokedTokenRepository() domain.RevokedTokenRepository
}
