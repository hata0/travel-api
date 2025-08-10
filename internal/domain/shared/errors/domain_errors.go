package errors

import "errors"

var (
	ErrInvalidUUID   = errors.New("invalid uuid format")
	ErrTripNotFound  = errors.New("trip not found")
	ErrUserNotFound  = errors.New("user not found")
	ErrTokenNotFound = errors.New("token not found")
)

func IsTripNotFound(err error) bool {
	return errors.Is(err, ErrTripNotFound)
}

func IsUserNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}

func IsTokenNotFound(err error) bool {
	return errors.Is(err, ErrTokenNotFound)
}
