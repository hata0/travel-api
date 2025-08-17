package user

import apperr "github.com/hata0/travel-api/internal/domain/errors"

const (
	CodeUserNotFound = "USER_NOT_FOUND"
)

func NewUserNotFoundError(opts ...apperr.AppErrorOption) *apperr.AppError {
	return apperr.NewAppError(CodeUserNotFound, "User not found", opts...)
}
