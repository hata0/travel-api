package usecase

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
	"travel-api/internal/config"
	"travel-api/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	mock_domain "travel-api/internal/domain/mock"
)

func TestAuthInteractor_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockUserRepository(ctrl)
	mockClock := mock_domain.NewMockClock(ctrl)
	mockUUIDGenerator := mock_domain.NewMockUUIDGenerator(ctrl)
	interactor := NewAuthInteractor(mockRepo, mock_domain.NewMockRefreshTokenRepository(ctrl), mock_domain.NewMockRevokedTokenRepository(ctrl), mockClock, mockUUIDGenerator, mock_domain.NewMockTransactionManager(ctrl))

	username := "testuser"
	email := "test@example.com"
	password := "password123"
	generatedUUID := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("正常系: ユーザーが正常に登録される", func(t *testing.T) {
		// FindByUsernameとFindByEmailがUserNotFoundを返すことを期待
		mockRepo.EXPECT().FindByUsername(gomock.Any(), username).Return(domain.User{}, domain.ErrUserNotFound).Times(1)
		mockRepo.EXPECT().FindByEmail(gomock.Any(), email).Return(domain.User{}, domain.ErrUserNotFound).Times(1)

		// UUIDと時刻の生成を期待
		mockUUIDGenerator.EXPECT().NewUUID().Return(generatedUUID).Times(1)
		mockClock.EXPECT().Now().Return(now).Times(2)

		// Createがエラーを返さないことを期待
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).Times(1)

		output, err := interactor.Register(context.Background(), username, email, password)

		assert.NoError(t, err)
		assert.Equal(t, generatedUUID, output.UserID)
	})

	t.Run("異常系: ユーザー名が既に存在する", func(t *testing.T) {
		// FindByUsernameがエラーなしでユーザーを返すことを期待
		mockRepo.EXPECT().FindByUsername(gomock.Any(), username).Return(domain.User{}, nil).Times(1)

		_, err := interactor.Register(context.Background(), username, email, password)

		assert.ErrorIs(t, err, domain.ErrUsernameAlreadyExists)
	})

	t.Run("異常系: メールアドレスが既に存在する", func(t *testing.T) {
		// FindByUsernameがUserNotFoundを返すことを期待
		mockRepo.EXPECT().FindByUsername(gomock.Any(), username).Return(domain.User{}, domain.ErrUserNotFound).Times(1)
		// FindByEmailがエラーなしでユーザーを返すことを期待
		mockRepo.EXPECT().FindByEmail(gomock.Any(), email).Return(domain.User{}, nil).Times(1)

		_, err := interactor.Register(context.Background(), username, email, password)

		assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
	})

	t.Run("異常系: リポジトリCreate失敗", func(t *testing.T) {
		mockRepo.EXPECT().FindByUsername(gomock.Any(), username).Return(domain.User{}, domain.ErrUserNotFound).Times(1)
		mockRepo.EXPECT().FindByEmail(gomock.Any(), email).Return(domain.User{}, domain.ErrUserNotFound).Times(1)
		mockUUIDGenerator.EXPECT().NewUUID().Return(generatedUUID).Times(1)
		mockClock.EXPECT().Now().Return(now).Times(2)

		expectedErr := domain.ErrUserAlreadyExists
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(expectedErr).Times(1)

		_, err := interactor.Register(context.Background(), username, email, password)
		assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
	})
}

func TestAuthInteractor_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_domain.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mock_domain.NewMockRefreshTokenRepository(ctrl)
	mockClock := mock_domain.NewMockClock(ctrl)
	mockUUIDGenerator := mock_domain.NewMockUUIDGenerator(ctrl)
	mockTransactionManager := mock_domain.NewMockTransactionManager(ctrl)
	interactor := NewAuthInteractor(mockUserRepo, mockRefreshTokenRepo, mock_domain.NewMockRevokedTokenRepository(ctrl), mockClock, mockUUIDGenerator, mockTransactionManager)

	email := "test@example.com"
	password := "password123"
	userID := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	userDomainID, _ := domain.NewUserID(userID)
	expectedUser := domain.NewUser(userDomainID, "testuser", email, string(hashedPassword), now, now)

	t.Run("正常系: ユーザーが正常にログインできる", func(t *testing.T) {
		mockTransactionManager.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		).Times(1)
		// FindByEmailがユーザーを返すことを期待
		mockUserRepo.EXPECT().FindByEmail(gomock.Any(), email).Return(expectedUser, nil).Times(1)

		// UUIDGeneratorがリフレッシュトークンを生成することを期待
		refreshTokenString := "6afcbe27-c792-485c-969f-4313db93e4a3"
		mockUUIDGenerator.EXPECT().NewUUID().Return(refreshTokenString).Times(1)

		// Clockが現在時刻を返すことを期待
		mockClock.EXPECT().Now().Return(now).Times(3)

		// リフレッシュトークンが作成されることを期待
		refreshTokenID, err := domain.NewRefreshTokenID(refreshTokenString)
		assert.NoError(t, err)
		expectedRefreshToken := domain.NewRefreshToken(
			refreshTokenID,
			expectedUser.ID,
			refreshTokenString,
			now.Add(config.RefreshTokenExpiration()),
			now,
		)
		mockRefreshTokenRepo.EXPECT().Create(gomock.Any(), expectedRefreshToken).Return(nil).Times(1)

		// JWTSecretが秘密鍵を返すことを期待
		os.Setenv("JWT_SECRET", "your_jwt_secret_key")
		defer os.Unsetenv("JWT_SECRET")

		output, err := interactor.Login(context.Background(), email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, output.Token)
		assert.Equal(t, refreshTokenString, output.RefreshToken)

		// 生成されたトークンの検証（オプション）
		token, _ := jwt.Parse(output.Token, func(token *jwt.Token) (interface{}, error) {
			return []byte("your_jwt_secret_key"), nil
		})
		claims, ok := token.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, userID, claims["user_id"])
	})

	t.Run("異常系: ユーザーが見つからない", func(t *testing.T) {
		mockTransactionManager.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		).Times(1)
		// FindByEmailがUserNotFoundを返すことを期待
		mockUserRepo.EXPECT().FindByEmail(gomock.Any(), email).Return(domain.User{}, domain.ErrUserNotFound).Times(1)

		_, err := interactor.Login(context.Background(), email, password)

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("異常系: パスワードが間違っている", func(t *testing.T) {
		mockTransactionManager.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		).Times(1)
		// FindByEmailがユーザーを返すことを期待
		mockUserRepo.EXPECT().FindByEmail(gomock.Any(), email).Return(expectedUser, nil).Times(1)

		_, err := interactor.Login(context.Background(), email, "wrongpassword")

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("異常系: JWT秘密鍵が未設定", func(t *testing.T) {
		mockTransactionManager.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		).Times(1)
		mockUserRepo.EXPECT().FindByEmail(gomock.Any(), email).Return(expectedUser, nil).Times(1)
		mockClock.EXPECT().Now().Return(now).Times(1)

		// JWT_SECRETが設定されていない状態にする
		os.Unsetenv("JWT_SECRET")

		_, err := interactor.Login(context.Background(), email, password)
		assert.Error(t, err)
		var appErr *domain.Error
		assert.True(t, errors.As(err, &appErr) && appErr.Code == domain.InternalServerError, "expected internal server error")
	})
}

func TestAuthInteractor_VerifyRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_domain.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mock_domain.NewMockRefreshTokenRepository(ctrl)
	mockRevokedTokenRepo := mock_domain.NewMockRevokedTokenRepository(ctrl)
	mockClock := mock_domain.NewMockClock(ctrl)
	mockUUIDGenerator := mock_domain.NewMockUUIDGenerator(ctrl)
	mockTransactionManager := mock_domain.NewMockTransactionManager(ctrl)
	interactor := NewAuthInteractor(mockUserRepo, mockRefreshTokenRepo, mockRevokedTokenRepo, mockClock, mockUUIDGenerator, mockTransactionManager)

	refreshTokenString := "valid-refresh-token"
	userIDString := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	userID, _ := domain.NewUserID(userIDString)
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := now.Add(config.RefreshTokenExpiration())

	// テスト用のユーザーとリフレッシュトークン
	user := domain.NewUser(userID, "testuser", "test@example.com", "hashedpass", now, now)
	refreshTokenID, err := domain.NewRefreshTokenID(uuid.New().String())
	assert.NoError(t, err)
	refreshToken := domain.NewRefreshToken(refreshTokenID, userID, refreshTokenString, expiresAt, now)
	revokedTokenIDString := uuid.New().String()
	revokedTokenID, err := domain.NewRevokedTokenID(revokedTokenIDString)
	require.NoError(t, err)
	revokedToken := domain.NewRevokedToken(revokedTokenID, userID, refreshToken.Token, refreshToken.ExpiresAt, now)

	newRefreshTokenIDString := uuid.New().String()
	newRefreshTokenID, err := domain.NewRefreshTokenID(newRefreshTokenIDString)
	assert.NoError(t, err)
	newRefreshToken := domain.NewRefreshToken(
		newRefreshTokenID,
		user.ID,
		newRefreshTokenIDString,
		now.Add(config.RefreshTokenExpiration()),
		now,
	)

	t.Run("正常系: リフレッシュトークンが検証され、新しいアクセストークンとリフレッシュトークンが発行される", func(t *testing.T) {
		mockRevokedTokenRepo.EXPECT().FindByJTI(gomock.Any(), refreshTokenString).Return(domain.RevokedToken{}, domain.ErrTokenNotFound).Times(1)
		mockTransactionManager.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		).Times(1)
		// FindByTokenがリフレッシュトークンを返すことを期待
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(refreshToken, nil).Times(1)

		// Clockが現在時刻を返すことを期待
		mockClock.EXPECT().Now().Return(now).Times(5) // 有効期限チェックと新しいトークン生成時

		callCount := 0
		mockUUIDGenerator.EXPECT().NewUUID().DoAndReturn(func() string {
			callCount++
			if callCount == 1 {
				return revokedTokenIDString
			}
			return newRefreshTokenIDString
		}).Times(2)

		// 古いリフレッシュトークンが失効済みとして記録され、削除されることを期待
		mockRevokedTokenRepo.EXPECT().Create(gomock.Any(), revokedToken).Return(nil).Times(1)
		mockRefreshTokenRepo.EXPECT().Delete(gomock.Any(), refreshToken).Return(nil).Times(1)

		// FindByIDがユーザーを返すことを期待
		mockUserRepo.EXPECT().FindByID(gomock.Any(), userID).Return(user, nil).Times(1)

		// 新しいリフレッシュトークンが保存されることを期待
		mockRefreshTokenRepo.EXPECT().Create(gomock.Any(), newRefreshToken).Return(nil).Times(1)

		// JWTSecretが秘密鍵を返すことを期待
		os.Setenv("JWT_SECRET", "your_jwt_secret_key")
		defer os.Unsetenv("JWT_SECRET")

		output, err := interactor.VerifyRefreshToken(context.Background(), refreshTokenString)

		assert.NoError(t, err)
		assert.NotEmpty(t, output.Token)
		assert.Equal(t, newRefreshTokenIDString, output.RefreshToken)
	})

	t.Run("異常系: リフレッシュトークンが見つからない", func(t *testing.T) {
		// FindByJTIがTokenNotFoundを返すことを期待
		mockRevokedTokenRepo.EXPECT().FindByJTI(gomock.Any(), refreshTokenString).Return(domain.RevokedToken{}, domain.ErrTokenNotFound).Times(1)

		mockTransactionManager.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		).Times(1)
		// FindByTokenがTokenNotFoundを返すことを期待
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(domain.RefreshToken{}, domain.ErrTokenNotFound).Times(1)

		_, err := interactor.VerifyRefreshToken(context.Background(), refreshTokenString)

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("異常系: リフレッシュトークンの有効期限が切れている", func(t *testing.T) {
		// FindByJTIがTokenNotFoundを返すことを期待
		mockRevokedTokenRepo.EXPECT().FindByJTI(gomock.Any(), refreshTokenString).Return(domain.RevokedToken{}, domain.ErrTokenNotFound).Times(1)

		mockTransactionManager.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		).Times(1)
		// 期限切れのリフレッシュトークン
		expiredRefreshTokenID, err := domain.NewRefreshTokenID(uuid.New().String())
		assert.NoError(t, err)
		expiredRefreshToken := domain.NewRefreshToken(expiredRefreshTokenID, userID, refreshTokenString, now.Add(-time.Hour), now)

		// FindByTokenが期限切れトークンを返すことを期待
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(expiredRefreshToken, nil).Times(1)

		// Clockが現在時刻を返すことを期待
		mockClock.EXPECT().Now().Return(now).Times(1)

		// 期限切れトークンが削除されることを期待
		mockRefreshTokenRepo.EXPECT().Delete(gomock.Any(), expiredRefreshToken).Return(nil).Times(1)

		_, err = interactor.VerifyRefreshToken(context.Background(), refreshTokenString)

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("異常系: 古いリフレッシュトークンの削除に失敗", func(t *testing.T) {
		// FindByJTIがTokenNotFoundを返すことを期待
		mockRevokedTokenRepo.EXPECT().FindByJTI(gomock.Any(), refreshTokenString).Return(domain.RevokedToken{}, domain.ErrTokenNotFound).Times(1)

		mockTransactionManager.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		).Times(1)
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(refreshToken, nil).Times(1)
		mockClock.EXPECT().Now().Return(now).Times(2)
		mockUUIDGenerator.EXPECT().NewUUID().Return(revokedTokenIDString).Times(1)
		mockRevokedTokenRepo.EXPECT().Create(gomock.Any(), revokedToken).Return(nil).Times(1)
		// Deleteがエラーを返すことを期待
		expectedErr := domain.ErrInternalServerError
		mockRefreshTokenRepo.EXPECT().Delete(gomock.Any(), refreshToken).Return(expectedErr).Times(1)

		_, err := interactor.VerifyRefreshToken(context.Background(), refreshTokenString)
		assert.ErrorIs(t, err, domain.ErrInternalServerError)
	})

	t.Run("異常系: リフレッシュトークンが再利用された場合", func(t *testing.T) {
		// FindByJTIが失効済みトークンを返すことを期待
		mockRevokedTokenRepo.EXPECT().FindByJTI(gomock.Any(), refreshTokenString).Return(domain.RevokedToken{}, nil).Times(1)
		// FindByTokenがリフレッシュトークンを返すことを期待
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(refreshToken, nil).Times(1)
		// DeleteByUserIDが呼び出されることを期待
		mockRefreshTokenRepo.EXPECT().DeleteByUserID(gomock.Any(), userID).Return(nil).Times(1)

		_, err := interactor.VerifyRefreshToken(context.Background(), refreshTokenString)

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("異常系: FindByIDでエラー", func(t *testing.T) {
		mockRevokedTokenRepo.EXPECT().FindByJTI(gomock.Any(), refreshTokenString).Return(domain.RevokedToken{}, domain.ErrTokenNotFound).Times(1)
		mockTransactionManager.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		).Times(1)
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(refreshToken, nil).Times(1)
		mockClock.EXPECT().Now().Return(now).Times(2)
		mockUUIDGenerator.EXPECT().NewUUID().Return(revokedTokenIDString).Times(1)
		mockRevokedTokenRepo.EXPECT().Create(gomock.Any(), revokedToken).Return(nil).Times(1)
		mockRefreshTokenRepo.EXPECT().Delete(gomock.Any(), refreshToken).Return(nil).Times(1)
		mockUserRepo.EXPECT().FindByID(gomock.Any(), userID).Return(domain.User{}, domain.ErrUserNotFound).Times(1)

		_, err := interactor.VerifyRefreshToken(context.Background(), refreshTokenString)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("異常系: リフレッシュトークン保存でエラー", func(t *testing.T) {
		mockRevokedTokenRepo.EXPECT().FindByJTI(gomock.Any(), refreshTokenString).Return(domain.RevokedToken{}, domain.ErrTokenNotFound).Times(1)
		mockTransactionManager.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		).Times(1)
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(refreshToken, nil).Times(1)
		mockClock.EXPECT().Now().Return(now).Times(5)

		callCount := 0
		mockUUIDGenerator.EXPECT().NewUUID().DoAndReturn(func() string {
			callCount++
			if callCount == 1 {
				return revokedTokenIDString
			}
			return newRefreshTokenIDString
		}).Times(2)

		mockRevokedTokenRepo.EXPECT().Create(gomock.Any(), revokedToken).Return(nil).Times(1)
		mockRefreshTokenRepo.EXPECT().Delete(gomock.Any(), refreshToken).Return(nil).Times(1)
		mockUserRepo.EXPECT().FindByID(gomock.Any(), userID).Return(user, nil).Times(1)
		mockRefreshTokenRepo.EXPECT().Create(gomock.Any(), newRefreshToken).Return(errors.New("db error")).Times(1)
		os.Setenv("JWT_SECRET", "your_jwt_secret_key")
		defer os.Unsetenv("JWT_SECRET")

		_, err := interactor.VerifyRefreshToken(context.Background(), refreshTokenString)
		assert.Error(t, err)
		var appErr *domain.Error
		assert.True(t, errors.As(err, &appErr) && appErr.Code == domain.InternalServerError, "expected internal server error")
	})
}

func TestAuthInteractor_RevokeRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRefreshTokenRepo := mock_domain.NewMockRefreshTokenRepository(ctrl)
	interactor := NewAuthInteractor(mock_domain.NewMockUserRepository(ctrl), mockRefreshTokenRepo, mock_domain.NewMockRevokedTokenRepository(ctrl), mock_domain.NewMockClock(ctrl), mock_domain.NewMockUUIDGenerator(ctrl), mock_domain.NewMockTransactionManager(ctrl))

	refreshTokenString := "token-to-revoke"
	userID := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	userDomainID, _ := domain.NewUserID(userID)
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := now.Add(time.Hour * 24 * 7)
	refreshTokenID, err := domain.NewRefreshTokenID(uuid.New().String())
	assert.NoError(t, err)
	refreshToken := domain.NewRefreshToken(refreshTokenID, userDomainID, refreshTokenString, expiresAt, now)

	t.Run("正常系: リフレッシュトークンが正常に失効される", func(t *testing.T) {
		// FindByTokenがリフレッシュトークンを返すことを期待
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(refreshToken, nil).Times(1)
		// Deleteがエラーを返さないことを期待
		mockRefreshTokenRepo.EXPECT().Delete(gomock.Any(), refreshToken).Return(nil).Times(1)

		err := interactor.RevokeRefreshToken(context.Background(), refreshTokenString)

		assert.NoError(t, err)
	})

	t.Run("異常系: リフレッシュトークンが見つからない", func(t *testing.T) {
		// FindByTokenがTokenNotFoundを返すことを期待
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(domain.RefreshToken{}, domain.ErrTokenNotFound).Times(1)

		err := interactor.RevokeRefreshToken(context.Background(), refreshTokenString)

		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}
