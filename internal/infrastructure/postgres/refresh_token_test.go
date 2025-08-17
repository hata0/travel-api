package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	refreshtoken "github.com/hata0/travel-api/internal/domain/refresh_token"
	"github.com/hata0/travel-api/internal/domain/user"
	postgres "github.com/hata0/travel-api/internal/infrastructure/postgres/generated"
	"github.com/hata0/travel-api/internal/infrastructure/postgres/mapper"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testRefreshToken テスト用のRefreshToken構造体
type testRefreshToken struct {
	ID        refreshtoken.RefreshTokenID
	UserID    user.UserID
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// newTestRefreshToken テスト用のRefreshTokenを生成する
func newTestRefreshToken(token string, userID user.UserID) testRefreshToken {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return testRefreshToken{
		ID:        refreshtoken.NewRefreshTokenID(uuid.New().String()),
		UserID:    userID,
		Token:     token,
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now,
	}
}

// toDomainRefreshToken ドメインオブジェクトに変換する
func (trt testRefreshToken) toDomainRefreshToken() *refreshtoken.RefreshToken {
	return refreshtoken.NewRefreshToken(trt.ID, trt.UserID, trt.Token, trt.ExpiresAt, trt.CreatedAt)
}

// refreshTokenTestSuite テスト用の共通セットアップ
type refreshTokenTestSuite struct {
	ctx     context.Context
	tx      pgx.Tx
	repo    refreshtoken.RefreshTokenRepository
	queries *postgres.Queries
	mapper  *mapper.PostgreSQLTypeMapper
}

// newRefreshTokenTestSuite テストスイートを作成する（トランザクション分離）
func newRefreshTokenTestSuite(t *testing.T) *refreshTokenTestSuite {
	t.Helper()

	ctx := context.Background()
	db := setupDB(t, ctx)

	// サブテスト用のトランザクションを開始
	tx, err := db.Begin(ctx)
	require.NoError(t, err, "トランザクション開始に失敗")

	// サブテスト終了時にロールバック
	t.Cleanup(func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			t.Logf("トランザクションロールバック時の警告: %v", err)
		}
	})

	return &refreshTokenTestSuite{
		ctx:     ctx,
		tx:      tx,
		repo:    NewRefreshTokenPostgresRepository(tx), // トランザクションを渡す
		queries: postgres.New(tx),
		mapper:  mapper.NewPostgreSQLTypeMapper(),
	}
}

// createUserInDB データベースに直接Userを作成する (user_test.goからコピー)
func (s *refreshTokenTestSuite) createUserInDB(t *testing.T, user testUser) {
	t.Helper()

	pgUUID, err := s.mapper.ToUUID(user.ID.String())
	require.NoError(t, err, "UUID変換に失敗")
	pgCreatedAt, err := s.mapper.ToTimestamp(user.CreatedAt)
	require.NoError(t, err, "CreatedAt変換に失敗")
	pgUpdatedAt, err := s.mapper.ToTimestamp(user.UpdatedAt)
	require.NoError(t, err, "UpdatedAt変換に失敗")

	err = s.queries.CreateUser(s.ctx, postgres.CreateUserParams{
		ID:           pgUUID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    pgCreatedAt,
		UpdatedAt:    pgUpdatedAt,
	})
	require.NoError(t, err, "テストデータの作成に失敗")
}

// createRefreshTokenInDB データベースに直接RefreshTokenを作成する
func (s *refreshTokenTestSuite) createRefreshTokenInDB(t *testing.T, token testRefreshToken) {
	t.Helper()

	pgID, err := s.mapper.ToUUID(token.ID.String())
	require.NoError(t, err, "ID変換に失敗")
	pgUserID, err := s.mapper.ToUUID(token.UserID.String())
	require.NoError(t, err, "UserID変換に失敗")
	pgExpiresAt, err := s.mapper.ToTimestamp(token.ExpiresAt)
	require.NoError(t, err, "ExpiresAt変換に失敗")
	pgCreatedAt, err := s.mapper.ToTimestamp(token.CreatedAt)
	require.NoError(t, err, "CreatedAt変換に失敗")

	err = s.queries.CreateRefreshToken(s.ctx, postgres.CreateRefreshTokenParams{
		ID:        pgID,
		UserID:    pgUserID,
		Token:     token.Token,
		ExpiresAt: pgExpiresAt,
		CreatedAt: pgCreatedAt,
	})
	require.NoError(t, err, "テストデータの作成に失敗")
}

// getRefreshTokenFromDB データベースから直接RefreshTokenを取得する
func (s *refreshTokenTestSuite) getRefreshTokenFromDB(t *testing.T, token string) (*postgres.RefreshToken, error) {
	t.Helper()

	record, err := s.queries.FindRefreshTokenByToken(s.ctx, token)

	return &record, err
}

// assertRefreshTokenEquals RefreshTokenの等価性をアサートする
func assertRefreshTokenEquals(t *testing.T, expected testRefreshToken, actual *refreshtoken.RefreshToken) {
	t.Helper()
	assert.Equal(t, expected.ID, actual.ID(), "IDが一致すること")
	assert.Equal(t, expected.UserID, actual.UserID(), "UserIDが一致すること")
	assert.Equal(t, expected.Token, actual.Token(), "Tokenが一致すること")
	assert.WithinDuration(t, expected.ExpiresAt, actual.ExpiresAt(), time.Second,
		"ExpiresAtがほぼ一致すること (expected: %v, actual: %v)", expected.ExpiresAt, actual.ExpiresAt())
	assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt(), time.Second,
		"CreatedAtがほぼ一致すること (expected: %v, actual: %v)", expected.CreatedAt, actual.CreatedAt())
}

// assertRefreshTokenExistsInDB データベースにRefreshTokenが存在することをアサートする
func (s *refreshTokenTestSuite) assertRefreshTokenExistsInDB(t *testing.T, expected testRefreshToken) {
	t.Helper()

	record, err := s.getRefreshTokenFromDB(t, expected.Token)
	require.NoError(t, err, "データベースにRefreshTokenが存在すること")

	actualID, err := s.mapper.FromUUID(record.ID)
	require.NoError(t, err, "ID変換に失敗")
	assert.Equal(t, expected.ID.String(), actualID, "IDが一致すること")

	actualUserID, err := s.mapper.FromUUID(record.UserID)
	require.NoError(t, err, "UserID変換に失敗")
	assert.Equal(t, expected.UserID.String(), actualUserID, "UserIDが一致すること")

	assert.Equal(t, expected.Token, record.Token, "Tokenが一致すること")

	actualExpiresAt, err := s.mapper.FromTimestamp(record.ExpiresAt)
	require.NoError(t, err, "ExpiresAt変換に失敗")
	assert.WithinDuration(t, expected.ExpiresAt, actualExpiresAt, time.Second, "ExpiresAtがほぼ一致すること")

	actualCreatedAt, err := s.mapper.FromTimestamp(record.CreatedAt)
	require.NoError(t, err, "CreatedAt変換に失敗")
	assert.WithinDuration(t, expected.CreatedAt, actualCreatedAt, time.Second, "CreatedAtがほぼ一致すること")
}

// assertRefreshTokenNotExistsInDB データベースにRefreshTokenが存在しないことをアサートする
func (s *refreshTokenTestSuite) assertRefreshTokenNotExistsInDB(t *testing.T, token string) {
	t.Helper()

	_, err := s.getRefreshTokenFromDB(t, token)
	assert.ErrorIs(t, err, pgx.ErrNoRows,
		"データベースにRefreshTokenが存在しないこと")
}

func TestRefreshTokenPostgresRepository_NewRefreshTokenPostgresRepository(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t, ctx)

	repo := NewRefreshTokenPostgresRepository(db)
	assert.NotNil(t, repo, "リポジトリインスタンスがnilであってはならない")
}

func TestRefreshTokenPostgresRepository_Create(t *testing.T) {
	t.Run("新しいRefreshTokenを正常に作成できること", func(t *testing.T) {
		suite := newRefreshTokenTestSuite(t)

		// Given: 関連するUserと新しいRefreshToken
		testUser := newTestUser("testuser-for-refresh-create", "test-refresh-create@example.com")
		suite.createUserInDB(t, testUser) // Userを作成

		testToken := newTestRefreshToken("token-create-1", testUser.ID) // UserIDを渡す
		domainToken := testToken.toDomainRefreshToken()

		// When: RefreshTokenを作成する
		err := suite.repo.Create(suite.ctx, domainToken)

		// Then: RefreshTokenが正常に作成される
		require.NoError(t, err, "Createでエラーが発生してはならない")
		suite.assertRefreshTokenExistsInDB(t, testToken)
	})

	t.Run("nilのRefreshTokenでInternalErrorが返されること", func(t *testing.T) {
		suite := newRefreshTokenTestSuite(t)

		// When: nilのRefreshTokenを作成する
		err := suite.repo.Create(suite.ctx, nil)

		// Then: InternalErrorが返される
		assert.ErrorIs(t, err, apperr.NewInternalError(""),
			"InternalErrorが返されるべき")
	})

	t.Run("重複するTokenでエラーが返されること", func(t *testing.T) {
		suite := newRefreshTokenTestSuite(t)

		// Given: 関連するUserと既存のRefreshToken
		testUser := newTestUser("testuser-for-refresh-duplicate", "test-refresh-duplicate@example.com")
		suite.createUserInDB(t, testUser) // Userを作成

		existingToken := newTestRefreshToken("token-duplicate", testUser.ID)
		suite.createRefreshTokenInDB(t, existingToken)

		// When: 同じTokenで別のRefreshTokenを作成する
		duplicateToken := newTestRefreshToken("token-duplicate", testUser.ID)
		duplicateToken.ID = refreshtoken.NewRefreshTokenID(uuid.New().String())

		err := suite.repo.Create(suite.ctx, duplicateToken.toDomainRefreshToken())

		// Then: エラーが返される
		assert.Error(t, err, "重複Tokenでの作成はエラーになるべき")
		assert.ErrorIs(t, err, apperr.NewInternalError(""),
			"InternalErrorが返されるべき")
	})
}

func TestRefreshTokenPostgresRepository_FindByToken(t *testing.T) {
	t.Run("存在するTokenでRefreshTokenを取得できること", func(t *testing.T) {
		suite := newRefreshTokenTestSuite(t)

		// Given: 関連するUserとデータベースにRefreshTokenが存在する
		testUser := newTestUser("testuser-for-refresh-find", "test-refresh-find@example.com")
		suite.createUserInDB(t, testUser) // Userを作成

		testToken := newTestRefreshToken("token-find-1", testUser.ID)
		suite.createRefreshTokenInDB(t, testToken)

		// When: FindByTokenでRefreshTokenを取得する
		foundToken, err := suite.repo.FindByToken(suite.ctx, testToken.Token)

		// Then: RefreshTokenが正常に取得できる
		require.NoError(t, err, "FindByTokenでエラーが発生してはならない")
		require.NotNil(t, foundToken, "取得したRefreshTokenがnilであってはならない")
		assertRefreshTokenEquals(t, testToken, foundToken)
	})

	t.Run("存在しないTokenでErrRefreshTokenNotFoundが返されること", func(t *testing.T) {
		suite := newRefreshTokenTestSuite(t)

		// Given: 存在しないToken
		nonExistentToken := "non-existent-token"

		// When: 存在しないTokenでRefreshTokenを取得する
		_, err := suite.repo.FindByToken(suite.ctx, nonExistentToken)

		// Then: RefreshTokenNotFoundが返される
		assert.ErrorIs(t, err, refreshtoken.NewRefreshTokenNotFoundError(),
			"RefreshTokenNotFoundが返されるべき")
	})
}

func TestRefreshTokenPostgresRepository_Delete(t *testing.T) {
	t.Run("存在するIDのRefreshTokenを正常に削除できること", func(t *testing.T) {
		suite := newRefreshTokenTestSuite(t)

		// Given: 関連するUserと既存のRefreshToken
		testUser := newTestUser("testuser-for-refresh-delete", "test-refresh-delete@example.com")
		suite.createUserInDB(t, testUser) // Userを作成

		existingToken := newTestRefreshToken("token-delete-1", testUser.ID)
		suite.createRefreshTokenInDB(t, existingToken)

		// When: RefreshTokenを削除する
		err := suite.repo.Delete(suite.ctx, existingToken.ID)

		// Then: RefreshTokenが正常に削除される
		require.NoError(t, err, "Deleteでエラーが発生してはならない")
		suite.assertRefreshTokenNotExistsInDB(t, existingToken.Token)
	})

	t.Run("存在しないIDでErrRefreshTokenNotFoundが返されること", func(t *testing.T) {
		suite := newRefreshTokenTestSuite(t)

		// Given: 存在しないIDを持つRefreshToken
		nonExistentID := refreshtoken.NewRefreshTokenID(uuid.New().String())

		// When: 存在しないIDでRefreshTokenを削除する
		err := suite.repo.Delete(suite.ctx, nonExistentID)

		// Then: RefreshTokenNotFoundが返される
		assert.ErrorIs(t, err, refreshtoken.NewRefreshTokenNotFoundError(),
			"RefreshTokenNotFoundが返されるべき")
	})
}

func TestRefreshTokenPostgresRepository_DeleteByUserID(t *testing.T) {
	t.Run("存在するUserIDのRefreshTokenを正常に削除できること", func(t *testing.T) {
		suite := newRefreshTokenTestSuite(t)

		// Given: 関連するUserと複数のRefreshToken
		testUser := newTestUser("testuser-for-deletebyuserid", "test-deletebyuserid@example.com")
		suite.createUserInDB(t, testUser) // Userを作成

		token1 := newTestRefreshToken("token-user-1", testUser.ID)
		token2 := newTestRefreshToken("token-user-2", testUser.ID)
		token3 := newTestRefreshToken("token-user-3", testUser.ID)
		suite.createRefreshTokenInDB(t, token1)
		suite.createRefreshTokenInDB(t, token2)
		suite.createRefreshTokenInDB(t, token3)

		// When: UserIDでRefreshTokenを削除する
		err := suite.repo.DeleteByUserID(suite.ctx, testUser.ID)

		// Then: RefreshTokenが正常に削除される
		require.NoError(t, err, "DeleteByUserIDでエラーが発生してはならない")
		suite.assertRefreshTokenNotExistsInDB(t, token1.Token)
		suite.assertRefreshTokenNotExistsInDB(t, token2.Token)
		suite.assertRefreshTokenNotExistsInDB(t, token3.Token)
	})

	t.Run("存在しないUserIDの場合エラーにならないこと", func(t *testing.T) {
		suite := newRefreshTokenTestSuite(t)

		// Given: 存在しないUserID
		nonExistentUserID := user.NewUserID(uuid.New().String())

		// When: 存在しないUserIDでRefreshTokenを削除する
		err := suite.repo.DeleteByUserID(suite.ctx, nonExistentUserID)

		// Then: エラーにならない
		assert.NoError(t, err, "存在しないUserIDの削除はエラーにならないべき")
	})

	t.Run("RefreshTokenが存在しないUserIDの場合エラーにならないこと", func(t *testing.T) {
		suite := newRefreshTokenTestSuite(t)

		// Given: RefreshTokenを持たないUser
		testUser := newTestUser("testuser-no-tokens", "test-no-tokens@example.com")
		suite.createUserInDB(t, testUser) // Userを作成

		// When: RefreshTokenを持たないUserIDで削除する
		err := suite.repo.DeleteByUserID(suite.ctx, testUser.ID)

		// Then: エラーにならない
		assert.NoError(t, err, "RefreshTokenを持たないUserIDの削除はエラーにならないべき")
	})
}
