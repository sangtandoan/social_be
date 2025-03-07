package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sangtandoan/social/internal/config"
	"github.com/sangtandoan/social/internal/middleware"
	"github.com/sangtandoan/social/internal/service"
	"github.com/sangtandoan/social/internal/service/cache"
	"github.com/sangtandoan/social/internal/store"
	"github.com/sangtandoan/social/internal/utils"

	"github.com/gin-gonic/gin"
)

type application struct {
	config *config.Config
	store  *store.Store
	mailer service.Mailer
	cache  *cache.CacheService
	srv    *http.Server
}

func (a *application) mount() http.Handler {
	r := gin.Default()

	api := r.Group("/api")
	api.Use(middleware.GlobalErrorHandler())
	{
		v1 := api.Group("/v1")
		{
			a.setupPostRoutes(v1)
			a.setupUserRoutes(v1)
			v1.GET("/feeds", a.getUserFeedHandler)
		}
	}

	return r
}

func (a *application) setupUserRoutes(group *gin.RouterGroup) {
	users := group.Group("/users")

	users.POST("", a.createUserHandler)
	users.PATCH("/activate", a.activateUserHandler)
	users.POST("/login", a.loginHandler)
}

func (a *application) setupPostRoutes(group *gin.RouterGroup) {
	posts := group.Group("/posts")

	posts.POST("", utils.MakeHandlerFunc(a.createPostHandler))
	posts.PATCH("/:id", utils.MakeHandlerFunc(a.updatePostHandler))
	posts.GET("/:id", utils.MakeHandlerFunc(a.getPostHandler))
	posts.GET("", utils.MakeHandlerFunc(a.getPostsHandler))
}

func (a *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         a.config.Addr,
		Handler:      mux,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Minute,
	}
	a.srv = srv

	utils.Log.Infof("server starts on port %s", a.config.Addr)

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		<-quit
		utils.Log.Info("Sever is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		shutdown <- srv.Shutdown(ctx)
	}()

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		utils.Log.Infof("Server error: $v", err)
		return err
	}

	err = <-shutdown
	if err != nil {
		utils.Log.Fatalf("Server forced to shutdown: %v", err)
		return err
	}

	utils.Log.Info("Server gratefull shutdown")
	return nil
}
