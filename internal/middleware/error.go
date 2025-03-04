package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/sangtandoan/social/internal/utils"
)

func GlobalErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				apiError := utils.NewApiError(
					http.StatusInternalServerError,
					gin.H{"msg": "An unexpected error occured"},
				)

				switch e := err.Err.(type) {
				case *utils.ApiError:
					apiError = e
				case *pq.Error:
					handlePostgresError(e, apiError)
				default:
				}

				c.JSON(apiError.StatusCode, apiError.Msg)
				utils.Log.Error(err.Error())
				return
			}
		}
	}
}

func handlePostgresError(err *pq.Error, fallback *utils.ApiError) {
	if err.Code == "23505" {
		fallback.StatusCode = http.StatusBadRequest
		fallback.Msg = gin.H{"msg": "username or email has existed"}
	}
}
