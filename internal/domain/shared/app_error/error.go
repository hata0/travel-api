package app_error

import (
	"fmt"
)

// Error はアプリケーション固有のエラーを表すカスタムエラー型です。
type Error struct {
	// Code はクライアントに返される機械可読なエラーコードです。
	Code ErrorCode
	// Message は開発者向けのエラーメッセージです。
	Message string
	// cause はエラーの根本原因です（オプション）。
	cause error
}

// Error はerrorインターフェースを実装します。
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.cause)
	}
	return e.Message
}

// Unwrap はエラーチェーンのためにcauseを返します。
func (e *Error) Unwrap() error {
	return e.cause
}

var (
	// ErrInternalServerError は、予期せぬ内部エラーが発生した場合に返されます。
	// このエラーは通常、具体的なエラー情報でラップして使用します。
	ErrInternalServerError = &Error{Code: InternalServerError, Message: "internal server error"}
	// ErrInvalidCredentials は、認証情報が無効な場合に返されます。
	ErrInvalidCredentials = &Error{Code: InvalidCredentials, Message: "invalid credentials"}
	// ErrConfiguration は、設定エラーが発生した場合に返されます。
	ErrConfiguration = &Error{Code: ConfigurationError, Message: "configuration error"}
)

// エラー変数の定義
var (
	// ErrInvalidUUID は、UUIDの形式が無効な場合に返されます。
	ErrInvalidUUID = &Error{Code: ValidationError, Message: "invalid uuid format"}
	// ErrTripNotFound は、Tripが見つからない場合に返されます。
	ErrTripNotFound = &Error{Code: TripNotFound, Message: "trip not found"}
	// ErrTripAlreadyExists は、Tripが既に存在する場合に返されます。
	ErrTripAlreadyExists = &Error{Code: TripAlreadyExists, Message: "trip already exists"}
	// ErrUserNotFound は、ユーザーが見つからない場合に返されます。
	ErrUserNotFound = &Error{Code: UserNotFound, Message: "user not found"}
	// ErrUserAlreadyExists は、ユーザーが既に存在する場合に返されます。
	ErrUserAlreadyExists = &Error{Code: UserAlreadyExists, Message: "user already exists"}
	// ErrUsernameAlreadyExists は、ユーザー名が既に存在する場合に返されます。
	ErrUsernameAlreadyExists = &Error{Code: UsernameAlreadyExists, Message: "username already exists"}
	// ErrEmailAlreadyExists は、メールアドレスが既に存在する場合に返されます。
	ErrEmailAlreadyExists = &Error{Code: EmailAlreadyExists, Message: "email already exists"}
	// ErrTokenNotFound は、トークンが見つからない場合に返されます。
	ErrTokenNotFound = &Error{Code: TokenNotFound, Message: "token not found"}
	// ErrTokenAlreadyExists は、トークンが既に存在する場合に返されます。
	ErrTokenAlreadyExists = &Error{Code: TokenAlreadyExists, Message: "token already exists"}
)

// NewInternalServerError は、具体的なエラー原因を含む内部サーバーエラーを生成します。
func NewInternalServerError(cause error) error {
	return &Error{
		Code:    InternalServerError,
		Message: "internal server error",
		cause:   cause,
	}
}
