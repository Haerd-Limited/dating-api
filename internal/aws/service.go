package aws

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	commonStorage "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/storage"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=aws
type Service interface {
	// Get image bytes for CompareFaces calls (or we can let Rekognition read directly from S3 if using Image{S3Object})
	GetObjectBytes(ctx context.Context, key string) ([]byte, error)
	// Optional: store best frame for a short retention
	StoreBestFrame(ctx context.Context, body []byte, userID string) (string, error)
	GenerateUploadURLs(ctx context.Context, userID string, count int, contentType string, ttl time.Duration, purpose *string) ([]commonStorage.UploadSlot, error)
}

type awsService struct {
	logger      *zap.Logger
	S3Uploader  *commonStorage.S3Uploader
	S3Presigner commonStorage.Presigner
	S3Reader    *commonStorage.S3Reader
	env         string
}

func NewAwsService(
	logger *zap.Logger,
	s3Uploader *commonStorage.S3Uploader,
	s3Presigner commonStorage.Presigner,
	s3Reader *commonStorage.S3Reader,
	env string,
) Service {
	return &awsService{
		logger:      logger,
		S3Uploader:  s3Uploader,
		S3Presigner: s3Presigner,
		S3Reader:    s3Reader,
		env:         env,
	}
}

func (s *awsService) StoreBestFrame(ctx context.Context, body []byte, userID string) (string, error) {
	bestKey := fmt.Sprintf("%s/users/%s/verification/photo/%d.jpg", s.env, userID, time.Now().UnixNano())

	err := s.putObject(ctx, bestKey, body)
	if err != nil {
		return "", err
	}

	return bestKey, nil
}

func (s *awsService) GenerateUploadURLs(ctx context.Context, userID string, count int, contentType string, ttl time.Duration, purpose *string) ([]commonStorage.UploadSlot, error) {
	return s.S3Presigner.GenerateUploadURLs(ctx, userID, count, contentType, ttl, purpose)
}

func (s *awsService) GetObjectBytes(ctx context.Context, key string) ([]byte, error) {
	return s.S3Reader.GetObjectBytes(ctx, key)
}

func (s *awsService) putObject(ctx context.Context, key string, body []byte) error {
	return s.S3Reader.PutObject(ctx, key, body)
}
