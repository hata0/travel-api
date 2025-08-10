package di

import (
	"context"
	"fmt"

	"github.com/hata0/travel-api/internal/infrastructure/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Factory はコンテナの生成を管理する
type Factory struct{}

// NewFactory はファクトリを作成する
func NewFactory() *Factory {
	return &Factory{}
}

// CreateProductionContainer は本番用コンテナを作成する
func (f *Factory) CreateProductionContainer(cfg config.Config) (*Container, error) {
	db, err := pgxpool.New(context.Background(), cfg.Database().URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	// 接続テスト
	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return NewContainer(db, cfg), nil
}

// CreateTestContainer はテスト用コンテナを作成する
func (f *Factory) CreateTestContainer(
	services ServiceProvider,
	repositories RepositoryProvider,
	cfg config.Config,
	db *pgxpool.Pool,
) *Container {
	return NewTestContainer(services, repositories, cfg, db)
}
