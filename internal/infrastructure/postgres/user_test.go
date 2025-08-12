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

// testUser テスト用のUser構造体
type testUser struct {
	ID           domain.UserID
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// newTestUser テスト用のUserを生成する
func newTestUser(username, email string) testUser {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return testUser{
		ID:           domain.NewUserID(uuid.New().String()),
		Username:     username,
		Email:        email,
		PasswordHash: "hashed_password_" + uuid.New().String(),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// toDomainUser ドメインオブジェクトに変換する
func (tu testUser) toDomainUser() *domain.User {
	return domain.NewUser(tu.ID, tu.Username, tu.Email, tu.PasswordHash, tu.CreatedAt, tu.UpdatedAt)
}

// userTestSuite テスト用の共通セットアップ
type userTestSuite struct {
	ctx     context.Context
	tx      pgx.Tx
	repo    domain.UserRepository
	queries *postgres.Queries
	mapper  *mapper.PostgreSQLTypeMapper
}

// newUserTestSuite テストスイートを作成する（トランザクション分離）
func newUserTestSuite(t *testing.T) *userTestSuite {
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

	return &userTestSuite{
		ctx:     ctx,
		tx:      tx,
		repo:    NewUserPostgresRepository(tx), // トランザクションを渡す
		queries: postgres.New(tx),
		mapper:  mapper.NewPostgreSQLTypeMapper(),
	}
}

// createUserInDB データベースに直接Userを作成する
func (s *userTestSuite) createUserInDB(t *testing.T, user testUser) {
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

// getUserFromDB データベースから直接Userを取得する
func (s *userTestSuite) getUserFromDB(t *testing.T, id domain.UserID) (*postgres.User, error) {
	t.Helper()

	pgUUID, err := s.mapper.ToUUID(id.String())
	require.NoError(t, err, "UUID変換に失敗")
	user, err := s.queries.GetUser(s.ctx, pgUUID)

	return &user, err
}

// assertUserEquals Userの等価性をアサートする
func assertUserEquals(t *testing.T, expected testUser, actual *domain.User) {
	t.Helper()
	assert.Equal(t, expected.ID, actual.ID(), "UserIDが一致すること")
	assert.Equal(t, expected.Username, actual.Username(), "Usernameが一致すること")
	assert.Equal(t, expected.Email, actual.Email(), "Emailが一致すること")
	assert.Equal(t, expected.PasswordHash, actual.PasswordHash(), "PasswordHashが一致すること")
	assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt(), time.Second,
		"CreatedAtがほぼ一致すること (expected: %v, actual: %v)", expected.CreatedAt, actual.CreatedAt())
	assert.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt(), time.Second,
		"UpdatedAtがほぼ一致すること (expected: %v, actual: %v)", expected.UpdatedAt, actual.UpdatedAt())
}

// assertUserExistsInDB データベースにUserが存在することをアサートする
func (s *userTestSuite) assertUserExistsInDB(t *testing.T, expected testUser) {
	t.Helper()

	record, err := s.getUserFromDB(t, expected.ID)
	require.NoError(t, err, "データベースにUserが存在すること")

	actualID, err := s.mapper.FromUUID(record.ID)
	require.NoError(t, err, "UUID変換に失敗")
	assert.Equal(t, expected.ID.String(), actualID, "IDが一致すること")
	assert.Equal(t, expected.Username, record.Username, "Usernameが一致すること")
	assert.Equal(t, expected.Email, record.Email, "Emailが一致すること")
	assert.Equal(t, expected.PasswordHash, record.PasswordHash, "PasswordHashが一致すること")

	actualCreatedAt, err := s.mapper.FromTimestamp(record.CreatedAt)
	require.NoError(t, err, "CreatedAt変換に失敗")
	assert.WithinDuration(t, expected.CreatedAt, actualCreatedAt, time.Second, "CreatedAtがほぼ一致すること")

	actualUpdatedAt, err := s.mapper.FromTimestamp(record.UpdatedAt)
	require.NoError(t, err, "UpdatedAt変換に失敗")
	assert.WithinDuration(t, expected.UpdatedAt, actualUpdatedAt, time.Second, "UpdatedAtがほぼ一致すること")
}

// assertUserNotExistsInDB データベースにUserが存在しないことをアサートする
func (s *userTestSuite) assertUserNotExistsInDB(t *testing.T, id domain.UserID) {
	t.Helper()

	_, err := s.getUserFromDB(t, id)
	assert.ErrorIs(t, err, pgx.ErrNoRows,
		"データベースにUserが存在しないこと")
}

func TestUserPostgresRepository_NewUserPostgresRepository(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t, ctx)

	repo := NewUserPostgresRepository(db)
	assert.NotNil(t, repo, "リポジトリインスタンスがnilであってはならない")
}

func TestUserPostgresRepository_Create(t *testing.T) {
	t.Run("新しいUserを正常に作成できること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// Given: 新しいUser
		testUser := newTestUser("testuser", "test@example.com")
		domainUser := testUser.toDomainUser()

		// When: Userを作成する
		err := suite.repo.Create(suite.ctx, domainUser)

		// Then: Userが正常に作成される
		require.NoError(t, err, "Createでエラーが発生してはならない")
		suite.assertUserExistsInDB(t, testUser)
	})

	t.Run("nilのUserでInternalErrorが返されること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// When: nilのUserを作成する
		err := suite.repo.Create(suite.ctx, nil)

		// Then: InternalErrorが返される
		assert.ErrorIs(t, err, apperr.NewInternalError("", nil),
			"InternalErrorが返されるべき")
	})

	t.Run("重複するIDでエラーが返されること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// Given: 既存のUser
		existingUser := newTestUser("existinguser", "existing@example.com")
		suite.createUserInDB(t, existingUser)

		// When: 同じIDで別のUserを作成する
		duplicateUser := testUser{
			ID:           existingUser.ID, // 同じID
			Username:     "duplicateuser",
			Email:        "duplicate@example.com",
			PasswordHash: "new_hashed_password",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		err := suite.repo.Create(suite.ctx, duplicateUser.toDomainUser())

		// Then: エラーが返される
		assert.Error(t, err, "重複IDでの作成はエラーになるべき")
		assert.ErrorIs(t, err, apperr.NewInternalError("", nil),
			"InternalErrorが返されるべき")
	})

	t.Run("重複するEmailでエラーが返されること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// Given: 既存のUser
		existingUser := newTestUser("user1", "duplicate@example.com")
		suite.createUserInDB(t, existingUser)

		// When: 同じEmailで別のUserを作成する
		duplicateEmailUser := newTestUser("user2", "duplicate@example.com")
		err := suite.repo.Create(suite.ctx, duplicateEmailUser.toDomainUser())

		// Then: エラーが返される
		assert.Error(t, err, "重複Emailでの作成はエラーになるべき")
		assert.ErrorIs(t, err, apperr.NewInternalError("", nil),
			"InternalErrorが返されるべき")
	})

	t.Run("重複するUsernameでエラーが返されること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// Given: 既存のUser
		existingUser := newTestUser("duplicateusername", "user1@example.com")
		suite.createUserInDB(t, existingUser)

		// When: 同じUsernameで別のUserを作成する
		duplicateUsernameUser := newTestUser("duplicateusername", "user2@example.com")
		err := suite.repo.Create(suite.ctx, duplicateUsernameUser.toDomainUser())

		// Then: エラーが返される
		assert.Error(t, err, "重複Usernameでの作成はエラーになるべき")
		assert.ErrorIs(t, err, apperr.NewInternalError("", nil),
			"InternalErrorが返されるべき")
	})

	t.Run("複数のUserを連続作成できること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// Given: 複数の新しいUser
		users := []testUser{
			newTestUser("user_seq_1", "seq1@example.com"),
			newTestUser("user_seq_2", "seq2@example.com"),
			newTestUser("user_seq_3", "seq3@example.com"),
		}

		// When: 複数のUserを連続で作成する
		for _, user := range users {
			err := suite.repo.Create(suite.ctx, user.toDomainUser())
			require.NoError(t, err, "連続作成でエラーが発生してはならない")
		}

		// Then: すべてのUserが正常に作成される
		for _, user := range users {
			suite.assertUserExistsInDB(t, user)
		}
	})
}

func TestUserPostgresRepository_FindByEmail(t *testing.T) {
	t.Run("存在するEmailでUserを取得できること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// Given: データベースにUserが存在する
		testUser := newTestUser("findbyemailuser", "findbyemail@example.com")
		suite.createUserInDB(t, testUser)

		// When: FindByEmailでUserを取得する
		foundUser, err := suite.repo.FindByEmail(suite.ctx, testUser.Email)

		// Then: Userが正常に取得できる
		require.NoError(t, err, "FindByEmailでエラーが発生してはならない")
		require.NotNil(t, foundUser, "取得したUserがnilであってはならない")
		assertUserEquals(t, testUser, foundUser)
	})

	t.Run("存在しないEmailでErrUserNotFoundが返されること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// Given: 存在しないEmail
		nonExistentEmail := "nonexistent@example.com"

		// When: 存在しないEmailでUserを取得する
		_, err := suite.repo.FindByEmail(suite.ctx, nonExistentEmail)

		// Then: ErrUserNotFoundが返される
		assert.ErrorIs(t, err, apperr.ErrUserNotFound,
			"ErrUserNotFoundが返されるべき")
	})
}

func TestUserPostgresRepository_FindByUsername(t *testing.T) {
	t.Run("存在するUsernameでUserを取得できること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// Given: データベースにUserが存在する
		testUser := newTestUser("findbyusernameuser", "findbyusername@example.com")
		suite.createUserInDB(t, testUser)

		// When: FindByUsernameでUserを取得する
		foundUser, err := suite.repo.FindByUsername(suite.ctx, testUser.Username)

		// Then: Userが正常に取得できる
		require.NoError(t, err, "FindByUsernameでエラーが発生してはならない")
		require.NotNil(t, foundUser, "取得したUserがnilであってはならない")
		assertUserEquals(t, testUser, foundUser)
	})

	t.Run("存在しないUsernameでErrUserNotFoundが返されること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// Given: 存在しないUsername
		nonExistentUsername := "nonexistentusername"

		// When: 存在しないUsernameでUserを取得する
		_, err := suite.repo.FindByUsername(suite.ctx, nonExistentUsername)

		// Then: ErrUserNotFoundが返される
		assert.ErrorIs(t, err, apperr.ErrUserNotFound,
			"ErrUserNotFoundが返されるべき")
	})
}

func TestUserPostgresRepository_FindByID(t *testing.T) {
	t.Run("存在するIDでUserを取得できること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// Given: データベースにUserが存在する
		testUser := newTestUser("findbyiduser", "findbyid@example.com")
		suite.createUserInDB(t, testUser)

		// When: FindByIDでUserを取得する
		foundUser, err := suite.repo.FindByID(suite.ctx, testUser.ID)

		// Then: Userが正常に取得できる
		require.NoError(t, err, "FindByIDでエラーが発生してはならない")
		require.NotNil(t, foundUser, "取得したUserがnilであってはならない")
		assertUserEquals(t, testUser, foundUser)
	})

	t.Run("存在しないIDでErrUserNotFoundが返されること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// Given: 存在しないID
		nonExistentID := domain.NewUserID(uuid.New().String())

		// When: 存在しないIDでUserを取得する
		_, err := suite.repo.FindByID(suite.ctx, nonExistentID)

		// Then: ErrUserNotFoundが返される
		assert.ErrorIs(t, err, apperr.ErrUserNotFound,
			"ErrUserNotFoundが返されるべき")
	})

	t.Run("不正な形式のIDでInternalErrorが返されること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// Given: 不正な形式のID
		invalidID := domain.NewUserID("invalid-uuid-format")

		// When: 不正なIDでUserを取得する
		_, err := suite.repo.FindByID(suite.ctx, invalidID)

		// Then: InternalErrorが返される
		assert.ErrorIs(t, err, apperr.NewInternalError("", nil),
			"InternalErrorが返されるべき")
	})

	t.Run("空文字列のIDでInternalErrorが返されること", func(t *testing.T) {
		suite := newUserTestSuite(t)

		// Given: 空文字列のID
		emptyID := domain.NewUserID("")

		// When: 空のIDでUserを取得する
		_, err := suite.repo.FindByID(suite.ctx, emptyID)

		// Then: InternalErrorが返される
		assert.ErrorIs(t, err, apperr.NewInternalError("", nil), "InternalErrorが返されるべき")
	})
}
