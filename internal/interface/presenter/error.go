package presenter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	shared_errors "travel-api/internal/shared/errors"

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

var httpStatusMap = map[string]int{
	shared_errors.CodeInvalidCredentials: http.StatusUnauthorized,
	shared_errors.CodeNotFound:           http.StatusNotFound,
	shared_errors.CodeConflict:           http.StatusConflict,
	shared_errors.CodeInternalError:      http.StatusInternalServerError,
}

func getHTTPStatus(code string) int {
	if status, ok := httpStatusMap[code]; ok {
		return status
	}

	// 未知のErrorCodeが渡された場合は、ログに記録し、500エラーを返す
	slog.Error("unknown error code", "code", code)
	return http.StatusInternalServerError
}

var safeMessageMap = map[string]string{
	shared_errors.CodeInvalidCredentials: "invalid credentials",
	shared_errors.CodeInternalError:      "an internal error occurred",
}

func ConvertToHTTPError(err error) (int, Error) {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		// バリデーションエラー: クライアント開発者向けに詳細な情報を提供します。
		return http.StatusBadRequest, Error{
			Code:    shared_errors.CodeValidationError,
			Message: "input validation failed. please check the details field for more information.",
			Details: formatValidationErrors(validationErrs),
		}
	}

	var unmarshalTypeError *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeError) {
		// JSONの型エラー: どのフィールドで問題があったかを具体的に伝えます。
		message := fmt.Sprintf("invalid json type provided for field '%s'", unmarshalTypeError.Field)
		return http.StatusBadRequest, Error{
			Code:    shared_errors.CodeValidationError,
			Message: message,
		}
	}

	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		// JSON構文エラー: クライアントには一般的なメッセージを返します。
		return http.StatusBadRequest, Error{
			Code:    shared_errors.CodeValidationError,
			Message: "the request body contains badly-formed json",
		}
	}

	var appErr *shared_errors.AppError
	if errors.As(err, &appErr) {
		// アプリケーションで定義されたドメインエラー。
		// 内部サーバーエラーの場合は、運用者が追跡できるよう詳細をログに出力します。
		if appErr.Code == shared_errors.CodeInternalError {
			slog.Error("Internal server error occurred", "details", appErr.Error())
		}

		if message, ok := safeMessageMap[appErr.Code]; ok {
			return getHTTPStatus(appErr.Code), Error{
				Code:    appErr.Code,
				Message: message,
			}
		}

		return getHTTPStatus(appErr.Code), Error{
			Code:    appErr.Code,
			Message: appErr.Message,
		}
	}

	// 上記のいずれにも当てはまらない、予期せぬエラー。
	// 詳細をログに記録し、クライアントには一般的なメッセージを返して、内部実装の詳細が漏洩しないようにします。
	slog.Error("An unexpected error occurred", "error", err)
	return http.StatusInternalServerError, Error{
		Code:    shared_errors.CodeInternalError,
		Message: "an unexpected internal server error has occurred. please contact support if the problem persists.",
	}
}
