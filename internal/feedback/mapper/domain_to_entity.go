package mapper

import (
	"time"

	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	feedbackdomain "github.com/Haerd-Limited/dating-api/internal/feedback/domain"
)

func FeedbackToEntity(feedback feedbackdomain.Feedback) *entity.Feedback {
	now := time.Now().UTC()

	out := entity.Feedback{
		ID:        feedback.ID,
		UserID:    feedback.UserID,
		Type:      feedback.Type,
		Text:      feedback.Text,
		CreatedAt: feedback.CreatedAt,
		UpdatedAt: feedback.UpdatedAt,
	}

	if feedback.CreatedAt.IsZero() {
		out.CreatedAt = now
	}

	if feedback.UpdatedAt.IsZero() {
		out.UpdatedAt = now
	}

	if feedback.Title != nil {
		out.Title = null.StringFrom(*feedback.Title)
	}

	return &out
}

func FeedbackAttachmentToEntity(attachment feedbackdomain.FeedbackAttachment) *entity.FeedbackAttachment {
	now := time.Now().UTC()

	out := entity.FeedbackAttachment{
		ID:         attachment.ID,
		FeedbackID: attachment.FeedbackID,
		URL:        attachment.URL,
		MediaType:  attachment.MediaType,
		CreatedAt:  attachment.CreatedAt,
	}

	if attachment.CreatedAt.IsZero() {
		out.CreatedAt = now
	}

	return &out
}
