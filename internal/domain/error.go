package domain

import "fmt"

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

// エラー変数の定義
var (
	// ErrInvalidUUID は、UUIDの形式が無効な場合に返されます。
	ErrInvalidUUID = &Error{Code: ValidationError, Message: "invalid uuid format"}
	// ErrTripNotFound は、Tripが見つからない場合に返されます。
	ErrTripNotFound = &Error{Code: TripNotFound, Message: "trip not found"}
	// ErrInternalServerError は、予期せぬ内部エラーが発生した場合に返されます。
	// このエラーは通常、具体的なエラー情報でラップして使用します。
	ErrInternalServerError = &Error{Code: InternalServerError, Message: "internal server error"}
)

// NewInternalServerError は、具体的なエラー原因を含む内部サーバーエラーを生成します。
func NewInternalServerError(cause error) error {
	return &Error{
		Code:    InternalServerError,
		Message: "internal server error",
		cause:   cause,
	}
}
