package errors

import "errors"

var (
	ErrInvalidUUID           = errors.New("invalid uuid format")
	ErrTripNotFound          = errors.New("trip not found")
	ErrTripAlreadyExists     = errors.New("trip already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrTokenNotFound         = errors.New("token not found")
	ErrTokenAlreadyExists    = errors.New("token already exists")
)
