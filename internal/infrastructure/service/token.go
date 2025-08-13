package service

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/hata0/travel-api/internal/usecase/service"
)

// TokenSettings はトークン関連の設定を保持する
type TokenSettings struct {
	Audience              string
	Issuer                string
	AccessTokenExpiration time.Duration
	JTIBytes              int
	ECDSAPrivateKey       *ecdsa.PrivateKey
	RefreshTokenBytes     int
}

// TokenServiceImpl はservice.TokenServiceの実装
type TokenServiceImpl struct {
	timeService   service.TimeService
	tokenSettings *TokenSettings
}

// NewTokenService は新しいTokenServiceを作成する
func NewTokenService(timeService service.TimeService, tokenSettings *TokenSettings) service.TokenService {
	return &TokenServiceImpl{
		timeService:   timeService,
		tokenSettings: tokenSettings,
	}
}

// GenerateAccessToken はアクセストークンを生成する
func (t *TokenServiceImpl) GenerateAccessToken(userID domain.UserID) (string, error) {
	now := t.timeService.Now()

	jti, err := t.generateJTI()
	if err != nil {
		return "", apperr.NewInternalError("Failed to generate JTI for access token", apperr.WithCause(err))
	}

	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		Audience:  jwt.ClaimStrings{t.tokenSettings.Audience},
		Issuer:    t.tokenSettings.Issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(t.tokenSettings.AccessTokenExpiration)),
		ID:        jti,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	signed, err := token.SignedString(t.tokenSettings.ECDSAPrivateKey)
	if err != nil {
		return "", apperr.NewInternalError("Failed to sign access token", apperr.WithCause(err))
	}

	return signed, nil
}

// GenerateRefreshToken はリフレッシュトークンを生成する
func (t *TokenServiceImpl) GenerateRefreshToken() (string, error) {
	b := make([]byte, t.tokenSettings.RefreshTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", apperr.NewInternalError("Failed to generate random bytes for refresh token", apperr.WithCause(err))
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateJTI はJTI（JWT ID）を生成する
func (t *TokenServiceImpl) generateJTI() (string, error) {
	b := make([]byte, t.tokenSettings.JTIBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
