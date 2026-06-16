package domain

import "time"

const (
	BroadcastStatusSent   = "sent"
	BroadcastStatusFailed = "failed"
)

type WaitlistUser struct {
	ID             string
	FirstName      string
	Phone          string
	OnboardingStep string
	CreatedAt      time.Time
	Contacted      bool
}

type BroadcastLog struct {
	UserID  string
	Phone   string
	Message string
	Status  string
	Error   *string
}

type RecipientResult struct {
	UserID string
	Phone  string
	Status string
	Error  *string
}

type BroadcastResult struct {
	Total      int
	Sent       int
	Failed     int
	Recipients []RecipientResult
}
