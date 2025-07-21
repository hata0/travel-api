package injector

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Factory はコンテナの生成を管理する
type Factory struct{}

// NewFactory はファクトリを作成する
func NewFactory() *Factory {
	return &Factory{}
}

// CreateProductionContainer は本番用コンテナを作成する
func (f *Factory) CreateProductionContainer(databaseURL, jwtSecret string) (*Container, error) {
	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	// 接続テスト
	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return NewContainer(db, jwtSecret), nil
}

// CreateTestContainer はテスト用コンテナを作成する
func (f *Factory) CreateTestContainer(
	services ServiceProvider,
	repositories RepositoryProvider,
	jwtSecret string,
	db *pgxpool.Pool,
) *Container {
	config := NewConfig(jwtSecret)
	return NewTestContainer(services, repositories, config, db)
}
