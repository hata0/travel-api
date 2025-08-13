package errors

type AppError struct {
	Code    string
	Message string
	Cause   error
}

func NewAppError(code, message string, opts ...AppErrorOption) *AppError {
	e := &AppError{
		Code:    code,
		Message: message,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Code == t.Code
	}
	return false
}

type AppErrorOption func(*AppError)

func WithCause(cause error) AppErrorOption {
	return func(e *AppError) {
		e.Cause = cause
	}
}
