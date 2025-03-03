package store

import (
	"context"
	"database/sql"
)

type followerStore struct {
	db *sql.DB
}

func NewFollowerStore(db *sql.DB) *followerStore {
	return &followerStore{db}
}

type FollowParams struct {
	UserID     int64
	FollowerID int64
}

func (s *followerStore) Follow(ctx context.Context, arg *FollowParams) error {
	query := "INSERT INTO followers (user_id, follower_id) VALUES ($1, $2)"

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, arg.UserID, arg.FollowerID)

	return err
}

type UnfollowParams struct {
	UserID     int64
	FollowerID int64
}

func (s *followerStore) Unfollow(ctx context.Context, arg *UnfollowParams) error {
	query := "DELETE FROM followers WHERE user_id = $1 and follower_id = $2"

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, arg.UserID, arg.FollowerID)

	return err
}
