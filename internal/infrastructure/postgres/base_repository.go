package postgres

import (
	"context"
	postgres "travel-api/internal/infrastructure/postgres/generated"
)

// BaseRepository はすべてのPostgresリポジトリに共通の機能を提供します。
type BaseRepository struct {
	db postgres.DBTX
}

// NewBaseRepository は新しいBaseRepositoryのインスタンスを作成します。
func NewBaseRepository(db postgres.DBTX) *BaseRepository {
	return &BaseRepository{db: db}
}

// getQueries はコンテキストからトランザクションを取得し、それに応じたQueriesインスタンスを返します。
func (r *BaseRepository) getQueries(ctx context.Context) *postgres.Queries {
	if tx, ok := GetTxFromContext(ctx); ok {
		return postgres.New(tx) // トランザクションがあればトランザクション対応のQueriesを返す
	}
	return postgres.New(r.db) // なければ通常のDBプール対応のQueriesを返す
}
