package main

import (
	"fmt"
	"log"

	"github.com/sangtandoan/social/internal/config"
	"github.com/sangtandoan/social/internal/db"
	"github.com/sangtandoan/social/internal/store"
)

func main() {
	config := config.LoadCfg()

	db, err := db.New(
		config.DbConfig.Addr,
		config.DbConfig.MaxOpenConns,
		config.DbConfig.MaxIdleConns,
		config.DbConfig.MaxLifeTime,
	)
	if err != nil {
		log.Panic("Failed to connect to database: ", err)
	}

	defer db.Close()
	fmt.Println("database connected")

	store := store.NewStore(db)

	app := application{config, store}

	mux := app.mount()

	log.Fatal(app.run(mux))
}
