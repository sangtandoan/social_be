package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"

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

	query := "INSERT INTO invitations (user_id, token, expires_at) VALUES ($1, $2, $3)"

	_, err := executor.ExecContext(ctx, query, arg.UserID, arg.Token, arg.ExpiresAt)

	return err
}

func (s *invitationStore) GetUserIDFromInvitation(
	ctx context.Context,
	token string,
) (int64, error) {
	executor := GetExecutor(ctx, s.db)

	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	query := "SELECT user_id FROM invitations WHERE token = $1"

	row := executor.QueryRowContext(ctx, query, hashedToken)

	var userID int64
	err := row.Scan(&userID)
	if err != nil {
		return -1, err
	}

	return userID, nil
}
