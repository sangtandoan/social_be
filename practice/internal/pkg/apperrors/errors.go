package apperrors

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var (
	ErrInvalidJSON = NewApiError(http.StatusBadRequest, "invalid json format")
	ErrNotFound    = NewApiError(http.StatusNotFound, "resource not found")
)

type ValidationError struct {
	Field string `json:"field,omitempty"`
	Msg   string `json:"msg,omitempty"`
}

type ApiError struct {
	Msg        any `json:"msg"`
	StatusCode int `json:"status_code"`
}

func (err *ApiError) Error() string {
	return fmt.Sprintf("api error code: %d", err.StatusCode)
}

func NewApiError(StatusCode int, Msg any) *ApiError {
	return &ApiError{
		Msg,
		StatusCode,
	}
}

func InvalidRequestData(errors []*ValidationError) *ApiError {
	return &ApiError{
		StatusCode: http.StatusUnprocessableEntity,
		Msg:        errors,
	}
}

func GetErroMsg(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Should be at least " + err.Param() + " characters long"
	case "max":
		return "Should be at most " + err.Param() + " characters long"
	case "gte":
		return "Should be greater than or equal to " + err.Param()
	default:
		return "Invalid value"
	}
}
