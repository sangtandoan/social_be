package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangtandoan/practice/internal/handler"
	"github.com/sangtandoan/practice/internal/middleware"
	"github.com/sangtandoan/practice/internal/utils"
)

type router struct {
	h *handler.Handler
}

func NewRouter(h *handler.Handler) *router {
	return &router{h}
}

func (r *router) SetupRouter() http.Handler {
	router := gin.New()

	router.Use(utils.ErrorHandler())
	api := router.Group("/api")
	{
		api.Use(utils.MakeHandlerFunc(middleware.MaxBodySize(1 << 20)))
		v1 := api.Group("/v1")
		{
			r.setupUserRoutes(v1)
			r.setupPostRoutes(v1)
		}
	}

	return router
}

func (r *router) setupUserRoutes(group *gin.RouterGroup) {
	users := group.Group("/users")
	{
		users.GET("", utils.MakeHandlerFunc(r.h.User.GetUsers))
		users.POST("", utils.MakeHandlerFunc(r.h.User.CreateUser))
	}
}

func (r *router) setupPostRoutes(group *gin.RouterGroup) {
	posts := group.Group("/posts")
	posts.GET("", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
}
