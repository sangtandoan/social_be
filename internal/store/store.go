package store

import (
	"context"
	"database/sql"
)

type Store struct {
	Posts interface {
		Create(context.Context) error
	}

	Users interface {
		Create(context.Context) error
	}
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Posts: &PostsStore{db},
		Users: &UsersStore{db},
	}
}
