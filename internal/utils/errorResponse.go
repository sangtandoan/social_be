package utils

import (
	"fmt"
	"net/http"
)

var (
	ErrInvalidJSON  = NewApiError(http.StatusBadRequest, "invalid json format")
	ErrNotFound     = NewApiError(http.StatusNotFound, "resource not found")
	ErrUnauthorized = NewApiError(http.StatusUnauthorized, "unauthorized")
)

type ApiError struct {
	Msg        any
	StatusCode int
}

func (err *ApiError) Error() string {
	return fmt.Sprintf("api error code: %d", err.StatusCode)
}

func NewApiError(statusCode int, msg any) *ApiError {
	return &ApiError{
		StatusCode: statusCode,
		Msg:        msg,
	}
}

func InvalidRequestData(errors []error) *ApiError {
	return &ApiError{
		StatusCode: http.StatusUnprocessableEntity,
		Msg:        errors,
	}
}
