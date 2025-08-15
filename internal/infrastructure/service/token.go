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

type TokenSettings struct {
	Audience              string
	Issuer                string
	AccessTokenExpiration time.Duration
	JTIBytes              int
	ECDSAPrivateKey       *ecdsa.PrivateKey
	RefreshTokenBytes     int
}

type TokenServiceImpl struct {
	timeService service.TimeService
	settings    *TokenSettings
}

func NewTokenService(timeService service.TimeService, settings *TokenSettings) service.TokenService {
	return &TokenServiceImpl{
		timeService: timeService,
		settings:    settings,
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
		Audience:  jwt.ClaimStrings{t.settings.Audience},
		Issuer:    t.settings.Issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(t.settings.AccessTokenExpiration)),
		ID:        jti,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	signed, err := token.SignedString(t.settings.ECDSAPrivateKey)
	if err != nil {
		return "", apperr.NewInternalError("Failed to sign access token", apperr.WithCause(err))
	}

	return signed, nil
}

// GenerateRefreshToken はリフレッシュトークンを生成する
func (t *TokenServiceImpl) GenerateRefreshToken() (string, error) {
	b := make([]byte, t.settings.RefreshTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", apperr.NewInternalError("Failed to generate random bytes for refresh token", apperr.WithCause(err))
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateJTI はJTI（JWT ID）を生成する
func (t *TokenServiceImpl) generateJTI() (string, error) {
	b := make([]byte, t.settings.JTIBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
