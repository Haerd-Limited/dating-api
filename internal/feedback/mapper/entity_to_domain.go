package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/entity"
	feedbackdomain "github.com/Haerd-Limited/dating-api/internal/feedback/domain"
)

func FeedbackToDomain(feedback *entity.Feedback, attachments []*entity.FeedbackAttachment) feedbackdomain.Feedback {
	if feedback == nil {
		return feedbackdomain.Feedback{}
	}

	out := feedbackdomain.Feedback{
		ID:        feedback.ID,
		UserID:    feedback.UserID,
		Type:      feedback.Type,
		Text:      feedback.Text,
		CreatedAt: feedback.CreatedAt,
		UpdatedAt: feedback.UpdatedAt,
	}

	if feedback.Title.Valid {
		title := feedback.Title.String
		out.Title = &title
	}

	if len(attachments) > 0 {
		out.Attachments = make([]feedbackdomain.FeedbackAttachment, 0, len(attachments))
		for _, att := range attachments {
			out.Attachments = append(out.Attachments, FeedbackAttachmentToDomain(att))
		}
	}

	return out
}

func FeedbackAttachmentToDomain(attachment *entity.FeedbackAttachment) feedbackdomain.FeedbackAttachment {
	if attachment == nil {
		return feedbackdomain.FeedbackAttachment{}
	}

	return feedbackdomain.FeedbackAttachment{
		ID:         attachment.ID,
		FeedbackID: attachment.FeedbackID,
		URL:        attachment.URL,
		MediaType:  attachment.MediaType,
		CreatedAt:  attachment.CreatedAt,
	}
}
