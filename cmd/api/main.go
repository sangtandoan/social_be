package main

import (
	"github.com/go-playground/validator/v10"
	"github.com/sangtandoan/social/internal/config"
	"github.com/sangtandoan/social/internal/db"
	"github.com/sangtandoan/social/internal/service"
	"github.com/sangtandoan/social/internal/service/cache"
	"github.com/sangtandoan/social/internal/store"
	"github.com/sangtandoan/social/internal/utils"
	"go.uber.org/zap"
)

func main() {
	config := config.LoadCfg()

	utils.Log = zap.Must(zap.NewDevelopment()).Sugar()
	defer utils.Log.Sync()

	utils.Validator = validator.New()

	db, err := db.New(
		config.DbConfig.Addr,
		config.DbConfig.MaxOpenConns,
		config.DbConfig.MaxIdleConns,
		config.DbConfig.MaxLifeTime,
	)
	if err != nil {
		utils.Log.Panic("Failed to connect to database: ", err)
	}

	defer db.Close()
	utils.Log.Info("database connected")

	mailer := service.NewSMTPMailer(config.MailerConfig)

	redisClient := cache.NewRedisClient(config.CacheConfig)
	cache := cache.NewCacheService(redisClient)

	store := store.NewStore(db)

	app := application{config, store, mailer, cache, nil}

	mux := app.mount()

	utils.Log.Fatal(app.run(mux))
}
