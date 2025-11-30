package mapper

import (
	dto "github.com/Haerd-Limited/dating-api/internal/api/feedback/dto"
	feedbackdomain "github.com/Haerd-Limited/dating-api/internal/feedback/domain"
)

func CreateFeedbackRequestToDomain(req dto.CreateFeedbackRequest, userID string) feedbackdomain.CreateFeedbackRequest {
	return feedbackdomain.CreateFeedbackRequest{
		UserID:         userID,
		Type:           req.Type,
		Title:          req.Title,
		Text:           req.Text,
		AttachmentUrls: req.AttachmentUrls,
	}
}
