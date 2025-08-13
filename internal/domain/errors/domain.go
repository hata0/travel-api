package errors

import "errors"

var (
	ErrTripNotFound         = errors.New("trip not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRevokedTokenNotFound = errors.New("revoked token not found")
)

func IsTripNotFound(err error) bool {
	return errors.Is(err, ErrTripNotFound)
}

func IsUserNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}

func IsRefreshTokenNotFound(err error) bool {
	return errors.Is(err, ErrRefreshTokenNotFound)
}

func IsRevokedTokenNotFound(err error) bool {
	return errors.Is(err, ErrRevokedTokenNotFound)
}
