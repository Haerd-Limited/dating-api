package domain

import "time"

// Entry is a single admin API access record.
type Entry struct {
	OccurredAt time.Time
	ActorIP    *string
	TokenFP    string
	Method     string
	Path       string
	TargetID   *string
	StatusCode int
}
