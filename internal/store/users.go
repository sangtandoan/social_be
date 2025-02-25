package store

import (
	"context"
	"database/sql"
)

type UsersStore struct {
	db *sql.DB
}

func (s *UsersStore) Create(c context.Context) error {
	return nil
}
