package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   *ServerConfig
	Database *DBConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DBConfig struct {
	Port         string
	User         string
	Password     string
	DBName       string
	MaxLifeTime  string
	MaxIdleTime  string
	MaxOpenConns int
	MaxIdleConns int
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	serverConfig := &ServerConfig{
		Port: getEnv("SERVER_PORT", ":8080"),
		Env:  getEnv("ENV", "development"),
	}

	dbConfig := &DBConfig{
		Port:         getEnv("DB_PORT", ":5432"),
		User:         getEnv("DB_USER", "admin"),
		Password:     getEnv("DB_PASSWORD", "secret"),
		DBName:       getEnv("DB_NAME", "app"),
		MaxLifeTime:  getEnv("DB_MAX_LIFE_TIME", "1h"),
		MaxIdleTime:  getEnv("DB_MAX_IDLE_TIME", "30m"),
		MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
	}

	return &Config{
		Server:   serverConfig,
		Database: dbConfig,
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if valueAsInt, err := strconv.Atoi(value); err == nil {
			return valueAsInt
		}
	}

	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if valueAsBool, err := strconv.ParseBool(value); err == nil {
			return valueAsBool
		}
	}

	return fallback
}

func getEnvAsSlice(key string, fallback []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.Split(value, ",")
	}

	return fallback
}
