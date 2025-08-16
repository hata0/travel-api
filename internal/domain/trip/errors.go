package trip

import apperr "github.com/hata0/travel-api/internal/domain/errors"

const (
	CodeTripNotFound = "TRIP_NOT_FOUND"
)

func NewTripNotFoundError(message string, opts ...apperr.AppErrorOption) *apperr.AppError {
	return apperr.NewAppError(CodeTripNotFound, message, opts...)
}
