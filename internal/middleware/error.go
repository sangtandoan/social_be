package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangtandoan/social/internal/utils"
)

func GlobalErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				switch e := err.Err.(type) {
				case *utils.ApiError:
					c.JSON(e.StatusCode, e)
				default:
					c.JSON(http.StatusInternalServerError, gin.H{"msg": "An unexpected error occured"})
				}

				utils.Log.Error(err.Error())
				return
			}
		}
	}
}
