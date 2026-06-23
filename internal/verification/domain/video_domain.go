package domain

import (
	"time"

	profiledomain "github.com/Haerd-Limited/dating-api/internal/profile/domain"
)

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
	RejectionReason  *string
	Photos           []profiledomain.Photo
	ReviewedByName   *string
	ReviewedAt       *time.Time
	ReviewNotes      *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type VideoReviewInfo struct {
	ReviewedByName      string
	ReviewedBySessionID string
	ReviewNotes         *string
}

type ApproveVideoRequest struct {
	Notes *string
}

type RejectVideoRequest struct {
	RejectionReason string
	Notes           *string
}
