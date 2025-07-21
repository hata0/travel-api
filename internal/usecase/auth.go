package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"travel-api/internal/config"
	"travel-api/internal/domain"
	"travel-api/internal/domain/shared/app_error"
	"travel-api/internal/domain/shared/clock"
	"travel-api/internal/domain/shared/transaction_manager"
	"travel-api/internal/domain/shared/uuid"
	"travel-api/internal/usecase/output"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -destination mock/auth.go travel-api/internal/usecase AuthUsecase
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

func (i *AuthInteractor) Register(ctx context.Context, username, email, password string) (output.RegisterOutput, error) {
	if err := i.checkUserExistence(ctx, username, email); err != nil {
		return output.RegisterOutput{}, err
	}

	// パスワードのハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return output.RegisterOutput{}, app_error.NewInternalServerError(err)
	}

	// 新しいユーザーIDを生成
	newUUID := i.uuidGenerator.NewUUID()
	userID, err := domain.NewUserID(newUUID)
	if err != nil {
		return output.RegisterOutput{}, app_error.NewInternalServerError(err)
	}

	// ユーザーエンティティを作成
	user := domain.NewUser(
		userID,
		username,
		email,
		string(hashedPassword),
		i.clock.Now(),
		i.clock.Now(),
	)

	// ユーザーを保存
	err = i.userRepository.Create(ctx, user)
	if err != nil {
		return output.RegisterOutput{}, err
	}

	return output.RegisterOutput{UserID: userID.String()}, nil
}

// checkUserExistence はユーザー名またはメールアドレスが既に存在するかをチェックします。
func (i *AuthInteractor) checkUserExistence(ctx context.Context, username, email string) error {
	// ユーザー名が既に存在するか確認
	_, err := i.userRepository.FindByUsername(ctx, username)
	if err == nil {
		return app_error.ErrUsernameAlreadyExists
	}

	// ユーザー名が見つからない場合は、メールアドレスの存在チェックに進む
	if !errors.Is(err, app_error.ErrUserNotFound) {
		return err
	}

	// メールアドレスが既に存在するか確認
	_, err = i.userRepository.FindByEmail(ctx, email)
	if err == nil {
		return app_error.ErrEmailAlreadyExists
	}

	if !errors.Is(err, app_error.ErrUserNotFound) {
		return err
	}

	// メールアドレスも見つからない場合は、ユーザーは存在しない
	return nil
}

func (i *AuthInteractor) Login(ctx context.Context, email, password string) (output.TokenPairOutput, error) {
	var tokenPair output.TokenPairOutput
	err := i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error {
		// ユーザーをメールアドレスで検索
		user, err := i.userRepository.FindByEmail(txCtx, email)
		if err != nil {
			// ユーザーが見つからない場合、認証情報が無効であると返す（セキュリティのため、ユーザーが存在しないことを直接伝えない）
			var appErr *app_error.Error
			if errors.As(err, &appErr) && appErr.Code == app_error.UserNotFound {
				return app_error.ErrInvalidCredentials
			}
			return err
		}

		// パスワードの検証
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				return app_error.ErrInvalidCredentials
			}
			return app_error.NewInternalServerError(err)
		}

		accessTokenString, refreshTokenString, err := i.generateTokenPair(user.ID, i.clock.Now())
		if err != nil {
			return err
		}

		// リフレッシュトークンをデータベースに保存
		refreshTokenID, err := domain.NewRefreshTokenID(refreshTokenString)
		if err != nil {
			return app_error.NewInternalServerError(err)
		}
		newRefreshToken := domain.NewRefreshToken(
			refreshTokenID,
			user.ID,
			refreshTokenString,
			i.clock.Now().Add(config.RefreshTokenExpiration()),
			i.clock.Now(),
		)
		if err := i.refreshTokenRepository.Create(txCtx, newRefreshToken); err != nil {
			return app_error.NewInternalServerError(err)
		}

		tokenPair = output.TokenPairOutput{Token: accessTokenString, RefreshToken: refreshTokenString}
		return nil
	})

	return tokenPair, err
}

func (i *AuthInteractor) VerifyRefreshToken(ctx context.Context, refreshToken string) (output.TokenPairOutput, error) {
	var tokenPair output.TokenPairOutput

	// まず失効済みトークンでないか確認
	_, err := i.revokedTokenRepository.FindByJTI(ctx, refreshToken)
	if err == nil {
		// 失効済みトークンが見つかった場合、再利用攻撃の可能性がある
		// このトークンに紐づくユーザーのセッションをすべて無効化する
		foundToken, findErr := i.refreshTokenRepository.FindByToken(ctx, refreshToken)
		if findErr == nil {
			// ユーザーIDに紐づくすべてのリフレッシュトークンを削除
			if delErr := i.refreshTokenRepository.DeleteByUserID(ctx, foundToken.UserID); delErr != nil {
				slog.Error("Failed to delete all refresh tokens for user after reuse detection", "error", delErr, "userID", foundToken.UserID.String())
			}
		}
		return output.TokenPairOutput{}, app_error.ErrInvalidCredentials
	}

	err = i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error {
		// リフレッシュトークンをデータベースから検索
		foundToken, err := i.refreshTokenRepository.FindByToken(txCtx, refreshToken)
		if err != nil {
			var appErr *app_error.Error
			if errors.As(err, &appErr) && appErr.Code == app_error.TokenNotFound {
				return app_error.ErrInvalidCredentials
			}
			return app_error.NewInternalServerError(err)
		}

		// リフレッシュトークンの有効期限をチェック
		if i.clock.Now().After(foundToken.ExpiresAt) {
			// 期限切れの場合は削除
			_ = i.refreshTokenRepository.Delete(txCtx, foundToken)
			return app_error.ErrInvalidCredentials
		}

		// 古いリフレッシュトークンを失効済みとして記録し、削除する
		revokedTokenID, uuidErr := domain.NewRevokedTokenID(i.uuidGenerator.NewUUID())
		if uuidErr != nil {
			return app_error.NewInternalServerError(uuidErr)
		}
		revokedToken := domain.NewRevokedToken(
			revokedTokenID,
			foundToken.UserID,
			foundToken.Token,
			foundToken.ExpiresAt,
			i.clock.Now(),
		)
		if err := i.revokedTokenRepository.Create(txCtx, revokedToken); err != nil {
			return app_error.NewInternalServerError(err)
		}
		if err := i.refreshTokenRepository.Delete(txCtx, foundToken); err != nil {
			return app_error.NewInternalServerError(err)
		}

		accessTokenString, newRefreshTokenString, err := i.generateTokenPair(foundToken.UserID, i.clock.Now())
		if err != nil {
			return err
		}

		newRefreshTokenID, err := domain.NewRefreshTokenID(newRefreshTokenString)
		if err != nil {
			return app_error.NewInternalServerError(err)
		}

		newRefreshToken := domain.NewRefreshToken(
			newRefreshTokenID,
			foundToken.UserID,
			newRefreshTokenString,
			i.clock.Now().Add(config.RefreshTokenExpiration()),
			i.clock.Now(),
		)
		if err := i.refreshTokenRepository.Create(txCtx, newRefreshToken); err != nil {
			return app_error.NewInternalServerError(err)
		}

		tokenPair = output.TokenPairOutput{Token: accessTokenString, RefreshToken: newRefreshTokenString}
		return nil
	})

	return tokenPair, err
}

func (i *AuthInteractor) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	// リフレッシュトークンをデータベースから削除
	foundToken, err := i.refreshTokenRepository.FindByToken(ctx, refreshToken)
	if err != nil {
		return err
	}
	return i.refreshTokenRepository.Delete(ctx, foundToken)
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
		return "", "", app_error.NewInternalServerError(err)
	}

	// リフレッシュトークンの生成
	refreshTokenString := i.uuidGenerator.NewUUID()

	return accessTokenString, refreshTokenString, nil
}
