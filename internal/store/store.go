package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/sangtandoan/social/internal/models/dto"
	"github.com/sangtandoan/social/internal/models/params"
)

var (
	ErrNotFound  = errors.New("resource not found")
	QueryTimeOut = time.Second * 5
)

type Executor interface {
	QueryRowContext(ctx context.Context, query string, arg ...any) *sql.Row
	QueryContext(ctx context.Context, query string, arg ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, arg ...any) (sql.Result, error)
}

type Tx interface {
	WithTx(ctx context.Context, f func(txCtx context.Context) error) error
}

type tx struct{ db *sql.DB }

type Store struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(ctx context.Context, id int64) (*Post, error)
		GetAll(ctx context.Context) ([]*Post, error)
		UpdatePost(ctx context.Context, arg *UpdatePostParams) (*Post, error)
		GetUserFeed(ctx context.Context, arg *dto.UserFeedRequest) ([]*PostResponse, error)
	}

	Users interface {
		Create(ctx context.Context, arg *dto.CreateUserRequest) (*User, error)
		GetByID(ctx context.Context, ID int64) (*User, error)
	}

	Followers interface {
		Follow(ctx context.Context, arg *FollowParams) error
		Unfollow(ctx context.Context, arg *UnfollowParams) error
	}

	Invitations interface {
		CreateInvitation(ctx context.Context, arg *params.CreateInvitationParams) error
	}

	Tx Tx
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Posts:       &PostsStore{db},
		Users:       &UsersStore{db},
		Followers:   NewFollowerStore(db),
		Invitations: NewInvitationStore(db),
		Tx:          &tx{db},
	}
}

type TxKey struct{}

func (t *tx) WithTx(ctx context.Context, f func(ctx context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	txContext := context.WithValue(ctx, TxKey{}, tx)
	err = f(txContext)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			panic("can not rollback")
		}
	}

	return tx.Commit()
}

func GetExecutor(ctx context.Context, db *sql.DB) Executor {
	tx, ok := ctx.Value(TxKey{}).(*sql.Tx)
	if !ok {
		return db
	}

	return tx
}
