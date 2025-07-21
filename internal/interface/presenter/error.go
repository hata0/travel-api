package presenter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"travel-api/internal/domain/shared/app_error"

	"github.com/go-playground/validator/v10"
)

type Error struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func formatValidationErrors(errs validator.ValidationErrors) []ValidationErrorDetail {
	var details []ValidationErrorDetail
	for _, err := range errs {
		field := strings.ToLower(err.Field())
		var message string
		switch err.Tag() {
		case "required":
			message = fmt.Sprintf("%s is a required field", field)
		default:
			message = fmt.Sprintf("%s is not valid", field)
		}
		details = append(details, ValidationErrorDetail{
			Field:   field,
			Message: message,
		})
	}
	return details
}

func mapErrorCodeToHTTPStatus(code app_error.ErrorCode) int {
	switch code {
	case app_error.ValidationError:
		return http.StatusBadRequest
	case app_error.TripNotFound:
		return http.StatusNotFound
	case app_error.TripAlreadyExists:
		return http.StatusConflict
	case app_error.UserNotFound:
		return http.StatusNotFound
	case app_error.UserAlreadyExists,
		app_error.UsernameAlreadyExists,
		app_error.EmailAlreadyExists:
		return http.StatusConflict
	case app_error.InvalidCredentials:
		return http.StatusUnauthorized
	case app_error.InternalServerError:
		return http.StatusInternalServerError
	case app_error.TokenNotFound:
		return http.StatusNotFound
	case app_error.TokenAlreadyExists:
		return http.StatusConflict
	case app_error.ConfigurationError:
		return http.StatusInternalServerError
	default:
		// 未知のErrorCodeが渡された場合は、ログに記録し、500エラーを返す
		slog.Error("unknown error code", "code", code)
		return http.StatusInternalServerError
	}
}

func ConvertToHTTPError(err error) (int, Error) {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		// バリデーションエラー: クライアント開発者向けに詳細な情報を提供します。
		return http.StatusBadRequest, Error{
			Code:    app_error.ValidationError.String(),
			Message: "Input validation failed. Please check the details field for more information.",
			Details: formatValidationErrors(validationErrs),
		}
	}

	var unmarshalTypeError *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeError) {
		// JSONの型エラー: どのフィールドで問題があったかを具体的に伝えます。
		message := fmt.Sprintf("Invalid JSON type provided for field '%s'.", unmarshalTypeError.Field)
		return http.StatusBadRequest, Error{
			Code:    app_error.ValidationError.String(),
			Message: message,
		}
	}

	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		// JSON構文エラー: クライアントには一般的なメッセージを返します。
		return http.StatusBadRequest, Error{
			Code:    app_error.ValidationError.String(),
			Message: "The request body contains badly-formed JSON.",
		}
	}

	var appErr *app_error.Error
	if errors.As(err, &appErr) {
		// アプリケーションで定義されたドメインエラー。
		// 内部サーバーエラーの場合は、運用者が追跡できるよう詳細をログに出力します。
		if errors.As(err, &appErr) && appErr.Code == app_error.InternalServerError {
			slog.Error("Internal server error occurred", "details", appErr.Error())
		}
		// クライアントには、エラーの原因(cause)を含まない、公開可能なメッセージのみを返します。
		return mapErrorCodeToHTTPStatus(appErr.Code), Error{
			Code:    appErr.Code.String(),
			Message: appErr.Message,
		}
	}

	// 上記のいずれにも当てはまらない、予期せぬエラー。
	// 詳細をログに記録し、クライアントには一般的なメッセージを返して、内部実装の詳細が漏洩しないようにします。
	slog.Error("An unexpected error occurred", "error", err)
	return http.StatusInternalServerError, Error{
		Code:    app_error.InternalServerError.String(),
		Message: "An unexpected internal server error has occurred. Please contact support if the problem persists.",
	}
}
