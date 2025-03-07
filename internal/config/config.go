package config

import (
	"fmt"
	"log"
	"time"

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

type RedisConfig struct {
	Addr        string        `mapstrucutre:"REDIS_ADDR"`
	Password    string        `mapstrucutre:"REDIS_PASSWORD"`
	MaxRetires  int           `mapstrucutre:"REDIS_RETRIES"`  // redis has already implement retry mechanism default = 3, if you want to implement your own, set = 0
	PoolSize    int           `mapstrucutre:"REDIS_POOLSIZE"` // = maxOpenConns
	DB          int           `mapstrucutre:"REDIS_DB"`
	IdleTimeout time.Duration `mapstrucutre:"REDIS_IDLE_TIMEOUT"` // = MaxIdleLifeTime
}

type MailerConfig struct {
	Host       string `mapstructure:"MAIL_HOST"`
	From       string `mapstructure:"MAIL_FROM"`
	Username   string `mapstructure:"MAIL_USERNAME"`
	Password   string `mapstructure:"MAIL_PASSWORD"`
	ServerAddr string `mapstructure:"ADDR"`
	Port       int    `mapstructure:"MAIL_PORT"`
}

type Config struct {
	DbConfig     *dbConfig
	MailerConfig *MailerConfig
	CacheConfig  *RedisConfig
	Addr         string `mapstructure:"ADDR"`
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

	var mailerConfig MailerConfig
	err = viper.Unmarshal(&mailerConfig)
	if err != nil {
		log.Fatal("Can not Unmarshal cfg file!")
	}

	var redisConfig RedisConfig
	err = viper.Unmarshal(&redisConfig)
	if err != nil {
		log.Fatal("can not unmarshal cfg file")
	}

	dbConfig.Addr = fmt.Sprintf(
		"postgres://%s:%s@localhost:5432/social?sslmode=disable",
		dbConfig.User,
		dbConfig.Password,
	)

	cfg.DbConfig = &dbConfig
	cfg.MailerConfig = &mailerConfig
	cfg.CacheConfig = &redisConfig

	return &cfg
}
