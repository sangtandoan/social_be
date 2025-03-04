package params

import "time"

type CreateInvitationParams struct {
	Token     string
	UserID    int64
	ExpiresAt time.Time
}
