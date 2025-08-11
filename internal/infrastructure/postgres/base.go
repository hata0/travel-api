package postgres

import (
	"context"

	postgres "github.com/hata0/travel-api/internal/infrastructure/postgres/generated"
	"github.com/hata0/travel-api/internal/infrastructure/postgres/mapper"
)

// BasePostgresRepository はすべてのPostgresリポジトリに共通の機能を提供します。
type BasePostgresRepository struct {
	db         postgres.DBTX
	typeMapper *mapper.PostgreSQLTypeMapper
}

// NewBasePostgresRepository は新しいBaseRepositoryのインスタンスを作成します。
func NewBasePostgresRepository(db postgres.DBTX) *BasePostgresRepository {
	return &BasePostgresRepository{
		db:         db,
		typeMapper: mapper.NewPostgreSQLTypeMapper(),
	}
}

// getQueries はコンテキストからトランザクションを取得し、それに応じたQueriesインスタンスを返します。
func (r *BasePostgresRepository) GetQueries(ctx context.Context) *postgres.Queries {
	if tx, ok := GetTxFromContext(ctx); ok {
		return postgres.New(tx) // トランザクションがあればトランザクション対応のQueriesを返す
	}
	return postgres.New(r.db) // なければ通常のDBプール対応のQueriesを返す
}

func (r *BasePostgresRepository) GetTypeMapper() *mapper.PostgreSQLTypeMapper {
	return r.typeMapper
}
