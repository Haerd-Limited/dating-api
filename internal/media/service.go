package media

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/media/domain"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
)

type Service interface {
	GeneratePhotoUploadUrl(ctx context.Context, userID string) (domain.UploadUrl, error)
	GenerateVoiceNoteUploadUrl(ctx context.Context, userID string, purpose string) (domain.UploadUrl, error)
	GenerateUploadURLsForProfilePhotos(ctx context.Context, userID string) ([]domain.UploadUrl, error)
	GenerateUploadURLsForProfilePrompts(ctx context.Context, userID string) ([]domain.UploadUrl, error)
}

const (
	maxUploadCountPhotos     = 6
	minUploadCountPhotos     = 1
	maxUploadCountPrompts    = 6
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
	url, err := s.awsService.GenerateUploadURLs(ctx, userID, minUploadCountPhotos, mimeJPEG, presignTTL, nil)
	if err != nil {
		return domain.UploadUrl{}, commonlogger.LogError(s.logger, "failed to generate photo upload url", err, zap.String("userID", userID))
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

func (s *service) GenerateVoiceNoteUploadUrl(ctx context.Context, userID string, purpose string) (domain.UploadUrl, error) {
	url, err := s.awsService.GenerateUploadURLs(ctx, userID, minUploadCountVoiceNotes, mimeM4A, presignTTL, &purpose)
	if err != nil {
		return domain.UploadUrl{}, commonlogger.LogError(s.logger, "failed to generate voicenote upload url", err, zap.String("userID", userID), zap.String("purpose", purpose))
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

func (s *service) GenerateUploadURLsForProfilePhotos(ctx context.Context, userID string) ([]domain.UploadUrl, error) {
	urls, err := s.awsService.GenerateUploadURLs(ctx, userID, maxUploadCountPhotos, mimeJPEG, presignTTL, nil)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "failed to generate upload urls", err, zap.String("userID", userID))
	}

	var photoUploadUrls []domain.UploadUrl
	for _, url := range urls {
		photoUploadUrls = append(photoUploadUrls, domain.UploadUrl{
			Key:       url.Key,
			UploadUrl: url.URL,
			Headers:   url.Headers,
			MaxBytes:  maxUploadBytes,
		})
	}

	return photoUploadUrls, nil
}

func (s *service) GenerateUploadURLsForProfilePrompts(ctx context.Context, userID string) ([]domain.UploadUrl, error) {
	urls, err := s.awsService.GenerateUploadURLs(ctx, userID, maxUploadCountPrompts, mimeM4A, presignTTL, nil)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "failed to generate upload urls", err, zap.String("userID", userID))
	}

	var voicePromptUploadUrls []domain.UploadUrl
	for _, url := range urls {
		voicePromptUploadUrls = append(voicePromptUploadUrls, domain.UploadUrl{
			Key:       url.Key,
			UploadUrl: url.URL,
			Headers:   url.Headers,
			MaxBytes:  maxUploadBytes,
		})
	}

	return voicePromptUploadUrls, nil
}
