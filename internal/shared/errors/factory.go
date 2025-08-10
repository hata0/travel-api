package errors

import "fmt"

func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    CodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

func NewInvalidCredentialsError(message string) *AppError {
	return &AppError{
		Code:    CodeInvalidCredentials,
		Message: message,
	}
}

func NewConflictError(resource string) *AppError {
	return &AppError{
		Code:    CodeConflict,
		Message: fmt.Sprintf("%s already exists", resource),
	}
}

func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Code:    CodeInternalError,
		Message: message,
		Cause:   cause,
	}
}
