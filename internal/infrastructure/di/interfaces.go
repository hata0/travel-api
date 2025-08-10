package di

import (
	"github.com/hata0/travel-api/internal/adapter/handler"
	"github.com/hata0/travel-api/internal/domain"
	"github.com/hata0/travel-api/internal/domain/shared/clock"
	"github.com/hata0/travel-api/internal/domain/shared/transaction_manager"
	"github.com/hata0/travel-api/internal/domain/shared/uuid"
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
