package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

var Validator *validator.Validate

func ReadJSON(c *gin.Context, data any) error {
	// Prevent from reading large data from request.body, can prevent ddos attack
	maxBytes := 1_048_578 // 1MB = 1 << 20

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(maxBytes))

	decoder := json.NewDecoder(c.Request.Body)
	return decoder.Decode(data)
}

type ApiFunc func(c *gin.Context) error

func MakeHandlerFunc(f ApiFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := f(c)
		if err != nil {
			var apiError *ApiError
			if errors.As(err, &apiError) {
				c.JSON(apiError.StatusCode, apiError.Msg)
			} else {
				c.JSON(http.StatusInternalServerError, "internal server error")
			}
			Log.Error(err)
		}
	}
}
