package usecase

import (
	"context"
	"errors"
	"log/slog"
	"travel-api/internal/config"
	"travel-api/internal/domain"
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
	clock                  domain.Clock
	uuidGenerator          domain.UUIDGenerator
	transactionManager     domain.TransactionManager
}

func NewAuthInteractor(userRepository domain.UserRepository, refreshTokenRepository domain.RefreshTokenRepository, revokedTokenRepository domain.RevokedTokenRepository, clock domain.Clock, uuidGenerator domain.UUIDGenerator, transactionManager domain.TransactionManager) *AuthInteractor {
	return &AuthInteractor{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		revokedTokenRepository: revokedTokenRepository,
		clock:                  clock,
		uuidGenerator:          uuidGenerator,
		transactionManager:     transactionManager,
	}
}

func (i *AuthInteractor) Register(ctx context.Context, username, email, password string) (output.RegisterOutput, error) {
	if err := i.checkUserExistence(ctx, username, email); err != nil {
		return output.RegisterOutput{}, err
	}

	// パスワードのハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return output.RegisterOutput{}, domain.NewInternalServerError(err)
	}

	// 新しいユーザーIDを生成
	newUUID := i.uuidGenerator.NewUUID()
	userID, err := domain.NewUserID(newUUID)
	if err != nil {
		return output.RegisterOutput{}, domain.NewInternalServerError(err)
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
		return domain.ErrUsernameAlreadyExists
	}
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Code != domain.UserNotFound {
		return err
	}

	// メールアドレスが既に存在するか確認
	_, err = i.userRepository.FindByEmail(ctx, email)
	if err == nil {
		return domain.ErrEmailAlreadyExists
	}
	if !errors.As(err, &appErr) || appErr.Code != domain.UserNotFound {
		return err
	}
	return nil
}

func (i *AuthInteractor) Login(ctx context.Context, email, password string) (output.TokenPairOutput, error) {
	var tokenPair output.TokenPairOutput
	err := i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error {
		// ユーザーをメールアドレスで検索
		user, err := i.userRepository.FindByEmail(txCtx, email)
		if err != nil {
			// ユーザーが見つからない場合、認証情報が無効であると返す（セキュリティのため、ユーザーが存在しないことを直接伝えない）
			var appErr *domain.Error
			if errors.As(err, &appErr) && appErr.Code == domain.UserNotFound {
				return domain.ErrInvalidCredentials
			}
			return err
		}

		// パスワードの検証
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				return domain.ErrInvalidCredentials
			}
			return domain.NewInternalServerError(err)
		}

		// JWTトークンの生成
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID.String(),
			"exp":     i.clock.Now().Add(config.AccessTokenExpiration()).Unix(),
		})

		jwtSecret, err := config.JWTSecret()
		if err != nil {
			return domain.NewInternalServerError(err)
		}

		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			return domain.NewInternalServerError(err)
		}

		// リフレッシュトークンの生成
		refreshTokenString := i.uuidGenerator.NewUUID()
		refreshTokenID, err := domain.NewRefreshTokenID(refreshTokenString)
		if err != nil {
			return domain.NewInternalServerError(err)
		}

		// リフレッシュトークンをデータベースに保存
		newRefreshToken := domain.NewRefreshToken(
			refreshTokenID,
			user.ID,
			refreshTokenString,
			i.clock.Now().Add(config.RefreshTokenExpiration()),
			i.clock.Now(),
		)
		err = i.refreshTokenRepository.Create(txCtx, newRefreshToken)
		if err != nil {
			return domain.NewInternalServerError(err)
		}

		tokenPair = output.TokenPairOutput{Token: tokenString, RefreshToken: refreshTokenString}
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
		return output.TokenPairOutput{}, domain.ErrInvalidCredentials
	}

	err = i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error {
		// リフレッシュトークンをデータベースから検索
		foundToken, err := i.refreshTokenRepository.FindByToken(txCtx, refreshToken)
		if err != nil {
			var appErr *domain.Error
			if errors.As(err, &appErr) && appErr.Code == domain.TokenNotFound {
				return domain.ErrInvalidCredentials
			}
			return domain.NewInternalServerError(err)
		}

		// リフレッシュトークンの有効期限をチェック
		if i.clock.Now().After(foundToken.ExpiresAt) {
			// 期限切れの場合は削除
			_ = i.refreshTokenRepository.Delete(txCtx, foundToken)
			return domain.ErrInvalidCredentials
		}

		// 古いリフレッシュトークンを失効済みとして記録し、削除する
		revokedTokenID, uuidErr := domain.NewRevokedTokenID(i.uuidGenerator.NewUUID())
		if uuidErr != nil {
			return domain.NewInternalServerError(uuidErr)
		}
		revokedToken := domain.NewRevokedToken(
			revokedTokenID,
			foundToken.UserID,
			foundToken.Token,
			foundToken.ExpiresAt,
			i.clock.Now(),
		)
		if err := i.revokedTokenRepository.Create(txCtx, revokedToken); err != nil {
			return domain.NewInternalServerError(err)
		}
		if err := i.refreshTokenRepository.Delete(txCtx, foundToken); err != nil {
			return domain.NewInternalServerError(err)
		}

		// 新しいアクセストークンとリフレッシュトークンを生成
		user, err := i.userRepository.FindByID(txCtx, foundToken.UserID)
		if err != nil {
			return err
		}

		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID.String(),
			"exp":     i.clock.Now().Add(config.AccessTokenExpiration()).Unix(),
		})

		jwtSecret, err := config.JWTSecret()
		if err != nil {
			return domain.NewInternalServerError(err)
		}

		accessTokenString, err := accessToken.SignedString([]byte(jwtSecret))
		if err != nil {
			return domain.NewInternalServerError(err)
		}

		newRefreshTokenString := i.uuidGenerator.NewUUID()
		newRefreshTokenID, err := domain.NewRefreshTokenID(newRefreshTokenString)
		if err != nil {
			return domain.NewInternalServerError(err)
		}

		newRefreshToken := domain.NewRefreshToken(
			newRefreshTokenID,
			user.ID,
			newRefreshTokenString,
			i.clock.Now().Add(config.RefreshTokenExpiration()),
			i.clock.Now(),
		)
		if err := i.refreshTokenRepository.Create(txCtx, newRefreshToken); err != nil {
			return domain.NewInternalServerError(err)
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
