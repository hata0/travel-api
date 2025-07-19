package usecase

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
	"travel-api/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
	interactor := NewAuthInteractor(mockRepo, mock_domain.NewMockRefreshTokenRepository(ctrl), mockClock, mockUUIDGenerator)

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

		assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
	})

	t.Run("異常系: メールアドレスが既に存在する", func(t *testing.T) {
		// FindByUsernameがUserNotFoundを返すことを期待
		mockRepo.EXPECT().FindByUsername(gomock.Any(), username).Return(domain.User{}, domain.ErrUserNotFound).Times(1)
		// FindByEmailがエラーなしでユーザーを返すことを期待
		mockRepo.EXPECT().FindByEmail(gomock.Any(), email).Return(domain.User{}, nil).Times(1)

		_, err := interactor.Register(context.Background(), username, email, password)

		assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
	})

	t.Run("異常系: リポジトリCreate失敗", func(t *testing.T) {
		mockRepo.EXPECT().FindByUsername(gomock.Any(), username).Return(domain.User{}, domain.ErrUserNotFound).Times(1)
		mockRepo.EXPECT().FindByEmail(gomock.Any(), email).Return(domain.User{}, domain.ErrUserNotFound).Times(1)
		mockUUIDGenerator.EXPECT().NewUUID().Return(generatedUUID).Times(1)
		mockClock.EXPECT().Now().Return(now).Times(2)

		expectedErr := errors.New("db error")
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(expectedErr).Times(1)

		_, err := interactor.Register(context.Background(), username, email, password)
		assert.Error(t, err)
		var appErr *domain.Error
		assert.True(t, errors.As(err, &appErr) && appErr.Code == domain.InternalServerError, "expected internal server error")
	})
}

func TestAuthInteractor_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_domain.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mock_domain.NewMockRefreshTokenRepository(ctrl)
	mockClock := mock_domain.NewMockClock(ctrl)
	mockUUIDGenerator := mock_domain.NewMockUUIDGenerator(ctrl)
	interactor := NewAuthInteractor(mockUserRepo, mockRefreshTokenRepo, mockClock, mockUUIDGenerator)

	email := "test@example.com"
	password := "password123"
	userID := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	userDomainID, _ := domain.NewUserID(userID)
	expectedUser := domain.NewUser(userDomainID, "testuser", email, string(hashedPassword), now, now)

	t.Run("正常系: ユーザーが正常にログインできる", func(t *testing.T) {
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
			now.Add(time.Hour*24*7),
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
		// FindByEmailがUserNotFoundを返すことを期待
		mockUserRepo.EXPECT().FindByEmail(gomock.Any(), email).Return(domain.User{}, domain.ErrUserNotFound).Times(1)

		_, err := interactor.Login(context.Background(), email, password)

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("異常系: パスワードが間違っている", func(t *testing.T) {
		// FindByEmailがユーザーを返すことを期待
		mockUserRepo.EXPECT().FindByEmail(gomock.Any(), email).Return(expectedUser, nil).Times(1)

		_, err := interactor.Login(context.Background(), email, "wrongpassword")

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("異常系: JWT秘密鍵が未設定", func(t *testing.T) {
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
	mockClock := mock_domain.NewMockClock(ctrl)
	mockUUIDGenerator := mock_domain.NewMockUUIDGenerator(ctrl)
	interactor := NewAuthInteractor(mockUserRepo, mockRefreshTokenRepo, mockClock, mockUUIDGenerator)

	refreshTokenString := "valid-refresh-token"
	userID := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	userDomainID, _ := domain.NewUserID(userID)
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := now.Add(time.Hour * 24 * 7)

	// テスト用のユーザーとリフレッシュトークン
	expectedUser := domain.NewUser(userDomainID, "testuser", "test@example.com", "hashedpass", now, now)
	refreshTokenID, err := domain.NewRefreshTokenID(uuid.New().String())
	assert.NoError(t, err)
	refreshToken := domain.NewRefreshToken(refreshTokenID, userDomainID, refreshTokenString, expiresAt, now)

	t.Run("正常系: リフレッシュトークンが検証され、新しいアクセストークンとリフレッシュトークンが発行される", func(t *testing.T) {
		// FindByTokenがリフレッシュトークンを返すことを期待
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(refreshToken, nil).Times(1)

		// Clockが現在時刻を返すことを期待
		mockClock.EXPECT().Now().Return(now).Times(4) // 有効期限チェックと新しいトークン生成時

		// 古いリフレッシュトークンが削除されることを期待
		mockRefreshTokenRepo.EXPECT().Delete(gomock.Any(), refreshTokenString).Return(nil).Times(1)

		// FindByIDがユーザーを返すことを期待
		mockUserRepo.EXPECT().FindByID(gomock.Any(), userDomainID).Return(expectedUser, nil).Times(1)

		newRefreshTokenString := uuid.New().String()
		// UUIDGeneratorが新しいリフレッシュトークンを生成することを期待
		mockUUIDGenerator.EXPECT().NewUUID().Return(newRefreshTokenString).Times(1)

		// 新しいリフレッシュトークンが保存されることを期待
		newRefreshTokenID, err := domain.NewRefreshTokenID(newRefreshTokenString)
		assert.NoError(t, err)
		expectedNewRefreshToken := domain.NewRefreshToken(
			newRefreshTokenID,
			expectedUser.ID,
			newRefreshTokenString,
			now.Add(time.Hour*24*7),
			now,
		)
		mockRefreshTokenRepo.EXPECT().Create(gomock.Any(), expectedNewRefreshToken).Return(nil).Times(1)

		// JWTSecretが秘密鍵を返すことを期待
		os.Setenv("JWT_SECRET", "your_jwt_secret_key")
		defer os.Unsetenv("JWT_SECRET")

		output, err := interactor.VerifyRefreshToken(context.Background(), refreshTokenString)

		assert.NoError(t, err)
		assert.NotEmpty(t, output.Token)
		assert.Equal(t, newRefreshTokenString, output.RefreshToken)
	})

	t.Run("異常系: リフレッシュトークンが見つからない", func(t *testing.T) {
		// FindByTokenがTokenNotFoundを返すことを期待
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(domain.RefreshToken{}, domain.ErrTokenNotFound).Times(1)

		_, err := interactor.VerifyRefreshToken(context.Background(), refreshTokenString)

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("異常系: リフレッシュトークンの有効期限が切れている", func(t *testing.T) {
		// 期限切れのリフレッシュトークン
		expiredRefreshTokenID, err := domain.NewRefreshTokenID(uuid.New().String())
		assert.NoError(t, err)
		expiredRefreshToken := domain.NewRefreshToken(expiredRefreshTokenID, userDomainID, refreshTokenString, now.Add(-time.Hour), now)

		// FindByTokenが期限切れトークンを返すことを期待
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(expiredRefreshToken, nil).Times(1)

		// Clockが現在時刻を返すことを期待
		mockClock.EXPECT().Now().Return(now).Times(1)

		// 期限切れトークンが削除されることを期待
		mockRefreshTokenRepo.EXPECT().Delete(gomock.Any(), refreshTokenString).Return(nil).Times(1)

		_, err = interactor.VerifyRefreshToken(context.Background(), refreshTokenString)

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("異常系: FindByIDでエラー", func(t *testing.T) {
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(refreshToken, nil).Times(1)
		mockClock.EXPECT().Now().Return(now).Times(1)
		mockRefreshTokenRepo.EXPECT().Delete(gomock.Any(), refreshTokenString).Return(nil).Times(1)
		mockUserRepo.EXPECT().FindByID(gomock.Any(), userDomainID).Return(domain.User{}, errors.New("db error")).Times(1)

		_, err := interactor.VerifyRefreshToken(context.Background(), refreshTokenString)
		assert.Error(t, err)
		var appErr *domain.Error
		assert.True(t, errors.As(err, &appErr) && appErr.Code == domain.InternalServerError, "expected internal server error")
	})

	t.Run("異常系: リフレッシュトークン保存でエラー", func(t *testing.T) {
		mockRefreshTokenRepo.EXPECT().FindByToken(gomock.Any(), refreshTokenString).Return(refreshToken, nil).Times(1)
		mockClock.EXPECT().Now().Return(now).Times(4)
		mockRefreshTokenRepo.EXPECT().Delete(gomock.Any(), refreshTokenString).Return(nil).Times(1)
		mockUserRepo.EXPECT().FindByID(gomock.Any(), userDomainID).Return(expectedUser, nil).Times(1)
		mockUUIDGenerator.EXPECT().NewUUID().Return(uuid.New().String()).Times(1)
		mockRefreshTokenRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("db error")).Times(1)
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

	mockUserRepo := mock_domain.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mock_domain.NewMockRefreshTokenRepository(ctrl)
	mockClock := mock_domain.NewMockClock(ctrl)
	mockUUIDGenerator := mock_domain.NewMockUUIDGenerator(ctrl)
	interactor := NewAuthInteractor(mockUserRepo, mockRefreshTokenRepo, mockClock, mockUUIDGenerator)

	refreshTokenString := "token-to-revoke"

	t.Run("正常系: リフレッシュトークンが正常に失効される", func(t *testing.T) {
		// Deleteがエラーを返さないことを期待
		mockRefreshTokenRepo.EXPECT().Delete(gomock.Any(), refreshTokenString).Return(nil).Times(1)

		err := interactor.RevokeRefreshToken(context.Background(), refreshTokenString)

		assert.NoError(t, err)
	})

	t.Run("異常系: リポジトリからエラーが返された場合", func(t *testing.T) {
		expectedErr := errors.New("db error")
		mockRefreshTokenRepo.EXPECT().Delete(gomock.Any(), refreshTokenString).Return(expectedErr).Times(1)

		err := interactor.RevokeRefreshToken(context.Background(), refreshTokenString)

		assert.ErrorIs(t, err, expectedErr)
	})
}
