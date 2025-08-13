package output

import "github.com/hata0/travel-api/internal/domain"

type RegisterOutput struct {
	UserID string
}

func NewRegisterOutput(userID domain.UserID) *RegisterOutput {
	return &RegisterOutput{
		UserID: userID.String(),
	}
}

type TokenPairOutput struct {
	AccessToken  string
	RefreshToken string
}

func NewTokenPairOutput(accessToken, refreshToken string) *TokenPairOutput {
	return &TokenPairOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}
