package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	feedbackdomain "github.com/Haerd-Limited/dating-api/internal/feedback/domain"
)

type CreateFeedbackRequest struct {
	Type           string   `json:"type" validate:"required,oneof=positive negative"`
	Title          *string  `json:"title,omitempty"`
	Text           string   `json:"text" validate:"required"`
	AttachmentUrls []string `json:"attachment_urls,omitempty"`
}

func (cfr CreateFeedbackRequest) Validate() error {
	if err := validator.New().Struct(cfr); err != nil {
		return err
	}

	// Custom validation: title required for negative feedback
	if cfr.Type == feedbackdomain.FeedbackTypeNegative {
		if cfr.Title == nil || *cfr.Title == "" {
			return fmt.Errorf("title is required for negative feedback")
		}
	}

	// Custom validation: attachments only allowed for negative feedback
	if cfr.Type == feedbackdomain.FeedbackTypePositive && len(cfr.AttachmentUrls) > 0 {
		return fmt.Errorf("attachments are only allowed for negative feedback")
	}

	return nil
}
