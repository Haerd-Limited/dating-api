package domain

import "time"

const (
	FeedbackTypePositive = "positive"
	FeedbackTypeNegative = "negative"
)

const (
	MediaTypeImage = "image"
	MediaTypeVideo = "video"
)

type Feedback struct {
	ID          string
	UserID      string
	Type        string
	Title       *string
	Text        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Attachments []FeedbackAttachment
}

type FeedbackAttachment struct {
	ID         string
	FeedbackID string
	URL        string
	MediaType  string
	CreatedAt  time.Time
}

type CreateFeedbackRequest struct {
	UserID         string
	Type           string
	Title          *string
	Text           string
	AttachmentUrls []string
}
