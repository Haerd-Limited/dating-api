package domain

import "time"

const SessionTTL = 7 * 24 * time.Hour

type Session struct {
	ID          string
	DisplayName string
	TokenHash   string
	APIKeyFP    string
	IP          *string
	CreatedAt   time.Time
	LastSeenAt  time.Time
	ExpiresAt   time.Time
}

type CreateSessionRequest struct {
	DisplayName string
	APIKeyFP    string
	IP          *string
}

type SessionResult struct {
	SessionToken string
	DisplayName  string
	ExpiresAt    time.Time
}
