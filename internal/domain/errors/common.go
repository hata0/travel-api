package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type AppError struct {
	code       string
	message    string
	cause      error
	stackTrace string
}

type AppErrorOption func(*AppError)

func NewAppError(code, message string, opts ...AppErrorOption) *AppError {
	e := &AppError{
		code:       code,
		message:    message,
		stackTrace: captureStackTrace(2),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.code, e.message, e.cause)
	}
	return fmt.Sprintf("[%s] %s", e.code, e.message)
}

func (e *AppError) Unwrap() error {
	return e.cause
}

func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.code == t.code
	}
	return false
}

func (e *AppError) String() string {
	var sb strings.Builder
	sb.WriteString(e.Error())

	if e.stackTrace != "" {
		sb.WriteString("\nStack Trace:\n")
		sb.WriteString(e.stackTrace)
	}

	return sb.String()
}

// Getters
func (e *AppError) Code() string       { return e.code }
func (e *AppError) Message() string    { return e.message }
func (e *AppError) Cause() error       { return e.cause }
func (e *AppError) StackTrace() string { return e.stackTrace }

func WithCause(cause error) AppErrorOption {
	return func(e *AppError) {
		e.cause = cause
	}
}

func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

func IsAppErrorWithCode(err error, code string) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.code == code
	}
	return false
}

func GetAppError(err error) *AppError {
	var appErr *AppError
	if ok := errors.As(err, &appErr); ok {
		return appErr
	}
	return nil
}

func captureStackTrace(skip int) string {
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(skip, pcs[:])

	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	var sb strings.Builder

	for {
		frame, more := frames.Next()

		if strings.Contains(frame.Function, "runtime.") {
			if !more {
				break
			}
			continue
		}

		sb.WriteString(fmt.Sprintf("  %s:%d %s\n",
			frame.File, frame.Line, frame.Function))

		if !more {
			break
		}
	}

	return sb.String()
}
