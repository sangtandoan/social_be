package main

import (
	"github.com/go-playground/validator/v10"
	"github.com/sangtandoan/social/internal/config"
	"github.com/sangtandoan/social/internal/db"
	"github.com/sangtandoan/social/internal/service"
	"github.com/sangtandoan/social/internal/store"
	"github.com/sangtandoan/social/internal/utils"
	"go.uber.org/zap"
)

func main() {
	config := config.LoadCfg()

	utils.Log = zap.Must(zap.NewProduction()).Sugar()
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

	store := store.NewStore(db)

	app := application{config, store, mailer}

	mux := app.mount()

	utils.Log.Fatal(app.run(mux))
}
