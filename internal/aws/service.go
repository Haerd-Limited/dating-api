package aws

import (
	"context"
	"time"

	"go.uber.org/zap"

	commonStorage "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/storage"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=aws
type Service interface {
	GenerateUploadURLs(ctx context.Context, userID string, count int, contentType string, ttl time.Duration, purpose *string) ([]commonStorage.UploadSlot, error)
}

type awsService struct {
	logger      *zap.Logger
	S3Uploader  *commonStorage.S3Uploader
	S3Presigner commonStorage.Presigner
}

func NewAwsService(
	logger *zap.Logger,
	s3Uploader *commonStorage.S3Uploader,
	s3Presigner commonStorage.Presigner,
) Service {
	return &awsService{
		logger:      logger,
		S3Uploader:  s3Uploader,
		S3Presigner: s3Presigner,
	}
}

func (s *awsService) GenerateUploadURLs(ctx context.Context, userID string, count int, contentType string, ttl time.Duration, purpose *string) ([]commonStorage.UploadSlot, error) {
	return s.S3Presigner.GenerateUploadURLs(ctx, userID, count, contentType, ttl, purpose)
}
