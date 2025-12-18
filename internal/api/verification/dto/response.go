package dto

import "time"

// Result is what handlers return to clients
type StartResponse struct {
	SessionID string `json:"session_id"`
	Region    string `json:"region"`
}

type CompleteResponse struct {
	Status        string   `json:"status"` // passed / failed / needs_review
	MatchScore    float64  `json:"match_score,omitempty"`
	Reasons       []string `json:"reasons,omitempty"`
	PhotoVerified bool     `json:"photo_verified"`
}

type StartVideoResponse struct {
	Code      string `json:"code"`
	UploadURL string `json:"upload_url"`
	UploadKey string `json:"upload_key"`
}

type SubmitVideoResponse struct {
	Status string `json:"status"`
}

type VideoAttemptResponse struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	VerificationCode string    `json:"verification_code"`
	VideoS3Key       string    `json:"video_s3_key"`
	Status           string    `json:"status"`
	RejectionReason  *string   `json:"rejection_reason,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type VideoDownloadURLResponse struct {
	VideoURL  string `json:"video_url"`
	ExpiresIn int    `json:"expires_in"` // seconds
}

type ApproveVideoResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type RejectVideoResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
