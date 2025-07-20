package domain

type ErrorCode string

func (e ErrorCode) String() string {
	return string(e)
}

const (
	InternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
	ValidationError     ErrorCode = "VALIDATION_ERROR"
	TripNotFound        ErrorCode = "TRIP_NOT_FOUND"
	TripAlreadyExists   ErrorCode = "TRIP_ALREADY_EXISTS"
	UserNotFound        ErrorCode = "USER_NOT_FOUND"
	UserAlreadyExists   ErrorCode = "USER_ALREADY_EXISTS"
	InvalidCredentials  ErrorCode = "INVALID_CREDENTIALS"
	TokenNotFound       ErrorCode = "TOKEN_NOT_FOUND"
	TokenAlreadyExists  ErrorCode = "TOKEN_ALREADY_EXISTS"
)
