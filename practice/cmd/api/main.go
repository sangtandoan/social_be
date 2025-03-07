package main

import (
	"log"
	"net/http"

	"github.com/sangtandoan/practice/internal/config"
	"github.com/sangtandoan/practice/internal/handler"
	"github.com/sangtandoan/practice/internal/pkg/logger"
	"github.com/sangtandoan/practice/internal/pkg/validator"
	"github.com/sangtandoan/practice/internal/repo"
	"github.com/sangtandoan/practice/internal/router"
	"github.com/sangtandoan/practice/internal/service"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		panic("can not load config")
	}

	logger.Init(config.Server.Env)
	defer logger.Log.Sync()

	repo := repo.NewRepo()

	service := service.NewService(repo)
	validator := validator.New()

	handler := handler.NewHanlder(service, validator)

	router := router.NewRouter(handler)

	err = http.ListenAndServe(":8080", router.SetupRouter())
	if err != nil {
		log.Fatal(err)
	}
}
