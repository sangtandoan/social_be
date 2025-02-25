package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sangtandoan/social/internal/config"
	"github.com/sangtandoan/social/internal/store"

	"github.com/gin-gonic/gin"
)

type application struct {
	config *config.Config
	store  *store.Store
}

func (a *application) mount() http.Handler {
	r := gin.Default()

	api := r.Group("/api")

	{
		v1 := api.Group("/v1")
		{
			v1.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "OK",
				})
			})
		}
	}

	return r
}

func (a *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         a.config.Addr,
		Handler:      mux,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Minute,
	}

	fmt.Printf("Server starts on port %s", a.config.Addr)

	return srv.ListenAndServe()
}
