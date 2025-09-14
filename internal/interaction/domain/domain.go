package domain

import (
	"time"
)

type Swipe struct {
	TargetUserID   string
	Action         string
	UserID         string
	IdempotencyKey *string
}

type Match struct {
	ID         string
	UserA      string
	UserB      string
	CreatedAt  time.Time
	RevealedAt time.Time
}
