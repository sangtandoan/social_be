package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func New(addr string, maxOpenConns, maxIdleConns int, maxLifeTime string) (*sql.DB, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	duration, err := time.ParseDuration(maxLifeTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(duration)

	context, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := db.PingContext(context); err != nil {
		return nil, err
	}

	return db, nil
}
