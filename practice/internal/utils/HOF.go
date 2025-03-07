package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sangtandoan/practice/internal/pkg/apperrors"
	"github.com/sangtandoan/practice/internal/pkg/logger"
)

type ApiFunc func(c *gin.Context) error

// There are 3 ways to handle global errors
// First is making HOF and apply it to every handler and every middleware
func MakeHandlerFunc(f ApiFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := f(c); err != nil {
			c.Error(err)
			return

			var appErr *apperrors.ApiError

			switch e := err.(type) {
			case *apperrors.ApiError:
				appErr = e
			case validator.ValidationErrors:
				appErr = HandleValidationErros(e)
			default:
				appErr = apperrors.NewApiError(http.StatusInternalServerError, "An unexpected error occurred")
			}

			c.JSON(appErr.StatusCode, appErr.Msg)
		}
	}
}

// Second is makeing a HandleError and calls it when need handle error
func HandleError(c *gin.Context, err error) {
	var appError *apperrors.ApiError
	switch e := err.(type) {
	case *apperrors.ApiError:
		appError = e
	case validator.ValidationErrors:
		appError = HandleValidationErros(e)
	default:
		appError = apperrors.NewApiError(http.StatusInternalServerError, "An unexpected error occured")
	}

	c.JSON(appError.StatusCode, appError.Msg)
}

// Third is using Gin feature to create global error middleware
// Using c.Error(err) to attach error to gin.Context
// Using c.Errors to get all attached errors in gin.Context
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			var appError *apperrors.ApiError
			for _, err := range c.Errors {
				switch e := err.Err.(type) {
				case *apperrors.ApiError:
					appError = e
				case validator.ValidationErrors:
					appError = HandleValidationErros(e)
				default:
					appError = apperrors.NewApiError(http.StatusInternalServerError, "An unexpected error occured")
				}

				c.AbortWithStatusJSON(appError.StatusCode, appError)
				logger.Log.Error(err)

				return
			}
		}
	}
}

func HandleValidationErros(err validator.ValidationErrors) *apperrors.ApiError {
	var arr []*apperrors.ValidationError

	for _, e := range err {
		var element apperrors.ValidationError
		element.Field = strings.ToLower(e.Field())
		element.Msg = apperrors.GetErroMsg(e)

		arr = append(arr, &element)
	}

	return apperrors.NewApiError(http.StatusBadRequest, arr)
}
