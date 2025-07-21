package postgres

import (
	"context"
	"testing"
	"time"
	"travel-api/internal/domain"
	"travel-api/internal/domain/shared/app_error"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestUser はテスト用のUserドメインオブジェクトを生成するヘルパー関数
func createTestUser(t *testing.T, username, email, passwordHash string, createdAt, updatedAt time.Time) domain.User {
	t.Helper()
	id, err := domain.NewUserID(uuid.New().String()) // 動的にUUIDを生成
	require.NoError(t, err)
	return domain.NewUser(id, username, email, passwordHash, createdAt, updatedAt)
}

// insertTestUser はテスト用のUserドメインオブジェクトをデータベースに挿入するヘルパー関数
func insertTestUser(t *testing.T, ctx context.Context, db DBTX, user domain.User) {
	t.Helper()
	queries := New(db)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(user.ID.String())

	var validatedCreatedAt pgtype.Timestamptz
	_ = validatedCreatedAt.Scan(user.CreatedAt)

	var validatedUpdatedAt pgtype.Timestamptz
	_ = validatedUpdatedAt.Scan(user.UpdatedAt)

	err := queries.CreateUser(ctx, CreateUserParams{
		ID:           validatedId,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    validatedCreatedAt,
		UpdatedAt:    validatedUpdatedAt,
	})
	require.NoError(t, err)
}

// getUserFromDB はデータベースから直接Userレコードを取得するヘルパー関数
func getUserFromDB(t *testing.T, ctx context.Context, db DBTX, idStr string) (User, error) {
	t.Helper()
	queries := New(db)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(idStr)

	return queries.GetUser(ctx, validatedId) // GetUserはsqlcで自動生成される想定
}

func TestUserPostgresRepository_Create(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewUserPostgresRepository(dbConn)

	// テスト用のユーザーを作成し、DBに挿入
	user := createTestUser(t, "testuser", "test@example.com", "hashedpass", time.Now(), time.Now())
	insertTestUser(t, ctx, dbConn, user)

	t.Run("正常系: 新しいレコードが作成される", func(t *testing.T) {
		username := "testuser_new"
		email := "test_new@example.com"
		passwordHash := "hashedpassword_new"
		now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		user := createTestUser(t, username, email, passwordHash, now, now)

		err := repo.Create(ctx, user)
		assert.NoError(t, err)

		// DBから直接取得して検証
		createdRecord, err := getUserFromDB(t, ctx, dbConn, user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, user.ID.String(), createdRecord.ID.String())
		assert.Equal(t, user.Username, createdRecord.Username)
		assert.Equal(t, user.Email, createdRecord.Email)
		assert.Equal(t, user.PasswordHash, createdRecord.PasswordHash)
		assert.True(t, user.CreatedAt.Equal(createdRecord.CreatedAt.Time))
		assert.True(t, user.UpdatedAt.Equal(createdRecord.UpdatedAt.Time))
	})

	t.Run("異常系: 重複するメールアドレスで作成", func(t *testing.T) {
		username1 := "user1"
		email := "duplicate@example.com"
		passwordHash := "hashedpassword"
		now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		user1 := createTestUser(t, username1, email, passwordHash, now, now)
		insertTestUser(t, ctx, dbConn, user1)

		username2 := "user2"
		user2 := createTestUser(t, username2, email, passwordHash, now, now)

		err := repo.Create(ctx, user2)
		assert.ErrorIs(t, err, app_error.ErrUserAlreadyExists)
	})

	t.Run("異常系: 重複するユーザー名で作成", func(t *testing.T) {
		username := "duplicate_username"
		email1 := "email1@example.com"
		passwordHash := "hashedpassword"
		now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		user1 := createTestUser(t, username, email1, passwordHash, now, now)
		insertTestUser(t, ctx, dbConn, user1)

		email2 := "email2@example.com"
		user2 := createTestUser(t, username, email2, passwordHash, now, now)

		err := repo.Create(ctx, user2)
		assert.ErrorIs(t, err, app_error.ErrUserAlreadyExists)
	})
}

func TestUserPostgresRepository_FindByEmail(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewUserPostgresRepository(dbConn)

	username := "testuser_email"
	email := "test_email@example.com"
	passwordHash := "hashedpassword_email"
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	user := createTestUser(t, username, email, passwordHash, now, now)
	insertTestUser(t, ctx, dbConn, user)

	t.Run("正常系: メールアドレスでユーザーが見つかる", func(t *testing.T) {
		foundUser, err := repo.FindByEmail(ctx, email)
		assert.NoError(t, err)
		assert.Equal(t, user.ID.String(), foundUser.ID.String())
		assert.Equal(t, user.Email, foundUser.Email)
	})

	t.Run("異常系: メールアドレスでユーザーが見つからない", func(t *testing.T) {
		_, err := repo.FindByEmail(ctx, "nonexistent@example.com")
		assert.ErrorIs(t, err, app_error.ErrUserNotFound)
	})
}

func TestUserPostgresRepository_FindByUsername(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewUserPostgresRepository(dbConn)

	username := "testuser_username"
	email := "test_username@example.com"
	passwordHash := "hashedpassword_username"
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	user := createTestUser(t, username, email, passwordHash, now, now)
	insertTestUser(t, ctx, dbConn, user)

	t.Run("正常系: ユーザー名でユーザーが見つかる", func(t *testing.T) {
		foundUser, err := repo.FindByUsername(ctx, username)
		assert.NoError(t, err)
		assert.Equal(t, user.ID.String(), foundUser.ID.String())
		assert.Equal(t, user.Username, foundUser.Username)
	})

	t.Run("異常系: ユーザー名でユーザーが見つからない", func(t *testing.T) {
		_, err := repo.FindByUsername(ctx, "nonexistentuser")
		assert.ErrorIs(t, err, app_error.ErrUserNotFound)
	})
}

func TestUserPostgresRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewUserPostgresRepository(dbConn)

	username := "testuser_id"
	email := "test_id@example.com"
	passwordHash := "hashedpassword_id"
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	user := createTestUser(t, username, email, passwordHash, now, now)
	insertTestUser(t, ctx, dbConn, user)

	t.Run("正常系: IDでユーザーが見つかる", func(t *testing.T) {
		foundUser, err := repo.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, user.ID.String(), foundUser.ID.String())
		assert.Equal(t, user.Username, foundUser.Username)
	})

	t.Run("異常系: IDでユーザーが見つからない", func(t *testing.T) {
		id, err := domain.NewUserID(uuid.New().String())
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, id)
		assert.ErrorIs(t, err, app_error.ErrUserNotFound)
	})
}
