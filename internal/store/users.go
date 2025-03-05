package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/sangtandoan/social/internal/models/dto"
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

func (s *UsersStore) Create(ctx context.Context, arg *dto.CreateUserRequest) (*User, error) {
	executor := GetExecutor(ctx, s.db)
	query := "INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id, password, created_at"

	row := executor.QueryRowContext(ctx, query, arg.Username, arg.Email, arg.Password)

	var user User
	err := row.Scan(&user.ID, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	user.Username = arg.Username
	user.Email = arg.Email

	return &user, nil
}

func (s *UsersStore) GetByID(ctx context.Context, id int64) (*User, error) {
	executor := GetExecutor(ctx, s.db)
	query := "SELECT id, username, email, password, created_at FROM users WHERE id = $1"

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	row := executor.QueryRowContext(ctx, query, id)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UsersStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	executor := GetExecutor(ctx, s.db)
	query := "SELECT id, username, email, password, created_at FROM users WHERE email = $1"

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	row := executor.QueryRowContext(ctx, query, email)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UsersStore) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM users WHERE id = $1"

	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *UsersStore) Activate(ctx context.Context, id int64) error {
	executor := GetExecutor(ctx, s.db)
	query := "UPDATE users SET active = true WHERE id = $1"

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	_, err := executor.ExecContext(ctx, query, id)
	return err
}
