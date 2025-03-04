package store

import (
	"context"
	"database/sql"

	"github.com/sangtandoan/social/internal/models/params"
)

type invitationStore struct {
	db *sql.DB
}

func NewInvitationStore(db *sql.DB) *invitationStore {
	return &invitationStore{db}
}

func (s *invitationStore) CreateInvitation(
	ctx context.Context,
	arg *params.CreateInvitationParams,
) error {
	executor := GetExecutor(ctx, s.db)

	query := "INSERT INTO invitations (user_id, token) VALUES ($1, $2)"

	_, err := executor.ExecContext(ctx, query, arg.UserID, arg.Token)

	return err
}
