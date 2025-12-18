package domain

import "time"

type StartVideoResult struct {
	Code      string
	UploadURL string
	UploadKey string
}

type SubmitVideoResult struct {
	Status string // "submitted"
}

type VideoAttemptFilter struct {
	Status []string // pending, needs_review, passed, failed
	UserID *string
	Limit  int
	Offset int
}

type VideoAttempt struct {
	ID               string
	UserID           string
	VerificationCode string
	VideoS3Key       string
	Status           string
	RejectionReason  *string // Extracted from ReasonCodes JSON field
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ApproveVideoRequest struct {
	Notes *string
}

type RejectVideoRequest struct {
	RejectionReason string
	Notes           *string
}
