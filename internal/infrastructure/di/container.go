package di

import (
	"github.com/hata0/travel-api/internal/adapter/handler"
	"github.com/hata0/travel-api/internal/domain/shared/clock"
	"github.com/hata0/travel-api/internal/domain/shared/transaction_manager"
	"github.com/hata0/travel-api/internal/domain/shared/uuid"
	"github.com/hata0/travel-api/internal/infrastructure/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Container は全ての依存関係を管理するメインコンテナ
type Container struct {
	config       config.Config
	db           *pgxpool.Pool
	services     ServiceProvider
	repositories RepositoryProvider
	usecases     *Usecases
	handlers     *Handlers
}

// NewContainer は本番用のコンテナを作成する
func NewContainer(db *pgxpool.Pool, cfg config.Config) *Container {
	services := NewServices(db)
	repositories := NewRepositories(db)
	usecases := NewUsecases(repositories, services, cfg)
	handlers := NewHandlers(usecases)

	return &Container{
		config:       cfg,
		db:           db,
		services:     services,
		repositories: repositories,
		usecases:     usecases,
		handlers:     handlers,
	}
}

// NewTestContainer はテスト用のコンテナを作成する（依存関係を外から注入可能）
func NewTestContainer(
	services ServiceProvider,
	repositories RepositoryProvider,
	cfg config.Config,
	db *pgxpool.Pool,
) *Container {
	usecases := NewUsecases(repositories, services, cfg)
	handlers := NewHandlers(usecases)

	return &Container{
		config:       cfg,
		db:           db,
		services:     services,
		repositories: repositories,
		usecases:     usecases,
		handlers:     handlers,
	}
}

// Close はコンテナが管理するリソースをクローズします。
func (c *Container) Close() error {
	if c.db != nil {
		c.db.Close()
	}
	return nil
}

// HandlerProvider インターフェースの実装
func (c *Container) TripHandler() *handler.TripHandler {
	return c.handlers.TripHandler()
}

func (c *Container) AuthHandler() *handler.AuthHandler {
	return c.handlers.AuthHandler()
}

// ServiceProvider インターフェースの実装
func (c *Container) Clock() clock.Clock {
	return c.services.Clock()
}

func (c *Container) UUIDGenerator() uuid.UUIDGenerator {
	return c.services.UUIDGenerator()
}

func (c *Container) TransactionManager() transaction_manager.TransactionManager {
	return c.services.TransactionManager()
}
