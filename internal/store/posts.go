package store

import (
	"context"
	"database/sql"
)

type PostsStore struct {
	db *sql.DB
}

func (s *PostsStore) Create(c context.Context) error {
	return nil
}
