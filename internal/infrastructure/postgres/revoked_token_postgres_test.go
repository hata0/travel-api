package postgres

import (
	"context"
	"testing"
	"time"
	"travel-api/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestRevokedToken はテスト用のRevokedTokenドメインオブジェクトを生成するヘルパー関数
func createTestRevokedToken(t *testing.T, jti string, expiresAt, revokedAt time.Time) domain.RevokedToken {
	t.Helper()
	id, err := domain.NewRevokedTokenID(uuid.New().String()) // 動的にUUIDを生成
	require.NoError(t, err)
	return domain.NewRevokedToken(id, jti, expiresAt, revokedAt)
}

// insertTestRevokedToken はテスト用のRevokedTokenドメインオブジェクトをデータベースに挿入するヘルパー関数
func insertTestRevokedToken(t *testing.T, ctx context.Context, db DBTX, token domain.RevokedToken) {
	t.Helper()
	queries := New(db)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(token.ID.String())

	var validatedExpiresAt pgtype.Timestamptz
	_ = validatedExpiresAt.Scan(token.ExpiresAt)

	var validatedRevokedAt pgtype.Timestamptz
	_ = validatedRevokedAt.Scan(token.RevokedAt)

	err := queries.CreateRevokedToken(ctx, CreateRevokedTokenParams{
		ID:        validatedId,
		TokenJti:  token.TokenJTI,
		ExpiresAt: validatedExpiresAt,
		RevokedAt: validatedRevokedAt,
	})
	require.NoError(t, err)
}

// getRevokedTokenFromDB はデータベースから直接RevokedTokenレコードを取得するヘルパー関数
func getRevokedTokenFromDB(t *testing.T, ctx context.Context, db DBTX, jti string) (RevokedToken, error) {
	t.Helper()
	queries := New(db)
	return queries.GetRevokedTokenByJTI(ctx, jti)
}

func TestRevokedTokenPostgresRepository_Create(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewRevokedTokenPostgresRepository(dbConn)

	t.Run("正常系: 新しいレコードが作成される", func(t *testing.T) {
		jti := "test-jti-1"
		expiresAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		revokedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		token := createTestRevokedToken(t, jti, expiresAt, revokedAt)

		err := repo.Create(ctx, token)
		assert.NoError(t, err)

		// DBから直接取得して検証
		createdRecord, err := getRevokedTokenFromDB(t, ctx, dbConn, jti)
		assert.NoError(t, err)
		assert.Equal(t, token.ID.String(), createdRecord.ID.String())
		assert.Equal(t, token.TokenJTI, createdRecord.TokenJti)
		assert.True(t, token.ExpiresAt.Equal(createdRecord.ExpiresAt.Time))
		assert.True(t, token.RevokedAt.Equal(createdRecord.RevokedAt.Time))
	})

	t.Run("異常系: 重複するJTIで作成", func(t *testing.T) {
		jti := "duplicate-jti"
		expiresAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		revokedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		token1 := createTestRevokedToken(t, jti, expiresAt, revokedAt)
		insertTestRevokedToken(t, ctx, dbConn, token1)

		token2 := createTestRevokedToken(t, jti, expiresAt, revokedAt)

		err := repo.Create(ctx, token2)
		assert.Error(t, err)
		// PostgreSQLの重複キーエラーはpgx.PgErrorとして返されることが多い
		// assert.IsType(t, &pgx.PgError{}, err)
	})
}

func TestRevokedTokenPostgresRepository_FindByJTI(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewRevokedTokenPostgresRepository(dbConn)

	jti := "test-jti-2"
	expiresAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	revokedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	token := createTestRevokedToken(t, jti, expiresAt, revokedAt)
	insertTestRevokedToken(t, ctx, dbConn, token)

	t.Run("正常系: JTIでトークンが見つかる", func(t *testing.T) {
		foundToken, err := repo.FindByJTI(ctx, jti)
		assert.NoError(t, err)
		assert.Equal(t, token.ID.String(), foundToken.ID.String())
		assert.Equal(t, token.TokenJTI, foundToken.TokenJTI)
	})

	t.Run("異常系: JTIでトークンが見つからない", func(t *testing.T) {
		_, err := repo.FindByJTI(ctx, "nonexistent-jti")
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}
