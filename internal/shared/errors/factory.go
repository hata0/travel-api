package errors

func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Code:    CodeInternalError,
		Message: message,
		Cause:   cause,
	}
}
