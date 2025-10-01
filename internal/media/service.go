package media

import (
	"context"
	"fmt"
	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/media/domain"
	"go.uber.org/zap"
	"time"
)

type Service interface {
	GeneratePhotoUploadUrl(ctx context.Context, userID string) (domain.UploadUrl, error)
	GenerateVoiceNoteUploadUrl(ctx context.Context, userID string) (domain.UploadUrl, error)
}

const (
	maxUploadCountPhotos     = 6
	minUploadCountPhotos     = 1
	maxUploadCountPrompts    = 6
	minUploadCountPrompts    = 1
	minUploadCountVoiceNotes = 1
	maxUploadBytes           = 5 << 20 // 5 MiB
	presignTTL               = 20 * time.Minute
	mimeJPEG                 = "image/jpeg"
	mimeM4A                  = "audio/mp4" // m4a is an MP4 container; "audio/m4a" also seen but "audio/mp4" is safer
)

type service struct {
	logger     *zap.Logger
	awsService aws.Service
}

func NewMediaService(
	logger *zap.Logger,
	awsService aws.Service,
) Service {
	return &service{
		logger:     logger,
		awsService: awsService,
	}
}
func (s *service) GeneratePhotoUploadUrl(ctx context.Context, userID string) (domain.UploadUrl, error) {
	url, err := s.awsService.GenerateUploadURLs(ctx, userID, minUploadCountPhotos, mimeJPEG, presignTTL)
	if err != nil {
		return domain.UploadUrl{}, fmt.Errorf("failed to generate photo upload url: %w", err)
	}
	if len(url) != minUploadCountPhotos {
		return domain.UploadUrl{}, fmt.Errorf("failed to generate photo upload url: expected %d urls, got %d", minUploadCountPhotos, len(url))
	}

	return domain.UploadUrl{
		Key:       url[0].Key,
		UploadUrl: url[0].URL,
		Headers:   url[0].Headers,
		MaxBytes:  maxUploadBytes,
	}, nil
}

func (s *service) GenerateVoiceNoteUploadUrl(ctx context.Context, userID string) (domain.UploadUrl, error) {
	url, err := s.awsService.GenerateUploadURLs(ctx, userID, minUploadCountVoiceNotes, mimeM4A, presignTTL)
	if err != nil {
		return domain.UploadUrl{}, fmt.Errorf("failed to generate voicenote upload url: %w", err)
	}
	if len(url) != minUploadCountVoiceNotes {
		return domain.UploadUrl{}, fmt.Errorf("failed to generate voicenote upload url: expected %d urls, got %d", minUploadCountVoiceNotes, len(url))
	}

	return domain.UploadUrl{
		Key:       url[0].Key,
		UploadUrl: url[0].URL,
		Headers:   url[0].Headers,
		MaxBytes:  maxUploadBytes,
	}, nil
}
