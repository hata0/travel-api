package errors

func NewInvalidCredentialsError(message string, opts ...AppErrorOption) *AppError {
	return NewAppError(CodeInvalidCredentials, message, opts...)
}

func NewConflictError(message string, opts ...AppErrorOption) *AppError {
	return NewAppError(CodeConflict, message, opts...)
}

func NewInternalError(message string, opts ...AppErrorOption) *AppError {
	return NewAppError(CodeInternalError, message, opts...)
}
