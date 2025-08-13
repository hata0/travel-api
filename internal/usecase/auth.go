package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/hata0/travel-api/internal/infrastructure/config"
	"github.com/hata0/travel-api/internal/usecase/output"
	"github.com/hata0/travel-api/internal/usecase/services"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -destination mock/auth.go github.com/hata0/travel-api/internal/usecase AuthUsecase
type AuthUsecase interface {
	Register(ctx context.Context, username, email, password string) (*output.RegisterOutput, error)
	Login(ctx context.Context, email, password string) (*output.TokenPairOutput, error)
	VerifyRefreshToken(ctx context.Context, refreshToken string) (*output.TokenPairOutput, error)
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
}

type AuthInteractor struct {
	userRepository         domain.UserRepository
	refreshTokenRepository domain.RefreshTokenRepository
	revokedTokenRepository domain.RevokedTokenRepository
	timeProvider           services.TimeProvider
	idGenerator            services.IDGenerator
	transactionManager     services.TransactionManager
	jwtSecret              string
}

func NewAuthInteractor(userRepository domain.UserRepository, refreshTokenRepository domain.RefreshTokenRepository, revokedTokenRepository domain.RevokedTokenRepository, timeProvider services.TimeProvider, idGenerator services.IDGenerator, transactionManager services.TransactionManager, jwtSecret string) *AuthInteractor {
	return &AuthInteractor{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		revokedTokenRepository: revokedTokenRepository,
		timeProvider:           timeProvider,
		idGenerator:            idGenerator,
		transactionManager:     transactionManager,
		jwtSecret:              jwtSecret,
	}
}

// Register は新しいユーザーを登録します
func (i *AuthInteractor) Register(ctx context.Context, username, email, password string) (*output.RegisterOutput, error) {
	now := i.timeProvider.Now()

	// ユーザーの重複チェック
	if err := i.checkUserExistence(ctx, username, email); err != nil {
		return nil, err
	}

	// パスワードのハッシュ化
	hashedPassword, err := i.hashPassword(password)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to hash password", err)
	}

	// 新しいユーザーIDを生成
	genUserID := i.idGenerator.Generate()
	newUserID := domain.NewUserID(genUserID)

	// ユーザーエンティティを作成
	newUser := domain.NewUser(newUserID, username, email, hashedPassword, now, now)

	// ユーザーを保存
	err = i.userRepository.Create(ctx, newUser)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to create user", err)
	}

	return output.NewRegisterOutput(newUserID), nil
}

// Login はユーザーの認証を行い、トークンペアを返します
func (i *AuthInteractor) Login(ctx context.Context, email, password string) (*output.TokenPairOutput, error) {
	now := i.timeProvider.Now()

	var tokenPair *output.TokenPairOutput

	err := i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error {
		// ユーザーをメールアドレスで検索
		foundUser, err := i.userRepository.FindByEmail(txCtx, email)
		if err != nil {
			if apperr.IsUserNotFound(err) {
				return apperr.NewInvalidCredentialsError("Invalid email or password")
			}
			return apperr.NewInternalError("Failed to retrieve user", err)
		}

		// パスワードの検証
		if err := i.verifyPassword(foundUser.PasswordHash(), password); err != nil {
			// verifyPasswordで既に適切なAppErrorが返されるため、そのまま返す
			return err
		}

		// トークンペアの生成
		newAccessTokenString, newRefreshTokenString, err := i.generateTokenPair(foundUser.ID(), now)
		if err != nil {
			return apperr.NewInternalError("Failed to generate token pair", err)
		}

		// リフレッシュトークンをデータベースに保存
		if err := i.storeRefreshToken(txCtx, foundUser.ID(), newRefreshTokenString, now); err != nil {
			return err
		}

		tokenPair = output.NewTokenPairOutput(newAccessTokenString, newRefreshTokenString)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// VerifyRefreshToken はリフレッシュトークンを検証し、新しいトークンペアを返します
func (i *AuthInteractor) VerifyRefreshToken(ctx context.Context, refreshToken string) (*output.TokenPairOutput, error) {
	now := i.timeProvider.Now()

	// 失効済みトークンのチェック
	if err := i.checkRevokedToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	var tokenPair *output.TokenPairOutput

	err := i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error {
		// リフレッシュトークンの検索と検証
		foundRefreshToken, err := i.findAndValidateRefreshToken(txCtx, refreshToken, now)
		if err != nil {
			return err
		}

		// 古いトークンの失効処理
		if err := i.revokeToken(txCtx, foundRefreshToken, now); err != nil {
			return err
		}

		// 新しいトークンペアの生成
		newAccessTokenString, newRefreshTokenString, err := i.generateTokenPair(foundRefreshToken.UserID(), now)
		if err != nil {
			return apperr.NewInternalError("Failed to generate token pair", err)
		}

		if err := i.storeRefreshToken(txCtx, foundRefreshToken.UserID(), newRefreshTokenString, now); err != nil {
			return err
		}

		tokenPair = output.NewTokenPairOutput(newAccessTokenString, newRefreshTokenString)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// TODO: DeleteByRefreshTokenを作りたい
// RevokeRefreshToken はリフレッシュトークンを失効させます
func (i *AuthInteractor) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	foundToken, err := i.refreshTokenRepository.FindByToken(ctx, refreshToken)
	if err != nil {
		if apperr.IsRefreshTokenNotFound(err) {
			// 既に存在しないトークンの場合は成功とする（冪等性）
			return nil
		}

		return apperr.NewInternalError("Failed to retrieve refresh token", err)
	}

	if err := i.refreshTokenRepository.Delete(ctx, foundToken.ID()); err != nil {
		return apperr.NewInternalError("Failed to delete refresh token", err)
	}

	return nil
}

// checkUserExistence はユーザー名またはメールアドレスが既に存在するかをチェックします。
func (i *AuthInteractor) checkUserExistence(ctx context.Context, username, email string) error {
	// ユーザー名が既に存在するか確認
	_, err := i.userRepository.FindByUsername(ctx, username)
	if err == nil {
		return apperr.NewConflictError("Username already exists")
	}

	// ユーザー名が見つからない場合は、メールアドレスの存在チェックに進む
	if !apperr.IsUserNotFound(err) {
		return apperr.NewInternalError("Failed to check username existence", err)
	}

	// メールアドレスが既に存在するか確認
	_, err = i.userRepository.FindByEmail(ctx, email)
	if err == nil {
		return apperr.NewConflictError("Email already exists")
	}

	if !apperr.IsUserNotFound(err) {
		return apperr.NewInternalError("Failed to check email existence", err)
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
			return apperr.NewInvalidCredentialsError("Invalid email or password")
		}
		// bcryptライブラリの内部エラー（ハッシュ形式が不正など）
		return apperr.NewInternalError("Failed to verify password", err)
	}
	return nil
}

// storeRefreshToken はリフレッシュトークンをデータベースに保存します
func (i *AuthInteractor) storeRefreshToken(ctx context.Context, userID domain.UserID, refreshTokenString string, now time.Time) error {
	genRefreshTokenID := i.idGenerator.Generate()
	newRefreshTokenID := domain.NewRefreshTokenID(genRefreshTokenID)

	newRefreshToken := domain.NewRefreshToken(
		newRefreshTokenID,
		userID,
		refreshTokenString,
		// TODO: config di する
		now.Add(config.RefreshTokenExpiration()),
		now,
	)

	if err := i.refreshTokenRepository.Create(ctx, newRefreshToken); err != nil {
		return apperr.NewInternalError("Failed to store refresh token", err)
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
			// TODO: ログをとる
		}

		return apperr.NewInvalidCredentialsError("Token has been revoked")
	}

	// 失効済みトークンが見つからない場合は正常
	if apperr.IsRevokedTokenNotFound(err) {
		return nil
	}

	// その他のエラー
	return apperr.NewInternalError("Failed to check revoked token", err)
}

// handleTokenReuseAttack はトークン再利用攻撃を検出した際の処理を行います
func (i *AuthInteractor) handleTokenReuseAttack(ctx context.Context, refreshToken string) error {
	foundToken, err := i.refreshTokenRepository.FindByToken(ctx, refreshToken)
	if err != nil {
		if apperr.IsRefreshTokenNotFound(err) {
			return nil // 既に削除済み
		}
		return err
	}

	// ユーザーIDに紐づくすべてのリフレッシュトークンを削除
	if err := i.refreshTokenRepository.DeleteByUserID(ctx, foundToken.UserID()); err != nil {
		return err
	}

	return nil
}

// findAndValidateRefreshToken はリフレッシュトークンを検索して検証します
func (i *AuthInteractor) findAndValidateRefreshToken(ctx context.Context, refreshToken string, now time.Time) (*domain.RefreshToken, error) {
	foundRefreshToken, err := i.refreshTokenRepository.FindByToken(ctx, refreshToken)
	if err != nil {
		if apperr.IsRefreshTokenNotFound(err) {
			return nil, apperr.NewInvalidCredentialsError("Invalid refresh token")
		}
		return nil, apperr.NewInternalError("Failed to retrieve refresh token for validation", err)
	}

	// 有効期限のチェック
	if now.After(foundRefreshToken.ExpiresAt()) {
		// 期限切れの場合は削除
		// TODO: not foundのケアは不要か？
		if err := i.refreshTokenRepository.Delete(ctx, foundRefreshToken.ID()); err != nil {
			// TODO: ログをとる
		}

		return nil, apperr.NewInvalidCredentialsError("Refresh token expired")
	}

	return foundRefreshToken, nil
}

// revokeToken はトークンを失効済みとして記録し削除します
func (i *AuthInteractor) revokeToken(ctx context.Context, refreshToken *domain.RefreshToken, now time.Time) error {
	// 失効済みトークンとして記録
	genRevokedTokenID := i.idGenerator.Generate()
	newRevokedTokenID := domain.NewRevokedTokenID(genRevokedTokenID)

	newRevokedToken := domain.NewRevokedToken(
		newRevokedTokenID,
		refreshToken.UserID(),
		refreshToken.Token(),
		refreshToken.ExpiresAt(),
		now,
	)

	if err := i.revokedTokenRepository.Create(ctx, newRevokedToken); err != nil {
		return apperr.NewInternalError("Failed to record revoked token", err)
	}

	// 元のリフレッシュトークンを削除
	if err := i.refreshTokenRepository.Delete(ctx, refreshToken.ID()); err != nil {
		if apperr.IsRefreshTokenNotFound(err) {
			// 既に削除済みの場合は無視
			return nil
		}
		// その他のエラーは内部エラーとして扱う
		return apperr.NewInternalError("Failed to delete refresh token", err)
	}

	return nil
}

// generateTokenPair はアクセストークンとリフレッシュトークン文字列を生成します。
func (i *AuthInteractor) generateTokenPair(userID domain.UserID, now time.Time) (string, string, error) {
	// JWTトークンの生成
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		// TODO: config di する
		"exp": now.Add(config.AccessTokenExpiration()).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(i.jwtSecret))
	if err != nil {
		return "", "", apperr.NewInternalError("Failed to sign access token", err)
	}

	// リフレッシュトークンの生成
	refreshTokenString := i.idGenerator.Generate()

	return accessTokenString, refreshTokenString, nil
}