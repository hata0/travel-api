package usecase

import (
	"context"
	"time"

	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/hata0/travel-api/internal/usecase/output"
	"github.com/hata0/travel-api/internal/usecase/service"
)

//go:generate mockgen -destination mock/auth.go github.com/hata0/travel-api/internal/usecase AuthUsecase
type AuthUsecase interface {
	Register(ctx context.Context, username, email, password string) (*output.RegisterOutput, error)
	Login(ctx context.Context, email, password string) (*output.TokenPairOutput, error)
	VerifyRefreshToken(ctx context.Context, refreshToken string) (*output.TokenPairOutput, error)
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
}

type AuthSettings struct {
	RefreshTokenExpiration time.Duration
	PasswordHashCost       int
}

type AuthInteractor struct {
	userRepository         domain.UserRepository
	refreshTokenRepository domain.RefreshTokenRepository
	revokedTokenRepository domain.RevokedTokenRepository
	timeService            service.TimeService
	idService              service.IDService
	transactionManager     service.TransactionManager
	tokenService           service.TokenService
	authSettings           *AuthSettings
}

func NewAuthInteractor(
	userRepository domain.UserRepository,
	refreshTokenRepository domain.RefreshTokenRepository,
	revokedTokenRepository domain.RevokedTokenRepository,
	timeService service.TimeService,
	idService service.IDService,
	transactionManager service.TransactionManager,
	tokenService service.TokenService,
	authSettings *AuthSettings,
) *AuthInteractor {
	return &AuthInteractor{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		revokedTokenRepository: revokedTokenRepository,
		timeService:            timeService,
		idService:              idService,
		transactionManager:     transactionManager,
		tokenService:           tokenService,
		authSettings:           authSettings,
	}
}

// Register はユーザーを登録する
func (i *AuthInteractor) Register(ctx context.Context, username, email, password string) (*output.RegisterOutput, error) {
	now := i.timeService.Now()

	if err := i.checkUserExistence(ctx, username, email); err != nil {
		return nil, err
	}

	userIDStr := i.idService.Generate()
	userID := domain.NewUserID(userIDStr)

	user, err := domain.NewUserWithPassword(userID, username, email, password, i.authSettings.PasswordHashCost, now)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to create user", apperr.WithCause(err))
	}

	if err := i.userRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	return output.NewRegisterOutput(userID), nil
}

// Login はユーザーをログインさせ、トークンペアを生成する
func (i *AuthInteractor) Login(ctx context.Context, email, password string) (*output.TokenPairOutput, error) {
	now := i.timeService.Now()

	var tokenPair *output.TokenPairOutput

	err := i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error {
		user, err := i.findUserByEmail(txCtx, email)
		if err != nil {
			return err
		}

		if err := user.VerifyPassword(password); err != nil {
			return apperr.NewInvalidCredentialsError("Invalid email or password", apperr.WithCause(err))
		}

		accessToken, refreshToken, err := i.generateTokenPair(user.ID())
		if err != nil {
			return err
		}

		if err := i.storeRefreshToken(txCtx, user.ID(), refreshToken, now); err != nil {
			return err
		}

		tokenPair = output.NewTokenPairOutput(accessToken, refreshToken)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// VerifyRefreshToken はリフレッシュトークンを検証し、新しいトークンペアを生成する
func (i *AuthInteractor) VerifyRefreshToken(ctx context.Context, refreshToken string) (*output.TokenPairOutput, error) {
	now := i.timeService.Now()

	if err := i.checkRevokedToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	var tokenPair *output.TokenPairOutput

	err := i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error {
		foundRefreshToken, err := i.findAndValidateRefreshToken(txCtx, refreshToken, now)
		if err != nil {
			return err
		}

		if err := i.revokeToken(txCtx, foundRefreshToken, now); err != nil {
			return err
		}

		newAccessToken, newRefreshToken, err := i.generateTokenPair(foundRefreshToken.UserID())
		if err != nil {
			return err
		}

		if err := i.storeRefreshToken(txCtx, foundRefreshToken.UserID(), newRefreshToken, now); err != nil {
			return err
		}

		tokenPair = output.NewTokenPairOutput(newAccessToken, newRefreshToken)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// RevokeRefreshToken はリフレッシュトークンを失効させる
func (i *AuthInteractor) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	now := i.timeService.Now()

	return i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error {
		foundToken, err := i.refreshTokenRepository.FindByToken(txCtx, refreshToken)
		if err != nil {
			if apperr.IsRefreshTokenNotFound(err) {
				return nil
			}
			return err
		}

		return i.revokeToken(txCtx, foundToken, now)
	})
}

// checkUserExistence はユーザー名とメールアドレスの重複をチェックする
func (i *AuthInteractor) checkUserExistence(ctx context.Context, username, email string) error {
	_, err := i.userRepository.FindByUsername(ctx, username)
	if err == nil {
		return apperr.NewConflictError("Username already exists")
	}
	if !apperr.IsUserNotFound(err) {
		return err
	}

	_, err = i.userRepository.FindByEmail(ctx, email)
	if err == nil {
		return apperr.NewConflictError("Email already exists")
	}
	if !apperr.IsUserNotFound(err) {
		return err
	}

	return nil
}

// findUserByEmail はメールアドレスでユーザーを検索する
func (i *AuthInteractor) findUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := i.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if apperr.IsUserNotFound(err) {
			return nil, apperr.NewInvalidCredentialsError("Invalid email or password")
		}
		return nil, err
	}
	return user, nil
}

// storeRefreshToken はリフレッシュトークンを保存する
func (i *AuthInteractor) storeRefreshToken(ctx context.Context, userID domain.UserID, refreshTokenStr string, now time.Time) error {
	refreshTokenIDStr := i.idService.Generate()
	refreshTokenID := domain.NewRefreshTokenID(refreshTokenIDStr)

	refreshToken := domain.NewRefreshToken(
		refreshTokenID,
		userID,
		refreshTokenStr,
		now.Add(i.authSettings.RefreshTokenExpiration),
		now,
	)

	if err := i.refreshTokenRepository.Create(ctx, refreshToken); err != nil {
		return err
	}

	return nil
}

// checkRevokedToken は失効済みトークンかどうかをチェックします
func (i *AuthInteractor) checkRevokedToken(ctx context.Context, refreshToken string) error {
	_, err := i.revokedTokenRepository.FindByJTI(ctx, refreshToken)
	if err == nil {
		if err := i.handleTokenReuseAttack(ctx, refreshToken); err != nil {
			// TODO: ログをとる
			return apperr.NewInvalidCredentialsError("Token has been revoked", apperr.WithCause(err))
		}
		return apperr.NewInvalidCredentialsError("Token has been revoked")
	}

	if apperr.IsRevokedTokenNotFound(err) {
		return nil
	}

	return err
}

// handleTokenReuseAttack はトークン再利用攻撃を検出した際の最適化された処理を行います
func (i *AuthInteractor) handleTokenReuseAttack(ctx context.Context, refreshToken string) error {
	return i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error {
		foundToken, err := i.refreshTokenRepository.FindByToken(txCtx, refreshToken)
		if err != nil {
			if apperr.IsRefreshTokenNotFound(err) {
				return nil
			}
			return err
		}

		if err := i.refreshTokenRepository.DeleteByUserID(txCtx, foundToken.UserID()); err != nil {
			return err
		}

		return nil
	})
}

// findAndValidateRefreshToken はリフレッシュトークンを検索し、検証する
func (i *AuthInteractor) findAndValidateRefreshToken(ctx context.Context, refreshToken string, now time.Time) (*domain.RefreshToken, error) {
	foundRefreshToken, err := i.refreshTokenRepository.FindByToken(ctx, refreshToken)
	if err != nil {
		if apperr.IsRefreshTokenNotFound(err) {
			return nil, apperr.NewInvalidCredentialsError("Invalid refresh token")
		}
		return nil, err
	}

	if foundRefreshToken.IsExpired(now) {
		if err := i.refreshTokenRepository.Delete(ctx, foundRefreshToken.ID()); err != nil {
			// TODO: ログをとる
			return nil, apperr.NewInvalidCredentialsError("Refresh token expired", apperr.WithCause(err))
		}
		return nil, apperr.NewInvalidCredentialsError("Refresh token expired")
	}

	return foundRefreshToken, nil
}

// revokeToken はトークンを失効させる
func (i *AuthInteractor) revokeToken(ctx context.Context, refreshToken *domain.RefreshToken, now time.Time) error {
	revokedTokenIDStr := i.idService.Generate()
	revokedTokenID := domain.NewRevokedTokenID(revokedTokenIDStr)

	revokedToken := domain.NewRevokedToken(
		revokedTokenID,
		refreshToken.UserID(),
		refreshToken.Token(),
		refreshToken.ExpiresAt(),
		now,
	)

	if err := i.revokedTokenRepository.Create(ctx, revokedToken); err != nil {
		return err
	}

	if err := i.refreshTokenRepository.Delete(ctx, refreshToken.ID()); err != nil {
		if !apperr.IsRefreshTokenNotFound(err) {
			return err
		}
	}

	return nil
}

// generateTokenPair はアクセストークンとリフレッシュトークンのペアを生成する
func (i *AuthInteractor) generateTokenPair(userID domain.UserID) (string, string, error) {
	accessToken, err := i.tokenService.GenerateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := i.tokenService.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
