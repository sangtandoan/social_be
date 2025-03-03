package store

import (
	"context"
	"database/sql"
	"time"
)

type UsersStore struct {
	db *sql.DB
}

type User struct {
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username,omitempty"`
	Email     string    `json:"email,omitempty"`
	Password  string    `json:"password,omitempty"`
	ID        int64     `json:"id,omitempty"`
}

func (s *UsersStore) Create(c context.Context) error {
	return nil
}

func (s *UsersStore) GetByID(ctx context.Context, ID int64) (*User, error) {
	query := "SELECT id, username, email, password, created_at FROM users WHERE id = $1"

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	row := s.db.QueryRowContext(ctx, query, ID)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
