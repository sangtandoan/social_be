package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type dbConfig struct {
	MaxLifeTime string `mapstructure:"DB_MAX_LIFE_TIME"`
	Password    string `mapstructure:"DB_PASSWORD"`
	User        string `mapstructure:"DB_USER"`

	Addr string

	MaxOpenConns int `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns int `mapstructure:"DB_MAX_IDLE_CONNS"`
}

type Config struct {
	DbConfig *dbConfig

	Addr string `mapstructure:"ADDR"`
}

var cfg Config

func LoadCfg() *Config {
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Can not file cfg file!")
	}

	viper.AutomaticEnv()

	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal("Can not Unmarshal cfg file!")
	}

	var dbConfig dbConfig
	err = viper.Unmarshal(&dbConfig)
	if err != nil {
		log.Fatal("Can not Unmarshal cfg file!")
	}

	fmt.Println(dbConfig.User, dbConfig.Password)
	dbConfig.Addr = fmt.Sprintf(
		"postgres://%s:%s@localhost:5432/social?sslmode=disable",
		dbConfig.User,
		dbConfig.Password,
	)

	cfg.DbConfig = &dbConfig

	return &cfg
}
