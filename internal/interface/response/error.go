package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"travel-api/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Error struct {
	StatusCode int
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
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

func mapErrorCodeToHTTPStatus(code domain.ErrorCode) int {
	switch code {
	case domain.ValidationError:
		return http.StatusBadRequest
	case domain.TripNotFound:
		return http.StatusNotFound
	case domain.UserNotFound:
		return http.StatusNotFound
	case domain.UserAlreadyExists:
		return http.StatusConflict
	case domain.InvalidCredentials:
		return http.StatusUnauthorized
	case domain.InternalServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func NewError(err error) Error {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return Error{
			StatusCode: http.StatusBadRequest,
			Code:       domain.ValidationError.String(),
			Message:    "validation failed",
			Details:    formatValidationErrors(validationErrs),
		}
	}

	var unmarshalTypeError *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeError) {
		return Error{
			StatusCode: http.StatusBadRequest,
			Code:       domain.ValidationError.String(),
			Message:    fmt.Sprintf("invalid type for field %s: expected %s, got %s", unmarshalTypeError.Field, unmarshalTypeError.Type.String(), unmarshalTypeError.Value),
		}
	}

	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) {
		return Error{
			StatusCode: http.StatusBadRequest,
			Code:       domain.ValidationError.String(),
			Message:    "invalid json syntax",
		}
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return Error{
			StatusCode: http.StatusBadRequest,
			Code:       domain.ValidationError.String(),
			Message:    "invalid json format",
		}
	}

	var appErr *domain.Error
	if errors.As(err, &appErr) {
		return Error{
			StatusCode: mapErrorCodeToHTTPStatus(appErr.Code),
			Code:       appErr.Code.String(),
			Message:    appErr.Error(),
		}
	}

	return Error{
		StatusCode: http.StatusInternalServerError,
		Code:       domain.InternalServerError.String(),
		Message:    err.Error(),
	}
}

func (e Error) JSON(c *gin.Context) {
	c.JSON(e.StatusCode, gin.H{
		"code":    e.Code,
		"message": e.Message,
		"details": e.Details,
	})
}
