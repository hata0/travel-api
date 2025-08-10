package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/hata0/travel-api/internal/domain/shared/clock"
	"github.com/hata0/travel-api/internal/domain/shared/transaction_manager"
	"github.com/hata0/travel-api/internal/domain/shared/uuid"
	"github.com/hata0/travel-api/internal/infrastructure/config"
	"github.com/hata0/travel-api/internal/usecase/output"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -destination mock/auth.go github.com/hata0/travel-api/internal/usecase AuthUsecase
type AuthUsecase interface {
	Register(ctx context.Context, username, email, password string) (output.RegisterOutput, error)
	Login(ctx context.Context, email, password string) (output.TokenPairOutput, error)
	VerifyRefreshToken(ctx context.Context, refreshToken string) (output.TokenPairOutput, error)
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
}

type AuthInteractor struct {
	userRepository         domain.UserRepository
	refreshTokenRepository domain.RefreshTokenRepository
	revokedTokenRepository domain.RevokedTokenRepository
	clock                  clock.Clock
	uuidGenerator          uuid.UUIDGenerator
	transactionManager     transaction_manager.TransactionManager
	jwtSecret              string
}

func NewAuthInteractor(userRepository domain.UserRepository, refreshTokenRepository domain.RefreshTokenRepository, revokedTokenRepository domain.RevokedTokenRepository, clock clock.Clock, uuidGenerator uuid.UUIDGenerator, transactionManager transaction_manager.TransactionManager, jwtSecret string) *AuthInteractor {
	return &AuthInteractor{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		revokedTokenRepository: revokedTokenRepository,
		clock:                  clock,
		uuidGenerator:          uuidGenerator,
		transactionManager:     transactionManager,
		jwtSecret:              jwtSecret,
	}
}

// Register は新しいユーザーを登録します
func (i *AuthInteractor) Register(ctx context.Context, username, email, password string) (output.RegisterOutput, error) {
	// ユーザーの重複チェック
	if err := i.checkUserExistence(ctx, username, email); err != nil {
		return output.RegisterOutput{}, err
	}

	// パスワードのハッシュ化
	hashedPassword, err := i.hashPassword(password)
	if err != nil {
		return output.RegisterOutput{}, apperr.NewInternalError("password hashing failed", err)
	}

	// 新しいユーザーIDを生成
	userID, err := i.generateUserID()
	if err != nil {
		return output.RegisterOutput{}, apperr.NewInternalError("user id generation failed", err)
	}

	// ユーザーエンティティを作成
	now := i.clock.Now()
	user := domain.NewUser(userID, username, email, hashedPassword, now, now)

	// ユーザーを保存
	err = i.userRepository.Create(ctx, user)
	if err != nil {
		return output.RegisterOutput{}, apperr.NewInternalError("failed to create user", err)
	}

	return output.RegisterOutput{UserID: userID.String()}, nil
}

// Login はユーザーの認証を行い、トークンペアを返します
func (i *AuthInteractor) Login(ctx context.Context, email, password string) (output.TokenPairOutput, error) {
	var tokenPair output.TokenPairOutput

	err := i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error {
		// ユーザーをメールアドレスで検索
		user, err := i.userRepository.FindByEmail(txCtx, email)
		if err != nil {
			if apperr.IsUserNotFound(err) {
				return apperr.NewInvalidCredentialsError("user not found")
			}
			return apperr.NewInternalError("failed to find user by email", err)
		}

		// パスワードの検証
		if err := i.verifyPassword(user.PasswordHash, password); err != nil {
			// verifyPasswordで既に適切なAppErrorが返されるため、そのまま返す
			return err
		}

		// トークンペアの生成
		accessTokenString, refreshTokenString, err := i.generateTokenPair(user.ID, i.clock.Now())
		if err != nil {
			return apperr.NewInternalError("token generation failed", err)
		}

		// リフレッシュトークンをデータベースに保存
		if err := i.storeRefreshToken(txCtx, refreshTokenString, user.ID); err != nil {
			return err
		}

		tokenPair = output.TokenPairOutput{Token: accessTokenString, RefreshToken: refreshTokenString}
		return nil
	})

	if err != nil {
		return output.TokenPairOutput{}, err
	}

	return tokenPair, nil
}

// VerifyRefreshToken はリフレッシュトークンを検証し、新しいトークンペアを返します
func (i *AuthInteractor) VerifyRefreshToken(ctx context.Context, refreshToken string) (output.TokenPairOutput, error) {
	// 失効済みトークンのチェック
	if err := i.checkRevokedToken(ctx, refreshToken); err != nil {
		return output.TokenPairOutput{}, err
	}

	var tokenPair output.TokenPairOutput

	err := i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error {
		// リフレッシュトークンの検索と検証
		foundToken, err := i.findAndValidateRefreshToken(txCtx, refreshToken)
		if err != nil {
			return err
		}

		// 古いトークンの失効処理
		if err := i.revokeToken(txCtx, foundToken); err != nil {
			return err
		}

		// 新しいトークンペアの生成
		accessTokenString, newRefreshTokenString, err := i.generateTokenPair(foundToken.UserID, i.clock.Now())
		if err != nil {
			return apperr.NewInternalError("token generation failed", err)
		}

		if err := i.storeRefreshToken(txCtx, newRefreshTokenString, foundToken.UserID); err != nil {
			return err
		}

		tokenPair = output.TokenPairOutput{Token: accessTokenString, RefreshToken: newRefreshTokenString}
		return nil
	})

	if err != nil {
		return output.TokenPairOutput{}, err
	}

	return tokenPair, nil
}

// RevokeRefreshToken はリフレッシュトークンを失効させます
func (i *AuthInteractor) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	foundToken, err := i.refreshTokenRepository.FindByToken(ctx, refreshToken)
	if err != nil {
		if apperr.IsTokenNotFound(err) {
			// 既に存在しないトークンの場合は成功とする（冪等性）
			return nil
		}

		return apperr.NewInternalError("failed to find refresh token by token", err)
	}

	if err := i.refreshTokenRepository.Delete(ctx, foundToken); err != nil {
		return apperr.NewInternalError("failed to delete refresh token", err)
	}

	return nil
}

// checkUserExistence はユーザー名またはメールアドレスが既に存在するかをチェックします。
func (i *AuthInteractor) checkUserExistence(ctx context.Context, username, email string) error {
	// ユーザー名が既に存在するか確認
	_, err := i.userRepository.FindByUsername(ctx, username)
	if err == nil {
		return apperr.NewConflictError("username")
	}

	// ユーザー名が見つからない場合は、メールアドレスの存在チェックに進む
	if !apperr.IsUserNotFound(err) {
		return apperr.NewInternalError("failed to check username existence", err)
	}

	// メールアドレスが既に存在するか確認
	_, err = i.userRepository.FindByEmail(ctx, email)
	if err == nil {
		return apperr.NewConflictError("email")
	}

	if !apperr.IsUserNotFound(err) {
		return apperr.NewInternalError("failed to check email existence", err)
	}

	// メールアドレスも見つからない場合は、ユーザーは存在しない
	return nil
}

// hashPassword はパスワードをハッシュ化します
func (i *AuthInteractor) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword はパスワードを検証します
func (i *AuthInteractor) verifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			// パスワード不一致の場合は認証エラーとして扱う
			return apperr.NewInvalidCredentialsError("password mismatch")
		}
		// bcryptライブラリの内部エラー（ハッシュ形式が不正など）
		return apperr.NewInternalError("password verification failed", err)
	}
	return nil
}

// generateUserID は新しいユーザーIDを生成します
func (i *AuthInteractor) generateUserID() (domain.UserID, error) {
	newUUID := i.uuidGenerator.NewUUID()
	return domain.NewUserID(newUUID)
}

// storeRefreshToken はリフレッシュトークンをデータベースに保存します
func (i *AuthInteractor) storeRefreshToken(ctx context.Context, refreshTokenString string, userID domain.UserID) error {
	refreshTokenID, err := domain.NewRefreshTokenID(refreshTokenString)
	if err != nil {
		return apperr.NewInternalError("refresh token id creation failed", err)
	}

	now := i.clock.Now()
	refreshToken := domain.NewRefreshToken(
		refreshTokenID,
		userID,
		refreshTokenString,
		now.Add(config.RefreshTokenExpiration()),
		now,
	)

	if err := i.refreshTokenRepository.Create(ctx, refreshToken); err != nil {
		return apperr.NewInternalError("failed to create refresh token", err)
	}

	return nil
}

// checkRevokedToken は失効済みトークンかどうかをチェックします
func (i *AuthInteractor) checkRevokedToken(ctx context.Context, refreshToken string) error {
	_, err := i.revokedTokenRepository.FindByJTI(ctx, refreshToken)
	if err == nil {
		// 失効済みトークンが見つかった場合、再利用攻撃の可能性
		// このトークンに紐づくユーザーのセッションをすべて無効化
		if err := i.handleTokenReuseAttack(ctx, refreshToken); err != nil {
			slog.Error("Failed to handle token reuse attack", "error", err)
		}

		return apperr.NewInvalidCredentialsError("token has been revoked")
	}

	// 失効済みトークンが見つからない場合は正常
	if apperr.IsTokenNotFound(err) {
		return nil
	}

	// その他のエラー
	return apperr.NewInternalError("Failed to check revoked token", err)
}

// handleTokenReuseAttack はトークン再利用攻撃を検出した際の処理を行います
func (i *AuthInteractor) handleTokenReuseAttack(ctx context.Context, refreshToken string) error {
	foundToken, err := i.refreshTokenRepository.FindByToken(ctx, refreshToken)
	if err != nil {
		if apperr.IsTokenNotFound(err) {
			return nil // 既に削除済み
		}
		return err
	}

	// ユーザーIDに紐づくすべてのリフレッシュトークンを削除
	if err := i.refreshTokenRepository.DeleteByUserID(ctx, foundToken.UserID); err != nil {
		return err
	}

	slog.Warn("All refresh tokens deleted due to potential reuse attack", "userID", foundToken.UserID.String())
	return nil
}

// findAndValidateRefreshToken はリフレッシュトークンを検索して検証します
func (i *AuthInteractor) findAndValidateRefreshToken(ctx context.Context, refreshToken string) (domain.RefreshToken, error) {
	foundToken, err := i.refreshTokenRepository.FindByToken(ctx, refreshToken)
	if err != nil {
		if apperr.IsTokenNotFound(err) {
			return domain.RefreshToken{}, apperr.NewInvalidCredentialsError("Invalid refresh token")
		}
		return domain.RefreshToken{}, apperr.NewInternalError("Failed to find refresh token", err)
	}

	// 有効期限のチェック
	if i.clock.Now().After(foundToken.ExpiresAt) {
		// 期限切れの場合は削除
		if err := i.refreshTokenRepository.Delete(ctx, foundToken); err != nil {
			slog.Error("Failed to delete expired refresh token", "error", err, "tokenID", foundToken.ID.String())
		}

		return domain.RefreshToken{}, apperr.NewInvalidCredentialsError("Refresh token expired")
	}

	return foundToken, nil
}

// revokeToken はトークンを失効済みとして記録し削除します
func (i *AuthInteractor) revokeToken(ctx context.Context, token domain.RefreshToken) error {
	// 失効済みトークンとして記録
	revokedTokenID, err := domain.NewRevokedTokenID(i.uuidGenerator.NewUUID())
	if err != nil {
		return apperr.NewInternalError("Revoked token ID creation failed", err)
	}

	revokedToken := domain.NewRevokedToken(
		revokedTokenID,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		i.clock.Now(),
	)

	if err := i.revokedTokenRepository.Create(ctx, revokedToken); err != nil {
		return apperr.NewInternalError("Failed to create revoked token", err)
	}

	// 元のリフレッシュトークンを削除
	if err := i.refreshTokenRepository.Delete(ctx, token); err != nil {
		return apperr.NewInternalError("Failed to delete refresh token", err)
	}

	return nil
}

// generateTokenPair はアクセストークンとリフレッシュトークン文字列を生成します。
func (i *AuthInteractor) generateTokenPair(userID domain.UserID, now time.Time) (string, string, error) {
	// JWTトークンの生成
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     now.Add(config.AccessTokenExpiration()).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(i.jwtSecret))
	if err != nil {
		return "", "", apperr.NewInternalError("", err)
	}

	// リフレッシュトークンの生成
	refreshTokenString := i.uuidGenerator.NewUUID()

	return accessTokenString, refreshTokenString, nil
}
