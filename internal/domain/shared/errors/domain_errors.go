package errors

import "errors"

var (
	ErrInvalidUUID        = errors.New("invalid uuid format")
	ErrTripNotFound       = errors.New("trip not found")
	ErrTripAlreadyExists  = errors.New("trip already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrTokenNotFound      = errors.New("token not found")
	ErrTokenAlreadyExists = errors.New("token already exists")
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
