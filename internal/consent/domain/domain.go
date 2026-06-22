package domain

import "time"

type Consent struct {
	UserID     string
	Type       string
	Version    string
	Accepted   bool
	AcceptedAt time.Time
	RevokedAt  *time.Time
	IP         *string
	UserAgent  *string
}

type RecordRequest struct {
	UserID    string
	Type      string
	Version   string
	Accepted  bool
	IP        *string
	UserAgent *string
}
