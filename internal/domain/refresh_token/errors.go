package refreshtoken

import apperr "github.com/hata0/travel-api/internal/domain/errors"

const (
	CodeRefreshTokenNotFound = "REFRESH_TOKEN_NOT_FOUND"
)

func NewRefreshTokenNotFoundError(message string, opts ...apperr.AppErrorOption) *apperr.AppError {
	return apperr.NewAppError(CodeRefreshTokenNotFound, message, opts...)
}
