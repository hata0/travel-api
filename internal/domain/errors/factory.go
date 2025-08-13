package errors

func NewInvalidCredentialsError(message string) *AppError {
	return &AppError{
		Code:    CodeInvalidCredentials,
		Message: message,
	}
}

func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:    CodeNotFound,
		Message: message,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		Code:    CodeConflict,
		Message: message,
	}
}

func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Code:    CodeInternalError,
		Message: message,
		Cause:   cause,
	}
}
