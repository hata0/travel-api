package postgres

import (
	"context"
	"testing"
	"time"
	"travel-api/internal/domain"
	apperr "travel-api/internal/domain/errors"
	postgres "travel-api/internal/infrastructure/postgres/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestRevokedToken はテスト用のRevokedTokenドメインオブジェクトを生成するヘルパー関数
func createTestRevokedToken(t *testing.T, userID domain.UserID, jti string, expiresAt, revokedAt time.Time) domain.RevokedToken {
	t.Helper()
	id, err := domain.NewRevokedTokenID(uuid.New().String()) // 動的にUUIDを生成
	require.NoError(t, err)
	return domain.NewRevokedToken(id, userID, jti, expiresAt, revokedAt)
}

// insertTestRevokedToken はテスト用のRevokedTokenドメインオブジェクトをデータベースに挿入するヘルパー関数
func insertTestRevokedToken(t *testing.T, ctx context.Context, db postgres.DBTX, token domain.RevokedToken) {
	t.Helper()
	queries := postgres.New(db)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(token.ID.String())

	var validatedUserID pgtype.UUID
	_ = validatedUserID.Scan(token.UserID.String())

	var validatedExpiresAt pgtype.Timestamptz
	_ = validatedExpiresAt.Scan(token.ExpiresAt)

	var validatedRevokedAt pgtype.Timestamptz
	_ = validatedRevokedAt.Scan(token.RevokedAt)

	err := queries.CreateRevokedToken(ctx, postgres.CreateRevokedTokenParams{
		ID:        validatedId,
		UserID:    validatedUserID,
		TokenJti:  token.TokenJTI,
		ExpiresAt: validatedExpiresAt,
		RevokedAt: validatedRevokedAt,
	})
	require.NoError(t, err)
}

// getRevokedTokenFromDB はデータベースから直接RevokedTokenレコードを取得するヘルパー関数
func getRevokedTokenFromDB(t *testing.T, ctx context.Context, db postgres.DBTX, jti string) (postgres.RevokedToken, error) {
	t.Helper()
	queries := postgres.New(db)
	return queries.GetRevokedTokenByJTI(ctx, jti)
}

func TestRevokedTokenPostgresRepository_Create(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewRevokedTokenPostgresRepository(dbConn)

	// テスト用のユーザーを作成し、DBに挿入
	user := createTestUser(t, "testuser_revoked", "revoked@example.com", "hashedpass_revoked", time.Now(), time.Now())
	insertTestUser(t, ctx, dbConn, user)

	t.Run("正常系: 新しいレコードが作成される", func(t *testing.T) {
		jti := "test-jti-1"
		expiresAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		revokedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		token := createTestRevokedToken(t, user.ID, jti, expiresAt, revokedAt)

		err := repo.Create(ctx, token)
		assert.NoError(t, err)

		// DBから直接取得して検証
		createdRecord, err := getRevokedTokenFromDB(t, ctx, dbConn, jti)
		assert.NoError(t, err)
		assert.Equal(t, token.ID.String(), createdRecord.ID.String())
		assert.Equal(t, token.UserID.String(), createdRecord.UserID.String())
		assert.Equal(t, token.TokenJTI, createdRecord.TokenJti)
		assert.True(t, token.ExpiresAt.Equal(createdRecord.ExpiresAt.Time))
		assert.True(t, token.RevokedAt.Equal(createdRecord.RevokedAt.Time))
	})
}

func TestRevokedTokenPostgresRepository_FindByJTI(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewRevokedTokenPostgresRepository(dbConn)

	// テスト用のユーザーを作成し、DBに挿入
	user := createTestUser(t, "testuser_find", "find@example.com", "hashedpass_find", time.Now(), time.Now())
	insertTestUser(t, ctx, dbConn, user)

	jti := "test-jti-2"
	expiresAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	revokedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	token := createTestRevokedToken(t, user.ID, jti, expiresAt, revokedAt)
	insertTestRevokedToken(t, ctx, dbConn, token)

	t.Run("正常系: JTIでトークンが見つかる", func(t *testing.T) {
		foundToken, err := repo.FindByJTI(ctx, jti)
		assert.NoError(t, err)
		assert.Equal(t, token.ID.String(), foundToken.ID.String())
		assert.Equal(t, token.UserID.String(), foundToken.UserID.String())
		assert.Equal(t, token.TokenJTI, foundToken.TokenJTI)
	})

	t.Run("異常系: JTIでトークンが見つからない", func(t *testing.T) {
		_, err := repo.FindByJTI(ctx, "nonexistent-jti")
		assert.ErrorIs(t, err, apperr.ErrTokenNotFound)
	})
}
