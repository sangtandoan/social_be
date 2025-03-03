package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/sangtandoan/social/internal/models/dto"
)

var (
	ErrNotFound  = errors.New("resource not found")
	QueryTimeOut = time.Second * 5
)

type Store struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(ctx context.Context, id int64) (*Post, error)
		GetAll(ctx context.Context) ([]*Post, error)
		UpdatePost(ctx context.Context, arg *UpdatePostParams) (*Post, error)
		GetUserFeed(ctx context.Context, arg *dto.UserFeedRequest) ([]*PostResponse, error)
	}

	Users interface {
		Create(context.Context) error
	}

	Followers interface {
		Follow(ctx context.Context, arg *FollowParams) error
		Unfollow(ctx context.Context, arg *UnfollowParams) error
	}
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Posts:     &PostsStore{db},
		Users:     &UsersStore{db},
		Followers: NewFollowerStore(db),
	}
}
