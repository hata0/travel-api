package domain

import "errors"

var (
	ErrInternalServerError = errors.New("internal server error")
	ErrInvalidUUID         = errors.New("invalid uuid format")
)
