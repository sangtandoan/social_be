package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sangtandoan/practice/internal/config"
)

func NewConnection(config *config.DBConfig) (*sql.DB, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@localhost:%s/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Port,
		config.DBName,
	)

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(parseStringToDuration(config.MaxLifeTime))
	db.SetConnMaxIdleTime(parseStringToDuration(config.MaxIdleTime))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected database!")
	return db, nil
}

func parseStringToDuration(from string) time.Duration {
	duration, err := time.ParseDuration(from)
	if err != nil {
		panic("invalid duration format")
	}

	return duration
}
