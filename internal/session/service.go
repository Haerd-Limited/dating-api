package session

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/session/domain"
	commonanalytics "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/analytics"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
)

type Service interface {
	TrackAppOpen(ctx context.Context, userID string) error
}

type service struct {
	logger *zap.Logger
}

var ErrEmptyUserID = errors.New("userID cannot be empty")

func NewService(logger *zap.Logger) Service {
	return &service{
		logger: logger,
	}
}

func (s *service) TrackAppOpen(ctx context.Context, userID string) error {
	if userID == "" {
		return commonlogger.LogError(s.logger, "track app open", ErrEmptyUserID, zap.String("reason", "userID is empty"))
	}

	// Create session open domain model
	sessionOpen := domain.SessionOpen{
		UserID:    userID,
		Timestamp: time.Now().UTC(),
	}

	// Track the event using analytics infrastructure
	props := map[string]any{
		"timestamp": sessionOpen.Timestamp,
		"user_id":   userID,
	}

	// Track the app_opened event
	commonanalytics.Track(ctx, "app_opened", &userID, nil, props)

	return nil
}
