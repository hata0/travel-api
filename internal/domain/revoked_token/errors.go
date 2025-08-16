package revokedtoken

import apperr "github.com/hata0/travel-api/internal/domain/errors"

const (
	CodeRevokedTokenNotFound = "REVOKED_TOKEN_NOT_FOUND"
)

func NewRevokedTokenNotFoundError(message string, opts ...apperr.AppErrorOption) *apperr.AppError {
	return apperr.NewAppError(CodeRevokedTokenNotFound, message, opts...)
}
