package feedback

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	feedbackdomain "github.com/Haerd-Limited/dating-api/internal/feedback/domain"
	feedbackmapper "github.com/Haerd-Limited/dating-api/internal/feedback/mapper"
	feedbackstorage "github.com/Haerd-Limited/dating-api/internal/feedback/storage"
	"github.com/Haerd-Limited/dating-api/internal/uow"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
)

type Service interface {
	CreateFeedback(ctx context.Context, req feedbackdomain.CreateFeedbackRequest) error
}

type service struct {
	logger *zap.Logger
	repo   feedbackstorage.Repository
	uow    uow.UoW
}

func NewService(
	logger *zap.Logger,
	repo feedbackstorage.Repository,
	uow uow.UoW,
) Service {
	return &service{
		logger: logger,
		repo:   repo,
		uow:    uow,
	}
}

var (
	ErrInvalidFeedbackType        = errors.New("invalid feedback type, must be 'positive' or 'negative'")
	ErrTitleRequiredForNegative   = errors.New("title is required for negative feedback")
	ErrAttachmentsOnlyForNegative = errors.New("attachments are only allowed for negative feedback")
	ErrTextRequired               = errors.New("text is required")
)

func (s *service) CreateFeedback(ctx context.Context, req feedbackdomain.CreateFeedbackRequest) error {
	if err := s.validateFeedbackRequest(req); err != nil {
		return commonlogger.LogError(s.logger, "validate feedback request", err,
			zap.String("userID", req.UserID),
			zap.String("type", req.Type))
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return commonlogger.LogError(s.logger, "begin tx", err)
	}

	defer func() { _ = tx.Rollback() }()

	// Generate feedback ID
	feedbackID := uuid.New().String()

	// Create feedback domain model
	feedback := feedbackdomain.Feedback{
		ID:     feedbackID,
		UserID: req.UserID,
		Type:   req.Type,
		Title:  req.Title,
		Text:   req.Text,
	}

	// Map to entity and insert
	feedbackEntity := feedbackmapper.FeedbackToEntity(feedback)
	if err := s.repo.CreateFeedback(ctx, feedbackEntity, tx.Raw()); err != nil {
		return commonlogger.LogError(s.logger, "create feedback", err,
			zap.String("userID", req.UserID),
			zap.String("feedbackID", feedbackID))
	}

	// Create attachments if provided
	if len(req.AttachmentUrls) > 0 {
		attachmentEntities := make([]*entity.FeedbackAttachment, 0, len(req.AttachmentUrls))

		for _, url := range req.AttachmentUrls {
			attachmentID := uuid.New().String()
			// Determine media type from URL or default to image
			mediaType := feedbackdomain.MediaTypeImage
			// You could add logic here to detect video URLs if needed

			attachment := feedbackdomain.FeedbackAttachment{
				ID:         attachmentID,
				FeedbackID: feedbackID,
				URL:        url,
				MediaType:  mediaType,
			}

			attachmentEntities = append(attachmentEntities, feedbackmapper.FeedbackAttachmentToEntity(attachment))
		}

		if err := s.repo.CreateFeedbackAttachments(ctx, attachmentEntities, tx.Raw()); err != nil {
			return commonlogger.LogError(s.logger, "create feedback attachments", err,
				zap.String("userID", req.UserID),
				zap.String("feedbackID", feedbackID))
		}
	}

	if err := tx.Commit(); err != nil {
		return commonlogger.LogError(s.logger, "commit tx", err)
	}

	return nil
}

func (s *service) validateFeedbackRequest(req feedbackdomain.CreateFeedbackRequest) error {
	if req.Text == "" {
		return ErrTextRequired
	}

	if req.Type != feedbackdomain.FeedbackTypePositive && req.Type != feedbackdomain.FeedbackTypeNegative {
		return ErrInvalidFeedbackType
	}

	if req.Type == feedbackdomain.FeedbackTypeNegative {
		if req.Title == nil || *req.Title == "" {
			return ErrTitleRequiredForNegative
		}
	}

	if req.Type == feedbackdomain.FeedbackTypePositive && len(req.AttachmentUrls) > 0 {
		return ErrAttachmentsOnlyForNegative
	}

	return nil
}
