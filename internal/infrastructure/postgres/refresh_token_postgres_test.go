package postgres

import (
	"context"
	"testing"
	"time"
	"travel-api/internal/domain"
	"travel-api/internal/domain/shared/app_error"
	postgres "travel-api/internal/infrastructure/postgres/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestRefreshToken はテスト用のRefreshTokenドメインオブジェクトを生成するヘルパー関数
func createTestRefreshToken(t *testing.T, userID domain.UserID, token string, expiresAt, createdAt time.Time) domain.RefreshToken {
	t.Helper()
	id, err := domain.NewRefreshTokenID(uuid.New().String()) // 動的にUUIDを生成
	require.NoError(t, err)
	return domain.NewRefreshToken(id, userID, token, expiresAt, createdAt)
}

// insertTestRefreshToken はテスト用のRefreshTokenドメインオブジェクトをデータベースに挿入するヘルパー関数
func insertTestRefreshToken(t *testing.T, ctx context.Context, db postgres.DBTX, token domain.RefreshToken) {
	t.Helper()
	queries := postgres.New(db)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(token.ID.String())

	var validatedUserID pgtype.UUID
	_ = validatedUserID.Scan(token.UserID.String())

	var validatedExpiresAt pgtype.Timestamptz
	_ = validatedExpiresAt.Scan(token.ExpiresAt)

	var validatedCreatedAt pgtype.Timestamptz
	_ = validatedCreatedAt.Scan(token.CreatedAt)

	err := queries.CreateRefreshToken(ctx, postgres.CreateRefreshTokenParams{
		ID:        validatedId,
		UserID:    validatedUserID,
		Token:     token.Token,
		ExpiresAt: validatedExpiresAt,
		CreatedAt: validatedCreatedAt,
	})
	require.NoError(t, err)
}

// getRefreshTokenFromDB はデータベースから直接RefreshTokenレコードを取得するヘルパー関数
func getRefreshTokenFromDB(t *testing.T, ctx context.Context, db postgres.DBTX, token string) (postgres.RefreshToken, error) {
	t.Helper()
	queries := postgres.New(db)
	return queries.FindRefreshTokenByToken(ctx, token)
}

func TestRefreshTokenPostgresRepository_Create(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewRefreshTokenPostgresRepository(dbConn)

	// テスト用のユーザーを作成し、DBに挿入
	user := createTestUser(t, "testuser", "test@example.com", "hashedpass", time.Now(), time.Now())
	insertTestUser(t, ctx, dbConn, user)

	t.Run("正常系: 新しいレコードが作成される", func(t *testing.T) {
		tokenStr := "test-refresh-token-1"
		expiresAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		token := createTestRefreshToken(t, user.ID, tokenStr, expiresAt, createdAt)

		err := repo.Create(ctx, token)
		assert.NoError(t, err)

		// DBから直接取得して検証
		createdRecord, err := getRefreshTokenFromDB(t, ctx, dbConn, tokenStr)
		assert.NoError(t, err)
		assert.Equal(t, token.ID.String(), createdRecord.ID.String())
		assert.Equal(t, token.UserID.String(), createdRecord.UserID.String())
		assert.Equal(t, token.Token, createdRecord.Token)
		assert.True(t, token.ExpiresAt.Equal(createdRecord.ExpiresAt.Time))
		assert.True(t, token.CreatedAt.Equal(createdRecord.CreatedAt.Time))
	})

	t.Run("異常系: 重複するトークンで作成", func(t *testing.T) {
		tokenStr := "duplicate-refresh-token"
		expiresAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		token1 := createTestRefreshToken(t, user.ID, tokenStr, expiresAt, createdAt)
		insertTestRefreshToken(t, ctx, dbConn, token1)

		token2 := createTestRefreshToken(t, user.ID, tokenStr, expiresAt, createdAt)

		err := repo.Create(ctx, token2)
		assert.ErrorIs(t, err, app_error.ErrTokenAlreadyExists)
	})
}

func TestRefreshTokenPostgresRepository_FindByToken(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewRefreshTokenPostgresRepository(dbConn)

	// テスト用のユーザーを作成し、DBに挿入
	user := createTestUser(t, "testuser2", "test2@example.com", "hashedpass2", time.Now(), time.Now())
	insertTestUser(t, ctx, dbConn, user)

	tokenStr := "test-refresh-token-2"
	expiresAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	token := createTestRefreshToken(t, user.ID, tokenStr, expiresAt, createdAt)
	insertTestRefreshToken(t, ctx, dbConn, token)

	t.Run("正常系: トークンでリフレッシュトークンが見つかる", func(t *testing.T) {
		foundToken, err := repo.FindByToken(ctx, tokenStr)
		assert.NoError(t, err)
		assert.Equal(t, token.ID.String(), foundToken.ID.String())
		assert.Equal(t, token.Token, foundToken.Token)
	})

	t.Run("異常系: トークンでリフレッシュトークンが見つからない", func(t *testing.T) {
		_, err := repo.FindByToken(ctx, "nonexistent-token")
		assert.ErrorIs(t, err, app_error.ErrTokenNotFound)
	})
}

func TestRefreshTokenPostgresRepository_Delete(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewRefreshTokenPostgresRepository(dbConn)

	// テスト用のユーザーを作成し、DBに挿入
	user := createTestUser(t, "testuser3", "test3@example.com", "hashedpass3", time.Now(), time.Now())
	insertTestUser(t, ctx, dbConn, user)

	tokenStr := "test-refresh-token-3"
	expiresAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	token := createTestRefreshToken(t, user.ID, tokenStr, expiresAt, createdAt)
	insertTestRefreshToken(t, ctx, dbConn, token)

	t.Run("正常系: リフレッシュトークンが削除される", func(t *testing.T) {
		err := repo.Delete(ctx, token)
		assert.NoError(t, err)

		// DBから直接取得して削除されたことを検証
		_, err = getRefreshTokenFromDB(t, ctx, dbConn, tokenStr)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestRefreshTokenPostgresRepository_DeleteByUserID(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewRefreshTokenPostgresRepository(dbConn)

	// テスト用のユーザーを作成し、DBに挿入
	user1 := createTestUser(t, "testuser4", "test4@example.com", "hashedpass4", time.Now(), time.Now())
	insertTestUser(t, ctx, dbConn, user1)
	user2 := createTestUser(t, "testuser5", "test5@example.com", "hashedpass5", time.Now(), time.Now())
	insertTestUser(t, ctx, dbConn, user2)

	// ユーザー1のトークンを2つ作成
	token1 := createTestRefreshToken(t, user1.ID, "user1-token1", time.Now().Add(time.Hour), time.Now())
	token2 := createTestRefreshToken(t, user1.ID, "user1-token2", time.Now().Add(time.Hour), time.Now())
	insertTestRefreshToken(t, ctx, dbConn, token1)
	insertTestRefreshToken(t, ctx, dbConn, token2)

	// ユーザー2のトークンを1つ作成
	token3 := createTestRefreshToken(t, user2.ID, "user2-token1", time.Now().Add(time.Hour), time.Now())
	insertTestRefreshToken(t, ctx, dbConn, token3)

	t.Run("正常系: 指定したユーザーIDのトークンのみが削除される", func(t *testing.T) {
		err := repo.DeleteByUserID(ctx, user1.ID)
		assert.NoError(t, err)

		// user1のトークンが削除されたことを確認
		_, err = getRefreshTokenFromDB(t, ctx, dbConn, "user1-token1")
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		_, err = getRefreshTokenFromDB(t, ctx, dbConn, "user1-token2")
		assert.ErrorIs(t, err, pgx.ErrNoRows)

		// user2のトークンが削除されていないことを確認
		_, err = getRefreshTokenFromDB(t, ctx, dbConn, "user2-token1")
		assert.NoError(t, err)
	})
}
