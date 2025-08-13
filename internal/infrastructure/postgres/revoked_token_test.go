package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	postgres "github.com/hata0/travel-api/internal/infrastructure/postgres/generated"
	"github.com/hata0/travel-api/internal/infrastructure/postgres/mapper"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testRevokedToken テスト用のRevokedToken構造体
type testRevokedToken struct {
	ID        domain.RevokedTokenID
	UserID    domain.UserID
	TokenJTI  string
	ExpiresAt time.Time
	RevokedAt time.Time
}

// newTestRevokedToken テスト用のRevokedTokenを生成する (UserIDを引数で受け取るように変更)
func newTestRevokedToken(jti string, userID domain.UserID) testRevokedToken {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return testRevokedToken{
		ID:        domain.NewRevokedTokenID(uuid.New().String()),
		UserID:    userID, // ここを修正
		TokenJTI:  jti,
		ExpiresAt: now.Add(time.Hour),
		RevokedAt: now,
	}
}

// toDomainRevokedToken ドメインオブジェクトに変換する
func (trt testRevokedToken) toDomainRevokedToken() *domain.RevokedToken {
	return domain.NewRevokedToken(trt.ID, trt.UserID, trt.TokenJTI, trt.ExpiresAt, trt.RevokedAt)
}

// revokedTokenTestSuite テスト用の共通セットアップ
type revokedTokenTestSuite struct {
	ctx     context.Context
	tx      pgx.Tx
	repo    domain.RevokedTokenRepository
	queries *postgres.Queries
	mapper  *mapper.PostgreSQLTypeMapper
}

// newRevokedTokenTestSuite テストスイートを作成する（トランザクション分離）
func newRevokedTokenTestSuite(t *testing.T) *revokedTokenTestSuite {
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

	return &revokedTokenTestSuite{
		ctx:     ctx,
		tx:      tx,
		repo:    NewRevokedTokenPostgresRepository(tx), // トランザクションを渡す
		queries: postgres.New(tx),
		mapper:  mapper.NewPostgreSQLTypeMapper(),
	}
}

// createUserInDB データベースに直接Userを作成する (user_test.goからコピー)
func (s *revokedTokenTestSuite) createUserInDB(t *testing.T, user testUser) {
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

// createRevokedTokenInDB データベースに直接RevokedTokenを作成する
func (s *revokedTokenTestSuite) createRevokedTokenInDB(t *testing.T, token testRevokedToken) {
	t.Helper()

	pgID, err := s.mapper.ToUUID(token.ID.String())
	require.NoError(t, err, "ID変換に失敗")
	pgUserID, err := s.mapper.ToUUID(token.UserID.String())
	require.NoError(t, err, "UserID変換に失敗")
	pgExpiresAt, err := s.mapper.ToTimestamp(token.ExpiresAt)
	require.NoError(t, err, "ExpiresAt変換に失敗")
	pgRevokedAt, err := s.mapper.ToTimestamp(token.RevokedAt)
	require.NoError(t, err, "RevokedAt変換に失敗")

	err = s.queries.CreateRevokedToken(s.ctx, postgres.CreateRevokedTokenParams{
		ID:        pgID,
		UserID:    pgUserID,
		TokenJti:  token.TokenJTI,
		ExpiresAt: pgExpiresAt,
		RevokedAt: pgRevokedAt,
	})
	require.NoError(t, err, "テストデータの作成に失敗")
}

// getRevokedTokenFromDB データベースから直接RevokedTokenを取得する
func (s *revokedTokenTestSuite) getRevokedTokenFromDB(t *testing.T, jti string) (*postgres.RevokedToken, error) {
	t.Helper()

	token, err := s.queries.FindRevokedTokenByJTI(s.ctx, jti)

	return &token, err
}

// assertRevokedTokenEquals RevokedTokenの等価性をアサートする
func assertRevokedTokenEquals(t *testing.T, expected testRevokedToken, actual *domain.RevokedToken) {
	t.Helper()
	assert.Equal(t, expected.ID, actual.ID(), "IDが一致すること")
	assert.Equal(t, expected.UserID, actual.UserID(), "UserIDが一致すること")
	assert.Equal(t, expected.TokenJTI, actual.TokenJTI(), "TokenJTIが一致すること")
	assert.WithinDuration(t, expected.ExpiresAt, actual.ExpiresAt(), time.Second,
		"ExpiresAtがほぼ一致すること (expected: %v, actual: %v)", expected.ExpiresAt, actual.ExpiresAt())
	assert.WithinDuration(t, expected.RevokedAt, actual.RevokedAt(), time.Second,
		"RevokedAtがほぼ一致すること (expected: %v, actual: %v)", expected.RevokedAt, actual.RevokedAt())
}

// assertRevokedTokenExistsInDB データベースにRevokedTokenが存在することをアサートする
func (s *revokedTokenTestSuite) assertRevokedTokenExistsInDB(t *testing.T, expected testRevokedToken) {
	t.Helper()

	record, err := s.getRevokedTokenFromDB(t, expected.TokenJTI)
	require.NoError(t, err, "データベースにRevokedTokenが存在すること")

	actualID, err := s.mapper.FromUUID(record.ID)
	require.NoError(t, err, "ID変換に失敗")
	assert.Equal(t, expected.ID.String(), actualID, "IDが一致すること")

	actualUserID, err := s.mapper.FromUUID(record.UserID)
	require.NoError(t, err, "UserID変換に失敗")
	assert.Equal(t, expected.UserID.String(), actualUserID, "UserIDが一致すること")

	assert.Equal(t, expected.TokenJTI, record.TokenJti, "TokenJTIが一致すること")

	actualExpiresAt, err := s.mapper.FromTimestamp(record.ExpiresAt)
	require.NoError(t, err, "ExpiresAt変換に失敗")
	assert.WithinDuration(t, expected.ExpiresAt, actualExpiresAt, time.Second, "ExpiresAtがほぼ一致すること")

	actualRevokedAt, err := s.mapper.FromTimestamp(record.RevokedAt)
	require.NoError(t, err, "RevokedAt変換に失敗")
	assert.WithinDuration(t, expected.RevokedAt, actualRevokedAt, time.Second, "RevokedAtがほぼ一致すること")
}

// assertRevokedTokenNotExistsInDB データベースにRevokedTokenが存在しないことをアサートする
// func (s *revokedTokenTestSuite) assertRevokedTokenNotExistsInDB(t *testing.T, jti string) {
// 	t.Helper()

// 	_, err := s.getRevokedTokenFromDB(t, jti)
// 	assert.ErrorIs(t, err, pgx.ErrNoRows,
// 		"データベースにRevokedTokenが存在しないこと")
// }

func TestRevokedTokenPostgresRepository_NewRevokedTokenPostgresRepository(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t, ctx)

	repo := NewRevokedTokenPostgresRepository(db)
	assert.NotNil(t, repo, "リポジトリインスタンスがnilであってはならない")
}

func TestRevokedTokenPostgresRepository_Create(t *testing.T) {
	t.Run("新しいRevokedTokenを正常に作成できること", func(t *testing.T) {
		suite := newRevokedTokenTestSuite(t)

		// Given: 関連するUserと新しいRevokedToken
		testUser := newTestUser("testuser-for-revoked-create", "test-revoked-create@example.com")
		suite.createUserInDB(t, testUser) // Userを作成

		testToken := newTestRevokedToken("jti-create-1", testUser.ID) // UserIDを渡す
		domainToken := testToken.toDomainRevokedToken()

		// When: RevokedTokenを作成する
		err := suite.repo.Create(suite.ctx, domainToken)

		// Then: RevokedTokenが正常に作成される
		require.NoError(t, err, "Createでエラーが発生してはならない")
		suite.assertRevokedTokenExistsInDB(t, testToken)
	})

	t.Run("nilのRevokedTokenでInternalErrorが返されること", func(t *testing.T) {
		suite := newRevokedTokenTestSuite(t)

		// When: nilのRevokedTokenを作成する
		err := suite.repo.Create(suite.ctx, nil)

		// Then: InternalErrorが返される
		assert.ErrorIs(t, err, apperr.NewInternalError(""),
			"InternalErrorが返されるべき")
	})

	t.Run("重複するJTIでエラーが返されること", func(t *testing.T) {
		suite := newRevokedTokenTestSuite(t)

		// Given: 関連するUserと既存のRevokedToken
		testUser := newTestUser("testuser-for-revoked-duplicate", "test-revoked-duplicate@example.com")
		suite.createUserInDB(t, testUser) // Userを作成

		existingToken := newTestRevokedToken("jti-duplicate", testUser.ID)
		suite.createRevokedTokenInDB(t, existingToken)

		// When: 同じJTIで別のRevokedTokenを作成する
		duplicateToken := newTestRevokedToken("jti-duplicate", testUser.ID)
		// IDとUserIDは異なるがJTIが同じ
		duplicateToken.ID = domain.NewRevokedTokenID(uuid.New().String())
		// duplicateToken.UserID = domain.NewUserID(uuid.New().String()) // UserIDは既存のUserIDを使うためコメントアウト

		err := suite.repo.Create(suite.ctx, duplicateToken.toDomainRevokedToken())

		// Then: エラーが返される
		assert.Error(t, err, "重複JTIでの作成はエラーになるべき")
		assert.ErrorIs(t, err, apperr.NewInternalError(""),
			"InternalErrorが返されるべき")
	})
}

func TestRevokedTokenPostgresRepository_FindByJTI(t *testing.T) {
	t.Run("存在するJTIでRevokedTokenを取得できること", func(t *testing.T) {
		suite := newRevokedTokenTestSuite(t)

		// Given: 関連するUserとデータベースにRevokedTokenが存在する
		testUser := newTestUser("testuser-for-revoked-find", "test-revoked-find@example.com")
		suite.createUserInDB(t, testUser) // Userを作成

		testToken := newTestRevokedToken("jti-find-1", testUser.ID)
		suite.createRevokedTokenInDB(t, testToken)

		// When: FindByJTIでRevokedTokenを取得する
		foundToken, err := suite.repo.FindByJTI(suite.ctx, testToken.TokenJTI)

		// Then: RevokedTokenが正常に取得できる
		require.NoError(t, err, "FindByJTIでエラーが発生してはならない")
		require.NotNil(t, foundToken, "取得したRevokedTokenがnilであってはならない")
		assertRevokedTokenEquals(t, testToken, foundToken)
	})

	t.Run("存在しないJTIでErrRevokedTokenNotFoundが返されること", func(t *testing.T) {
		suite := newRevokedTokenTestSuite(t)

		// Given: 存在しないJTI
		nonExistentJTI := "non-existent-jti"

		// When: 存在しないJTIでRevokedTokenを取得する
		_, err := suite.repo.FindByJTI(suite.ctx, nonExistentJTI)

		// Then: ErrRevokedTokenNotFoundが返される
		assert.ErrorIs(t, err, apperr.ErrRevokedTokenNotFound,
			"ErrRevokedTokenNotFoundが返されるべき")
	})
}
