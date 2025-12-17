package domain

import "time"

// SessionOpen represents a session tracking event when the app is opened
type SessionOpen struct {
	UserID    string
	SessionID *string
	Timestamp time.Time
}
