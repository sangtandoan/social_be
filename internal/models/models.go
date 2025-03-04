package models

import "time"

type Invitation struct {
	CreatedAt time.Time
	Token     string
	UserID    int64
}
